package user

import (
	"context"
	"errors"

	"github.com/SeeXWH/pr-reviewer-service/internal/model"

	"gorm.io/gorm"
)

type Service struct {
	repo Storer
}

func NewService(repo Storer) *Service {
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

func (s *Service) GetReviewCandidates(ctx context.Context, teamName string, excludeUserID string) ([]model.User, error) {
	return s.repo.GetReviewCandidates(ctx, teamName, excludeUserID)
}

func (s *Service) GetByID(ctx context.Context, id string) (*model.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetReplacementCandidate(ctx context.Context, teamName string, excludeUserIDs []string) (*model.User, error) {
	return s.repo.GetReplacementCandidate(ctx, teamName, excludeUserIDs)
}
