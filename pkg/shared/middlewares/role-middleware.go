package middlewares

import (
	"microservices/pkg/shared/utils"
	"net/http"
)

func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := r.Context().Value("role")
			for _, role := range allowedRoles {
				if role == userRole {
					next.ServeHTTP(w, r)
					return
				}
			}
			utils.JsonError(w, "Unauthorized role", http.StatusForbidden, nil)
		})
	}
}
