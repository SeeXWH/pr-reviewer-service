package team

import (
	"context"

	"github.com/SeeXWH/pr-reviewer-service/internal/model"
	"github.com/SeeXWH/pr-reviewer-service/pkg/db"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *db.PostgresDb
}

func NewRepository(db *db.PostgresDb) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(ctx context.Context, team *model.Team) error {
	return r.db.PostgresDb.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model.Team{Name: team.Name}).Error; err != nil {
			return err
		}
		if len(team.Members) > 0 {
			err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "user_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"username", "is_active", "team_name"}),
			}).Create(&team.Members).Error

			if err != nil {
				return err
			}
		}
		return nil
	})
}
