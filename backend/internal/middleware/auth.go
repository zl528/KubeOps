package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/kubeops/ops-kubernetes/internal/service"
)

type contextKey string

const (
	UserIDKey   contextKey = "userID"
	UsernameKey contextKey = "username"
	UserRoleKey contextKey = "userRole"
)

func AuthMiddleware(authSvc *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for login, health, and websocket endpoints
			if r.URL.Path == "/api/auth/login" ||
				r.URL.Path == "/api/health" ||
				r.URL.Path == "/api/auth/register" {
				next.ServeHTTP(w, r)
				return
			}

			// Get token from header or query parameter (for websocket)
			var tokenString string
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					tokenString = parts[1]
				}
			}
			if tokenString == "" {
				tokenString = r.URL.Query().Get("token")
			}
			if tokenString == "" {
				http.Error(w, `{"code":401,"message":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			claims, err := authSvc.ValidateToken(tokenString)
			if err != nil {
				http.Error(w, `{"code":401,"message":"invalid token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UsernameKey, claims.Username)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserFromContext(ctx context.Context) (int64, string, string) {
	userID, _ := ctx.Value(UserIDKey).(int64)
	username, _ := ctx.Value(UsernameKey).(string)
	role, _ := ctx.Value(UserRoleKey).(string)
	return userID, username, role
}

func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, _ := r.Context().Value(UserRoleKey).(string)
		if role != "admin" {
			http.Error(w, `{"code":403,"message":"forbidden: admin required"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
