package models

import (
	"e-commerce-backend/order/dbs"
	"e-commerce-backend/shared/utils"
	"gorm.io/gorm"
	"log"
	"time"
)

type OldOrder struct {
	Id                 int             `json:"id"`
	UserId             int             `json:"user_id"`
	Carts              json.RawMessage `json:"carts"`
	TotalPrice         float64         `json:"total_price"`
	IsPay              bool            `json:"is_pay"`
	ShippingAddress    string          `json:"shipping_address"`
	ShippingMethod     string          `json:"shipping_method"`
	BillingAddress     string          `json:"billing_address"`
	PaymentMethod      string          `json:"payment_method"`
	PaymentTransaction string          `json:"payment_transaction"`
	OrderDate          string          `json:"order_date"`
	EstimatedDelivery  string          `json:"estimated_delivery"`
	OrderStatus        string          `json:"order_status"`
	PromoCode          string          `json:"promo_code,omitempty"`
	Discount           float64         `json:"discount,omitempty"`
	Tax                float64         `json:"tax"`
}

type Order struct {
	OrderID         int        `gorm:"primaryKey;autoIncrement" json:"order_id"`  // Unique identifier for the order
	CustomerID      int        `gorm:"not null" json:"customer_id"`               // Foreign key linking to the customer
	OrderDate       time.Time  `gorm:"type:timestamp;not null" json:"order_date"` // Date and time when the order was placed
	ShippingAddress string     `gorm:"not null" json:"shipping_address"`          // Address where the order will be shipped
	BillingAddress  string     `gorm:"not null" json:"billing_address"`           // Address used for billing
	TotalAmount     float64    `gorm:"not null" json:"total_amount"`              // Total cost of the order
	PaymentMethod   string     `gorm:"not null" json:"payment_method"`            // Payment method used (e.g., credit card)
	PaymentStatus   string     `gorm:"not null" json:"payment_status"`            // Payment status (e.g., Pending, Completed)
	OrderStatus     string     `gorm:"not null" json:"order_status"`              // Order status (e.g., Pending, Shipped)
	ShippingMethod  string     `gorm:"not null" json:"shipping_method"`           // Shipping method (e.g., Standard, Express)
	TrackingNumber  string     `gorm:"default:null" json:"tracking_number"`       // Tracking number for the shipment
	DiscountCode    string     `gorm:"default:null" json:"discount_code"`         // Discount code applied (if any)
	TaxAmount       float64    `gorm:"not null" json:"tax_amount"`                // Tax amount applied to the order
	CreatedAt       time.Time  `gorm:"autoCreateTime;not null" json:"created_at"` // Timestamp when the order was created
	UpdatedAt       time.Time  `gorm:"autoUpdateTime;not null" json:"updated_at"` // Timestamp when the order was last updated
	CancelledAt     *time.Time `gorm:"default:null" json:"cancelled_at"`          // Timestamp when the order was canceled (nullable)
}

func InitOrderSchemas() {
	if dbs.DB == nil {
		log.Fatalf("Database connection is nil")
		return
	}

	if err := dbs.DB.AutoMigrate(&Order{}); err != nil {
		log.Fatalf(utils.DatabaseMigrationError, "Order", err)
	} else {
		log.Printf(utils.SchemaMigrationSuccess, "Order")
	}
}

type OrderInterface interface {
	GetOrders(db *gorm.DB) ([]Order, error)
	GetOrderById(db *gorm.DB, id int) error
	GetOrdersByUserId(db *gorm.DB, userId int) ([]Order, error)
	CreateOrder(db *gorm.DB, order *Order) error
}

func (o *Order) GetOrders(db *gorm.DB) ([]Order, error) {
	var orders []Order
	if err := db.Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (o *Order) GetOrderById(db *gorm.DB, id int) error {
	if err := db.First(&o, id).Error; err != nil {
		return err
	}
	return nil
}

func (o *Order) GetOrdersByUserId(db *gorm.DB, userId int) ([]Order, error) {
	var orders []Order
	if err := db.Where("user_id = ?", userId).Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (o *Order) CreateOrder(db *gorm.DB) error {
	if err := db.Create(&o).Error; err != nil {
		return err
	}
	return nil
}
