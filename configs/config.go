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
			Password: os.Getenv("12345"),
			Dbname:   os.Getenv("PrDB"),
			Host:     os.Getenv("localhost"),
			Port:     os.Getenv("5432"),
		},
	}
}
