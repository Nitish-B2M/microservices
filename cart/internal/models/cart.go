package models

import (
	"e-commerce-backend/cart/dbs"
	"e-commerce-backend/shared/utils"
	"log"

	"gorm.io/gorm"
)

type Cart struct {
	Id        string `json:"id"`
	UserId    string `json:"userId"`
	ProductId string `json:"productId"`
	Quantity  int    `json:"quantity" default:"1"`
	gorm.Model
}

func InitCartSchema() {
	db := dbs.DB
	if err := db.AutoMigrate(&Cart{}); err != nil {
		log.Fatalf(utils.DatabaseMigrationError, "Cart", err)
	} else {
		log.Printf(utils.SchemaMigrationSuccess, "Cart")
	}
}
