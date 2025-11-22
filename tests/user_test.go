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
	"github.com/SeeXWH/pr-reviewer-service/internal/user"
	"github.com/SeeXWH/pr-reviewer-service/pkg/db"
	"github.com/SeeXWH/pr-reviewer-service/pkg/logger"
	
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

func TestUserSuite(t *testing.T) {
	suite.Run(t, new(UserSuite))
}

type UserSuite struct {
	suite.Suite
	rawDB     *gorm.DB
	dbWrapper *db.PostgresDB
	router    http.Handler
	cleanUp   func()
}

func (s *UserSuite) SetupSuite() {
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

	userRepo := user.NewRepository(s.dbWrapper)
	userService := user.NewService(userRepo, log)

	user.NewHandler(mux, userService, cfg)

	s.router = mux
}

func (s *UserSuite) TearDownSuite() {
	if s.cleanUp != nil {
		s.cleanUp()
	}
}

func (s *UserSuite) SetupTest() {
	s.rawDB.Exec("TRUNCATE TABLE pr_reviewers CASCADE")
	s.rawDB.Exec("TRUNCATE TABLE pull_requests CASCADE")
	s.rawDB.Exec("TRUNCATE TABLE users CASCADE")
	s.rawDB.Exec("TRUNCATE TABLE teams CASCADE")
}

func (s *UserSuite) TestSetIsActive() {
	team := model.Team{Name: "backend"}
	s.Require().NoError(s.rawDB.Create(&team).Error)

	u1 := model.User{ID: "u1", Username: "Alice", IsActive: true, TeamName: "backend"}
	s.Require().NoError(s.rawDB.Create(&u1).Error)

	reqBody := user.SetActiveRequestDTO{
		UserID:   "u1",
		IsActive: false,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewBuffer(bodyBytes))
	rr := httptest.NewRecorder()

	s.router.ServeHTTP(rr, req)

	s.Equal(http.StatusOK, rr.Code)

	var resp user.ResponseWrapper
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	s.Require().NoError(err)

	s.Equal(false, resp.User.IsActive)
	s.Equal("u1", resp.User.UserID)

	var dbUser model.User
	s.rawDB.First(&dbUser, "user_id = ?", "u1")
	s.False(dbUser.IsActive)
}

func (s *UserSuite) TestMassDeactivate() {
	team := model.Team{Name: "backend"}
	s.Require().NoError(s.rawDB.Create(&team).Error)

	users := []model.User{
		{ID: "u1", Username: "Author", IsActive: true, TeamName: "backend"},
		{ID: "u2", Username: "FiredGuy", IsActive: true, TeamName: "backend"},
		{ID: "u3", Username: "Hero", IsActive: true, TeamName: "backend"},
	}
	s.Require().NoError(s.rawDB.Create(&users).Error)

	pr := model.PullRequest{
		ID:        "pr-1",
		Status:    "OPEN",
		AuthorID:  "u1",
		Reviewers: []*model.User{&users[1]},
	}
	s.Require().NoError(s.rawDB.Create(&pr).Error)

	reqBody := user.MassDeactivateRequestDTO{
		TeamName: "backend",
		UserIDs:  []string{"u2"},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodPost, "/users/massDeactivate", bytes.NewBuffer(bodyBytes))
	rr := httptest.NewRecorder()

	s.router.ServeHTTP(rr, req)

	s.Equal(http.StatusOK, rr.Code)

	var resp user.MassDeactivateResponseDTO
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	s.Require().NoError(err)

	s.Equal(1, resp.DeactivatedCount)
	s.Equal(1, resp.ReassignedPRs)

	var u2FromDB model.User
	s.rawDB.First(&u2FromDB, "user_id = ?", "u2")
	s.False(u2FromDB.IsActive)

	var prFromDB model.PullRequest
	s.rawDB.Preload("Reviewers").First(&prFromDB, "pull_request_id = ?", "pr-1")

	s.Require().Len(prFromDB.Reviewers, 1)
	s.Equal("u3", prFromDB.Reviewers[0].ID)
}

func (s *UserSuite) TestGetReview() {
	team := model.Team{Name: "backend"}
	s.Require().NoError(s.rawDB.Create(&team).Error)
	users := []model.User{
		model.User{ID: "u1", Username: "Author", IsActive: true, TeamName: "backend"},
		model.User{ID: "u2", Username: "Anton", IsActive: true, TeamName: "backend"},
		model.User{ID: "u3", Username: "Kirill", IsActive: true, TeamName: "backend"},
	}
	s.Require().NoError(s.rawDB.Create(&users).Error)

	pr1 := model.PullRequest{
		ID:        "pr-1",
		Name:      "No way",
		Status:    "OPEN",
		AuthorID:  "u1",
		Reviewers: []*model.User{&users[1], &users[2]},
	}
	s.Require().NoError(s.rawDB.Create(&pr1).Error)

	pr2 := model.PullRequest{
		ID:        "pr-2",
		Name:      "Фиксики",
		Status:    "OPEN",
		AuthorID:  "u1",
		Reviewers: []*model.User{&users[2]},
	}
	s.Require().NoError(s.rawDB.Create(&pr2).Error)

	req, _ := http.NewRequest(http.MethodGet, "/users/getReview?user_id=u2", nil)
	rr := httptest.NewRecorder()

	s.router.ServeHTTP(rr, req)

	s.Equal(http.StatusOK, rr.Code)

	var resp user.ReviewsResponseDTO
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	s.Require().NoError(err)

	s.Equal("u2", resp.UserID)

	s.Require().Len(resp.PullRequests, 1)

	foundPR := resp.PullRequests[0]
	s.Equal("pr-1", foundPR.ID)
	s.Equal("No way", foundPR.Name)
	s.Equal("u1", foundPR.AuthorID)
}
