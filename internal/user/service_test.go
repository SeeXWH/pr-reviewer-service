package user

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/SeeXWH/pr-reviewer-service/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type MockStorer struct {
	mock.Mock
}

func (m *MockStorer) UpdateActiveStatus(ctx context.Context, userID string, isActive bool) (*model.User, error) {
	args := m.Called(ctx, userID, isActive)
	if val, ok := args.Get(0).(*model.User); ok {
		return val, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStorer) GetUserReviews(ctx context.Context, userID string) ([]model.PullRequest, error) {
	args := m.Called(ctx, userID)
	if val, ok := args.Get(0).([]model.PullRequest); ok {
		return val, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStorer) GetByID(ctx context.Context, id string) (*model.User, error) {
	args := m.Called(ctx, id)
	if val, ok := args.Get(0).(*model.User); ok {
		return val, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStorer) GetReviewCandidates(
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

func (m *MockStorer) GetReplacementCandidate(
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

func (m *MockStorer) MassDeactivateAndReassign(
	ctx context.Context,
	teamName string,
	userIDs []string,
) (MassDeactivateResult, error) {
	args := m.Called(ctx, teamName, userIDs)
	return args.Get(0).(MassDeactivateResult), args.Error(1)
}

func setupService() (*Service, *MockStorer) {
	mockRepo := new(MockStorer)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := NewService(mockRepo, logger)
	return svc, mockRepo
}

func TestService_SetIsActive(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		svc, mockRepo := setupService()
		expectedUser := &model.User{ID: "u1", IsActive: true}

		mockRepo.On("UpdateActiveStatus", ctx, "u1", true).Return(expectedUser, nil)

		res, err := svc.SetIsActive(ctx, "u1", true)

		require.NoError(t, err)
		assert.Equal(t, expectedUser, res)
		mockRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		svc, mockRepo := setupService()

		mockRepo.On("UpdateActiveStatus", ctx, "missing", false).Return(nil, gorm.ErrRecordNotFound)

		_, err := svc.SetIsActive(ctx, "missing", false)

		assert.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("db error", func(t *testing.T) {
		svc, mockRepo := setupService()
		unexpectedErr := errors.New("oops")

		mockRepo.On("UpdateActiveStatus", ctx, "u1", true).Return(nil, unexpectedErr)

		_, err := svc.SetIsActive(ctx, "u1", true)

		assert.ErrorIs(t, err, unexpectedErr)
	})
}

func TestService_GetReviews(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		svc, mockRepo := setupService()
		expectedPRs := []model.PullRequest{{ID: "pr1"}, {ID: "pr2"}}

		mockRepo.On("GetUserReviews", ctx, "u1").Return(expectedPRs, nil)

		res, err := svc.GetReviews(ctx, "u1")

		require.NoError(t, err)
		assert.Len(t, res, 2)
	})

	t.Run("user not found (no reviews records)", func(t *testing.T) {
		svc, mockRepo := setupService()

		mockRepo.On("GetUserReviews", ctx, "u1").Return(nil, gorm.ErrRecordNotFound)

		_, err := svc.GetReviews(ctx, "u1")

		assert.ErrorIs(t, err, ErrUserNotFound)
	})
}

func TestService_GetReviewCandidates(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		svc, mockRepo := setupService()
		expectedUsers := []model.User{{ID: "u2"}, {ID: "u3"}}

		mockRepo.On("GetReviewCandidates", ctx, "TeamA", "u1").Return(expectedUsers, nil)

		res, err := svc.GetReviewCandidates(ctx, "TeamA", "u1")

		require.NoError(t, err)
		assert.Equal(t, expectedUsers, res)
	})

	t.Run("db error", func(t *testing.T) {
		svc, mockRepo := setupService()
		mockRepo.On("GetReviewCandidates", ctx, "TeamA", "u1").Return(nil, errors.New("db fail"))

		_, err := svc.GetReviewCandidates(ctx, "TeamA", "u1")

		assert.Error(t, err)
	})
}

func TestService_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		svc, mockRepo := setupService()
		expectedUser := &model.User{ID: "u1"}

		mockRepo.On("GetByID", ctx, "u1").Return(expectedUser, nil)

		res, err := svc.GetByID(ctx, "u1")

		require.NoError(t, err)
		assert.Equal(t, expectedUser, res)
	})

	t.Run("not found returns original error", func(t *testing.T) {
		svc, mockRepo := setupService()

		mockRepo.On("GetByID", ctx, "u1").Return(nil, gorm.ErrRecordNotFound)

		_, err := svc.GetByID(ctx, "u1")

		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func TestService_GetReplacementCandidate(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		svc, mockRepo := setupService()
		excludes := []string{"u1", "u2"}
		candidate := &model.User{ID: "u3"}

		mockRepo.On("GetReplacementCandidate", ctx, "TeamA", excludes).Return(candidate, nil)

		res, err := svc.GetReplacementCandidate(ctx, "TeamA", excludes)

		require.NoError(t, err)
		assert.Equal(t, candidate, res)
	})

	t.Run("not found returns original error", func(t *testing.T) {
		svc, mockRepo := setupService()
		excludes := []string{"u1"}

		mockRepo.On("GetReplacementCandidate", ctx, "TeamA", excludes).Return(nil, gorm.ErrRecordNotFound)

		_, err := svc.GetReplacementCandidate(ctx, "TeamA", excludes)

		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func TestService_MassDeactivate(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		svc, mockRepo := setupService()
		ids := []string{"u1", "u2"}
		expectedResult := MassDeactivateResult{
			DeactivatedCount: 2,
			ReassignedCount:  5,
		}

		mockRepo.On("MassDeactivateAndReassign", ctx, "TeamA", ids).Return(expectedResult, nil)

		res, err := svc.MassDeactivate(ctx, "TeamA", ids)

		require.NoError(t, err)
		assert.Equal(t, 2, res.DeactivatedCount)
		assert.Equal(t, 5, res.ReassignedCount)
	})

	t.Run("error", func(t *testing.T) {
		svc, mockRepo := setupService()
		ids := []string{"u1"}

		mockRepo.On("MassDeactivateAndReassign", ctx, "TeamA", ids).
			Return(MassDeactivateResult{}, errors.New("tx failed"))

		_, err := svc.MassDeactivate(ctx, "TeamA", ids)

		assert.Error(t, err)
	})
}
