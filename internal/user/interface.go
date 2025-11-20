package user

import (
	"context"

	"github.com/SeeXWH/pr-reviewer-service/internal/model"
)

type Provider interface {
	SetIsActive(context.Context, string, bool) (*model.User, error)
	GetReviews(context.Context, string) ([]model.PullRequest, error)
}

type Storer interface {
	UpdateActiveStatus(context.Context, string, bool) (*model.User, error)
	GetUserReviews(context.Context, string) ([]model.PullRequest, error)
	GetByID(context.Context, string) (*model.User, error)
	GetReviewCandidates(context.Context, string, string) ([]model.User, error)
	GetReplacementCandidate(context.Context, string, []string) (*model.User, error)
}
