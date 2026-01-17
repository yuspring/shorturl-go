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

func (s *Server) SetupRoutes(staticFS fs.FS) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	mux.HandleFunc("GET /{$}", s.rootHandler)
	mux.HandleFunc("POST /shorten", s.shortenHandler)
	mux.HandleFunc("GET /stats", s.basicAuthMiddleware(s.listAllStatsHandler))
	mux.HandleFunc("GET /{id}", s.redirectHandler)

	return SecurityHeaders(mux)
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
		http.Error(w, "An internal error occurred", http.StatusInternalServerError)
		return
	}

	go func() {
		ctx := context.WithoutCancel(r.Context())

		ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		if err := s.rdb.Incr(ctx, "hits:"+shortID).Err(); err != nil {
			slog.ErrorContext(ctx, "failed to increment hits",
				"shortID", shortID,
				"error", err)
			return
		}

		if err := s.rdb.Expire(ctx, "hits:"+shortID, 3*24*time.Hour).Err(); err != nil {
			slog.WarnContext(ctx, "failed to set expiration for hits",
				"shortID", shortID,
				"error", err)
		}
	}()

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func (s *Server) shortenHandler(w http.ResponseWriter, r *http.Request) {

	originalURL := strings.TrimSpace(r.FormValue("url"))
	customAlias := strings.TrimSpace(r.FormValue("alias"))

	if strings.Contains(originalURL, r.Host) {
		s.renderTemplate(w, PageData{Mode: "home", Error: "Cannot shorten URLs belonging to this service", OriginalURL: originalURL, CustomAlias: customAlias})
		return
	}

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
			http.Error(w, "Failed to set custom alias", http.StatusInternalServerError)
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
			http.Error(w, "An internal error occurred", http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, "/?id="+shortID, http.StatusFound)
}

func (s *Server) listAllStatsHandler(w http.ResponseWriter, r *http.Request) {
	keys, err := s.rdb.Keys(r.Context(), "url:*").Result()
	if err != nil {
		slog.Error("failed to list keys", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rows := make([]StatsRow, 0, len(keys))
	for _, key := range keys {
		shortID := strings.TrimPrefix(key, "url:")
		original, err := s.rdb.Get(r.Context(), key).Result()
		if err != nil {
			continue
		}
		hits, _ := s.rdb.Get(r.Context(), "hits:"+shortID).Int64()

		ttl, _ := s.rdb.TTL(r.Context(), key).Result()
		expiresAt := "Never"
		if ttl > 0 {
			taipei := time.FixedZone("UTC+8", 8*60*60)
			expiresAt = time.Now().In(taipei).Add(ttl).Format("2006-01-02 15:04")
		}

		rows = append(rows, StatsRow{
			ShortID:     shortID,
			OriginalURL: original,
			Hits:        hits,
			ExpiresAt:   expiresAt,
		})
	}

	s.renderTemplate(w, PageData{Mode: "statsAll", StatsRows: rows})
}
