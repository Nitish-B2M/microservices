// Package middlewares provides middleware functions for role-based access control.
package middlewares

import (
	"context"
	"e-commerce-backend/shared/utils"
	"e-commerce-backend/users/pkg/payloads"
	"errors"
	"net/http"

	"gorm.io/gorm"
)

// RoleMiddleware checks if the user's role is in the allowed roles list.
func RoleMiddleware(db *gorm.DB, allowedRoles ...string) func(http.Handler) http.Handler {
	if len(allowedRoles) == 0 {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := utils.GetUserIDFromContext(r)
			if userID == 0 {
				utils.JsonError(w, "User not authenticated", http.StatusUnauthorized, errors.New(utils.UnauthorizedError))
				return
			}

			var result []payloads.UserRoleInfoResp
			err := db.
				Table("user_roles").
				Select("user_roles.id as user_role_id, user_roles.user_id, user_roles.is_active, roles.name as role_name").
				Joins("JOIN roles ON roles.id = user_roles.role_id").
				Where("user_roles.user_id = ?", userID).
				Find(&result).Error

			if err != nil {
				utils.JsonError(w, "Error fetching user and role", http.StatusInternalServerError, err)
				return
			}

			roleAllowed := false
			for _, allowedRole := range allowedRoles {
				for _, role := range result {
					if role.RoleName == allowedRole && role.IsActive {
						roleAllowed = true
						break
					}
				}
			}

			if !roleAllowed {
				utils.JsonError(w, "Unauthorized role", http.StatusForbidden, errors.New(utils.UnauthorizedError))
				return
			}

			ctx := context.WithValue(r.Context(), utils.ActiveRoleNameKey, result)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
