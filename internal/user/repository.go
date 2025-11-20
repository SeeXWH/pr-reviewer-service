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

func (r *Repository) GetReviewCandidates(ctx context.Context, teamName string, authorID string) ([]model.User, error) {
	var candidates []model.User
	err := r.db.PostgresDb.WithContext(ctx).
		Where("team_name = ? AND is_active = ? AND user_id != ?", teamName, true, authorID).
		Order("RANDOM()").
		Limit(2).
		Find(&candidates).Error

	if err != nil {
		return nil, err
	}
	return candidates, nil
}

func (r *Repository) GetByID(ctx context.Context, userID string) (*model.User, error) {
	var user model.User
	err := r.db.PostgresDb.WithContext(ctx).First(&user, "user_id = ?", userID).Error
	return &user, err
}

func (r *Repository) GetReplacementCandidate(ctx context.Context, teamName string, excludeUserIDs []string) (*model.User, error) {
	var candidate model.User
	err := r.db.PostgresDb.WithContext(ctx).
		Where("team_name = ? AND is_active = ?", teamName, true).
		Not("user_id", excludeUserIDs).
		Order("RANDOM()").
		First(&candidate).Error

	if err != nil {
		return nil, err
	}
	return &candidate, nil
}
