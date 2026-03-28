package httpserver

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/htahta103/taskmanagerv2/internal/auth"
)

func (s *server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if s.cfg.CORSOrigin != "" && origin == s.cfg.CORSOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		raw = strings.TrimSpace(raw)
		if raw == "" {
			writeError(w, http.StatusUnauthorized, "missing bearer token", "unauthorized")
			return
		}
		uid, err := auth.ParseAccessToken(raw, s.cfg.JWTSecret)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired token", "unauthorized")
			return
		}
		next(w, r.WithContext(withUserID(r.Context(), uid)))
	}
}

func randomRefreshToken() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
