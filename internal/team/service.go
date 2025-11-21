package team

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
		log:  log.With("component", "teamService"),
	}
}

func (s *Service) Create(ctx context.Context, team *model.Team) (*model.Team, error) {
	log := s.log.With("op", "Create", "team_name", team.Name)

	err := s.repo.Create(ctx, team)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			log.WarnContext(ctx, "team already exists")
			return nil, ErrTeamExists
		}
		log.ErrorContext(ctx, "failed to create team", "error", err)
		return nil, err
	}

	log.InfoContext(ctx, "team created")
	return team, nil
}

func (s *Service) GetByName(ctx context.Context, name string) (*model.Team, error) {
	team, err := s.repo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTeamNotFound
		}
		s.log.ErrorContext(ctx, "failed to get team by name", "op", "GetByName", "team_name", name, "error", err)
		return nil, err
	}
	return team, nil
}
