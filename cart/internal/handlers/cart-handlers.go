package handlers

import (
	"e-commerce-backend/cart/dbs"
	"e-commerce-backend/cart/internal/services"
	"e-commerce-backend/shared/middlewares"
	"net/http"

	"github.com/gorilla/mux"
)

func CartHandler(r *mux.Router) {
	cartService := services.NewService(dbs.DB)

	r.Handle("/user/{id}/cart", middlewares.AuthMiddleware(http.HandlerFunc(cartService.GetCartItemByUserID))).Methods("GET")
	r.Handle("/user/{id}/cart/add", middlewares.AuthMiddleware(http.HandlerFunc(cartService.AddToCart))).Methods("POST")
	r.Handle("/user/{id}/cart/delete", middlewares.AuthMiddleware(http.HandlerFunc(cartService.RemoveFromCart))).Methods("POST")
	r.Handle("/user/{id}/cart/{cartID}/checkout", middlewares.AuthMiddleware(http.HandlerFunc(cartService.Checkout))).Methods("POST")
}
