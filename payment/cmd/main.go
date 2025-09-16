package main

import (
	"e-commerce-backend/payment/dbs"
	"e-commerce-backend/payment/internal/handlers"
	"e-commerce-backend/payment/internal/models"
	"e-commerce-backend/shared/middlewares"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	dbs.InitDB()
	defer dbs.CloseDB()
	db := dbs.DB

	InitSchemas()

	r := mux.NewRouter()
	//r.Use(middlewares.AuthMiddleware)
	//r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	//	utils.JsonResponse(nil, w, "Hello World", 0)
	//})

	r.Handle("/admin", middlewares.AuthMiddleware(middlewares.RoleMiddleware(db, "admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Admin")
	}))))

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"X-Requested-With", "Content-Type", "Authorization"},
	})

	handlers.PaymentHandler(r)

	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal(".env file not found from main.go")
	}

	port := os.Getenv("PAYMENT_PORT")
	if port == "" {
		port = "8084"
	}
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe("localhost:"+port, c.Handler(r)); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func InitSchemas() {
	models.InitPaymentSchema()
}
