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
	"github.com/SeeXWH/pr-reviewer-service/internal/pullrequest"
	"github.com/SeeXWH/pr-reviewer-service/internal/user"
	"github.com/SeeXWH/pr-reviewer-service/pkg/db"
	"github.com/SeeXWH/pr-reviewer-service/pkg/logger"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

func TestPRSuite(t *testing.T) {
	suite.Run(t, new(PRSuite))
}

type PRSuite struct {
	suite.Suite
	rawDB     *gorm.DB
	dbWrapper *db.PostgresDB
	router    http.Handler
	cleanUp   func()
}

func (s *PRSuite) SetupSuite() {
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
	prRepo := pullrequest.NewRepository(s.dbWrapper)
	prService := pullrequest.NewService(userService, prRepo, log)
	pullrequest.NewHandler(mux, prService, cfg)

	s.router = mux
}

func (s *PRSuite) TearDownSuite() {
	if s.cleanUp != nil {
		s.cleanUp()
	}
}

func (s *PRSuite) SetupTest() {
	s.rawDB.Exec("TRUNCATE TABLE pr_reviewers CASCADE")
	s.rawDB.Exec("TRUNCATE TABLE pull_requests CASCADE")
	s.rawDB.Exec("TRUNCATE TABLE users CASCADE")
	s.rawDB.Exec("TRUNCATE TABLE teams CASCADE")
}

func (s *PRSuite) TestCreatePR_Success() {
	team := model.Team{Name: "backend"}
	s.rawDB.Create(&team)

	users := []model.User{
		{ID: "u1", Username: "Author", IsActive: true, TeamName: "backend"},
		{ID: "u2", Username: "Reviewer1", IsActive: true, TeamName: "backend"},
		{ID: "u3", Username: "Reviewer2", IsActive: true, TeamName: "backend"},
	}
	s.rawDB.Create(&users)

	reqDTO := pullrequest.CreatePRRequestDTO{
		PRID:     "pr-100",
		Name:     "New Feature",
		AuthorID: "u1",
	}
	bodyBytes, _ := json.Marshal(reqDTO)

	req, _ := http.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewBuffer(bodyBytes))
	rr := httptest.NewRecorder()

	s.router.ServeHTTP(rr, req)

	s.Equal(http.StatusCreated, rr.Code)

	var resp pullrequest.PRResponseWrapper
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	s.Require().NoError(err)

	s.Equal("OPEN", resp.PR.Status)
	s.Len(resp.PR.Reviewers, 2)
}

func (s *PRSuite) TestMergePR_Success() {
	team := model.Team{Name: "backend"}
	s.rawDB.Create(&team)
	u1 := model.User{ID: "u1", TeamName: "backend"}
	s.rawDB.Create(&u1)

	pr := model.PullRequest{ID: "pr-200", Name: "Fix", AuthorID: "u1", Status: "OPEN"}
	s.rawDB.Create(&pr)

	reqDTO := pullrequest.MergePRRequestDTO{PRID: "pr-200"}
	bodyBytes, _ := json.Marshal(reqDTO)

	req1, _ := http.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewBuffer(bodyBytes))
	rr1 := httptest.NewRecorder()
	s.router.ServeHTTP(rr1, req1)

	s.Equal(http.StatusOK, rr1.Code)

	var prFromDB model.PullRequest
	s.rawDB.First(&prFromDB, "pull_request_id = ?", "pr-200")
	s.Equal("MERGED", prFromDB.Status)
	s.NotNil(prFromDB.MergedAt)

	req2, _ := http.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewBuffer(bodyBytes))
	rr2 := httptest.NewRecorder()
	s.router.ServeHTTP(rr2, req2)

	s.Equal(http.StatusOK, rr2.Code)
}

func (s *PRSuite) TestReassign_Success() {
	team := model.Team{Name: "backend"}
	s.rawDB.Create(&team)

	users := []model.User{
		{ID: "u1", Username: "Author", IsActive: true, TeamName: "backend"},
		{ID: "u2", Username: "OldRev", IsActive: true, TeamName: "backend"},
		{ID: "u3", Username: "NewRev", IsActive: true, TeamName: "backend"},
	}
	s.rawDB.Create(&users)

	pr := model.PullRequest{
		ID:        "pr-300",
		Status:    "OPEN",
		AuthorID:  "u1",
		Reviewers: []*model.User{&users[1]},
	}
	s.rawDB.Create(&pr)

	reqDTO := pullrequest.ReassignPRRequestDTO{
		PRID:      "pr-300",
		OldUserID: "u2",
	}
	bodyBytes, _ := json.Marshal(reqDTO)

	req, _ := http.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewBuffer(bodyBytes))
	rr := httptest.NewRecorder()

	s.router.ServeHTTP(rr, req)

	s.Equal(http.StatusOK, rr.Code)

	var resp pullrequest.ReassignResponseWrapper
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	s.Require().NoError(err)

	s.Equal("u3", resp.ReplacedBy)

	var prFromDB model.PullRequest
	s.rawDB.Preload("Reviewers").First(&prFromDB, "pull_request_id = ?", "pr-300")

	s.Len(prFromDB.Reviewers, 1)
	s.Equal("u3", prFromDB.Reviewers[0].ID)
}
