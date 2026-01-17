package app

import (
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"shorturl-go/web"

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

	tmpl, err := template.ParseFS(web.Assets, "template/index.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedded templates: %w", err)
	}

	return &Server{
		rdb:       rdb,
		tmpl:      tmpl,
		adminUser: adminUserInfo[0],
		adminPass: adminUserInfo[1],
	}, nil
}

func (s *Server) Run(ctx context.Context, port string) error {

	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	staticFS, err := fs.Sub(web.Assets, "static")
	if err != nil {
		return fmt.Errorf("failed to create static sub-fs: %w", err)
	}
	mux := s.SetupRoutes(staticFS)

	srv := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	go func() {
		PrintIPs(port)
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
