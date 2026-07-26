package middleware

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"golang.org/x/time/rate"
)

type PerProjectLimiter struct {
	mu      sync.Mutex
	limiters map[string]*rate.Limiter
}

func NewPerProjectLimiter() *PerProjectLimiter {
	return &PerProjectLimiter{
		limiters: make(map[string]*rate.Limiter),
	}
}

func (l *PerProjectLimiter) GetLimiter(slug string, rps, burst int) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	lim, exists := l.limiters[slug]
	if !exists {
		lim = rate.NewLimiter(rate.Limit(rps), burst)
		l.limiters[slug] = lim
		return lim
	}

	// Update if config changed
	if lim.Limit() != rate.Limit(rps) || lim.Burst() != burst {
		lim = rate.NewLimiter(rate.Limit(rps), burst)
		l.limiters[slug] = lim
	}

	return lim
}

func RateLimit(limiter *PerProjectLimiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slug := chi.URLParam(r, "slug")
			if slug == "" {
				slug = r.URL.Query().Get("slug")
			}
			if slug == "" {
				next.ServeHTTP(w, r)
				return
			}

			lim := limiter.GetLimiter(slug, 100, 200)
			if !lim.Allow() {
				slog.Warn("rate limit exceeded", "slug", slug)
				http.Error(w, `{"error":"too many requests"}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
