package team

import (
	"context"
	"errors"

	"github.com/SeeXWH/pr-reviewer-service/internal/model"
	"gorm.io/gorm"
)

type Service struct {
	repo TeamStorer
}

func NewService(repo TeamStorer) *Service {
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

func (s *Service) GetByName(ctx context.Context, name string) (*model.Team, error) {
	team, err := s.repo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTeamNotFound
		}
		return nil, err
	}
	return team, nil
}
