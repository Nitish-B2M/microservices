package middlewares

import (
	"e-commerce-backend/shared/utils"
	"net/http"
	"strings"

	"gorm.io/gorm"
)

// RoleMiddleware checks if the user's role is in the allowed roles list.
func RoleMiddleware(db *gorm.DB, allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value("userID").(uint)
			if !ok {
				utils.JsonError(w, "User not authenticated", http.StatusUnauthorized, nil)
				return
			}

			var role string
			if err := db.Table("roles").Where("user_id = ?", userID).Pluck("role", &role).Error; err != nil {
				// If there's an error fetching the role (e.g., role not found)
				utils.JsonError(w, "Role not found", http.StatusUnauthorized, nil)
				return
			}

			roleAllowed := false
			for _, allowedRole := range allowedRoles {
				if strings.EqualFold(strings.ToLower(role), strings.ToLower(allowedRole)) {
					roleAllowed = true
					break
				}
			}

			if !roleAllowed {
				utils.JsonError(w, "Unauthorized role", http.StatusForbidden, nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
