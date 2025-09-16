package models

import (
	"e-commerce-backend/order/dbs"
	"e-commerce-backend/shared/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
	"time"
)

type Order struct {
	OrderID        uuid.UUID     `gorm:"primaryKey;type:char(36)" json:"order_id"`
	UserID         int           `gorm:"not null" json:"user_id"`
	IsPaid         bool          `json:"is_paid"`
	OrderStatus    string        `json:"order_status" gorm:"default:pending"` // "pending", "delivered", "cancelled", "processing", "shipped"
	PromoCode      string        `gorm:"default:null" json:"promo_code"`
	DiscountAmount float64       `gorm:"default:null" json:"discount_amount"`
	TaxAmount      float64       `json:"tax_amount"`
	SubTotal       float64       `json:"sub_total"`
	TotalAmount    float64       `gorm:"not null" json:"total_amount"`
	OrderItems     []OrderItems  `gorm:"foreignKey:OrderID" json:"items"` // One-to-many relationship
	OrderShipping  OrderShipping `gorm:"foreignKey:OrderID" json:"shipping"`
	CreatedAt      time.Time     `json:"created_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time     `json:"updated_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
}

type OrderItems struct {
	CartId         int       `json:"cart_id"`
	OrderID        uuid.UUID `json:"order_id"`
	ProductID      int       `json:"product_id"`
	Quantity       int       `json:"quantity"`
	Price          float64   `json:"price"`
	ItemTotalPrice float64   `json:"item_total_price"`
	Discount       float64   `json:"discount"`
	CreatedAt      time.Time `json:"created_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
}

type OrderShipping struct {
	OrderID           uuid.UUID `json:"order_id"`
	BillingAddressId  uuid.UUID `json:"billing_address_id"`
	BillingUserId     int       `json:"billing_user_id"`
	BillingContact    string    `json:"billing_contact"`
	ShippingCost      float64   `json:"shipping_cost"`
	ShippingAddressId uuid.UUID `json:"shipping_address_id"`
	ShippingContact   string    `json:"shipping_contact"`
	ShippingMethod    string    `json:"shipping_method"`
	ShippingStatus    string    `json:"shipping_status" gorm:"default:pending"` // "pending", "delivered", "packing", "shipped"
	CreatedAt         time.Time `json:"created_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
}

//need to implement shipping history

func InitOrderSchemas() {
	if dbs.DB == nil {
		log.Fatalf("Database connection is nil")
		return
	}

	if err := dbs.DB.AutoMigrate(&Order{}, &OrderShipping{}, &OrderItems{}); err != nil {
		log.Fatalf(utils.DatabaseMigrationError, "Order/OrderShipping/OrderItems", err)
	} else {
		log.Printf(utils.SchemaMigrationSuccess, "Order/OrderShipping/OrderItems")
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
	if err := db.Preload("OrderItems").Preload("OrderShipping").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (o *Order) GetOrderById(db *gorm.DB, id string) error {
	if err := db.Preload("OrderItems").Preload("OrderShipping").Where("order_id = ?", id).First(&o).Error; err != nil {
		return err
	}
	return nil
}

func (o *Order) GetOrdersByUserId(db *gorm.DB, userId int) ([]Order, error) {
	var orders []Order
	if err := db.Preload("OrderItems").Preload("OrderShipping").Where("user_id = ?", userId).Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (o *Order) CreateOrder(db *gorm.DB) error {
	if err := db.Create(&o).Error; err != nil {
		return err
	}
	log.Println(o)

	if len(o.OrderItems) > 0 {
		for _, item := range o.OrderItems {
			item.OrderID = o.OrderID
			if err := db.Create(&item).Error; err != nil {
				return err
			}
		}
	}

	if (OrderShipping{}) != o.OrderShipping {
		if err := db.Create(&o.OrderShipping).Error; err != nil {
			return err
		}
	}
	return nil
}

func (o *Order) UpdateOrder(db *gorm.DB) error {
	if err := db.Model(&Order{}).Where("order_id = ?", o.OrderID).Updates(o).Error; err != nil {
		return err
	}

	if err := db.Model(&OrderShipping{}).Where("order_id = ?", o.OrderID).Updates(o.OrderShipping).Error; err != nil {
		return err
	}

	for _, item := range o.OrderItems {
		item.OrderID = o.OrderID
		if err := db.Model(&OrderItems{}).Where("order_id = ? AND product_id = ?", o.OrderID, item.ProductID).Updates(item).Error; err != nil {
			return err
		}
	}
	return nil
}
