package app

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func (s *Server) SetupRoutes(staticFS fs.FS) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	mux.HandleFunc("GET /{$}", s.rootHandler)
	mux.HandleFunc("POST /shorten", s.shortenHandler)
	mux.HandleFunc("GET /stats/{id}", s.statsHandler)
	mux.HandleFunc("GET /{id}", s.redirectHandler)

	return mux
}

func (s *Server) rootHandler(w http.ResponseWriter, r *http.Request) {
	resultID := r.URL.Query().Get("id")
	if resultID != "" {
		originalURL, err := s.rdb.Get(r.Context(), "url:"+resultID).Result()
		if err == nil {
			scheme := "http"
			if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
				scheme = "https"
			}
			fullShortURL := fmt.Sprintf("%s://%s/%s", scheme, r.Host, resultID)
			s.renderTemplate(w, PageData{
				Mode:        "result",
				OriginalURL: originalURL,
				ShortURL:    fullShortURL,
				ShortID:     resultID,
			})
			return
		}
	}

	s.renderTemplate(w, PageData{Mode: "home"})
}

func (s *Server) redirectHandler(w http.ResponseWriter, r *http.Request) {
	shortID := r.PathValue("id")

	if len(shortID) > 20 || !isAlphaNumeric(shortID) {
		slog.Warn("invalid short ID format", "id", shortID)
		http.NotFound(w, r)
		return
	}

	originalURL, err := s.rdb.Get(r.Context(), "url:"+shortID).Result()
	if err == redis.Nil {
		http.NotFound(w, r)
		return
	} else if err != nil {
		slog.Error("database error on redirect", "id", shortID, "error", fmt.Errorf("failed to get url from redis: %w", err))
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	go func() {
		ctx := context.Background()
		s.rdb.Incr(ctx, "hits:"+shortID)
		s.rdb.Expire(ctx, "hits:"+shortID, 3*24*time.Hour)
	}()

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func (s *Server) shortenHandler(w http.ResponseWriter, r *http.Request) {

	originalURL := strings.TrimSpace(r.FormValue("url"))
	customAlias := strings.TrimSpace(r.FormValue("alias"))

	normalizedURL, err := validateURL(originalURL)
	if err != nil {
		s.renderTemplate(w, PageData{Mode: "home", Error: err.Error(), OriginalURL: originalURL, CustomAlias: customAlias})
		return
	}
	originalURL = normalizedURL

	var shortID string

	if customAlias != "" {
		if len(customAlias) < 3 || len(customAlias) > 20 || !isAlphaNumeric(customAlias) {
			s.renderTemplate(w, PageData{Mode: "home", Error: "Custom alias must be 3-20 alphanumeric characters", OriginalURL: originalURL, CustomAlias: customAlias})
			return
		}
		success, err := s.rdb.SetNX(r.Context(), "url:"+customAlias, originalURL, 3*24*time.Hour).Result()
		if err != nil {
			slog.Error("database error on custom alias set", "alias", customAlias, "error", fmt.Errorf("failed to set alias in redis: %w", err))
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		if !success {
			s.renderTemplate(w, PageData{Mode: "home", Error: "Custom alias already in use", OriginalURL: originalURL, CustomAlias: customAlias})
			return
		}
		shortID = customAlias
	} else {
		shortID, err = s.generateUniqueKey(r.Context(), originalURL)
		if err != nil {
			slog.Error("id generation failed", "url", originalURL, "error", fmt.Errorf("failed to generate unique key: %w", err))
			http.Error(w, "Failed to generate short ID", http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, "/?id="+shortID, http.StatusFound)
}

func (s *Server) statsHandler(w http.ResponseWriter, r *http.Request) {
	shortID := r.PathValue("id")

	if shortID == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	exists, err := s.rdb.Exists(r.Context(), "url:"+shortID).Result()
	if err != nil || exists == 0 {
		s.renderTemplate(w, PageData{Mode: "home", Error: "Short URL not found or expired"})
		return
	}

	count, err := s.rdb.Get(r.Context(), "hits:"+shortID).Int64()
	if err != nil && err != redis.Nil {
		count = 0
	}

	s.renderTemplate(w, PageData{Mode: "stats", ShortID: shortID, StatsCount: count})
}
