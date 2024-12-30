package services

import (
	"e-commerce-backend/payment/internal/models"
	"e-commerce-backend/shared/utils"
	"fmt"
	"net/http"

	"gorm.io/gorm"
)

type Service struct {
	DB *gorm.DB
}

type PaymentService interface {
	HandlePaymentRequest(w http.ResponseWriter, r *http.Request)
	GetPayment(w http.ResponseWriter, r *http.Request)
	AddToPayment(w http.ResponseWriter, r *http.Request)
	UpdatePayment(w http.ResponseWriter, r *http.Request)
	DeletePayment(w http.ResponseWriter, r *http.Request)
}

func NewPaymentService(db *gorm.DB) *Service {
	return &Service{
		DB: db,
	}
}

func (s *Service) HandlePaymentRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, Payment!")
}

func (s *Service) GetPayment(w http.ResponseWriter, r *http.Request) {
	var cart models.Payment

	utils.JsonResponse(cart, w, utils.ProductCategoryError, http.StatusCreated)
}

func (s *Service) AddToPayment(w http.ResponseWriter, r *http.Request) {
	var cart models.Payment

	utils.JsonResponse(cart, w, utils.ProductCategoryError, http.StatusCreated)
}

func (s *Service) UpdatePayment(w http.ResponseWriter, r *http.Request) {
	var cart models.Payment

	utils.JsonResponse(cart, w, utils.ProductCategoryError, http.StatusCreated)
}

func (s *Service) DeletePayment(w http.ResponseWriter, r *http.Request) {
	var cart models.Payment

	utils.JsonResponse(cart, w, utils.ProductCategoryError, http.StatusCreated)
}
