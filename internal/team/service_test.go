package team

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

func (m *MockStorer) Create(ctx context.Context, team *model.Team) error {
	args := m.Called(ctx, team)
	return args.Error(0)
}

func (m *MockStorer) GetByName(ctx context.Context, name string) (*model.Team, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Team), args.Error(1)
}

// --- Helpers ---

func setupService() (*Service, *MockStorer) {
	mockRepo := new(MockStorer)
	// Используем Discard handler, чтобы логи не засоряли вывод тестов
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := NewService(mockRepo, logger)
	return svc, mockRepo
}

// --- Tests ---

func TestService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		svc, mockRepo := setupService()
		inputTeam := &model.Team{Name: "Backend"}

		// Ожидаем успешное создание
		mockRepo.On("Create", ctx, inputTeam).Return(nil)

		result, err := svc.Create(ctx, inputTeam)

		require.NoError(t, err)
		assert.Equal(t, "Backend", result.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("team already exists", func(t *testing.T) {
		svc, mockRepo := setupService()
		inputTeam := &model.Team{Name: "Backend"}

		// Симулируем ошибку дубликата от GORM
		mockRepo.On("Create", ctx, inputTeam).Return(gorm.ErrDuplicatedKey)

		_, err := svc.Create(ctx, inputTeam)

		// Проверяем, что сервис вернул ErrTeamExists
		require.ErrorIs(t, err, ErrTeamExists)
		mockRepo.AssertExpectations(t)
	})

	t.Run("generic error", func(t *testing.T) {
		svc, mockRepo := setupService()
		inputTeam := &model.Team{Name: "Backend"}
		unexpectedErr := errors.New("db connection failed")

		mockRepo.On("Create", ctx, inputTeam).Return(unexpectedErr)

		_, err := svc.Create(ctx, inputTeam)

		// Проверяем, что возвращается исходная ошибка
		assert.ErrorIs(t, err, unexpectedErr)
	})
}

func TestService_GetByName(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		svc, mockRepo := setupService()
		expectedTeam := &model.Team{
			Name: "Frontend",
			// Members могут быть nil или пустым слайсом, это не влияет на тест
		}

		mockRepo.On("GetByName", ctx, "Frontend").Return(expectedTeam, nil)

		result, err := svc.GetByName(ctx, "Frontend")

		require.NoError(t, err)
		assert.Equal(t, expectedTeam.Name, result.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("team not found", func(t *testing.T) {
		svc, mockRepo := setupService()

		// Симулируем ошибку записи не найдено
		mockRepo.On("GetByName", ctx, "Unknown").Return(nil, gorm.ErrRecordNotFound)

		_, err := svc.GetByName(ctx, "Unknown")

		// Проверяем, что сервис конвертирует ошибку в ErrTeamNotFound
		require.ErrorIs(t, err, ErrTeamNotFound)
		mockRepo.AssertExpectations(t)
	})

	t.Run("generic error", func(t *testing.T) {
		svc, mockRepo := setupService()
		unexpectedErr := errors.New("timeout")

		mockRepo.On("GetByName", ctx, "Broken").Return(nil, unexpectedErr)

		_, err := svc.GetByName(ctx, "Broken")

		assert.ErrorIs(t, err, unexpectedErr)
	})
}
