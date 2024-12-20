package main

import (
	"fmt"
	"log"
	"microservices/pkg/services/product"
	"microservices/pkg/shared/dbs"
	"microservices/pkg/shared/models"
	"net/http"
	"os"
)

func main() {
	dbs.InitDB()
	defer dbs.CloseDB()

	InitSchemas()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})

	product.HandleProductRequest()

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
}
