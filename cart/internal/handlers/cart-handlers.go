package handlers

import (
	"e-commerce-backend/cart/dbs"
	"e-commerce-backend/cart/internal/services"

	"github.com/gorilla/mux"
)

func CartHandler(r *mux.Router) {
	cartService := services.NewService(dbs.DB)
	_ = cartService
	// r.Handle("/user/{id}/cart", middlewares.AuthMiddleware(http.HandlerFunc(cartService.GetCart))).Methods("POST")
	// r.Handle("/user/{id}/cart/add", middlewares.AuthMiddleware(http.HandlerFunc(cartService.AddToCart))).Methods("POST")
	// r.Handle("/user/{id}/cart/update", middlewares.AuthMiddleware(http.HandlerFunc(cartService.UpdateCart))).Methods("POST")
	// r.Handle("/user/{id}/cart/delete", middlewares.AuthMiddleware(http.HandlerFunc(cartService.DeleteCart))).Methods("POST")
}
