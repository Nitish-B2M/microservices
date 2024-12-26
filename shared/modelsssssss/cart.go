package models

import (
	"log"
	"microservices/pkg/shared/dbs"
	"microservices/pkg/shared/utils"

	"gorm.io/gorm"
)

type Cart struct {
	Id        string `json:"id"`
	UserId    string `json:"userId"`
	ProductId string `json:"productId"`
	Quantity  int    `json:"quantity"`
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
