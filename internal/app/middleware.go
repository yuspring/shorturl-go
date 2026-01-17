package app

import (
	"crypto/subtle"
	"net/http"
)

func (s *Server) basicAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()

		if !ok || !secureCompare(user, s.adminUser) || !secureCompare(pass, s.adminPass) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted Statistics"`)
			http.Error(w, "401 Unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

func secureCompare(given, expected string) bool {
	return subtle.ConstantTimeCompare([]byte(given), []byte(expected)) == 1
}
