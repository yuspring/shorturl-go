package app

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type Server struct {
	rdb       *redis.Client
	tmpl      *template.Template
	adminUser string
	adminPass string
}

type PageData struct {
	OriginalURL string
	ShortURL    string
	ShortID     string
	CustomAlias string
	StatsCount  int64
	Error       string
	Mode        string
}

func NewServer(rdb *redis.Client, adminUserInfo [2]string) (*Server, error) {
	tmpl, err := template.ParseFiles("web/template/index.html")
	if err != nil {
		return nil, err
	}

	return &Server{
		rdb:       rdb,
		tmpl:      tmpl,
		adminUser: adminUserInfo[0],
		adminPass: adminUserInfo[1],
	}, nil
}

func (s *Server) Run(ctx context.Context, port string) error {
	mux := s.SetupRoutes("web/static")

	srv := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	go func() {
		PrintBanner(port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	slog.Info("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return srv.Shutdown(shutdownCtx)
}
