package main

import (
	"e-commerce-backend/shared/middlewares"
	"e-commerce-backend/shared/utils"
	"e-commerce-backend/users/dbs"
	"e-commerce-backend/users/internal/handlers"
	"e-commerce-backend/users/internal/models"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	dbs.InitDB()
	defer dbs.CloseDB()
	db := dbs.UserDB

	InitSchemas()

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		utils.JsonResponse(nil, w, "Hello World", 0)
	})
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"X-Requested-With", "Content-Type", "Authorization"},
	})

	r.Handle("/admin", middlewares.AuthMiddleware(middlewares.RoleMiddleware(db, "admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Admin")
	}))))

	handlers.UserHandler(r)

	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal(".env file not found from main.go")
	}
	port := os.Getenv("USER_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe("localhost:"+port, c.Handler(r)); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func InitSchemas() {
	models.InitUserSchema()
	models.InitAddressSchemas()
}
