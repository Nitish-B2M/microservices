package main

import (
	"e-commerce-backend/cart/dbs"
	"e-commerce-backend/cart/internal/handlers"
	"e-commerce-backend/cart/internal/models"
	"e-commerce-backend/shared/utils"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	dbs.InitDB()
	defer dbs.CloseDB()

	InitSchemas()

	r := mux.NewRouter()
	//r.Use(middlewares.AuthMiddleware)
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		utils.JsonResponse(nil, w, "Hello World", 0)
	})

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"X-Requested-With", "Content-Type", "Authorization"},
	})

	handlers.CartHandler(r)

	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal(".env file not found from main.go")
	}
	port := os.Getenv("CART_PORT")
	if port == "" {
		port = "8082"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe("localhost:"+port, c.Handler(r)); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func InitSchemas() {
	models.InitCartSchema()
}
