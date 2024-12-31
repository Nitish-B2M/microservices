package dbs

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error

	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	if dbUser == "" || dbPass == "" || dbHost == "" || dbName == "" || dbPort == "" {
		log.Fatal("Missing required environment variables")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPass, dbHost, dbPort, dbName)
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error opening DB connection: %v", err)
	}

	log.Println("Successfully connected to MySQL database")
}

func CloseDB() {
	// Close the database connection
	sqlDB, err := DB.DB()
	if err != nil {
		log.Printf("Error closing DB connection: %v", err)
		return
	}

	if err := sqlDB.Close(); err != nil {
		log.Printf("Error closing DB connection: %v", err)
	}
	log.Println("Successfully closed the database connection")
}
