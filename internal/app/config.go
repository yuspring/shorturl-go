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
	godotenv.Load()

	adminUser := os.Getenv("ADMIN_USER")
	adminPass := os.Getenv("ADMIN_PASS")
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	conf := &Config{
		AdminUser: adminUser,
		AdminPass: adminPass,
		RedisAddr: redisAddr,
		Port:      port,
	}

	if conf.AdminUser == "" || conf.AdminPass == "" {
		slog.Error("Critical security error: ADMIN_USER or ADMIN_PASS environment variables are not set")
		return nil, os.ErrInvalid
	}

	return conf, nil
}
