package models

import (
	"log"
	"microservices/pkg/shared/dbs"
	"microservices/pkg/shared/utils"
	"time"
)

type Order struct {
	ID            int       `json:"id"`
	OrderId       int       `json:"order_id"`
	OrderType     int       `json:"order_type"` // order type is
	OrderCart     []Cart    `json:"order_cart"`
	OrderedAt     time.Time `json:"ordered_at"`
	Price         int       `json:"price"`
	Discount      int       `json:"discount"`
	PaymentMethod Payment   `json:"payment_method"`
}

func InitOrderSchema() {
	db := dbs.DB
	if err := db.AutoMigrate(&Order{}); err != nil {
		log.Fatalf(utils.DatabaseMigrationError, "Order", err)
	} else {
		log.Printf(utils.SchemaMigrationSuccess, "Order")
	}
}
