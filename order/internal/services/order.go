package services

import (
	"bytes"
	"e-commerce-backend/order/internal/models"
	"e-commerce-backend/order/pkg/constants"
	"e-commerce-backend/order/pkg/payloads"
	"e-commerce-backend/shared/utils"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type Service struct {
	DB *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		DB: db,
	}
}

type OrderInterface interface {
	GetOrders(c *gin.Context)
	CreateOrder(c *gin.Context)
	GetOrderById(c *gin.Context)
	Checkout(c *gin.Context)
}

func (db *Service) GetOrders(c *gin.Context) {
	log.Println("GetOrder")
	c.JSON(200, gin.H{"status": "ok"})
}

func (db *Service) CreateOrder(c *gin.Context) {
	log.Println("CreateOrder")
}

func (db *Service) GetOrderById(c *gin.Context) {
	log.Println("GetOrderById")
}

func (db *Service) Checkout(c *gin.Context) {
	// need cart
	// - cart id, price, product id also include product details
	// need user details
	// when go for checkout then it will first call create order then use order id to proceed for checkout

	userId, err := constants.GetUserIdFromParams(c)
	if err != nil {
		utils.GinError(c, fmt.Sprintf(utils.UserNotFoundError, userId), http.StatusNotFound, err)
		return
	}

	//how to fetch body
	var body payloads.RequestCart
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.GinError(c, utils.InvalidJSONBody, http.StatusBadRequest, err)
		return
	}
	carts := body.Carts
	var cartIds []int
	totalPrice := 0.0
	var order models.Order

	for _, cart := range carts {
		cartId := int(cart["cart_id"].(float64))

		cart, err := fetchCartDetails(c, userId, cartId)
		if err != nil {
			utils.GinError(c, err.Error(), http.StatusBadRequest, err)
			return
		}
		if cart == nil {
			utils.GinError(c, fmt.Sprintf(utils.CartItemNotFoundError, cartId), http.StatusNotFound, err)
			return
		}
		cartData := cart["data"].(map[string]interface{})

		productId := int(cartData["product_id"].(float64))
		product := fetchProductDetails(c, productId)
		if product == nil {
			utils.GinError(c, fmt.Sprintf(utils.ProductNotFoundError, productId), http.StatusNotFound, err)
			return
		}
		productData := product["data"].(map[string]interface{})
		if ok := verifyQuantity(int(cartData["quantity"].(float64)), int(productData["quantity"].(float64))); !ok {
			utils.GinError(c, fmt.Sprintf(utils.CartOutOfStockError, productId), http.StatusBadRequest, nil)
			return
		}

		eachTotalPrice := calculatePrice(productData["price"].(float64), int(cartData["quantity"].(float64)))
		totalPrice += eachTotalPrice

		cartIds = append(cartIds, cartId)
	}

	cartItemsJSON, err := json.Marshal(cartIds)
	if err != nil {
		utils.GinError(c, err.Error(), http.StatusBadRequest, err)
		return
	}
	//temporary basis
	order.OrderID = 1
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	order.CustomerID = userId
	order.TotalAmount = totalPrice
	order.Carts = string(cartItemsJSON)
	//create order
	if err := order.CreateOrder(db.DB); err != nil {
		utils.GinError(c, err.Error(), http.StatusBadRequest, err)
		return
	}
	log.Println("Order created: ", order)

	payResp, err := proceedForPayment(c, order)
	if err != nil {
		utils.GinError(c, err.Error(), http.StatusBadRequest, err)
		return
	}

	if payResp["data"] != nil {
		respData := payResp["data"].(map[string]interface{})
		paymentId := int(respData["payment_id"].(float64))
		if paymentId > 0 {
			order.IsPaid = true
			//	update payment
			if err := order.UpdateOrder(db.DB); err != nil {
				utils.GinError(c, "while updating"+err.Error(), http.StatusBadRequest, err)
				return
			}
		}
	}

	//invoices.Invoices()

	resp := map[string]interface{}{
		"data": order,
	}

	utils.GinResponse(resp, c, utils.OrderSuccessful, http.StatusOK)
}

