package services

import (
	"e-commerce-backend/payment/internal/models"
	"e-commerce-backend/shared/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/paymentintent"
	"github.com/stripe/stripe-go/v81/price"
	"github.com/stripe/stripe-go/v81/product"
	"log"
	"net/http"
	"os"
	"time"

	"gorm.io/gorm"
)

type Service struct {
	DB *gorm.DB
}

type PaymentService interface {
	GetPayment(w http.ResponseWriter, r *http.Request)
	InitiatePayment(w http.ResponseWriter, r *http.Request)
	RefundPayment(w http.ResponseWriter, r *http.Request)
}

func NewPaymentService(db *gorm.DB) *Service {
	return &Service{
		DB: db,
	}
}

func initStripe() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
		return
	}
	stripe.Key = os.Getenv("PAYMENT_SECRET_KEY")
}

func (s *Service) GetPayment(w http.ResponseWriter, r *http.Request) {
	var pay models.Payment

	utils.JsonResponse(pay, w, utils.ProductCategoryError, http.StatusCreated)
}

func ValidatePaymentRequest(req map[string]interface{}) error {
	if int(req["order_id"].(float64)) <= 0 {
		return errors.New("order id is required")
	}
	if int(req["customer_id"].(float64)) <= 0 {
		return errors.New("customer id is required")
	}
	if int(req["total_amount"].(float64)) <= 0 {
		return errors.New("total amount is required")
	}
	return nil
}

func (s *Service) InitiatePayment(w http.ResponseWriter, r *http.Request) {

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JsonResponse(req, w, utils.InvalidPaymentRequest, http.StatusBadRequest)
		return
	}
	if err := ValidatePaymentRequest(req); err != nil {
		utils.JsonResponse(req, w, utils.PaymentValidationFailed, http.StatusBadRequest)
		return
	}
	log.Println("request:", req)
	orderId := int(req["order_id"].(float64))
	amount := req["total_amount"].(float64)

	//for now using demoPayment
	payment := demoPayment(orderId, amount)
	//store payment
	if err := payment.CreatePayment(s.DB); err != nil {
		utils.JsonResponse(req, w, utils.PaymentFailed, http.StatusInternalServerError)
		return
	}
	resp := map[string]interface{}{}
	if payment.PaymentID > 0 {
		resp["payment_id"] = payment.PaymentID
	}

	//this not working
	//intent := stripePayment(w)
	//response := map[string]interface{}{
	//	"client_secret": intent.ClientSecret,
	//}
	utils.JsonResponse(resp, w, utils.PaymentSuccessful, http.StatusCreated)
}

func demoPayment(orderId int, amount float64) models.Payment {
	resp := models.Payment{
		OrderID:              orderId,
		PaymentMethod:        "xyz",
		PaymentDate:          time.Now(),
		PaymentFailureReason: "",
		PaymentStatus:        utils.PaymentStatusPaid,
		PaymentRetryCount:    0,
		Amount:               amount,
	}
	return resp
}

func stripePayment(w http.ResponseWriter) *stripe.PaymentIntent {
	initStripe()
	productParams := &stripe.ProductParams{
		Name:        stripe.String("Starter Subscription"),
		Description: stripe.String("$12/Month subscription"),
	}
	starterProduct, err := product.New(productParams)
	if err != nil {
		log.Fatalf("Error creating product: %v", err)
		utils.JsonError(w, "Error creating product", http.StatusInternalServerError, err)
		return nil
	}

	// Create price for the product
	priceParams := &stripe.PriceParams{
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		Product:  stripe.String(starterProduct.ID),
		Recurring: &stripe.PriceRecurringParams{
			Interval: stripe.String(string(stripe.PriceRecurringIntervalMonth)),
		},
		UnitAmount: stripe.Int64(1),
	}
	starterPrice, err := price.New(priceParams)
	if err != nil {
		log.Fatalf("Error creating price: %v", err)
		utils.JsonError(w, "Error creating price", http.StatusInternalServerError, err)
		return nil
	}

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(starterPrice.UnitAmount),
		Currency: stripe.String(string(stripe.CurrencyUSD)),
	}
	intent, err := paymentintent.New(params)
	if err != nil {
		log.Fatalf("Error creating PaymentIntent: %v", err)
		utils.JsonError(w, "Error creating PaymentIntent", http.StatusInternalServerError, err)
		return nil
	}

	fmt.Println("Success! Here is your starter subscription product id: " + starterProduct.ID)
	fmt.Println("Success! Here is your starter subscription price id: " + starterPrice.ID)
	log.Println("Payment Intent: ", *intent)
	return intent
}

func (s *Service) RefundPayment(w http.ResponseWriter, r *http.Request) {
	//	logic here
}
