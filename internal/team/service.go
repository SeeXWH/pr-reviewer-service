package team

import (
	"context"
	"errors"

	"github.com/SeeXWH/pr-reviewer-service/internal/model"
	"gorm.io/gorm"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, team *model.Team) (*model.Team, error) {
	err := s.repo.Create(ctx, team)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrTeamExists
		}
		return nil, err
	}
	return team, nil
}