func fetchCartDetails(c *gin.Context, userId, cartId int) (map[string]interface{}, error) {
	links := constants.MicroserviceLinks()
	cartLink := links["cartMSCallByIdLink"]

	cartMicroserviceCall := fmt.Sprintf(cartLink, userId, cartId)
	log.Println(cartMicroserviceCall)
	req, err := http.NewRequest(http.MethodGet, cartMicroserviceCall, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request to microservice")
	}

	token := utils.GetTokenFromRequestUsingGin(c)
	if token == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	req.Header.Set("Authorization", token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(utils.ErrorCallingCartMicroservice)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//utils.GinErrorWithExtra(c, "utils.ErrorReadingResponseBody", http.StatusInternalServerError, err)
		return nil, fmt.Errorf(utils.ErrorCallingCartMicroservice)
	}

	var cart map[string]interface{}
	if err := utils.ParseJSON(body, &cart); err != nil {
		return nil, fmt.Errorf("failed to parse cart details")
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			if strings.Contains(cart["error"].(string), "not found") {
				return nil, fmt.Errorf(cart["error"].(string))
			}
		}
		return nil, fmt.Errorf(utils.ErrorCallingCartMicroservice)
	}

	return cart, nil
}

func fetchProductDetails(c *gin.Context, productId int) map[string]interface{} {
	links := constants.MicroserviceLinks()
	productLink := links["productMSCallByIdLink"]

	productMicroserviceCall := fmt.Sprintf(productLink, productId)
	log.Println(productMicroserviceCall)
	req, err := http.NewRequest(http.MethodGet, productMicroserviceCall, nil)
	if err != nil {
		utils.LogError("Failed to create request", map[string]interface{}{"error": err.Error()})
		return nil
	}

	token := utils.GetTokenFromRequestUsingGin(c)
	if token == "" {
		utils.GinError(c, utils.MissingAuthorizationHeader, http.StatusUnauthorized, nil)
		return nil
	}
	req.Header.Set("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		utils.GinErrorWithExtra(c, utils.ErrorProductMicroservices, http.StatusInternalServerError, err)
		return nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		utils.GinErrorWithExtra(c, "utils.ErrorReadingResponseBody", http.StatusInternalServerError, err)
		return nil
	}

	var product map[string]interface{}
	if err := utils.ParseJSON(body, &product); err != nil {
		log.Printf("Error parsing Cart response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse cart details"})
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			if strings.Contains(product["error"].(string), "not found") {
				utils.GinError(c, product["error"].(string), http.StatusNotFound, nil)
				return nil
			}
		}
		utils.GinError(c, utils.ErrorCallingCartMicroservice, resp.StatusCode, nil)
		return nil
	}

	return product
}

func verifyQuantity(requestQuantity, actualQuantity int) bool {
	if requestQuantity <= actualQuantity {
		return true
	}
	return false
}

func calculatePrice(price float64, quantity int) float64 {
	total := price * float64(quantity)
	return total
}

func proceedForPayment(c *gin.Context, order models.Order) (map[string]interface{}, error) {
	links := constants.MicroserviceLinks()
	paymentLink := links["paymentMSInitiateCallLink"]
	paymentMicroserviceCall := fmt.Sprintf(paymentLink, order.OrderID)
	log.Println(paymentMicroserviceCall)

	payload := map[string]interface{}{"customer_id": order.CustomerID, "total_amount": order.TotalAmount, "order_id": order.OrderID}
	jsonPayload, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, paymentMicroserviceCall, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request to microservice: %v", err)
	}
	token := utils.GetTokenFromRequestUsingGin(c)
	if token == "" {
		return nil, fmt.Errorf("missing authorization header")
	}
	req.Header.Set("Authorization", token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to microservice: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body")
	}
	var response map[string]interface{}
	if err := utils.ParseJSON(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response body")
	}
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusCreated {
		} else {
			return nil, fmt.Errorf("microservice responded with status code %d", resp.StatusCode)
		}
	}

	return response, nil
}
