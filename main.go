package main

import (
	"log"
	"microservices/pkg/handlers"
	"microservices/pkg/services/product"
	"microservices/pkg/shared/dbs"
	"microservices/pkg/shared/models"
	"microservices/pkg/shared/utils"
	"net/http"
	"os"
)

func main() {
	dbs.InitDB()
	defer dbs.CloseDB()

	InitSchemas()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		utils.JsonResponse(nil, w, "Hello World", 0)
	})

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
