package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/avito-internships/test-backend-1-untrik/internal/dto"
	"github.com/avito-internships/test-backend-1-untrik/internal/http/helper"
	auth "github.com/avito-internships/test-backend-1-untrik/internal/jwt"
)

type contextKey string

const ClaimsContextKey contextKey = "claims"

type jwtInterface interface {
	ParseToken(tokenString string) (*auth.Claims, error)
}

func AuthMiddleware(jwtService jwtInterface) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				slog.Info("invalid token")
				helper.WriteError(w, http.StatusUnauthorized, dto.ErrorCodeUnauthorized, "unauthorized")
				return
			}
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := jwtService.ParseToken(tokenString)
			if err != nil {
				slog.Info("invalid token", "error", err)
				helper.WriteError(w, http.StatusUnauthorized, dto.ErrorCodeUnauthorized, "unauthorized")
				return
			}
			ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
func IsAdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(ClaimsContextKey).(*auth.Claims)
		if !ok {
			helper.WriteError(w, http.StatusUnauthorized, dto.ErrorCodeUnauthorized, "unauthorized")
			return
		}

		if claims.Role != "admin" {
			helper.WriteError(w, http.StatusForbidden, dto.ErrorCodeForbidden, "forbidden")
			return
		}

		next.ServeHTTP(w, r)
	})
}
func IsUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(ClaimsContextKey).(*auth.Claims)
		if !ok {
			helper.WriteError(w, http.StatusUnauthorized, dto.ErrorCodeUnauthorized, "unauthorized")
			return
		}

		if claims.Role != "user" {
			helper.WriteError(w, http.StatusForbidden, dto.ErrorCodeForbidden, "forbidden")
			return
		}

		next.ServeHTTP(w, r)
	})
}
