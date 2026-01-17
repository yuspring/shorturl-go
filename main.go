package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"shorturl-go/internal/app"
)

func main() {
	cfg, err := app.LoadConfig()
	if err != nil {
		os.Exit(1)
	}

	rdb, err := app.InitRedis(cfg.RedisAddr)
	if err != nil {
		slog.Error("Redis initialization failed", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server, err := app.NewServer(rdb, [2]string{cfg.AdminUser, cfg.AdminPass})
	if err != nil {
		slog.Error("Server initialization failed", "error", err)
		os.Exit(1)
	}

	if err := server.Run(ctx, cfg.Port); err != nil {
		slog.Error("Server shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("Server stopped")
}
