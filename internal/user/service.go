package user

import (
	"context"
	"errors"
	"log/slog"

	"github.com/SeeXWH/pr-reviewer-service/internal/model"

	"gorm.io/gorm"
)

type Service struct {
	repo Storer
	log  *slog.Logger
}

func NewService(repo Storer, log *slog.Logger) *Service {
	return &Service{
		repo: repo,
		log:  log.With("component", "userService"),
	}
}

func (s *Service) SetIsActive(ctx context.Context, userID string, isActive bool) (*model.User, error) {
	log := s.log.With("op", "SetIsActive", "user_id", userID, "is_active", isActive)

	updatedUser, err := s.repo.UpdateActiveStatus(ctx, userID, isActive)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.WarnContext(ctx, "failed to update status: user not found")
			return nil, ErrUserNotFound
		}
		log.ErrorContext(ctx, "failed to update active status", "error", err)
		return nil, err
	}

	log.InfoContext(ctx, "user active status updated")
	return updatedUser, nil
}

func (s *Service) GetReviews(ctx context.Context, userID string) ([]model.PullRequest, error) {
	log := s.log.With("op", "GetReviews", "user_id", userID)

	prs, err := s.repo.GetUserReviews(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		log.ErrorContext(ctx, "failed to get user reviews", "error", err)
		return nil, err
	}
	return prs, nil
}

func (s *Service) GetReviewCandidates(
	ctx context.Context,
	teamName string,
	excludeUserID string,
) ([]model.User, error) {
	log := s.log.With("op", "GetReviewCandidates", "team", teamName)

	users, err := s.repo.GetReviewCandidates(ctx, teamName, excludeUserID)
	if err != nil {
		log.ErrorContext(ctx, "failed to fetch candidates", "error", err)
		return nil, err
	}
	return users, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			s.log.ErrorContext(ctx, "failed to get user by id", "op", "GetByID", "user_id", id, "error", err)
		}
		return nil, err
	}
	return user, nil
}

func (s *Service) GetReplacementCandidate(
	ctx context.Context,
	teamName string,
	excludeUserIDs []string,
) (*model.User, error) {
	log := s.log.With("op", "GetReplacementCandidate", "team", teamName, "excluded_count", len(excludeUserIDs))

	candidate, err := s.repo.GetReplacementCandidate(ctx, teamName, excludeUserIDs)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.WarnContext(ctx, "no replacement candidate found")
			return nil, err
		}
		log.ErrorContext(ctx, "failed to find replacement", "error", err)
		return nil, err
	}

	log.InfoContext(ctx, "replacement candidate selected", "candidate_id", candidate.ID)
	return candidate, nil
}
