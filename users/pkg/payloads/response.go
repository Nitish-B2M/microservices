// Package payloads
package payloads

import "time"

type UserResponse struct {
	ID         int          `json:"id"`
	FirstName  string       `json:"first_name"`
	LastName   string       `json:"last_name"`
	Email      string       `json:"email"`
	Gender     string       `json:"gender,omitempty"`
	IsVerified bool         `json:"is_verified"`
	IsDeleted  bool         `json:"is_deleted"`
	IsActive   bool         `json:"is_active"`
	CreatedAt  time.Time    `json:"created_at" update_only:"-"`
	UpdatedAt  time.Time    `json:"updated_at" update_only:"-"`
	Role       ResponseRole `json:"roles"`
}

type ResponseRole struct {
	ActiveRole Role   `json:"active_role"`
	Roles      []Role `json:"all_roles"`
}
type Role struct {
	UserId   int    `json:"-"`
	Username string `json:"username"`
	Role     string `json:"role"`
	RoleID   int    `json:"role_id"`
}

type LoginResponse struct {
	ID         int       `json:"id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      string    `json:"email"`
	Gender     string    `json:"gender,omitempty"`
	IsVerified bool      `json:"is_verified"`
	IsDeleted  bool      `json:"is_deleted"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`
}

type BaseModel struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserCreatedResponse struct {
	UserID int `json:"user_id"`
}

type UserProfileResp struct {
	ID         int    `json:"id"`
	FirstName  string `json:"f_nm"`
	LastName   string `json:"l_nm"`
	FullName   string `json:"fl_nm"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Gender     string `json:"gender,omitempty"`
	IsVerified bool   `json:"is_verified"`
	IsActive   bool   `json:"is_active"`
	BaseModel
}

type UserAuthResponse struct {
	Token      string       `json:"token"`
	ExpireTime int64        `json:"expire_time"`
	User       UserResponse `json:"user"`
}

type UserLoginResponse struct {
	Token      string          `json:"token"`
	ExpireTime int64           `json:"expire_time"`
	User       UserProfileResp `json:"user"`
}
