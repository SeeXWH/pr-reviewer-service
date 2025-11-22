//go:build integration

package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SeeXWH/pr-reviewer-service/configs"
	"github.com/SeeXWH/pr-reviewer-service/internal/analytics"
	"github.com/SeeXWH/pr-reviewer-service/internal/model"
	"github.com/SeeXWH/pr-reviewer-service/pkg/db"
	"github.com/SeeXWH/pr-reviewer-service/pkg/logger"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

func TestAnalyticsSuite(t *testing.T) {
	suite.Run(t, new(AnalyticsSuit))
}

type AnalyticsSuit struct {
	suite.Suite
	rawDB       *gorm.DB
	dbWrapper   *db.PostgresDB
	router      http.Handler
	cleanUpFunc func()
}

func (s *AnalyticsSuit) SetupSuite() {
	ctx := context.Background()
	log := logger.Setup()
	mux := http.NewServeMux()

	pgContainer, cleanup, err := SetupPostgresContainer()
	s.Require().NoError(err)
	s.cleanUpFunc = cleanup

	host, err := pgContainer.Host(ctx)
	s.Require().NoError(err)

	natPort, err := pgContainer.MappedPort(ctx, "5432")
	s.Require().NoError(err)

	testConfig := &configs.Config{
		DB: configs.DB{
			Username: "user",
			Password: "password",
			Dbname:   "testdb",
			Host:     host,
			Port:     natPort.Port(),
		},
		App: configs.App{
			TimeOut: 300 * time.Millisecond,
		},
	}

	var errDB error
	for i := 0; i < 10; i++ {
		s.dbWrapper, errDB = db.NewPostgresDB(testConfig)
		if errDB == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	s.Require().NoError(errDB)
	s.rawDB = s.dbWrapper.PostgresDB

	err = s.rawDB.AutoMigrate(&model.Team{}, &model.User{}, &model.PullRequest{})
	s.Require().NoError(err)

	analyticsRepo := analytics.NewRepository(s.dbWrapper)
	analyticsService := analytics.NewService(analyticsRepo, log)
	analytics.NewHandler(mux, analyticsService, testConfig)
	s.router = mux
}

func (s *AnalyticsSuit) TearDownSuite() {
	if s.cleanUpFunc != nil {
		s.cleanUpFunc()
	}
}

func (s *AnalyticsSuit) SetupTest() {
	s.rawDB.Exec("TRUNCATE TABLE pr_reviewers CASCADE")
	s.rawDB.Exec("TRUNCATE TABLE pull_requests CASCADE")
	s.rawDB.Exec("TRUNCATE TABLE users CASCADE")
	s.rawDB.Exec("TRUNCATE TABLE teams CASCADE")
}

func (s *AnalyticsSuit) TestGetStats() {
	team := model.Team{Name: "backend"}
	s.Require().NoError(s.rawDB.Create(&team).Error)

	u1 := model.User{ID: "u1", Username: "Alice", IsActive: true, TeamName: "backend"}
	u2 := model.User{ID: "u2", Username: "Bob", IsActive: true, TeamName: "backend"}
	u3 := model.User{ID: "u3_author", Username: "Author", IsActive: true, TeamName: "backend"}
	s.Require().NoError(s.rawDB.Create(&u1).Error)
	s.Require().NoError(s.rawDB.Create(&u2).Error)
	s.Require().NoError(s.rawDB.Create(&u3).Error)

	pr1 := model.PullRequest{ID: "pr-1", AuthorID: "u3_author", Reviewers: []*model.User{&u1, &u2}}
	pr2 := model.PullRequest{ID: "pr-2", AuthorID: "u3_author", Reviewers: []*model.User{&u1}}

	s.Require().NoError(s.rawDB.Create(&pr1).Error)
	s.Require().NoError(s.rawDB.Create(&pr2).Error)

	req, _ := http.NewRequest(http.MethodGet, "/analytics/pr", nil)
	rr := httptest.NewRecorder()

	s.router.ServeHTTP(rr, req)

	s.Equal(http.StatusOK, rr.Code)

	var resp analytics.StatsResponseDTO
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	s.Require().NoError(err)

	s.Require().Len(resp.Stats, 2)

	s.Equal("u1", resp.Stats[0].UserID)
	s.Equal(2, resp.Stats[0].ReviewCount)

	s.Equal("u2", resp.Stats[1].UserID)
	s.Equal(1, resp.Stats[1].ReviewCount)
}
