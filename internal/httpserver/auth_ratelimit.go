package httpserver

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// authMinuteLimiter enforces a simple fixed window per IP (register / login / refresh).
type authMinuteLimiter struct {
	max     int
	mu      sync.Mutex
	buckets map[string]*rlBucket
}

type rlBucket struct {
	reset time.Time
	n     int
}

func newAuthMinuteLimiter(maxPerMinute int) *authMinuteLimiter {
	if maxPerMinute <= 0 {
		return nil
	}
	return &authMinuteLimiter{
		max:     maxPerMinute,
		buckets: make(map[string]*rlBucket),
	}
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (l *authMinuteLimiter) allow(ip string, now time.Time) bool {
	if l == nil {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	b := l.buckets[ip]
	if b == nil || now.After(b.reset) {
		l.buckets[ip] = &rlBucket{reset: now.Add(time.Minute), n: 1}
		return true
	}
	if b.n >= l.max {
		return false
	}
	b.n++
	return true
}

func (s *server) withAuthRateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.authRL == nil {
			next(w, r)
			return
		}
		ip := clientIP(r)
		if !s.authRL.allow(ip, time.Now()) {
			writeError(w, http.StatusTooManyRequests, "too many authentication requests", "rate_limit")
			return
		}
		next(w, r)
	}
}
