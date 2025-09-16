package middlewares

import (
	"context"
	"e-commerce-backend/shared/utils"
	"e-commerce-backend/users/validator"
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.JsonError(w, utils.MissingAuthorizationHeader, http.StatusUnauthorized, errors.New(utils.MissingAuthorizationHeader))
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			utils.JsonError(w, utils.InvalidAuthorizationHeader, http.StatusUnauthorized, errors.New(utils.InvalidAuthorizationHeader))
			return
		}

		token, err := utils.ValidateJWT(tokenString)
		if err != nil || !token.Valid {
			utils.JsonError(w, utils.InvalidTokenError, http.StatusUnauthorized, err)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			utils.JsonError(w, utils.InvalidTokenClaims, http.StatusUnauthorized, errors.New(utils.InvalidTokenClaims))
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			utils.JsonError(w, utils.UserIdNotFoundInToken, http.StatusUnauthorized, errors.New(utils.UserIdNotFoundInToken))
			return
		}

		ctx := context.WithValue(r.Context(), utils.UserIDKey, int(userID))
		if claims["username"] != nil {
			ctx = context.WithValue(ctx, utils.UserNameKey, claims["username"].(string))
		}
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func HandleRoute(
	router *mux.Router,
	method string,
	path string,
	validatorName string,
	handler http.HandlerFunc,
	mws ...func(http.Handler) http.Handler,
) {
	var h http.Handler = handler

	validatorMiddleware := validator.Validator(validatorName)
	if validatorMiddleware != nil {
		h = validatorMiddleware(h)
	}

	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}

	router.Handle(path, h).Methods(method)
}

func HandleGet(r *mux.Router, path string, validatorName string, handler http.HandlerFunc, mws ...func(http.Handler) http.Handler) {
	HandleRoute(r, http.MethodGet, path, validatorName, handler, mws...)
}

func HandlePost(r *mux.Router, path string, validatorName string, handler http.HandlerFunc, mws ...func(http.Handler) http.Handler) {
	HandleRoute(r, http.MethodPost, path, validatorName, handler, mws...)
}

func HandlePut(r *mux.Router, path string, validatorName string, handler http.HandlerFunc, mws ...func(http.Handler) http.Handler) {
	HandleRoute(r, http.MethodPut, path, validatorName, handler, mws...)
}

func HandleDelete(r *mux.Router, path string, validatorName string, handler http.HandlerFunc, mws ...func(http.Handler) http.Handler) {
	HandleRoute(r, http.MethodDelete, path, validatorName, handler, mws...)
}

func HandlePatch(r *mux.Router, path string, validatorName string, handler http.HandlerFunc, mws ...func(http.Handler) http.Handler) {
	HandleRoute(r, http.MethodPatch, path, validatorName, handler, mws...)
}

func GinAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader { // If no "Bearer" prefix
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header"})
			c.Abort()
			return
		}

		token, err := utils.ValidateJWT(tokenString)
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
			c.Abort()
			return
		}

		c.Set(utils.UserIDKey, int(userID))
		c.Next()
	}
}
