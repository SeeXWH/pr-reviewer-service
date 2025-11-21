package analytics

import (
	"context"
	"log/slog"
)

type Service struct {
	repo Storer
	log  *slog.Logger
}

func NewService(repository Storer, log *slog.Logger) *Service {
	return &Service{
		repo: repository,
		log:  log.With("component", "analyticService"),
	}
}

func (s *Service) GetStats(ctx context.Context) ([]ReviewerStat, error) {
	stats, err := s.repo.GetReviewerStats(ctx)
	if err != nil {
		s.log.ErrorContext(ctx, "failed to fetch reviewer stats", "op", "GetStats", "error", err)
		return nil, err
	}

	return stats, nil
}
