package ingestwebapp

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type adapterIDContextKey struct{}

func DeviceIngestJWTMiddleware(secret string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/device-ingest/") {
				next.ServeHTTP(w, r)
				return
			}
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}
			tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			if tokenString == "" {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if token.Method != jwt.SigningMethodHS256 {
					return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "invalid token claims", http.StatusUnauthorized)
				return
			}
			adapterIDValue, ok := claims["adapterId"]
			if !ok {
				http.Error(w, "missing adapterId claim", http.StatusForbidden)
				return
			}
			adapterID, err := parseAdapterID(adapterIDValue)
			if err != nil || adapterID <= 0 {
				http.Error(w, "invalid adapterId claim", http.StatusForbidden)
				return
			}
			ctx := context.WithValue(r.Context(), adapterIDContextKey{}, adapterID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func parseAdapterID(value interface{}) (int, error) {
	switch v := value.(type) {
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	case int:
		return v, nil
	case int64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("unsupported adapterId type")
	}
}
