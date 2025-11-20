package user

import (
	"context"

	"github.com/SeeXWH/pr-reviewer-service/internal/model"
	"github.com/SeeXWH/pr-reviewer-service/pkg/db"
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
