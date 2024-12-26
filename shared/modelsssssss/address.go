package models

import (
	"log"
	"microservices/pkg/shared/dbs"
	"microservices/pkg/shared/utils"
)

type Address struct {
	ID            int    `json:"id"`
	House         string `json:"house"`
	Street        string `json:"street"`
	City          string `json:"city"`
	State         string `json:"state"`
	Zip           string `json:"zip"`
	Type          int8   `json:"type"` // home/work/other
	ReceiverId    int    `json:"receiver_id" default:"null"`
	ReceiverPhone string `json:"receiver_phone" default:"null"`
}

func InitAddressSchema() {
	db := dbs.DB
	if err := db.AutoMigrate(&Address{}); err != nil {
		log.Fatalf(utils.DatabaseMigrationError, "Address", err)
	} else {
		log.Printf(utils.SchemaMigrationSuccess, "Address")
	}
}
