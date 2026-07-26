package middleware

import (
	"net/http"
	"strings"
)

type CORS struct {
	allowedOrigins []string
}

func NewCORS(origins string) *CORS {
	parts := strings.Split(origins, ",")
	trimmed := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			trimmed = append(trimmed, p)
		}
	}
	return &CORS{allowedOrigins: trimmed}
}

func (c *CORS) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if c.isAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Max-Age", "86400")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (c *CORS) isAllowed(origin string) bool {
	if origin == "" {
		return false
	}
	for _, o := range c.allowedOrigins {
		if o == "*" || o == origin {
			return true
		}
	}
	return false
}
