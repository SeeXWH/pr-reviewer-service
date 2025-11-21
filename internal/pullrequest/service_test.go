package pullrequest

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/SeeXWH/pr-reviewer-service/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type MockUserProvider struct {
	mock.Mock
}

func (m *MockUserProvider) GetByID(ctx context.Context, id string) (*model.User, error) {
	args := m.Called(ctx, id)
	if val, ok := args.Get(0).(*model.User); ok {
		return val, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserProvider) GetReviewCandidates(
	ctx context.Context,
	teamName string,
	excludeUserID string,
) ([]model.User, error) {
	args := m.Called(ctx, teamName, excludeUserID)
	if val, ok := args.Get(0).([]model.User); ok {
		return val, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserProvider) GetReplacementCandidate(
	ctx context.Context,
	teamName string,
	excludeUserIDs []string,
) (*model.User, error) {
	args := m.Called(ctx, teamName, excludeUserIDs)
	if val, ok := args.Get(0).(*model.User); ok {
		return val, args.Error(1)
	}
	return nil, args.Error(1)
}

type MockPRStorer struct {
	mock.Mock
}

func (m *MockPRStorer) Create(ctx context.Context, pr *model.PullRequest) error {
	args := m.Called(ctx, pr)
	return args.Error(0)
}

func (m *MockPRStorer) GetByID(ctx context.Context, id string) (*model.PullRequest, error) {
	args := m.Called(ctx, id)
	if val, ok := args.Get(0).(*model.PullRequest); ok {
		return val, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPRStorer) Update(ctx context.Context, pr *model.PullRequest) error {
	args := m.Called(ctx, pr)
	return args.Error(0)
}

func (m *MockPRStorer) UpdateReviewers(ctx context.Context, pr *model.PullRequest) error {
	args := m.Called(ctx, pr)
	return args.Error(0)
}

func setupService() (*Service, *MockUserProvider, *MockPRStorer) {
	mockUser := new(MockUserProvider)
	mockRepo := new(MockPRStorer)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := NewService(mockUser, mockRepo, logger)
	return svc, mockUser, mockRepo
}

func TestService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		svc, mockUser, mockRepo := setupService()
		inputPR := model.PullRequest{AuthorID: "u1", Name: "Feature"}
		author := &model.User{ID: "u1", TeamName: "Alpha"}
		candidates := []model.User{{ID: "r1"}, {ID: "r2"}}

		mockUser.On("GetByID", ctx, "u1").Return(author, nil)
		mockUser.On("GetReviewCandidates", ctx, "Alpha", "u1").Return(candidates, nil)
		mockRepo.On("Create", ctx, mock.MatchedBy(func(pr *model.PullRequest) bool {
			return pr.Status == "OPEN" && len(pr.Reviewers) == 2 && pr.CreatedAt.After(time.Time{})
		})).Return(nil)

		res, err := svc.Create(ctx, inputPR)

		require.NoError(t, err)
		assert.Equal(t, "OPEN", res.Status)
		mockUser.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("author not found", func(t *testing.T) {
		svc, mockUser, _ := setupService()
		inputPR := model.PullRequest{AuthorID: "unknown"}

		mockUser.On("GetByID", ctx, "unknown").Return(nil, gorm.ErrRecordNotFound)

		_, err := svc.Create(ctx, inputPR)

		assert.ErrorIs(t, err, ErrAuthorNotFound)
	})

	t.Run("pr already exists", func(t *testing.T) {
		svc, mockUser, mockRepo := setupService()
		inputPR := model.PullRequest{AuthorID: "u1"}
		author := &model.User{ID: "u1", TeamName: "Alpha"}

		mockUser.On("GetByID", ctx, "u1").Return(author, nil)
		mockUser.On("GetReviewCandidates", ctx, "Alpha", "u1").Return([]model.User{}, nil)
		mockRepo.On("Create", ctx, mock.Anything).Return(gorm.ErrDuplicatedKey)

		_, err := svc.Create(ctx, inputPR)

		assert.ErrorIs(t, err, ErrPRExists)
	})
}

func TestService_Merge(t *testing.T) {
	ctx := context.Background()

	t.Run("success merge", func(t *testing.T) {
		svc, _, mockRepo := setupService()
		pr := &model.PullRequest{ID: "pr-1", Status: "OPEN"}

		mockRepo.On("GetByID", ctx, "pr-1").Return(pr, nil)
		mockRepo.On("Update", ctx, mock.MatchedBy(func(updated *model.PullRequest) bool {
			return updated.Status == MergeStatus && updated.MergedAt != nil
		})).Return(nil)

		res, err := svc.Merge(ctx, "pr-1")

		require.NoError(t, err)
		assert.Equal(t, MergeStatus, res.Status)
		assert.NotNil(t, res.MergedAt)
	})

	t.Run("already merged", func(t *testing.T) {
		svc, _, mockRepo := setupService()
		pr := &model.PullRequest{ID: "pr-1", Status: MergeStatus}

		mockRepo.On("GetByID", ctx, "pr-1").Return(pr, nil)

		res, err := svc.Merge(ctx, "pr-1")

		require.NoError(t, err)
		assert.Equal(t, MergeStatus, res.Status)
		mockRepo.AssertNotCalled(t, "Update", ctx, mock.Anything)
	})

	t.Run("pr not found", func(t *testing.T) {
		svc, _, mockRepo := setupService()

		mockRepo.On("GetByID", ctx, "missing").Return(nil, gorm.ErrRecordNotFound)

		_, err := svc.Merge(ctx, "missing")

		assert.ErrorIs(t, err, ErrPRNotFound)
	})
}

func TestService_ReassignReviewer(t *testing.T) {
	ctx := context.Background()

	t.Run("success reassign", func(t *testing.T) {
		svc, mockUser, mockRepo := setupService()

		oldRev := &model.User{ID: "old"}
		stayRev := &model.User{ID: "stay"}
		newRev := &model.User{ID: "new"}
		author := &model.User{ID: "author", TeamName: "Devs"}

		pr := &model.PullRequest{
			ID:        "pr-1",
			AuthorID:  "author",
			Status:    "OPEN",
			Reviewers: []*model.User{oldRev, stayRev},
		}

		mockRepo.On("GetByID", ctx, "pr-1").Return(pr, nil)
		mockUser.On("GetByID", ctx, "author").Return(author, nil)
		expectedExcludes := []string{"author", "old", "stay"}
		mockUser.On("GetReplacementCandidate", ctx, "Devs", expectedExcludes).Return(newRev, nil)

		mockRepo.On("UpdateReviewers", ctx, mock.MatchedBy(func(updated *model.PullRequest) bool {
			ids := make([]string, 0, len(updated.Reviewers))
			for _, r := range updated.Reviewers {
				ids = append(ids, r.ID)
			}
			return assert.Contains(t, ids, "new") &&
				assert.Contains(t, ids, "stay") &&
				assert.NotContains(t, ids, "old")
		})).Return(nil)

		resPR, resUser, err := svc.ReassignReviewer(ctx, "pr-1", "old")

		require.NoError(t, err)
		assert.Equal(t, "new", resUser.ID)
		assert.Len(t, resPR.Reviewers, 2)
	})

	t.Run("pr not found", func(t *testing.T) {
		svc, _, mockRepo := setupService()
		mockRepo.On("GetByID", ctx, "pr-1").Return(nil, gorm.ErrRecordNotFound)

		_, _, err := svc.ReassignReviewer(ctx, "pr-1", "any")
		assert.ErrorIs(t, err, ErrPRNotFound)
	})

	t.Run("pr is merged", func(t *testing.T) {
		svc, _, mockRepo := setupService()
		pr := &model.PullRequest{ID: "pr-1", Status: MergeStatus}
		mockRepo.On("GetByID", ctx, "pr-1").Return(pr, nil)

		_, _, err := svc.ReassignReviewer(ctx, "pr-1", "any")
		assert.ErrorIs(t, err, ErrPRMerged)
	})

	t.Run("reviewer not assigned", func(t *testing.T) {
		svc, _, mockRepo := setupService()
		pr := &model.PullRequest{
			ID:        "pr-1",
			Status:    "OPEN",
			Reviewers: []*model.User{{ID: "other"}},
		}
		mockRepo.On("GetByID", ctx, "pr-1").Return(pr, nil)

		_, _, err := svc.ReassignReviewer(ctx, "pr-1", "not-assigned-id")
		assert.ErrorIs(t, err, ErrNotAssigned)
	})

	t.Run("no replacement candidate", func(t *testing.T) {
		svc, mockUser, mockRepo := setupService()
		pr := &model.PullRequest{
			ID:        "pr-1",
			AuthorID:  "author",
			Status:    "OPEN",
			Reviewers: []*model.User{{ID: "old"}},
		}
		author := &model.User{ID: "author", TeamName: "Devs"}

		mockRepo.On("GetByID", ctx, "pr-1").Return(pr, nil)
		mockUser.On("GetByID", ctx, "author").Return(author, nil)
		mockUser.On("GetReplacementCandidate", ctx, "Devs", mock.Anything).Return(nil, gorm.ErrRecordNotFound)

		_, _, err := svc.ReassignReviewer(ctx, "pr-1", "old")
		assert.ErrorIs(t, err, ErrNoCandidate)
	})
}
