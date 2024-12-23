package payloads

import "time"

type ProductRequest struct {
	ID        int       `json:"id"`
	PName     string    `json:"product_name"`
	PDesc     string    `json:"product_desc"`
	Price     float64   `json:"price"`
	Quantity  int       `json:"quantity"`
	IsDeleted bool      `json:"is_deleted"`
	Discount  float64   `json:"discount"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `json:"updated_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	Tags      []string  `json:"tags" gorm:"-"`
	Rating    float64   `json:"rating"`
}

type UserUpdateRequest struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Gender    string `json:"gender"`
}
