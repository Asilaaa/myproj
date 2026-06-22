package httpx

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"myproj/internal/auth"
)

type RateLimiter struct {
	limit  int
	window time.Duration

	mu      sync.Mutex
	entries map[string]rateLimitEntry
}

type rateLimitEntry struct {
	count      int
	windowEnds time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:   limit,
		window:  window,
		entries: map[string]rateLimitEntry{},
	}
}

func (l *RateLimiter) Middleware(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if l == nil || l.limit <= 0 || l.window <= 0 {
				next.ServeHTTP(w, r)
				return
			}

			key := l.key(scope, r)
			if !l.allow(key) {
				WriteError(w, http.StatusTooManyRequests, "rate limit exceeded, please try again shortly")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (l *RateLimiter) key(scope string, r *http.Request) string {
	if userID, ok := auth.GetUserIDFromContext(r.Context()); ok && userID != "" {
		return fmt.Sprintf("%s:user:%s", scope, userID)
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	return fmt.Sprintf("%s:ip:%s", scope, host)
}

func (l *RateLimiter) allow(key string) bool {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	for existingKey, entry := range l.entries {
		if now.After(entry.windowEnds) {
			delete(l.entries, existingKey)
		}
	}

	entry, ok := l.entries[key]
	if !ok || now.After(entry.windowEnds) {
		l.entries[key] = rateLimitEntry{count: 1, windowEnds: now.Add(l.window)}
		return true
	}
	if entry.count >= l.limit {
		return false
	}

	entry.count++
	l.entries[key] = entry
	return true
}
