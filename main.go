package main

import (
	"fmt"
	"log"
	"microservices/pkg/handlers"
	"microservices/pkg/services/product"
	"microservices/pkg/shared/dbs"
	"microservices/pkg/shared/middlewares"
	"microservices/pkg/shared/models"
	"microservices/pkg/shared/utils"
	"net/http"
	"os"
)

func ProtectedHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(uint)
	fmt.Fprintf(w, "Hello, User ID: %d", userID)
}

func main() {
	dbs.InitDB()
	defer dbs.CloseDB()
	db := dbs.DB

	InitSchemas()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		utils.JsonResponse(nil, w, "Hello World", 0)
	})

	http.Handle("/protected", middlewares.AuthMiddleware(http.HandlerFunc(ProtectedHandler)))
	http.Handle("/admin", middlewares.AuthMiddleware(middlewares.RoleMiddleware(db, "admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Admin")
	}))))

	product.ProductHandler()
	handlers.UserHandler()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe("localhost:"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func InitSchemas() {
	models.InitProductSchema()
	models.InitTagSchema()
	models.InitUserSchema()
}
