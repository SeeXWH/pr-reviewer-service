package team

import (
	"context"

	"github.com/SeeXWH/pr-reviewer-service/internal/model"
)

type TeamProvider interface {
	Create(context.Context, *model.Team) (*model.Team, error)
	GetByName(context.Context, string) (*model.Team, error)
}

type TeamStorer interface {
	Create(context.Context, *model.Team) error
	GetByName(context.Context, string) (*model.Team, error)
}
