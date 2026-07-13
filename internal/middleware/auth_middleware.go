package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/samAlcazar/gama_api/internal/config"
	"github.com/samAlcazar/gama_api/internal/service"
)

type contextKey string

const (
	UserIDKey    contextKey = "userID"
	UserNickKey  contextKey = "userNick"
	UserRoleKey  contextKey = "userRole"
	UserPermsKey contextKey = "userPerms"
)

func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error": "cabecera Authorization faltante"}`))
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error": "formato de cabecera de autenticación inválido"}`))
				return
			}

			tokenString := parts[1]
			claims := &service.CustomClaims{}

			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, errors.New("método de firma inválido")
				}
				return cfg.PublicKey, nil
			})

			if err != nil || !token.Valid {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error": "token inválido o expirado"}`))
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.Subject)
			ctx = context.WithValue(ctx, UserNickKey, claims.Nick)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
			ctx = context.WithValue(ctx, UserPermsKey, claims.Permissions)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			perms, ok := r.Context().Value(UserPermsKey).([]string)
			if !ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error": "acceso prohibido: sin permisos asignados"}`))
				return
			}

			hasPerm := false
			for _, p := range perms {
				if p == permission {
					hasPerm = true
					break
				}
			}

			if !hasPerm {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error": "acceso prohibido: permiso insuficiente"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetUserID(ctx context.Context) string {
	if val, ok := ctx.Value(UserIDKey).(string); ok {
		return val
	}
	return ""
}

func GetUserNick(ctx context.Context) string {
	if val, ok := ctx.Value(UserNickKey).(string); ok {
		return val
	}
	return ""
}

func GetUserRole(ctx context.Context) string {
	if val, ok := ctx.Value(UserRoleKey).(string); ok {
		return val
	}
	return ""
}

func GetUserPermissions(ctx context.Context) []string {
	if val, ok := ctx.Value(UserPermsKey).([]string); ok {
		return val
	}
	return nil
}
