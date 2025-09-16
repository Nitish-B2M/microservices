package main

import (
	"e-commerce-backend/products/dbs"
	"e-commerce-backend/products/internal/handlers"
	"e-commerce-backend/shared/utils"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"gorm.io/gorm"

	"github.com/gorilla/mux"
)

func main() {
	dbs.InitDB()
	db := dbs.DB

	InitSchemas(db)

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

	handlers.ProductHandler(r)
	utils.RegisterSubRoutes("/product", r, handlers.ProductRoutes)
	utils.RegisterSubRoutes("/product/seller", r, handlers.SellerRoutes)

	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal(".env file not found from main.go")
	}
	port := os.Getenv("PRODUCT_PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Server starting on port %s", port)
	srv := &http.Server{
		Addr:    "localhost:" + port,
		Handler: c.Handler(r),
	}
	defer dbs.CloseDB()
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func InitSchemas(db *gorm.DB) {
	// models.InitProductSchema()
	//models.InitNewProductSchema(db)
}
