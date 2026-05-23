package server

import (
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// AuthMiddleware checks X-API-Key header.
// If API_KEY env is not set, all requests are denied (fail closed).
func AuthMiddleware(next http.Handler) http.Handler {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Println("FATAL: API_KEY not set — refusing to serve unauthenticated. Set API_KEY env var.")
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, `{"error":"server misconfigured: API_KEY not set"}`, http.StatusInternalServerError)
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/health" {
			next.ServeHTTP(w, r)
			return
		}

		key := r.Header.Get("X-API-Key")

		if key != apiKey {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware validates Origin against an allowlist and sets CORS headers.
func CORSMiddleware(next http.Handler) http.Handler {
	allowedOrigins := os.Getenv("CORS_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "http://localhost:5173,http://localhost:3000"
	}
	origins := strings.Split(allowedOrigins, ",")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/events" {
			origin := r.Header.Get("Origin")
			if originAllowed(origin, origins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func originAllowed(origin string, allowed []string) bool {
	for _, a := range allowed {
		a = strings.TrimSpace(a)
		if a == "*" || a == origin {
			return true
		}
	}
	return false
}

// RateLimitMiddleware limits requests per IP.
// Uses RemoteAddr only — ignores X-Forwarded-For to prevent spoofing.
type RateLimitMiddleware struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
	stopOnce sync.Once
	stopCh   chan struct{}
}

func NewRateLimitMiddleware(requestsPerMinute int) *RateLimitMiddleware {
	rl := &RateLimitMiddleware{
		requests: make(map[string][]time.Time),
		limit:    requestsPerMinute,
		window:   time.Minute,
		stopCh:   make(chan struct{}),
	}
	// Background cleanup every 5 minutes to prevent memory leak
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-rl.stopCh:
				return
			case <-ticker.C:
				rl.cleanup()
			}
		}
	}()
	return rl
}

func (rl *RateLimitMiddleware) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	cutoff := time.Now().Add(-rl.window)
	for ip, times := range rl.requests {
		var recent []time.Time
		for _, t := range times {
			if t.After(cutoff) {
				recent = append(recent, t)
			}
		}
		if len(recent) == 0 {
			delete(rl.requests, ip)
		} else {
			rl.requests[ip] = recent
		}
	}
}

func (rl *RateLimitMiddleware) Stop() {
	rl.stopOnce.Do(func() { close(rl.stopCh) })
}

func (rl *RateLimitMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use RemoteAddr only — do not trust X-Forwarded-For
		ip := r.RemoteAddr

		rl.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-rl.window)
		var recent []time.Time
		for _, t := range rl.requests[ip] {
			if t.After(cutoff) {
				recent = append(recent, t)
			}
		}
		rl.requests[ip] = append(recent, now)
		count := len(rl.requests[ip])
		rl.mu.Unlock()

		if count > rl.limit {
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
