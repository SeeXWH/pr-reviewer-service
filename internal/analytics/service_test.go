package analytics

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockStorer struct {
	mock.Mock
}

func (m *MockStorer) GetReviewerStats(ctx context.Context) ([]ReviewerStat, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ReviewerStat), args.Error(1)
}

func setupService() (*Service, *MockStorer) {
	mockRepo := new(MockStorer)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := NewService(mockRepo, logger)
	return svc, mockRepo
}

func TestService_GetStats(t *testing.T) {
	ctx := context.Background()
	dummyStats := []ReviewerStat{{}, {}}

	t.Run("success", func(t *testing.T) {
		svc, mockRepo := setupService()
		mockRepo.On("GetReviewerStats", ctx).Return(dummyStats, nil)

		stats, err := svc.GetStats(ctx)

		require.NoError(t, err)
		assert.Equal(t, dummyStats, stats)
		assert.Len(t, stats, 2)

		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		svc, mockRepo := setupService()
		expectedErr := errors.New("db connection failed")
		mockRepo.On("GetReviewerStats", ctx).Return(nil, expectedErr)
		stats, err := svc.GetStats(ctx)
		require.ErrorIs(t, err, expectedErr)
		assert.Nil(t, stats)

		mockRepo.AssertExpectations(t)
	})
}
