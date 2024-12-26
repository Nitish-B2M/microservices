package middlewares

import (
	"context"
	"e-commerce-backend/shared/utils"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.JsonError(w, "Authorization header is missing", http.StatusUnauthorized, nil)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := utils.ValidateJWT(tokenString)
		if err != nil || !token.Valid {
			utils.JsonError(w, "Invalid token", http.StatusUnauthorized, err)
			return
		}

		// Add claims to request context for downstream handlers
		claims, _ := token.Claims.(jwt.MapClaims)
		ctx := context.WithValue(r.Context(), "user_id", claims["user_id"])
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
