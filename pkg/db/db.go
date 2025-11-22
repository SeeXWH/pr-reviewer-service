package db

import (
	"fmt"

	"github.com/SeeXWH/pr-reviewer-service/configs"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresDB struct {
	PostgresDB *gorm.DB
}

func NewPostgresDB(conf *configs.Config) (*PostgresDB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s port=%s dbname=%s sslmode=%s",
		conf.DB.Host,
		conf.DB.Username,
		conf.DB.Password,
		conf.DB.Port,
		conf.DB.Dbname,
		"disable",
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err = sqlDB.Ping(); err != nil {
		return nil, err
	}

	return &PostgresDB{db}, nil
}
