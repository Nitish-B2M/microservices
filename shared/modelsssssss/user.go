package models

import (
	"errors"
	"log"
	"microservices/pkg/shared/dbs"
	"microservices/pkg/shared/payloads"
	"microservices/pkg/shared/utils"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID         int       `json:"id" gorm:"primaryKey;autoIncrement"`
	FirstName  string    `json:"first_name" gorm:"type:varchar(100);not null"`
	LastName   string    `json:"last_name" gorm:"type:varchar(100);not null"`
	Email      string    `json:"email" gorm:"type:varchar(100);not null"`
	Password   string    `json:"password" gorm:"type:varchar(255);not null"`
	Gender     string    `json:"gender,omitempty"`
	IsVerified bool      `json:"is_verified" gorm:"default:false"`
	IsDeleted  bool      `json:"is_deleted" gorm:"default:false"`
	IsActive   bool      `json:"is_active" gorm:"default:true"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type LoginUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserAuthResponse struct {
	Token string                `json:"token"`
	User  payloads.UserResponse `json:"user"`
}

type UserToken struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    int       `json:"user_id" gorm:"not null"`
	Token     string    `json:"token" gorm:"unique;not null"`
	Type      int       `json:"type" gorm:"not null"` // 1 for password_reset, 2 for email_verification
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type Role struct {
	ID     uint   `json:"id"gorm:"primaryKey"`
	Role   string `json:"role"gorm:"not null"` // Role name (admin, seller, user)
	UserID int    `json:"user_id"gorm:"not null;index"`
}

func InitUserSchema() {
	db := dbs.DB
	if err := db.AutoMigrate(&User{}, &UserToken{}, &Role{}); err != nil {
		log.Fatalf(utils.DatabaseMigrationError, "User, UserToken and UserRole", err)
	} else {
		log.Printf(utils.SchemaMigrationSuccess, "User, UserToken and UserRole")
	}
}

func (user *User) AddUser(db *gorm.DB) (int, error) {
	if err := db.Create(user).Error; err != nil {
		return 0, err
	}
	return user.ID, nil
}

func (user *User) GetUserByEmail(db *gorm.DB, email string) (*User, error) {
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (usr *User) FetchUserById(db *gorm.DB, id int) (*payloads.UserResponse, error) {
	if err := db.Where("id =? and is_deleted =?", id, false).First(&usr).Error; err != nil {
		return nil, err
	}

	userResponse := CopyUserToUserResponse(usr)
	return userResponse, nil
}

func (usr *User) GetAllUsers(db *gorm.DB) ([]payloads.UserResponse, error) {
	var user []User
	if err := db.Where("is_deleted =?", false).Find(&user).Error; err != nil {
		return nil, err
	}

	var userResponse []payloads.UserResponse
	for _, u := range user {
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

	if _, err := user.FetchUserById(db, id); err != nil {
		return errors.New("user not found")
	}

	if err := db.Model(&user).Where("id = ?", id).Update("is_verified", 1).Error; err != nil {
		return err
	}

	return nil
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

// ############# Role ############
func (role *Role) createDefaultUserRole(db *gorm.DB, userId int) (uint, error) {
	if err := db.Where("role = ?", "user").First(&role).Error; err != nil {
		role = &Role{
			Role:   "user",
			UserID: userId,
		}
		if err := db.Create(&role).Error; err != nil {
			return 0, err
		}
	}

	return role.ID, nil
}
