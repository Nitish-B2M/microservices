package models

import (
	"e-commerce-backend/order/dbs"
	"e-commerce-backend/shared/utils"
	"gorm.io/gorm"
	"log"
	"time"
)

type Order struct {
	OrderID        int       `gorm:"primaryKey;autoIncrement" json:"order_id"`
	CustomerID     int       `gorm:"not null" json:"customer_id"`
	IsPaid         bool      `json:"is_paid"`
	TotalAmount    float64   `gorm:"not null" json:"total_amount"`
	Carts          string    `gorm:"type:json" json:"carts"`
	OrderStatus    int       `json:"order_status" gorm:"default:0"`
	DiscountCode   string    `gorm:"default:null" json:"discount_code"`
	TaxAmount      float64   `json:"tax_amount"`
	ShippingMethod string    `json:"shipping_method"`
	CreatedAt      time.Time `json:"-" gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
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

func (o *Order) UpdateOrder(db *gorm.DB) error {
	if err := db.Save(&o).Error; err != nil {
		return err
	}
	return nil
}
