package team

import (
	"context"

	"github.com/SeeXWH/pr-reviewer-service/internal/model"
)

type Provider interface {
	Create(context.Context, *model.Team) (*model.Team, error)
	GetByName(context.Context, string) (*model.Team, error)
}

type Storer interface {
	Create(context.Context, *model.Team) error
	GetByName(context.Context, string) (*model.Team, error)
}
