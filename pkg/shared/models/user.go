package models

import (
	"log"
	"microservices/pkg/shared/dbs"
	"microservices/pkg/shared/payloads"
	"microservices/pkg/shared/utils"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	FirstName string    `json:"first_name" gorm:"type:varchar(100);not null"`
	LastName  string    `json:"last_name" gorm:"type:varchar(100);not null"`
	Email     string    `json:"email" gorm:"type:varchar(100);unique;not null"`
	Password  string    `json:"password" gorm:"type:varchar(255);not null"`
	Gender    string    `json:"gender,omitempty"`
	IsDeleted bool      `json:"is_deleted" gorm:"default:false"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type LoginUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserAuthResponse struct {
	Token string                `json:"token"`
	User  payloads.UserResponse `json:"user"`
}

type PasswordResetToken struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    int       `json:"user_id" gorm:"not null"`
	Token     string    `json:"token" gorm:"unique;not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func NewUser() *User {
	return &User{}
}

func InitUserSchema() {
	db := dbs.DB
	if err := db.AutoMigrate(&User{}, &PasswordResetToken{}); err != nil {
		log.Fatalf(utils.DatabaseMigrationError, "User and PasswordResetToken", err)
	} else {
		log.Printf(utils.SchemaMigrationSuccess, "User and PasswordResetToken")
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

func CopyUserToUserResponse(user *User) *payloads.UserResponse {
	return &payloads.UserResponse{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Gender:    user.Gender,
		IsDeleted: user.IsDeleted,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
