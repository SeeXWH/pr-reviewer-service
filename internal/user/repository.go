package user

import (
	"context"

	"github.com/SeeXWH/pr-reviewer-service/internal/model"
	"github.com/SeeXWH/pr-reviewer-service/pkg/db"
	"gorm.io/gorm"
)

type Repository struct {
	db *db.PostgresDb
}

func NewRepository(db *db.PostgresDb) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) UpdateActiveStatus(ctx context.Context, userID string, isActive bool) (*model.User, error) {
	var user model.User
	err := r.db.PostgresDb.WithContext(ctx).First(&user, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	err = r.db.PostgresDb.WithContext(ctx).Model(&user).Update("is_active", isActive).Error
	if err != nil {
		return nil, err
	}
	user.IsActive = isActive
	return &user, nil
}

func (r *Repository) GetUserReviews(ctx context.Context, userID string) ([]model.PullRequest, error) {
	var count int64
	err := r.db.PostgresDb.WithContext(ctx).Model(&model.User{}).Where("user_id = ?", userID).Count(&count).Error
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var prs []model.PullRequest
	err = r.db.PostgresDb.WithContext(ctx).
		Table("pull_requests").
		Joins("JOIN pr_reviewers ON pr_reviewers.pull_request_id = pull_requests.pull_request_id").
		Where("pr_reviewers.user_id = ?", userID).
		Find(&prs).Error
	if err != nil {
		return nil, err
	}
	return prs, nil
}
