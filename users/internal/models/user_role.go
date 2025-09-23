package models

import (
	"e-commerce-backend/shared/utils"
	"e-commerce-backend/users/dbs"
	"e-commerce-backend/users/pkg/constants"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// type Role struct {
// 	ID   uint   `json:"id" gorm:"primaryKey"`
// 	Role string `json:"role" gorm:"type:varchar(50);not null"` // e.g., "admin", "seller", "user"
// }
// type UserRole struct {
// 	ID       uint   `json:"id" gorm:"primaryKey"`
// 	UserID   int    `json:"user_id" gorm:"not null;index"`   // Foreign Key to User
// 	Username string `json:"username" gorm:"not null;unique"` // Username associated with the user
// 	//Password string `json:"password" gorm:"not null"`
// 	RoleID uint `json:"role_id" gorm:"not null;index"` // Foreign Key to Role
// }

type RoleChangeRequest struct {
	ID            int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID        int       `json:"user_id" gorm:"not null;index"`
	RequestedRole string    `json:"requested_role" gorm:"type:varchar(50);not null"`
	Status        string    `json:"status" gorm:"type:enum('pending', 'approved', 'rejected');default:'pending'"`
	UserComment   *string   `json:"user_comment,omitempty"`
	AdminID       *int      `json:"admin_id" gorm:"index"`
	AdminComment  *string   `json:"admin_comment,omitempty"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type RoleChangeCommand struct {
	UserID        int
	RequestedRole string
	AdminID       int
	Approve       bool
	Comment       string
	DB            *gorm.DB
}

func InitUserRoleSchema() {
	db := dbs.UserDB
	if err := db.AutoMigrate(&Role{}, &UserRole{}, &RoleChangeRequest{}); err != nil {
		log.Fatalf(utils.DatabaseMigrationError, "Role, UserRole, RoleChangeRequest", err)
	} else {
		log.Printf(utils.SchemaMigrationSuccess, "Role, UserRole, RoleChangeRequest")
	}
}

// func NewUserRoleService(userID, roleID int, username string) *UserRole {
// 	return &UserRole{
// 		UserID:   userID,
// 		RoleID:   uint(roleID),
// 		Username: username,
// 	}
// }

type RoleInterface interface {
	CheckRoleExists(db *gorm.DB, role string) (bool, error)
	CreateRole(db *gorm.DB) error
}

// ##### Role Logic #####

func (r *Role) CheckRoleExists(db *gorm.DB, role string) error {
	if err := db.Where("role = ?", role).First(&r).Error; err != nil {
		return err
	}
	return nil
}

func (r *Role) CreateRole(db *gorm.DB) error {
	if err := db.Create(&r).Error; err != nil {
		return err
	}
	return nil
}

func (r *Role) GetAllRoles(db *gorm.DB) ([]Role, error) {
	var roles []Role
	if err := db.Find(&roles).Error; err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		return nil, fmt.Errorf("no roles found")
	}
	return roles, nil
}

func GetUserRequestedRole(db *gorm.DB, userID int, requestedRole, status string) (RoleChangeRequest, error) {
	var request RoleChangeRequest
	if err := db.Where("user_id = ? AND requested_role = ? AND status = ?", userID, requestedRole, status).First(&request).Error; err != nil {
		return request, err
	}
	return request, nil
}

func RequestRoleChange(db *gorm.DB, userID int, requestedRole, comment string) error {
	// Check if the requested role exists
	var role Role
	if err := db.Where("role = ?", requestedRole).First(&role).Error; err != nil {
		return fmt.Errorf("role %s does not exist", requestedRole)
	}

	// Check if the user already has the requested role
	var userRole UserRole
	if err := db.Where("user_id = ? AND role_id = ?", userID, role.ID).First(&userRole).Error; err == nil {
		return fmt.Errorf("user already has the requested role")
	}

	request := RoleChangeRequest{
		UserID:        userID,
		RequestedRole: requestedRole,
		Status:        constants.RoleStatusPending,
		UserComment:   &comment,
	}
	if err := db.Create(&request).Error; err != nil {
		return fmt.Errorf("failed to create role change request: %w", err)
	}

	return nil
}

func (cmd *RoleChangeCommand) Execute() error {
	// Start a database transaction
	tx := cmd.DB.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Fetch the request
	var request RoleChangeRequest
	if err := tx.Where("user_id = ? AND requested_role = ? AND status = 'pending'", cmd.UserID, cmd.RequestedRole).First(&request).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("role change request not found: %w", err)
	}

	// Update the request status
	if cmd.Approve {
		request.Status = constants.RoleStatusApproved

		// Remove any existing roles
		if err := tx.Where("user_id = ?", cmd.UserID).Delete(&UserRole{}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to remove existing roles: %w", err)
		}

		// Assign the new role
		var role Role
		if err := tx.Where("role = ?", cmd.RequestedRole).First(&role).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("role not found: %w", err)
		}

		newUserRole := UserRole{
			// UserID: cmd.UserID,
			RoleID: role.ID,
		}
		if err := tx.Create(&newUserRole).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to assign new role: %w", err)
		}
	} else {
		request.Status = constants.RoleStatusRejected
	}

	// Update the request with admin details
	request.AdminID = &cmd.AdminID
	request.AdminComment = &cmd.Comment
	if err := tx.Save(&request).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update request status: %w", err)
	}

	// Commit the transaction
	tx.Commit()
	return nil
}
