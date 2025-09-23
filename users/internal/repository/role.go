package repository

import (
	"e-commerce-backend/users/internal/models"
	pkg_utils "e-commerce-backend/users/utils"
	"errors"
	"time"

	"gorm.io/gorm"
)

// AutoMigrateRoleTables automigrates the role tables
func AutoMigrateRoleTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Role{},
		&models.Permission{},
		&models.UserRole{},
		&models.RolePermission{},
	)
}

type UserRole struct {
	DB *gorm.DB
}

func NewUserRole(db *gorm.DB) *UserRole {
	return &UserRole{DB: db}
}

func (ur *UserRole) AddAdminRole(userID int) error {
	// check user exists
	var user models.User
	if err := ur.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	var role models.Role
	// Find the admin role
	if err := ur.DB.Where("name = ?", "admin").First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("admin role not found")
		}
		return err
	}

	// Check if user already has this role (optional)
	var userRoleDB models.UserRole
	var count int64
	ur.DB.Model(&userRoleDB).
		Where("user_id = ? AND role_id = ? AND is_active = ?", userID, role.ID, true).
		Count(&count)
	if count > 0 {
		// Role already assigned, no need to add again
		return nil
	}

	// Create the user role assignment
	userRole := models.UserRole{
		UserID:    userID,
		RoleID:    role.ID,
		GrantedAt: time.Now(),
		IsActive:  true,
		BaseModel: models.BaseModel{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	if err := ur.DB.Create(&userRole).Error; err != nil {
		return err
	}

	return nil
}

func (ur *UserRole) AddSellerRole(userID int) error {
	// check user exists
	var user models.User
	if err := ur.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	var role models.Role
	// Find the seller role
	if err := ur.DB.Where("name = ?", "seller").First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("seller role not found")
		}
		return err
	}

	// Check if user already has this role (optional)
	var userRoleDB models.UserRole
	var count int64
	ur.DB.Model(&userRoleDB).
		Where("user_id = ? AND role_id = ? AND is_active = ?", userID, role.ID, true).
		Count(&count)
	if count > 0 {
		return nil
	}

	// Create the user role assignment
	userRole := models.UserRole{
		UserID:    userID,
		RoleID:    role.ID,
		GrantedAt: time.Now(),
		IsActive:  true,
		BaseModel: models.BaseModel{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	if err := ur.DB.Create(&userRole).Error; err != nil {
		return err
	}

	return nil
}

// FetchAllRoles fetches all roles for a user
func (ur *UserRole) FetchAllRoles(userID int) ([]models.UserRoleInfo, error) {
	var roles []models.UserRoleInfo
	err := ur.DB.
		Table("user_roles").
		Select("user_roles.id as user_role_id, user_roles.user_id, user_roles.is_active, roles.name as role_name").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Scan(&roles).Error
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (ur *UserRole) IsUserWithRoleExists(userID int, roleID int) (bool, error) {
	var userRole models.UserRole
	if err := ur.DB.Where("id = ? AND user_id = ? AND is_active = ? AND deleted_at is null", roleID, userID, true).First(&userRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (ur *UserRole) SwitchRole(userID int, roleID int) (*models.User, []models.UserRoleInfo, error) {
	// check user exists
	repo := NewUserRepo(ur.DB)
	val, err := repo.IsUserExists(userID)
	if err != nil {
		return nil, nil, err
	} else if !val {
		return nil, nil, errors.New(pkg_utils.UserNotFoundError)
	}

	// check role exists
	roleExists, err := ur.IsUserWithRoleExists(userID, roleID)
	if err != nil {
		return nil, nil, err
	} else if !roleExists {
		return nil, nil, errors.New(pkg_utils.RoleNotExistsError)
	}

	// get user
	user, err := repo.FetchUserProfileByID(userID)
	if err != nil {
		return nil, nil, err
	}

	// switch role
	roles, err := ur.FetchAllRoles(userID)
	if err != nil {
		return nil, nil, err
	}

	return user, roles, nil
}
