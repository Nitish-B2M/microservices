package dbs

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var UserDB *gorm.DB

func InitDB() {
	var err error

	dbUser := "root"
	dbPassword := "root"
	dbHost := "localhost"
	dbPort := "3307"
	dbName := "ecomm"

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPassword, dbHost, dbPort, dbName)
	UserDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error opening DB connection: %v", err)
	}

	log.Println("Successfully connected to MySQL database")
}

func CloseDB() {
	DB, _ := UserDB.DB()
	DB.Close()
}
