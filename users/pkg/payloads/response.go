package payloads

import "time"

type UserResponse struct {
	ID         int       `json:"id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      string    `json:"email"`
	Gender     string    `json:"gender,omitempty"`
	IsVerified bool      `json:"is_verified"`
	IsDeleted  bool      `json:"is_deleted"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at" update_only:"-"`
	UpdatedAt  time.Time `json:"updated_at" update_only:"-"`
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
