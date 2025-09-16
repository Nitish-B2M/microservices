package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"e-commerce-backend/payment/internal/models"
	"e-commerce-backend/shared/utils"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/paymentintent"
	"github.com/stripe/stripe-go/v81/price"
	"github.com/stripe/stripe-go/v81/product"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Service struct {
	DB *gorm.DB
}

type RapydClient struct {
	c *http.Client
}
type Server struct {
	rapydClient *RapydClient
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
	if req["order_id"].(string) == "" {
		return errors.New("order id is required")
	}
	if int(req["user_id"].(float64)) <= 0 {
		return errors.New("user id is required")
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
	orderId := req["order_id"].(string)
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

func demoPayment(orderId string, amount float64) models.Payment {
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
		utils.JsonError(w, "error creating product", http.StatusInternalServerError, err)
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

var (
	html map[string]string = map[string]string{
		"home": `<!doctype html>
<html lang='en'>
    <head>
        <meta charset='utf-8'>
        <title>Rapyd Checkout Demo</title>
    </head>
    <body>
        <header>
            <h1>Rapyd Checkout Demo</h1>
        </header>
		<main>
			<form action='/payment/checkout' method='POST'>
				<div>
					<h2>Buy woolen socks</h2>
				</div>
				<div>
					<label>Number of items:</label>
					<input type='text' name='amount' value='1'>
				</div> 
				<input type='submit' value='Continue to checkout'> </div>
			</form>
		</main>
    </body>
</html>`,
		"complete": `<!doctype html>
<html lang='en'>
    <head>
        <meta charset='utf-8'>
        <title>Rapyd Checkout Demo</title>
    </head>
    <body>
        <header>
            <h1>Rapyd Checkout Demo</h1>
        </header>
        <nav>
            <a href='/'>Home</a>
        </nav>
        <main>
			<h2>Checkout complete</h2>
        </main>
    </body>
</html>`,
		"cancel": `<!doctype html>
<html lang='en'>
    <head>
        <meta charset='utf-8'>
        <title>Rapyd Checkout Demo</title>
    </head>
    <body>
        <header>
            <h1>Rapyd Checkout Demo</h1>
        </header>
        <nav>
            <a href='/'>Home</a>
        </nav>
        <main>
			<h2>Checkout canceled</h2>
        </main>
    </body>
</html>`,
	}
)

type CheckoutPage struct {
	Amount              float64 `json:"amount"`
	Country             string  `json:"country"`
	Currency            string  `json:"currency"`
	Customer            int     `json:"customer"`
	CompleteCheckoutURL string  `json:"complete_checkout_url"`
	CancelCheckoutURL   string  `json:"error_checkout_url"`
}

func NewRapydServer() *Server {
	return &Server{
		rapydClient: NewRapydClient(),
	}
}

func (s *Server) ListOutCountries(w http.ResponseWriter, r *http.Request) {
	body, err := s.rapydClient.Request("GET", "/v1/data/countries", nil)
	if err != nil {
		utils.JsonError(w, "error calling /v1/data/countries: %w", http.StatusInternalServerError, err)
	}

	var checkoutResponse map[string]any
	err = json.Unmarshal(body, &checkoutResponse)
	if err != nil {
		utils.JsonError(w, "cannot unmarshal response from /v1/data/countries", http.StatusInternalServerError, err)
		return
	}

	//data := checkoutResponse["data"].(map[string]any)
	utils.JsonResponse(checkoutResponse, w, "success", http.StatusOK)
}

func (s *Server) CreateCheckoutPage(host string, c float64) (string, error) {
	// create the Rapyd checkout page with data from the basket.
	checkoutPage := CheckoutPage{
		Amount:              1.0,
		Country:             "US",
		Currency:            "USD",
		CompleteCheckoutURL: "http://127.0.0.1:3000/payment/complete",
		CancelCheckoutURL:   "http://127.0.0.1:3000/payment/cancel",
	}

	reqBody, err := json.Marshal(checkoutPage)
	if err != nil {
		return "", fmt.Errorf("error marshalling json: %w", err)
	}

	body, err := s.rapydClient.Request("POST", "/v1/checkout", reqBody)
	if err != nil {
		return "", fmt.Errorf("error calling /v1/checkout: %w", err)
	}

	var checkoutResponse map[string]any
	err = json.Unmarshal(body, &checkoutResponse)
	if err != nil {
		return "", fmt.Errorf("cannot unmarshal response from /v1/checkout: %w, body: %s", err, string(body))
	}

	status := checkoutResponse["status"].(map[string]any)
	if status["error_code"] != "" {
		return "", fmt.Errorf("error creating checkout page: %s: %s",
			status["status"],
			status["message"],
		)
	}

	data := checkoutResponse["data"].(map[string]any)
	return data["redirect_url"].(string), nil
}

func raypdSignature(httpMethod string, urlPath string, salt string, timestamp string, accessKey string, secretKey string, body string) string {
	// create a sha256 HMAC with the secret key
	hash := hmac.New(sha256.New, []byte(secretKey))

	// sign the request
	hash.Write([]byte(strings.ToLower(httpMethod) + urlPath + salt + timestamp + accessKey + secretKey + body))

	// get the hex digest of the hash and base64-encode it
	hexdigest := make([]byte, hex.EncodedLen(hash.Size()))
	hex.Encode(hexdigest, hash.Sum(nil))
	return base64.StdEncoding.EncodeToString(hexdigest)
}

func NewRapydClient() *RapydClient {
	return &RapydClient{
		c: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func FetchRapydHeader(w http.ResponseWriter, r *http.Request) {
	body := []byte(`{"amount": 1, "country": "US", "currency": "USD", "complete_payment_url": "http://localhost:8080/payment/complete", "error_payment_url": "http://localhost:8080/payment/cancel"}`)
	b := bytes.NewReader(body)
	urlPath := "/v1/checkout"
	_, err := http.NewRequest("POST", "https://sandboxapi.rapyd.net"+urlPath, b)
	if err != nil {
		log.Println(fmt.Sprintf("failed to create request: %v", err))
		return
	}
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	salt := fmt.Sprintf("%016x", rand.Uint64())
	key := os.Getenv("RAPYD_ACCESS_KEY2")
	secret := os.Getenv("RAPYD_SECRET_KEY2")

	w.Write([]byte(fmt.Sprintf("access_key=%s&salt=%s&timestamp=%s&signature=%s", key, salt, timestamp, raypdSignature(http.MethodPost, urlPath, salt, timestamp, key, secret, string(body)))))
}

func (rc *RapydClient) Request(method string, urlPath string, body []byte) ([]byte, error) {
	// turn the body into an io.Reader
	b := bytes.NewReader(body)

	// create a request with all headers required for authentication.
	req, err := http.NewRequest(method, "https://sandboxapi.rapyd.net"+urlPath, b)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	salt := fmt.Sprintf("%016x", rand.Uint64())
	key := os.Getenv("RAPYD_ACCESS_KEY2")
	secret := os.Getenv("RAPYD_SECRET_KEY2")
	if key == "" || secret == "" {
		log.Fatalln("Please set the environment variables RAPYD_ACCESS_KEY and RAPYD_SECRET_KEY before starting the server.")
	}
	req.Header.Set("access_key", key)
	req.Header.Set("salt", salt)
	req.Header.Set("timestamp", timestamp)
	req.Header.Set("signature", raypdSignature(method, urlPath, salt, timestamp, key, secret, string(body)))

	// run the request and return the response body.
	resp, err := rc.c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	return respBody, nil
}

func (s *Server) HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html["home"])
}

type Request struct {
	Amount float64 `json:"amount"`
}

func (s *Server) CheckoutHandler(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Failed to parse request body:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create the checkout page using the Rapyd API
	rapydCheckoutPage, err := s.CreateCheckoutPage(r.Host, req.Amount)
	if err != nil {
		log.Println("Cannot create checkout page:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Return the redirect URL to the frontend
	response := map[string]string{
		"redirect_url": rapydCheckoutPage,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) CompleteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html["complete"])
}

func (s *Server) CancelHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html["cancel"])
}

type WebhookPayload struct {
	Event     string  `json:"event"`
	Reference string  `json:"reference"`
	Status    string  `json:"status"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
}

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r.Body); err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	body := buf.Bytes()

	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	fmt.Printf("Received webhook: %+v\n", payload)

	switch payload.Event {
	case "payment.success":
		fmt.Printf("Payment success for reference: %s\n", payload.Reference)
	case "payment.failed":
		fmt.Printf("Payment failed for reference: %s\n", payload.Reference)
	default:
		fmt.Printf("Unhandled event: %s\n", payload.Event)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Webhook received successfully"))
}
