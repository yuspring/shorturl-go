package app

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AdminUser string
	AdminPass string
	RedisAddr string
	Port      string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.Warn(".env file not found or could not be loaded", "error", err)
	}

	conf := &Config{
		AdminUser: os.Getenv("USER"),
		AdminPass: os.Getenv("PASSWORD"),
		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
		Port:      getEnv("PORT", "8080"),
	}

	if conf.AdminUser == "" || conf.AdminPass == "" {
		slog.Error("Critical security error: USER or PASSWORD environment variables are not set")
		return nil, os.ErrInvalid
	}

	return conf, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
