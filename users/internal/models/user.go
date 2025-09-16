// Package models for user-related operations.
package models

import (
	"e-commerce-backend/shared/utils"
	"e-commerce-backend/users/dbs"
	"e-commerce-backend/users/pkg/constants"
	"e-commerce-backend/users/pkg/payloads"
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID         int            `json:"id" gorm:"primaryKey;autoIncrement"`
	FirstName  string         `json:"first_name" gorm:"type:varchar(100);not null"`
	LastName   string         `json:"last_name" gorm:"type:varchar(100);not null"`
	Username   string         `json:"username" gorm:"type:varchar(100);not null"`
	Email      string         `json:"email" gorm:"type:varchar(100);not null"`
	Password   string         `json:"password" gorm:"type:varchar(255);not null"`
	Gender     string         `json:"gender,omitempty" gorm:"type:varchar(20)"` // Consider a custom type for validation
	IsVerified bool           `json:"is_verified" gorm:"default:false"`
	IsDeleted  bool           `json:"is_deleted" gorm:"default:false"`
	IsActive   bool           `json:"is_active" gorm:"default:true"`
	CreatedAt  time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"` // Soft delete support
}

type UserToken struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    int       `json:"user_id" gorm:"not null"`
	Token     string    `json:"token" gorm:"unique;not null"`
	Type      int       `json:"type" gorm:"not null"` // 1 for password_reset, 2 for email_verification
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func InitUserSchema() {
	db := dbs.UserDB
	if err := db.AutoMigrate(&User{}, &UserToken{}); err != nil {
		log.Fatalf(utils.DatabaseMigrationError, "User, UserToken", err)
	} else {
		log.Printf(utils.SchemaMigrationSuccess, "User, UserToken")
	}
}

type UserService interface {
	CreateUser(db *gorm.DB) (int, error)
	GetUserByEmail(db *gorm.DB, email string) (*User, error)
	GetUserByID(db *gorm.DB, id int) (*payloads.UserResponse, error)
	GetAllUsers(db *gorm.DB) ([]payloads.UserResponse, error)
	UpdateUser(db *gorm.DB, id int, updatedFields map[string]interface{}) (int, error)
	DeleteUser(db *gorm.DB, id int) error
	DeActivateUser(db *gorm.DB, id int) error
	ActivateUser(db *gorm.DB, id int) error
	GenerateUserToken(db *gorm.DB, tokenType int) (string, error)
	ValidateAndUseToken(db *gorm.DB, token string, tokenType int) (UserToken, error)
	ResetPassword(db *gorm.DB, id int, password string) error
	CheckUserEmailAlreadyVerified(db *gorm.DB, email string) bool
	VerifyUserEmail(db *gorm.DB, id int) error
	IsEmailVerified(db *gorm.DB, email string) bool
}

func (user *User) CreateUser(db *gorm.DB) (int, error) {
	tx := db.Begin()
	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	roleName := constants.RoleUser
	var role Role
	if err := tx.Where("role = ?", roleName).First(&role).Error; err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("role %s does not exist: %w", roleName, err)
	}

	userRole := NewUserRoleService(user.ID, int(role.ID), user.Username)
	if err := tx.Create(userRole).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return 0, err
	}
	return user.ID, nil
}

func LoggedInUser(db *gorm.DB, username, password string) (*User, error) {

	fmt.Println(username, password)

	var user User
	err := db.Table("users").
		Select("users.*").
		Where("users.username = ?", username).
		First(&user).Error

	if err != nil {
		return nil, err
	}

	fmt.Println(user)

	if ok, _ := utils.CompareHashedPassword(user.Password, password); !ok {
		return nil, errors.New("invalid username or password")
	}

	return &user, nil
}

func (user *User) GetUserUsingUsername(db *gorm.DB, username string) (*payloads.UserResponse, error) {
	var result payloads.ResponseRole
	err := db.Table("user_roles").
		Select("user_roles.user_id, roles.role, user_roles.role_id, user_roles.username").
		Joins("LEFT JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.username = ?", username).
		First(&result.ActiveRole).Error
	if err != nil {
		return nil, fmt.Errorf("error fetching user and role: %w", err)
	}

	userRes, err := user.GetUserByID(db, result.ActiveRole.UserId)
	if err != nil {
		return nil, fmt.Errorf("error fetching user: %w", err)
	}
	userRes.Role.ActiveRole = result.ActiveRole
	return userRes, nil
}

func (user *User) GetUserByEmail(db *gorm.DB, email string) (*User, error) {
	if err := db.Where("email = ? and is_deleted =?", email, false).First(&user).Error; err != nil {
		return nil, err
	}
	if !user.IsActive {
		return user, errors.New(utils.RequestUserIsDeactivated)
	}
	return user, nil
}

