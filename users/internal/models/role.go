package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role represents the roles table
type Role struct {
	ID          uuid.UUID `json:"id" gorm:"type:char(36);primaryKey"`
	Name        string    `json:"name" gorm:"type:varchar(50);unique;not null"`
	Description string    `json:"description" gorm:"type:text"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	BaseModelNoSoftDelete

	// Relationships
	UserRoles       []UserRole       `json:"-" gorm:"foreignKey:RoleID"`
	RolePermissions []RolePermission `json:"-" gorm:"foreignKey:RoleID"`
}

// UserRole represents the user_roles table (junction table)
type UserRole struct {
	ID       int        `json:"id" gorm:"autoIncrement"`
	UserID   int        `json:"user_id" gorm:"not null;primaryKey"`
	RoleID   uuid.UUID  `json:"role_id" gorm:"type:char(36);not null;primaryKey"`
	TenantID *uuid.UUID `json:"tenant_id,omitempty" gorm:"type:char(36)"`

	GrantedBy *uuid.UUID `json:"granted_by,omitempty" gorm:"type:char(36)"`
	GrantedAt time.Time  `json:"granted_at" gorm:"autoCreateTime"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	IsActive  bool       `json:"is_active" gorm:"default:true"`
	BaseModel

	User          User  `json:"-" gorm:"foreignKey:UserID;references:ID"`
	Role          Role  `json:"-" gorm:"foreignKey:RoleID;references:ID"`
	GrantedByUser *User `json:"-" gorm:"foreignKey:GrantedBy;references:ID"`
}

// Permission represents the permissions table
type Permission struct {
	ID          uuid.UUID `json:"id" gorm:"type:char(36);primaryKey"`
	Name        string    `json:"name" gorm:"type:varchar(100);unique;not null"`
	Resource    string    `json:"resource" gorm:"type:varchar(50);not null"`
	Action      string    `json:"action" gorm:"type:varchar(50);not null"`
	Description string    `json:"description" gorm:"type:text"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	BaseModel

	// Relationships
	RolePermissions []RolePermission `json:"-" gorm:"foreignKey:PermissionID"`
}

// RolePermission represents the role_permissions table (junction table)
type RolePermission struct {
	ID           int       `json:"role_permission_id"`
	RoleID       uuid.UUID `json:"role_id" gorm:"type:char(36);not null;primaryKey"`
	PermissionID uuid.UUID `json:"permission_id" gorm:"type:char(36);not null;primaryKey"`
	BaseModel

	Role       Role       `json:"-" gorm:"foreignKey:RoleID;references:ID"`
	Permission Permission `json:"-" gorm:"foreignKey:PermissionID;references:ID"`
}

// UserRoleInfo this can't be used for table creation
type UserRoleInfo struct {
	UserRoleID int    `json:"user_role_id"`
	UserID     int    `json:"user_id"`
	IsActive   bool   `json:"is_active"`
	RoleName   string `json:"role_name"`
}

// BeforeCreate hooks for UUID generation (if needed)
func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

func (ur *UserRole) BeforeCreate(tx *gorm.DB) error {
	// No need to set ur.ID manually; it will be auto-incremented by the database.
	return nil
}

func (p *Permission) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func (rp *RolePermission) BeforeCreate(tx *gorm.DB) error {
	if rp.RoleID == uuid.Nil {
		rp.RoleID = uuid.New()
	}
	return nil
}

func (Role) TableName() string {
	return "roles"
}

func (UserRole) TableName() string {
	return "user_roles"
}

func (Permission) TableName() string {
	return "permissions"
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

// func FetchAllUsers(db *gorm.DB) error {
// 	// var users []User
// 	// if err := db.Find(&users).Error; err != nil {
// 	// 	return err
// 	// }

// 	// for _, user := range users {
// 	// 	if err := AddDefaultCustomerRoleToUser(db, user.ID); err != nil {
// 	// 		return err
// 	// 	}
// 	// }

// 	// return nil

// 	// return AssignAllPermissionsToAdminRole(db)
// }

// func AddDefaultCustomerRoleToUser(db *gorm.DB, userID int) error {
// 	var role Role
// 	// Find the "customer" role
// 	if err := db.Where("name = ?", "customer").First(&role).Error; err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return errors.New("default customer role not found")
// 		}
// 		return err
// 	}

// 	// Check if user already has this role (optional)
// 	var count int64
// 	db.Model(&UserRole{}).
// 		Where("user_id = ? AND role_id = ? AND is_active = ?", userID, role.ID, true).
// 		Count(&count)
// 	if count > 0 {
// 		// Role already assigned, no need to add again
// 		return nil
// 	}

// 	// Create the user role assignment
// 	userRole := UserRole{
// 		UserID:    userID,
// 		RoleID:    role.ID,
// 		GrantedAt: time.Now(),
// 		IsActive:  true,
// 		BaseModel: BaseModel{
// 			CreatedAt: time.Now(),
// 			UpdatedAt: time.Now(),
// 		},
// 	}

// 	if err := db.Create(&userRole).Error; err != nil {
// 		return err
// 	}

// 	return nil
// }

// func AssignAllPermissionsToAdminRole(db *gorm.DB) error {
// 	var adminRole Role
// 	// Find the admin role
// 	if err := db.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return errors.New("admin role not found")
// 		}
// 		return err
// 	}

// 	var permissions []Permission
// 	// Fetch all active permissions
// 	if err := db.Where("is_active = ?", true).Find(&permissions).Error; err != nil {
// 		return err
// 	}

// 	// Prepare slice to hold RolePermission entries
// 	var rolePermissions []RolePermission

// 	var count int
// 	for _, perm := range permissions {
// 		count++
// 		rolePermissions = append(rolePermissions, RolePermission{
// 			ID:           count,
// 			RoleID:       adminRole.ID,
// 			PermissionID: perm.ID,
// 			BaseModel: BaseModel{
// 				CreatedAt: time.Now(),
// 				UpdatedAt: time.Now(),
// 			},
// 		})
// 	}

// 	// Use transaction for safety
// 	return db.Transaction(func(tx *gorm.DB) error {
// 		for _, rp := range rolePermissions {
// 			// Avoid duplicate entries if they exist
// 			var existing RolePermission
// 			err := tx.Where("role_id = ? AND permission_id = ?", rp.RoleID, rp.PermissionID).First(&existing).Error
// 			if err == nil {
// 				// Already exists, skip
// 				continue
// 			}
// 			if !errors.Is(err, gorm.ErrRecordNotFound) {
// 				return err
// 			}
// 			if err := tx.Create(&rp).Error; err != nil {
// 				return err
// 			}
// 		}
// 		return nil
// 	})
// }
