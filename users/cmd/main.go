package main

import (
	"e-commerce-backend/shared/utils"
	"e-commerce-backend/users/dbs"
	"e-commerce-backend/users/internal/handlers"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	dbs.InitDB()
	defer dbs.CloseDB()

	InitSchemas(dbs.UserDB)

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		utils.JsonResponse(nil, w, "Hello World", 0)
	})
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"X-Requested-With", "Content-Type", "Authorization"},
	})

	utils.RegisterSubRoutes("/admin", r, handlers.AdminRoutes)
	utils.RegisterSubRoutes("/user/auth", r, handlers.AuthRoutes)
	utils.RegisterSubRoutes("/user/role", r, handlers.RoleRoutes)
	utils.RegisterSubRoutes("/user/profile", r, handlers.ProfileRoutes)
	utils.RegisterSubRoutes("/user/address", r, handlers.AddressRoutes)

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

func InitSchemas(db *gorm.DB) {
	// models.InitUserSchema()
	// models.InitAddressSchemas()
	// models.InitUserRoleSchema()
	// repository.AutoMigrateRoleTables(db)
}