func (user *User) GetUserByID(db *gorm.DB, id int) (*payloads.UserResponse, error) {
	if err := db.Where("id =? and is_deleted =?", id, false).First(&user).Error; err != nil {
		return nil, err
	}

	var result payloads.ResponseRole
	err := db.Table("user_roles").
		Select("user_roles.user_id, roles.role, user_roles.role_id, user_roles.username").
		Joins("LEFT JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ?", id).
		Find(&result.Roles).Error

	if err != nil {
		return nil, fmt.Errorf("error fetching user and role: %w", err)
	}

	userRes := CopyUserToUserResponse(user)
	userRes.Role = result
	return userRes, nil
}

func (user *User) GetAllUsers(db *gorm.DB) ([]payloads.UserResponse, error) {
	var user1 []User
	if err := db.Where("is_deleted =?", false).Find(&user1).Error; err != nil {
		return nil, err
	}

	var userResponse []payloads.UserResponse
	for _, u := range user1 {
		userRes := CopyUserToUserResponse(&u)
		userResponse = append(userResponse, *userRes)
	}

	return userResponse, nil
}

func (user *User) UpdateUser(db *gorm.DB, id int, updatedFields map[string]interface{}) (int, error) {
	if err := db.Model(&user).Where("id = ?", id).Updates(updatedFields).Error; err != nil {
		return 0, err
	}
	return user.ID, nil
}

func (user *User) DeleteUser(db *gorm.DB, id int) error {
	if err := db.Model(&user).Where("id = ? and is_deleted = ?", id, false).Update("is_deleted", true).Error; err != nil {
		return err
	}
	return nil
}

func (user *User) DeActivateUser(db *gorm.DB, id int) error {
	if err := db.Model(&user).Where("id = ?", id).Update("is_active", false).Error; err != nil {
		return err
	}
	return nil
}

func (user *User) ActivateUser(db *gorm.DB, id int) error {
	if err := db.Model(&user).Where("id = ?", id).Update("is_active", true).Error; err != nil {
		return err
	}
	return nil
}

func (user *User) GenerateUserToken(db *gorm.DB, tokenType int) (string, error) {
	token := utils.GenerateRandomToken()
	expiresAt := time.Now().Add(time.Hour * 1)

	resetToken := UserToken{
		UserID:    user.ID,
		Token:     token,
		Type:      tokenType,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	if err := db.Create(&resetToken).Error; err != nil {
		return "", err
	}

	return token, nil
}

func (user *User) ValidateAndUseToken(db *gorm.DB, token string, tokenType int) (UserToken, error) {
	var userToken UserToken

	if err := db.Where("token = ? AND type = ? AND expires_at > ?", token, tokenType, time.Now()).First(&userToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return UserToken{}, errors.New("invalid or expired token")
		}
		return UserToken{}, err
	}

	// Delete the token after use
	if err := db.Delete(&userToken).Error; err != nil {
		return UserToken{}, errors.New("deleting reset token failed")
	}

	return userToken, nil
}

func (user *User) ResetPassword(db *gorm.DB, id int, password string) error {
	hashedPassword, err := utils.HashedPassword(password)
	if err != nil {
		return err
	}

	if err := db.Model(&user).Where("id = ?", id).Update("password", hashedPassword).Error; err != nil {
		return err
	}

	return nil
}

func (user *User) CheckUserEmailAlreadyVerified(db *gorm.DB, email string) bool {
	if err := db.Model(&user).Where("email =? and is_verified =?", email, 1).First(&user).Error; err != nil {
		return false
	}
	return true
}

func (user *User) VerifyUserEmail(db *gorm.DB, id int) error {

	if _, err := user.GetUserByID(db, id); err != nil {
		return errors.New("user not found")
	}

	if err := db.Model(&user).Where("id = ?", id).Update("is_verified", 1).Error; err != nil {
		return err
	}
	return nil
}

func (user *User) IsEmailVerified(db *gorm.DB, email string) bool {
	if err := db.Model(&user).Where("email = ? and is_verified = ?", email, 1).First(&User{}).Error; err != nil {
		return false
	}
	return true
}

func CopyUserToUserResponse(user *User) *payloads.UserResponse {
	return &payloads.UserResponse{
		ID:         user.ID,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Email:      user.Email,
		Gender:     user.Gender,
		IsVerified: user.IsVerified,
		IsDeleted:  user.IsDeleted,
		IsActive:   user.IsActive,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}
}

func GetUserRoleFromDB(db *gorm.DB, userID uint) (string, error) {
	var role string
	err := db.Table("users").
		Select("roles.role").
		Joins("LEFT JOIN user_roles ON user_roles.user_id = users.id").
		Joins("LEFT JOIN roles ON roles.id = user_roles.role_id").
		Where("users.id = ?", userID).
		First(&role).Error
	if err != nil {
		return "", fmt.Errorf("error fetching user role: %w", err)
	}
	return role, nil
}
