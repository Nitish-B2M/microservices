// Package repository provides database operations for product-related operations.
package repository

import (
	"e-commerce-backend/users/internal/models"
	pkg_utils "e-commerce-backend/users/utils"
	"errors"
	"time"

	"gorm.io/gorm"
)

type UserRepo struct {
	DB *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{DB: db}
}

func (repo *UserRepo) CreateUser(user *models.User) (*models.User, error) {
	if err := repo.DB.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (repo *UserRepo) IsEmailExists(email string) (bool, error) {
	var count int64
	if err := repo.DB.Model(&models.User{}).Where("email = ? and deleted_at is null", email).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repo *UserRepo) IsUserActive(email string) (bool, error) {
	var count int64
	if err := repo.DB.Model(&models.User{}).Where("email = ? and is_active = ? and deleted_at is null", email, true).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repo *UserRepo) IsUserExists(id int) (bool, error) {
	var count int64
	if err := repo.DB.Model(&models.User{}).Where("id = ? and is_active = ? and deleted_at is null", id, true).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repo *UserRepo) checkUserByEmail(email string) error {
	if email == "" {
		return errors.New(pkg_utils.UserNotFoundError)
	}

	val, err := repo.IsEmailExists(email)
	if err != nil {
		return err
	} else if !val {
		return errors.New(pkg_utils.UserNotFoundError)
	}

	active, err := repo.IsUserActive(email)
	if err != nil {
		return err
	} else if !active {
		return errors.New(pkg_utils.UserNotActiveError)
	}

	return nil
}

func (repo *UserRepo) GetUserByEmail(email string) (*models.User, error) {
	if err := repo.checkUserByEmail(email); err != nil {
		return nil, err
	}

	var user models.User
	if err := repo.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (repo *UserRepo) CheckRoleExists(roleName string) (models.Role, error) {
	var role models.Role
	err := repo.DB.Where("name = ?", roleName).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return role, errors.New(pkg_utils.RoleNotExistsError)
		}
		return role, err
	}
	return role, nil
}

func (repo *UserRepo) AttachDefaultRole(user *models.User) (*models.UserRoleInfo, error) {
	role, err := repo.CheckRoleExists("customer")
	if err != nil {
		return nil, err
	}

	err = repo.DB.Model(user).Association("Roles").Append(&role)
	if err != nil {
		return nil, err
	}

	roleInfo := models.UserRoleInfo{}
	err = repo.DB.
		Table("user_roles").
		Select("user_roles.id").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND roles.name = ?", user.ID, role.Name).
		First(&roleInfo).Error

	if err != nil {
		return nil, err
	}

	return &roleInfo, nil
}

func (repo *UserRepo) UpdateUserActiveRole(user *models.User, role models.UserRoleInfo) error {
	updateRole := map[string]interface{}{"active_role_id": role.UserRoleID, "active_role": role.RoleName}
	if err := repo.DB.Model(user).Updates(updateRole).Error; err != nil {
		return err
	}

	user.ActiveRoleID = role.UserRoleID
	user.ActiveRole = role.RoleName

	return nil
}

func (repo *UserRepo) SendVerificationEmail(user *models.User) error {
	return nil
}

func (repo *UserRepo) FetchUserProfileByID(id int) (*models.User, error) {
	val, err := repo.IsUserExists(id)
	if err != nil {
		return nil, err
	} else if !val {
		return nil, errors.New(pkg_utils.UserNotFoundError)
	}

	var user models.User
	if err := repo.DB.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (repo *UserRepo) StoreToken(userID int, token string, tokenType int) error {
	expiresAt := time.Now().Add(time.Hour * 1)

	storeToken := models.UserToken{
		UserID:    userID,
		Token:     token,
		Type:      tokenType,
		ExpiresAt: expiresAt,
		BaseModel: models.BaseModel{
			CreatedAt: time.Now(),
		},
	}

	if err := repo.DB.Create(&storeToken).Error; err != nil {
		return err
	}

	return nil
}

func (repo *UserRepo) VerifyToken(token string, tokenType int) (*models.UserToken, error) {
	var userToken models.UserToken
	var user models.User

	if err := repo.DB.Where("token = ? AND type = ? AND expires_at > ? AND is_verified = ? AND deleted_at is not null",
		token, tokenType, time.Now(), true).First(&userToken).Error; err == nil {
		return nil, pkg_utils.ErrTokenAlreadyVerified
	}

	if err := repo.DB.Where("token = ? AND type = ? AND expires_at > ? AND is_verified = ? AND deleted_at is null",
		token, tokenType, time.Now(), false).First(&userToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkg_utils.ErrTokenNotFound
		}
		return nil, err
	}

	userToken.IsVerified = true
	if err := repo.DB.Save(&userToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkg_utils.ErrTokenNotFound
		}
		return nil, err
	}

	// Soft delete the token
	if err := repo.DB.Delete(&userToken).Error; err != nil {
		return nil, err
	}

	// Mark user as verified
	if err := repo.DB.Model(&user).Where("id = ?", userToken.UserID).Update("is_verified", true).Error; err != nil {
		return nil, err
	}

	return &userToken, nil
}

func (repo *UserRepo) LoginUser(data models.LoginUser) (*models.User, error) {
	if err := repo.checkUserByEmail(data.Email); err != nil {
		return nil, err
	}

	var user models.User
	if err := repo.DB.Where("email = ?", data.Email).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
