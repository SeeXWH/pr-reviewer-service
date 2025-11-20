package configs

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DB DB
}

type DB struct {
	Username string
	Password string
	Dbname   string
	Host     string
	Port     string
}

func Load() *Config {
	_ = godotenv.Load(".env")
	return &Config{
		DB: DB{
			Username: os.Getenv("POSTGRES_USER"),
			Password: os.Getenv("POSTGRES_PASSWORD"),
			Dbname:   os.Getenv("POSTGRES_DB"),
			Host:     os.Getenv("POSTGRES_HOST"),
			Port:     os.Getenv("POSTGRES_PORT"),
		},
	}
}
