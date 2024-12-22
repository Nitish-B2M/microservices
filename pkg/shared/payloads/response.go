package payloads

import (
	"time"
)

type ProductResponse struct {
	ID        int         `json:"id"`
	PName     string      `json:"product_name"`
	PDesc     string      `json:"product_desc"`
	Price     float64     `json:"price"`
	Quantity  int         `json:"quantity"`
	IsDeleted bool        `json:"is_deleted"`
	Discount  float64     `json:"discount"`
	Rating    float64     `json:"rating"`
	Category  string      `json:"category"`
	Tags      interface{} `json:"tags"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type UserResponse struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Gender    string    `json:"gender,omitempty"`
	IsDeleted bool      `json:"is_deleted"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LoginResponse struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Gender    string    `json:"gender,omitempty"`
	IsDeleted bool      `json:"-"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
