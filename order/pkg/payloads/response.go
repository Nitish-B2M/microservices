package payloads

import (
	"e-commerce-backend/order/internal/models"
	"time"
)

type OrderResponse struct {
	OrderID        string               `json:"order_id"`
	Items          []ResponseOrderItems `json:"items"`
	IsPaid         bool                 `json:"is_paid"`
	TotalAmount    float64              `gorm:"not null" json:"total_amount"`
	OrderStatus    string               `json:"order_status" gorm:"default:pending"` // "pending", "delivered", "cancelled", "processing", "shipped"
	PromoCode      string               `gorm:"default:null" json:"promo_code"`
	DiscountAmount float64              `gorm:"default:null" json:"discount_amount"`
	TaxAmount      float64              `json:"tax_amount"`
	SubTotal       float64              `json:"sub_total"`
	Shipping       models.OrderShipping `json:"shipping"`
	Payment        PaymentResponse      `json:"payment"`
	CreatedAt      time.Time            `json:"created_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time            `json:"updated_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
}

type ResponseOrderItems struct {
	ProductID    int       `json:"product_id"`
	Quantity     int       `json:"quantity"`
	Price        float64   `json:"price"`
	TotalPrice   float64   `json:"total_price"`
	Discount     float64   `json:"discount"`
	Tax          float64   `json:"tax"`
	ProductName  string    `json:"product_name"`
	ProductDesc  string    `json:"product_desc"`
	ProductImage string    `json:"product_image"`
	CreatedAt    time.Time `json:"created_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
}

type PaymentResponse struct {
	PaymentID     string `json:"payment_id"`
	PaymentMethod string `json:"payment_method"`
}
