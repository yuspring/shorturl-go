package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

var alphaNumericRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func (s *Server) generateUniqueKey(ctx context.Context, originalURL string) (string, error) {
	for range 5 {
		id := randomBase62(5)
		success, err := s.rdb.SetNX(ctx, "url:"+id, originalURL, 3*24*time.Hour).Result()
		if err != nil {
			return "", err
		}
		if success {
			return id, nil
		}
	}
	return "", errors.New("failed to generate unique key")
}

func randomBase62(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = base62Chars[rand.N(len(base62Chars))]
	}
	return string(b)
}

func validateURL(rawURL string) (string, error) {
	if rawURL == "" {
		return "", errors.New("URL cannot be empty")
	}

	normalizedURL := rawURL
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		normalizedURL = "https://" + rawURL
	}

	u, err := url.Parse(normalizedURL)
	if err != nil || u.Host == "" || !strings.Contains(u.Host, ".") {
		return "", errors.New("invalid URL format (must be a valid domain or URL)")
	}

	ips, err := net.LookupIP(u.Hostname())
	if err != nil {
		slog.Warn("failed to parse host", "host", u.Hostname(), "error", err)
		return normalizedURL, nil
	}
	for _, ip := range ips {
		if ip.IsPrivate() || ip.IsLoopback() {
			return "", errors.New("shortening internal network URLs is prohibited")
		}
	}
	return normalizedURL, nil
}

func PrintIPs(port string) {
	fmt.Printf("➜  Local:   http://localhost%s\n", port)
	for _, ip := range GetLocalIPs() {
		fmt.Printf("➜  Network: http://%s%s\n", ip, port)
	}
	fmt.Println()
}

func GetLocalIPs() []string {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && !ipnet.IP.IsLinkLocalUnicast() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}
	return ips
}

func isAlphaNumeric(s string) bool {
	return alphaNumericRegex.MatchString(s)
}

func (s *Server) renderTemplate(w http.ResponseWriter, data PageData) {
	if err := s.tmpl.Execute(w, data); err != nil {
		http.Error(w, "Template rendering error", http.StatusInternalServerError)
		slog.Error("Template execution error", "error", err)
	}
}
