package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	claimsKey contextKey = "claims"
	subjectKey contextKey = "subject"
	roleKey    contextKey = "role"
)

func Middleware(j *JWT) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				http.Error(w, `{"error":"missing authorization"}`, http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(header, "Bearer ")
			if tokenStr == header {
				http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
				return
			}

			claims, err := j.Validate(tokenStr)
			if err != nil {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			ctx = context.WithValue(ctx, subjectKey, claims.Subject)
			ctx = context.WithValue(ctx, roleKey, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetSubject(ctx context.Context) string {
	v, _ := ctx.Value(subjectKey).(string)
	return v
}

func GetRole(ctx context.Context) string {
	v, _ := ctx.Value(roleKey).(string)
	return v
}

func GetClaims(ctx context.Context) *Claims {
	v, _ := ctx.Value(claimsKey).(*Claims)
	return v
}
