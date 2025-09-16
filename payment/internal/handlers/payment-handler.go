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
	s := services.NewRapydServer()

	//1. /payment/checkout
	//3. /payment/retry
	r.Handle("/order/{id}/payment", middlewares.AuthMiddleware(http.HandlerFunc(paymentService.GetPayment))).Methods("POST")
	r.Handle("/order/{id}/payment/initiate", middlewares.AuthMiddleware(http.HandlerFunc(paymentService.InitiatePayment))).Methods("POST")
	r.HandleFunc("/rapyd/list_out_countries", s.ListOutCountries).Methods("GET")
	r.HandleFunc("/payment/", s.HomeHandler).Methods("GET")
	r.HandleFunc("/payment/checkout", s.CheckoutHandler).Methods("POST")
	r.HandleFunc("/complete", s.CompleteHandler).Methods("GET")
	r.HandleFunc("/payment/cancel", s.CancelHandler).Methods("GET")
	r.HandleFunc("/payment/webhook", services.WebhookHandler).Methods("GET")
	r.HandleFunc("/payment/header", services.FetchRapydHeader).Methods("GET")
}
