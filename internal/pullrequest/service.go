package pullrequest

import (
	"context"
	"errors"
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
}

func NewService(userProvider UserProvider, repo PRStorer) *Service {
	return &Service{
		repo:         repo,
		userProvider: userProvider,
	}
}

func (s *Service) Create(ctx context.Context, pr model.PullRequest) (*model.PullRequest, error) {
	author, err := s.userProvider.GetByID(ctx, pr.AuthorID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAuthorNotFound
		}
		return nil, err
	}

	candidates, err := s.userProvider.GetReviewCandidates(ctx, author.TeamName, author.ID)
	if err != nil {
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
			return nil, ErrPRExists
		}
		return nil, err
	}

	return &pr, nil
}

func (s *Service) Merge(ctx context.Context, prID string) (*model.PullRequest, error) {
	pr, err := s.repo.GetByID(ctx, prID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPRNotFound
		}
		return nil, err
	}
	if pr.Status == MergeStatus {
		return pr, nil
	}
	pr.Status = MergeStatus
	now := time.Now()
	pr.MergedAt = &now
	if err = s.repo.Update(ctx, pr); err != nil {
		return nil, err
	}
	return pr, nil
}

func (s *Service) ReassignReviewer(ctx context.Context, prID string, oldUserID string) (*model.PullRequest, *model.User, error) {
	pr, err := s.getAndValidatePR(ctx, prID)
	if err != nil {
		return nil, nil, err
	}
	excludeIDs, err := s.validateAssignmentAndGetExclusions(pr, oldUserID)
	if err != nil {
		return nil, nil, err
	}
	author, err := s.userProvider.GetByID(ctx, pr.AuthorID)
	if err != nil {
		return nil, nil, err
	}
	newReviewer, err := s.findReplacement(ctx, author.TeamName, excludeIDs)
	if err != nil {
		return nil, nil, err
	}
	pr.Reviewers = replaceReviewerInSlice(pr.Reviewers, oldUserID, newReviewer)
	if err = s.repo.UpdateReviewers(ctx, pr); err != nil {
		return nil, nil, err
	}
	return pr, newReviewer, nil
}

func (s *Service) getAndValidatePR(ctx context.Context, prID string) (*model.PullRequest, error) {
	pr, err := s.repo.GetByID(ctx, prID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPRNotFound
		}
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
			return nil, ErrNoCandidate
		}
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
