package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/SeeXWH/pr-reviewer-service/internal/model"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("Start migration")
	temp := time.Now()
	_ = godotenv.Load()
	DSN := fmt.Sprintf("host=%s user=%s password=%s port=%s dbname=%s sslmode=%s",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
		"disable")
	log.Println(DSN)
	db, err := gorm.Open(postgres.Open(DSN), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Database connected. Running AutoMigrate...")
	err = db.AutoMigrate(&model.Team{}, &model.User{}, &model.PullRequest{})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Migration completed successfully in %.3fs", time.Since(temp).Seconds())
}
