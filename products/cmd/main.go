package main

import (
	"e-commerce-backend/products/dbs"
	"e-commerce-backend/products/internal/handlers"
	"e-commerce-backend/products/internal/models"
	"e-commerce-backend/shared/middlewares"
	"e-commerce-backend/shared/utils"
	"fmt"
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

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"X-Requested-With", "Content-Type", "Authorization"},
	})

	handlers.ProductHandler(r)

	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal(".env file not found from main.go")
	}
	port := os.Getenv("PRODUCT_PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe("localhost:"+port, c.Handler(r)); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func InitSchemas() {
	models.InitProductSchema()
}
