package user

import (
	"context"

	"github.com/SeeXWH/pr-reviewer-service/internal/model"
)

type UserProvider interface {
	SetIsActive(context.Context, string, bool) (*model.User, error)
	GetReviews(context.Context, string) ([]model.PullRequest, error)
}

type UserStorer interface {
	UpdateActiveStatus(context.Context, string, bool) (*model.User, error)
	GetUserReviews(context.Context, string) ([]model.PullRequest, error)
}
