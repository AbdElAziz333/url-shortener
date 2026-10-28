package configloader

import (
	"os"
	"path/filepath"
	"runtime"

	"aziz.dev/shortener/internal/config"
	"github.com/joho/godotenv"
)

func LoadFromEnv() (*config.AppConfig, error) {
	if env := os.Getenv("APP_ENV"); env == "" || env == "local" {
		loadDotEnv()
	}

	return &config.AppConfig{
		Service: config.ServiceConfig{
			Name:     os.Getenv("APP_NAME"),
			Port:     os.Getenv("APP_PORT"),
			GRPCPort: os.Getenv("GRPC_PORT"),
		},
		Postgres: config.PostgresConfig{
			Host:     os.Getenv("POSTGRES_HOST"),
			User:     os.Getenv("POSTGRES_USER"),
			Password: os.Getenv("POSTGRES_PASSWORD"),
			DBName:   os.Getenv("POSTGRES_DB"),
		},
	}, nil
}

func loadDotEnv() {
	if _, file, _, ok := runtime.Caller(0); ok {
		_ = godotenv.Load(filepath.Join(filepath.Dir(file), "..", "..", ".env"))
	}
}