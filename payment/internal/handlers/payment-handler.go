package handlers

import (
	"e-commerce-backend/payment/dbs"
	"e-commerce-backend/payment/internal/services"
	"e-commerce-backend/shared/middlewares"
	"net/http"

	"github.com/gorilla/mux"
)

func PaymentHandler(r *mux.Router) {
	paymentService := services.NewPaymentService(dbs.DB)

	r.Handle("/order/{id}/payment", middlewares.AuthMiddleware(http.HandlerFunc(paymentService.GetPayment))).Methods("POST")
	r.Handle("/order/{id}/payment/initiate", middlewares.AuthMiddleware(http.HandlerFunc(paymentService.InitiatePayment))).Methods("POST")
}
