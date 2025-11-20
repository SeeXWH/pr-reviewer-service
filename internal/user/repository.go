package user

import (
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
