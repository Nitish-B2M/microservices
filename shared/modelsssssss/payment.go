package models

import (
	"log"
	"microservices/pkg/shared/dbs"
	"microservices/pkg/shared/utils"
)

type Payment struct {
	ID                int    `json:"id"`
	PaymentMethod     int8   `json:"payment_method"`      // online or cod
	PaymentMethodName string `json:"payment_method_name"` // card, cash, upi, net-banking
}

func InitPaymentSchema() {
	db := dbs.DB
	if err := db.AutoMigrate(&Payment{}); err != nil {
		log.Fatalf(utils.DatabaseMigrationError, "Payment", err)
	} else {
		log.Printf(utils.SchemaMigrationSuccess, "Payment")
	}
}
