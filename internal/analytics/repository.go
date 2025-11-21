package analytics

import (
	"context"

	"github.com/SeeXWH/pr-reviewer-service/pkg/db"
)

type Repository struct {
	db *db.PostgresDB
}

func NewRepository(db *db.PostgresDB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetReviewerStats(ctx context.Context) ([]ReviewerStat, error) {
	var stats []ReviewerStat
	err := r.db.PostgresDB.WithContext(ctx).
		Table("pr_reviewers").
		Select("user_id, count(*) as count").
		Group("user_id").
		Order("count desc").
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	return stats, nil
}
