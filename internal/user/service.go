package user

import (
	"context"
	"errors"

	"github.com/SeeXWH/pr-reviewer-service/internal/model"
	"gorm.io/gorm"
)

type Service struct {
	repo UserStorer
}

func NewService(repo UserStorer) *Service {
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

func (s *Service) GetReviews(ctx context.Context, userID string) ([]model.PullRequest, error) {
	prs, err := s.repo.GetUserReviews(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return prs, nil
}
