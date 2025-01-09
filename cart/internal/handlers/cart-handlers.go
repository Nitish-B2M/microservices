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

	r.Handle("/user/cart", middlewares.AuthMiddleware(http.HandlerFunc(cartService.GetCartItemByUserID))).Methods("GET")
	r.Handle("/user/cart/add", middlewares.AuthMiddleware(http.HandlerFunc(cartService.AddToCart))).Methods("POST")
	r.Handle("/user/cart/{id}/update/qty", middlewares.AuthMiddleware(http.HandlerFunc(cartService.UpdateCartQty))).Methods("POST")
	r.Handle("/user/cart/delete", middlewares.AuthMiddleware(http.HandlerFunc(cartService.RemoveFromCart))).Methods("POST")
	r.Handle("/user/cart/{cart_id}/checkout", middlewares.AuthMiddleware(http.HandlerFunc(cartService.Checkout))).Methods("POST")
	r.Handle("/user/cart/{cart_id}", middlewares.AuthMiddleware(http.HandlerFunc(cartService.GetCartByCartId))).Methods("GET")
}
