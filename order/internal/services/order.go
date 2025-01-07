package services

import (
	"bytes"
	"e-commerce-backend/order/internal/models"
	"e-commerce-backend/order/pkg/constants"
	"e-commerce-backend/order/pkg/payloads"
	"e-commerce-backend/shared/invoices"
	"e-commerce-backend/shared/utils"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
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
	if err := constants.ValidateUserWithCtxUserId(c); err != nil {
		utils.GinError(c, err.Error(), http.StatusBadRequest, err)
		return
	}

	c.JSON(200, gin.H{"status": "ok"})
}

func (db *Service) GetOrderById(c *gin.Context) {
	if err := constants.ValidateUserWithCtxUserId(c); err != nil {
		utils.GinError(c, err.Error(), http.StatusBadRequest, err)
		return
	}

	orderIdStr := c.Param("order_id")
	if orderIdStr == "" {
		utils.GinError(c, utils.OrderIdRequired, http.StatusBadRequest, nil)
		return
	}
	orderId, err := strconv.Atoi(orderIdStr)
	if err != nil {
		utils.GinError(c, fmt.Sprintf(utils.OrderIdInvalid, orderIdStr), http.StatusBadRequest, nil)
		return
	}

	var order models.Order
	if err := order.GetOrderById(db.DB, orderId); err != nil {
		utils.GinError(c, err.Error(), http.StatusInternalServerError, err)
		return
	}

	utils.GinResponse(order, c, fmt.Sprintf(utils.OrderFetchSuccess, orderId), http.StatusOK)
}

func (db *Service) Checkout(c *gin.Context) {
	if err := constants.ValidateUserWithCtxUserId(c); err != nil {
		utils.GinError(c, err.Error(), http.StatusBadRequest, err)
		return
	}

	userId, err := constants.GetUserIdFromParams(c)
	if err != nil {
		utils.GinError(c, fmt.Sprintf(utils.UserNotFoundError, userId), http.StatusBadRequest, err)
		return
	}

	//how to fetch body
	var body payloads.RequestCart
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.GinError(c, utils.InvalidJSONBody, http.StatusBadRequest, err)
		return
	}

	//invoice list
	var invoiceList invoices.Invoice

	carts := body.Carts
	var cartIds []int
	subTotalPrice, totalDiscount := 0.0, 0.0
	var order models.Order

	for _, cart := range carts {
		cartId := int(cart["cart_id"].(float64))

		cart, err := fetchCartDetails(c, userId, cartId)
		if err != nil {
			utils.GinError(c, err.Error(), http.StatusBadRequest, err)
			return
		}
		if cart == nil {
			utils.GinError(c, fmt.Sprintf(utils.CartItemNotFoundError, cartId), http.StatusBadRequest, err)
			return
		}
		cartData := cart["data"].(map[string]interface{})

		productId := int(cartData["product_id"].(float64))
		product := fetchProductDetails(c, productId)
		if product == nil {
			utils.GinError(c, fmt.Sprintf(utils.ProductNotFoundError, productId), http.StatusBadRequest, err)
			return
		}
		productData := product["data"].(map[string]interface{})
		if ok := verifyQuantity(int(cartData["quantity"].(float64)), int(productData["quantity"].(float64))); !ok {
			utils.GinError(c, fmt.Sprintf(utils.CartOutOfStockError, productId), http.StatusBadRequest, nil)
			return
		}

		//calculating individual product price
		eachTotalPrice, eachDiscountAmt := calculatePrice(productData["price"].(float64), productData["discount"].(float64), int(cartData["quantity"].(float64)))
		subTotalPrice += eachTotalPrice
		totalDiscount += eachDiscountAmt

		cartIds = append(cartIds, cartId)

		//invoice data
		var invoiceItem invoices.InvoiceItem
		appendCartToInvoiceItem(&invoiceItem, cartData, productData, eachTotalPrice)
		invoiceList.InvoiceItemList = append(invoiceList.InvoiceItemList, invoiceItem)
	}

	//adding default tax(18%)
	taxAmt, subTotalPrice := calculateTotalWithTax(subTotalPrice)

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
	order.TaxAmount = taxAmt
	order.SubTotal = subTotalPrice
	order.TotalAmount = subTotalPrice + taxAmt
	order.DiscountAmount = totalDiscount
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

	userData, err := fetchUserDetails(c, userId)
	if err != nil {
		utils.LogError("from order services: Failed to fetch user data", map[string]interface{}{"error": err.Error()})
		return
	}

	//send data to invoice generator also this will send mail to user
	GenerateOrderInvoice(order, userData, invoiceList)
	go SendInvoice()

	utils.GinResponse(order, c, utils.OrderSuccessful, http.StatusOK)
}

func fetchUserDetails(c *gin.Context, userId int) (map[string]interface{}, error) {
	links := constants.MicroserviceLinks()
	userLinkById := links["userMSCallByIdLink"]

	userMicroserviceCall := fmt.Sprintf(userLinkById, userId)
	log.Println(userMicroserviceCall)
	req, err := http.NewRequest("GET", userMicroserviceCall, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to user microservices. error %w", err)
	}

	token := utils.GetTokenFromRequestUsingGin(c)
	req.Header.Add("Authorization", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to user microservices. error: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body. error: %w", err)
	}

	var userDetails map[string]interface{}
	err = json.Unmarshal(body, &userDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body. error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			if strings.Contains(userDetails["error"].(string), "not found") {
				return nil, fmt.Errorf(userDetails["error"].(string))
			}
		}
		return nil, fmt.Errorf("failed to fetch user details. error: %s", resp.Status)
	}

	userDetails = userDetails["data"].(map[string]interface{})
	return userDetails, nil
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
		return nil, fmt.Errorf(utils.ErrorCallingCartMicroservice) //change error message
	}

	var cart map[string]interface{}
	if err := utils.ParseJSON(body, &cart); err != nil {
		return nil, fmt.Errorf("failed to parse cart details") // add each error message into message file
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
		utils.GinErrorWithExtra(c, utils.ErrorProductMicroservices, http.StatusInternalServerError, err, "order.go")
		return nil
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	if _, err = io.Copy(&buf, resp.Body); err != nil {
		utils.GinErrorWithExtra(c, "utils.ErrorReadingResponseBody", http.StatusInternalServerError, err, "order.go")
		return nil
	}
	body := buf.Bytes()

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

func calculatePrice(price, discount float64, quantity int) (float64, float64) {
	p := price * float64(quantity)
	discountAmt := p * (discount / 100)
	total := p - discountAmt
	return total, discountAmt
}

func calculateTotalWithTax(totalAmt float64) (float64, float64) {
	taxPct := 18
	tax := float64(taxPct) / float64(100)
	taxAmt := totalAmt * tax
	return taxAmt, totalAmt
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

func appendCartToInvoiceItem(invoice *invoices.InvoiceItem, cart, product map[string]interface{}, itemTotalPrice float64) {
	invoice.Total = strconv.FormatFloat(itemTotalPrice, 'f', -1, 64)
	invoice.Quantity = strconv.FormatFloat(cart["quantity"].(float64), 'f', -1, 64)
	invoice.Price = strconv.FormatFloat(product["price"].(float64), 'f', -1, 64)
	invoice.Description = product["description"].(string)
	invoice.Item = product["product_name"].(string)
	invoice.DiscountedPrice = strconv.FormatFloat(product["discount"].(float64), 'f', -1, 64)
}
