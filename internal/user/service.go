package user

import (
	"context"
	"errors"

	"github.com/SeeXWH/pr-reviewer-service/internal/model"
	"gorm.io/gorm"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SetIsActive(ctx context.Context, userID string, isActive bool) (*model.User, error) {
	updatedUser, err := s.repo.UpdateActiveStatus(ctx, userID, isActive)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return updatedUser, nil
}
