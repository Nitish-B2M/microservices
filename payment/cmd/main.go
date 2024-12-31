package main

import (
	"e-commerce-backend/payment/dbs"
	"e-commerce-backend/payment/internal/handlers"
	"e-commerce-backend/payment/internal/models"
	"e-commerce-backend/shared/middlewares"
	"e-commerce-backend/shared/utils"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	dbs.InitDB()
	defer dbs.CloseDB()
	db := dbs.DB

	InitSchemas()

	r := mux.NewRouter()
	//r.Use(middlewares.AuthMiddleware)
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		utils.JsonResponse(nil, w, "Hello World", 0)
	})

	r.Handle("/admin", middlewares.AuthMiddleware(middlewares.RoleMiddleware(db, "admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Admin")
	}))))

	handlers.PaymentHandler(r)

	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal(".env file not found from main.go")
	}

	port := os.Getenv("PAYMENT_PORT")
	if port == "" {
		port = "8084"
	}
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe("localhost:"+port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func InitSchemas() {
	models.InitPaymentSchema()
}
