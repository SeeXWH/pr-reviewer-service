//go:build integration

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SeeXWH/pr-reviewer-service/configs"
	"github.com/SeeXWH/pr-reviewer-service/internal/model"
	"github.com/SeeXWH/pr-reviewer-service/internal/team"
	"github.com/SeeXWH/pr-reviewer-service/pkg/db"
	"github.com/SeeXWH/pr-reviewer-service/pkg/logger"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

func TestTeamSuite(t *testing.T) {
	suite.Run(t, new(TeamSuite))
}

type TeamSuite struct {
	suite.Suite
	rawDB     *gorm.DB
	dbWrapper *db.PostgresDB
	router    http.Handler
	cleanUp   func()
}

func (s *TeamSuite) SetupSuite() {
	ctx := context.Background()
	log := logger.Setup()
	mux := http.NewServeMux()

	pgContainer, cleanup, err := SetupPostgresContainer()
	s.Require().NoError(err)
	s.cleanUp = cleanup

	host, _ := pgContainer.Host(ctx)
	natPort, _ := pgContainer.MappedPort(ctx, "5432")

	cfg := &configs.Config{
		DB: configs.DB{
			Username: "user",
			Password: "password",
			Dbname:   "testdb",
			Host:     host,
			Port:     natPort.Port(),
		},
		App: configs.App{
			TimeOut: 500 * time.Millisecond,
		},
	}

	var errDB error
	for i := 0; i < 10; i++ {
		s.dbWrapper, errDB = db.NewPostgresDB(cfg)
		if errDB == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	s.Require().NoError(errDB)
	s.rawDB = s.dbWrapper.PostgresDB

	err = s.rawDB.AutoMigrate(&model.Team{}, &model.User{}, &model.PullRequest{})
	s.Require().NoError(err)

	teamRepo := team.NewRepository(s.dbWrapper)
	teamService := team.NewService(teamRepo, log)

	team.NewHandler(mux, teamService, cfg)

	s.router = mux
}

func (s *TeamSuite) TearDownSuite() {
	if s.cleanUp != nil {
		s.cleanUp()
	}
}

func (s *TeamSuite) SetupTest() {
	s.rawDB.Exec("TRUNCATE TABLE pr_reviewers CASCADE")
	s.rawDB.Exec("TRUNCATE TABLE pull_requests CASCADE")
	s.rawDB.Exec("TRUNCATE TABLE users CASCADE")
	s.rawDB.Exec("TRUNCATE TABLE teams CASCADE")
}

func (s *TeamSuite) TestCreateTeam_Success() {
	reqDTO := team.CreateRequestDTO{
		TeamName: "alpha-squad",
		Members: []team.UserCreateRequestDTO{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: true},
		},
	}
	bodyBytes, _ := json.Marshal(reqDTO)

	req, _ := http.NewRequest(http.MethodPost, "/team/add", bytes.NewBuffer(bodyBytes))
	rr := httptest.NewRecorder()

	s.router.ServeHTTP(rr, req)

	s.Equal(http.StatusCreated, rr.Code)

	var resp team.CreateTeamResponseDTO
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	s.Require().NoError(err)

	s.Equal("alpha-squad", resp.Team.TeamName)
	s.Len(resp.Team.Members, 2)

	var dbTeam model.Team
	err = s.rawDB.First(&dbTeam, "team_name = ?", "alpha-squad").Error
	s.NoError(err)

	var count int64
	s.rawDB.Model(&model.User{}).Where("team_name = ?", "alpha-squad").Count(&count)
	s.Equal(int64(2), count)
}

func (s *TeamSuite) TestCreateTeam_Duplicate() {
	existingTeam := model.Team{Name: "beta-squad"}
	s.Require().NoError(s.rawDB.Create(&existingTeam).Error)

	reqDTO := team.CreateRequestDTO{
		TeamName: "beta-squad",
		Members:  []team.UserCreateRequestDTO{},
	}
	bodyBytes, _ := json.Marshal(reqDTO)

	req, _ := http.NewRequest(http.MethodPost, "/team/add", bytes.NewBuffer(bodyBytes))
	rr := httptest.NewRecorder()

	s.router.ServeHTTP(rr, req)

	s.Equal(http.StatusBadRequest, rr.Code)

	s.Contains(rr.Body.String(), "TEAM_EXISTS")
}

func (s *TeamSuite) TestGetTeam_Success() {
	teamName := "gamma-squad"
	s.rawDB.Create(&model.Team{Name: teamName})
	s.rawDB.Create(&model.User{ID: "g1", Username: "Gus", TeamName: teamName, IsActive: true})

	req, _ := http.NewRequest(http.MethodGet, "/team/get?team_name="+teamName, nil)
	rr := httptest.NewRecorder()

	s.router.ServeHTTP(rr, req)

	s.Equal(http.StatusOK, rr.Code)

	var resp team.InfoDTO
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	s.Require().NoError(err)

	s.Equal(teamName, resp.TeamName)
	s.Len(resp.Members, 1)
	s.Equal("g1", resp.Members[0].UserID)
}

func (s *TeamSuite) TestGetTeam_NotFound() {
	req, _ := http.NewRequest(http.MethodGet, "/team/get?team_name=ghost-team", nil)
	rr := httptest.NewRecorder()

	s.router.ServeHTTP(rr, req)

	s.Equal(http.StatusNotFound, rr.Code)
	s.Contains(rr.Body.String(), "NOT_FOUND")
}
