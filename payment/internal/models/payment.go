package models

import (
	"e-commerce-backend/payment/dbs"
	"e-commerce-backend/shared/utils"
	"gorm.io/gorm"
	"log"
	"time"
)

type Payment struct {
	PaymentID            int       `gorm:"primaryKey;autoIncrement" json:"payment_id"`
	OrderID              int       `gorm:"not null" json:"order_id"`
	PaymentMethod        string    `gorm:"not null" json:"payment_method"`
	PaymentStatus        string    `gorm:"not null" json:"payment_status"` // "Pending", "Completed", "Failed"
	PaymentFailureReason string    `gorm:"type:text;default:null" json:"payment_failure_reason"`
	PaymentRetryCount    int       `gorm:"default:0" json:"payment_retry_count"`
	Amount               float64   `gorm:"not null" json:"amount"`
	PaymentDate          time.Time `gorm:"autoCreateTime" json:"payment_date"`
}

type PaymentInterface interface {
	CreatePayment(db *gorm.DB) error
	RefundPayment(db *gorm.DB) error
}

func InitPaymentSchema() {
	db := dbs.DB
	if err := db.AutoMigrate(&Payment{}); err != nil {
		log.Fatalf(utils.DatabaseMigrationError, "Payment", err)
	} else {
		log.Printf(utils.SchemaMigrationSuccess, "Payment")
	}
}

func (pay *Payment) CreatePayment(db *gorm.DB) error {
	if err := db.Create(&pay).Error; err != nil {
		return err
	}
	return nil
}
