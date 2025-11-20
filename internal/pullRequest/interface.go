package pullRequest

import (
	"context"

	"github.com/SeeXWH/pr-reviewer-service/internal/model"
)

type UserProvider interface {
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetReviewCandidates(ctx context.Context, teamName string, excludeUserID string) ([]model.User, error)
	GetReplacementCandidate(ctx context.Context, teamName string, excludeUserIDs []string) (*model.User, error)
}

type PRStorer interface {
	Create(context.Context, *model.PullRequest) error
	GetByID(context.Context, string) (*model.PullRequest, error)
	Update(context.Context, *model.PullRequest) error
	UpdateReviewers(context.Context, *model.PullRequest) error
}

type PRProvider interface {
	Create(context.Context, model.PullRequest) (*model.PullRequest, error)
	Merge(context.Context, string) (*model.PullRequest, error)
	ReassignReviewer(context.Context, string, string) (*model.PullRequest, *model.User, error)
}
