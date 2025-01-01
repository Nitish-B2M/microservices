package main

import (
	"e-commerce-backend/order/dbs"
	"e-commerce-backend/order/internal/handlers"
	"e-commerce-backend/order/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	dbs.InitDB()
	defer dbs.CloseDB()
	if dbs.DB == nil {
		log.Fatal("DB connection is nil after InitDB")
	}

	//Init Schemas
	InitSchemas()

	router := gin.Default()
	r := router.Group("user/:id/order")
	handlers.OrderHandler(r)

	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal("Error loading .env file from main.go")
	}
	port := os.Getenv("ORDER_PORT")
	if port == "" {
		port = "8083"
	}
	
	if err := router.Run("localhost:" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func InitSchemas() {
	models.InitOrderSchemas()
}
