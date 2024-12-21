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
