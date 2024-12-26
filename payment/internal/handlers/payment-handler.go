package handlers

import (
	"e-commerce-backend/payment/dbs"
	"e-commerce-backend/payment/internal/services/payment"
	"e-commerce-backend/shared/middlewares"
	"net/http"

	"github.com/gorilla/mux"
)

func PaymentHandler(r *mux.Router) {
	paymentService := payment.NewPaymentService(dbs.DB)

	r.Handle("/user/{id}/payment", middlewares.AuthMiddleware(http.HandlerFunc(paymentService.GetPayment))).Methods("POST")
	r.Handle("/user/{id}/payment/add", middlewares.AuthMiddleware(http.HandlerFunc(paymentService.AddToPayment))).Methods("POST")
	r.Handle("/user/{id}/payment/update", middlewares.AuthMiddleware(http.HandlerFunc(paymentService.UpdatePayment))).Methods("POST")
	r.Handle("/user/{id}/payment/delete", middlewares.AuthMiddleware(http.HandlerFunc(paymentService.DeletePayment))).Methods("POST")
}
