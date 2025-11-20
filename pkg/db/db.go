package db

import (
	"fmt"

	"github.com/SeeXWH/pr-reviewer-service/configs"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresDb struct {
	PostgresDb *gorm.DB
}

func NewPostgresDb(conf *configs.Config) *PostgresDb {
	DSN := fmt.Sprintf("host=%s user=%s password=%s port=%s dbname=%s sslmode=%s", conf.DB.Host,
		conf.DB.Username,
		conf.DB.Password,
		conf.DB.Port,
		conf.DB.Dbname,
		"disable")
	db, err := gorm.Open(postgres.Open(DSN), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		panic(err)
	}
	return &PostgresDb{db}
}
