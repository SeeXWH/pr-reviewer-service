package pullrequest

import (
	"context"

	"github.com/SeeXWH/pr-reviewer-service/internal/model"
	"github.com/SeeXWH/pr-reviewer-service/pkg/db"
)

type Repository struct {
	db *db.PostgresDb
}

func NewRepository(db *db.PostgresDb) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, pr *model.PullRequest) error {
	return r.db.PostgresDb.WithContext(ctx).Create(pr).Error
}

func (r *Repository) GetByID(ctx context.Context, id string) (*model.PullRequest, error) {
	var pr model.PullRequest
	err := r.db.PostgresDb.WithContext(ctx).
		Preload("Reviewers").
		First(&pr, "pull_request_id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

func (r *Repository) Update(ctx context.Context, pr *model.PullRequest) error {
	return r.db.PostgresDb.WithContext(ctx).Save(pr).Error
}

func (r *Repository) UpdateReviewers(ctx context.Context, pr *model.PullRequest) error {
	return r.db.PostgresDb.WithContext(ctx).Model(pr).Association("Reviewers").Replace(pr.Reviewers)
}
