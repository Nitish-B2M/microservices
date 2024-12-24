package main

import (
	"fmt"
	"github.com/gorilla/mux"
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

	product.ProductHandler(r)
	handlers.UserHandler(r)
	handlers.CartHandler(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe("localhost:"+port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func InitSchemas() {
	models.InitProductSchema()
	models.InitTagSchema()
	models.InitUserSchema()
	models.InitCartSchema()
}
