package pullrequest

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/SeeXWH/pr-reviewer-service/internal/model"

	"gorm.io/gorm"
)

const (
	MergeStatus = "MERGED"
)

type Service struct {
	repo         PRStorer
	userProvider UserProvider
	log          *slog.Logger
}

func NewService(userProvider UserProvider, repo PRStorer, log *slog.Logger) *Service {
	return &Service{
		repo:         repo,
		userProvider: userProvider,
		log:          log.With("component", "prService"),
	}
}

func (s *Service) Create(ctx context.Context, pr model.PullRequest) (*model.PullRequest, error) {
	log := s.log.With("op", "Create", "author_id", pr.AuthorID, "name", pr.Name)

	author, err := s.userProvider.GetByID(ctx, pr.AuthorID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.WarnContext(ctx, "failed to create pr: author not found")
			return nil, ErrAuthorNotFound
		}
		log.ErrorContext(ctx, "failed to fetch author", "error", err)
		return nil, err
	}

	candidates, err := s.userProvider.GetReviewCandidates(ctx, author.TeamName, author.ID)
	if err != nil {
		log.ErrorContext(ctx, "failed to fetch review candidates", "error", err)
		return nil, err
	}

	pr.Status = "OPEN"
	pr.CreatedAt = time.Now()

	pr.Reviewers = make([]*model.User, len(candidates))
	for i := range candidates {
		pr.Reviewers[i] = &candidates[i]
	}

	err = s.repo.Create(ctx, &pr)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			log.WarnContext(ctx, "pr already exists")
			return nil, ErrPRExists
		}
		log.ErrorContext(ctx, "failed to create pr", "error", err)
		return nil, err
	}

	log.InfoContext(ctx, "pr created", "pr_id", pr.ID, "reviewers_count", len(pr.Reviewers))
	return &pr, nil
}

func (s *Service) Merge(ctx context.Context, prID string) (*model.PullRequest, error) {
	log := s.log.With("op", "Merge", "pr_id", prID)

	pr, err := s.repo.GetByID(ctx, prID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPRNotFound
		}
		log.ErrorContext(ctx, "failed to fetch pr", "error", err)
		return nil, err
	}
	if pr.Status == MergeStatus {
		return pr, nil
	}
	pr.Status = MergeStatus
	now := time.Now()
	pr.MergedAt = &now
	if err = s.repo.Update(ctx, pr); err != nil {
		log.ErrorContext(ctx, "failed to update pr status", "error", err)
		return nil, err
	}

	log.InfoContext(ctx, "pr merged")
	return pr, nil
}

func (s *Service) ReassignReviewer(
	ctx context.Context,
	prID string,
	oldUserID string,
) (*model.PullRequest, *model.User, error) {
	log := s.log.With("op", "ReassignReviewer", "pr_id", prID, "old_user_id", oldUserID)

	pr, err := s.getAndValidatePR(ctx, prID)
	if err != nil {
		return nil, nil, err
	}
	excludeIDs, err := s.validateAssignmentAndGetExclusions(pr, oldUserID)
	if err != nil {
		log.WarnContext(ctx, "validation failed", "error", err)
		return nil, nil, err
	}
	author, err := s.userProvider.GetByID(ctx, pr.AuthorID)
	if err != nil {
		log.ErrorContext(ctx, "failed to fetch author details", "error", err)
		return nil, nil, err
	}
	newReviewer, err := s.findReplacement(ctx, author.TeamName, excludeIDs)
	if err != nil {
		return nil, nil, err
	}
	pr.Reviewers = replaceReviewerInSlice(pr.Reviewers, oldUserID, newReviewer)
	if err = s.repo.UpdateReviewers(ctx, pr); err != nil {
		log.ErrorContext(ctx, "failed to update reviewers list", "error", err)
		return nil, nil, err
	}

	log.InfoContext(ctx, "reviewer reassigned", "new_user_id", newReviewer.ID)
	return pr, newReviewer, nil
}

func (s *Service) getAndValidatePR(ctx context.Context, prID string) (*model.PullRequest, error) {
	pr, err := s.repo.GetByID(ctx, prID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPRNotFound
		}
		s.log.ErrorContext(ctx, "failed to fetch pr by id", "op", "getAndValidatePR", "pr_id", prID, "error", err)
		return nil, err
	}
	if pr.Status == MergeStatus {
		return nil, ErrPRMerged
	}
	return pr, nil
}

func (s *Service) validateAssignmentAndGetExclusions(pr *model.PullRequest, oldUserID string) ([]string, error) {
	isAssigned := false
	excludeIDs := []string{pr.AuthorID}
	for _, r := range pr.Reviewers {
		excludeIDs = append(excludeIDs, r.ID)
		if r.ID == oldUserID {
			isAssigned = true
		}
	}
	if !isAssigned {
		return nil, ErrNotAssigned
	}
	return excludeIDs, nil
}

func (s *Service) findReplacement(ctx context.Context, teamName string, excludeIDs []string) (*model.User, error) {
	newReviewer, err := s.userProvider.GetReplacementCandidate(ctx, teamName, excludeIDs)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.log.WarnContext(ctx, "no replacement candidate available", "op", "findReplacement", "team", teamName)
			return nil, ErrNoCandidate
		}
		s.log.ErrorContext(ctx, "failed to find replacement candidate", "op", "findReplacement", "error", err)
		return nil, err
	}
	return newReviewer, nil
}

func replaceReviewerInSlice(currentReviewers []*model.User, oldUserID string, newReviewer *model.User) []*model.User {
	updatedList := make([]*model.User, 0, len(currentReviewers))
	for _, r := range currentReviewers {
		if r.ID == oldUserID {
			updatedList = append(updatedList, newReviewer)
		} else {
			updatedList = append(updatedList, r)
		}
	}
	return updatedList
}
