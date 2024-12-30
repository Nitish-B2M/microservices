package middlewares

import (
	"context"
	"e-commerce-backend/shared/utils"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.JsonError(w, utils.MissingAuthorizationHeader, http.StatusUnauthorized, nil)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			utils.JsonError(w, utils.InvalidAuthorizationHeader, http.StatusUnauthorized, nil)
			return
		}

		token, err := utils.ValidateJWT(tokenString)
		if err != nil || !token.Valid {
			utils.JsonError(w, utils.InvalidTokenError, http.StatusUnauthorized, err)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			utils.JsonError(w, utils.InvalidTokenClaims, http.StatusUnauthorized, nil)
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			utils.JsonError(w, utils.UserIdNotFoundInToken, http.StatusUnauthorized, nil)
			return
		}

		ctx := context.WithValue(r.Context(), "userID", int(userID))
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
