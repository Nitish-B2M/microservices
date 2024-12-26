package main

import (
	"e-commerce-backend/cart/dbs"
	"e-commerce-backend/cart/internal/handlers"
	"e-commerce-backend/cart/internal/models"
	"e-commerce-backend/shared/middlewares"
	"e-commerce-backend/shared/utils"
	"fmt"
	"log"
	"net/http"

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

	handlers.CartHandler(r)

	port := "8083"
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe("localhost:"+port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func InitSchemas() {
	models.InitCartSchema()
}
