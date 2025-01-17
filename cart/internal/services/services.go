package services

import (
	"bytes"
	"e-commerce-backend/cart/internal/models"
	"e-commerce-backend/cart/pkg/constants"
	"e-commerce-backend/cart/pkg/payloads"
	"e-commerce-backend/shared/utils"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

const (
	CartQuantityAddMethod string = "add"
	CartQuantitySubMethod string = "subtract"
)

var cartMutex sync.Mutex

type Service struct {
	DB *gorm.DB
	mu sync.Mutex
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		DB: db,
	}
}

type CartInterface interface {
	GetCartItemByUserID(w http.ResponseWriter, r *http.Request) // user_id and cart_id
	AddToCart(w http.ResponseWriter, r *http.Request)
	UpdateCartQty(w http.ResponseWriter, r *http.Request)
	RemoveFromCart(w http.ResponseWriter, r *http.Request)
	Checkout(w http.ResponseWriter, r *http.Request)
	ValidateCart(cartItems []models.Cart) error
	CheckProductStock(productId int, quantity int) (bool, error)
	GetCartByCartId(w http.ResponseWriter, r *http.Request)
	ClearCart(w http.ResponseWriter, r *http.Request)
}

func verifyUserUsingIdAndCtxId(r *http.Request) (int, bool) {
	vars := mux.Vars(r)
	id := vars["id"]
	userID, _ := strconv.Atoi(id)
	ctxUserID := r.Context().Value("userID").(int)
	if userID != ctxUserID {
		return 0, false
	}
	return userID, true
}

func getCartIdFromParams(r *http.Request) (int, bool) {
	vars := mux.Vars(r)
	id := vars["cart_id"]
	cartId, _ := strconv.Atoi(id)
	return cartId, true
}

func fetchProductUsingMicroservices(productId int) (map[string]interface{}, error) {
	productServiceURL := fmt.Sprintf(constants.ProductMicroserviceGetProductCall, productId)
	resp, err := http.Get(productServiceURL)
	if err != nil {
		utils.LogError(fmt.Sprintf(constants.FailedToFetchProductDetails, productId), map[string]interface{}{"error": err.Error()})
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		utils.LogError(fmt.Sprintf(constants.ProductNotFound, productId), map[string]interface{}{})
		return nil, err
	}

	// Decode the product details
	var product map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		utils.LogError(fmt.Sprintf(constants.ErrorDecodingProductDetails, productId), map[string]interface{}{"error": err.Error()})
		return nil, err
	}

	if product == nil {
		utils.LogError(fmt.Sprintf(constants.ProductNotFound, productId), map[string]interface{}{"error": err.Error()})
		return nil, err
	}

	data, ok := product["data"].(map[string]interface{})
	if !ok {
		utils.LogError(constants.ProductDetailsNotEnough, map[string]interface{}{"error": err.Error()})
		return nil, err
	}

	return data, nil
}

func (db *Service) GetCartItemByUserID(w http.ResponseWriter, r *http.Request) {
	userId := utils.GetUserIdFromContext(r)

	var cart models.Cart
	cartItems, err := cart.GetCartByUserId(db.DB, userId)
	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.UserCartNotFoundError, userId), http.StatusNotFound, err)
		return
	}

	var cartResp payloads.CartResponse

	//fetching product details using product id
	for _, cartItem := range cartItems {
		product, err := fetchProductUsingMicroservices(cartItem.ProductId)
		if err != nil {
			return
		}

		price := product["price"].(float64) * float64(cartItem.Quantity)
		cartRespItem := payloads.CartItemResponse{
			Id:       cartItem.Id,
			Product:  product,
			Quantity: cartItem.Quantity,
			Price:    price,
		}

		cartResp.Items = append(cartResp.Items, cartRespItem)
	}

	utils.JsonResponse(cartResp, w, fmt.Sprintf(utils.CartFetchedSuccessfully, userId), http.StatusOK)
}

func (db *Service) AddToCart(w http.ResponseWriter, r *http.Request) {
	token := utils.GetTokenFromRequestHeader(r)
	if token == "" {
		utils.JsonError(w, utils.MissingAuthorizationHeader, http.StatusUnauthorized, nil)
		return
	}
	userId := utils.GetUserIdFromContext(r)

	var req payloads.CartRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utils.JsonError(w, utils.InvalidProductRequest, http.StatusBadRequest, err)
		return
	}
	//avoiding race condition
	cartMutex.Lock()
	defer cartMutex.Unlock()

	//response
	var cartResp payloads.CartResponse
	var errorMsg []error

	for _, item := range req.Items {
		product, err := fetchProductUsingMicroservices(item.ProductID)
		if err != nil || product == nil {
			errorMsg = append(errorMsg, fmt.Errorf(utils.ProductNotFoundError, item.ProductID))
			if err != nil {
				utils.LogError(fmt.Sprintf(utils.ProductNotFoundError, item.ProductID), map[string]interface{}{"error": err.Error()})
				continue
			}
			utils.LogError(fmt.Sprintf(utils.ProductNotFoundError, item.ProductID), map[string]interface{}{})
			continue
		}

		if int(product["quantity"].(float64)) < item.Quantity {
			utils.JsonError(w, constants.ProductQuantityOutOfStock, http.StatusBadRequest, err)
			continue
		}

		//insert into cart table logic int 4
		newCart := models.Cart{
			UserId:    userId,
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
		}

		cart, err := newCart.AddToCart(db.DB)
		if err != nil {
			utils.JsonError(w, utils.CartItemAdditionError, http.StatusInternalServerError, err)
			return
		}

		//cart response
		price := product["price"].(float64) * float64(cart.Quantity)
		cartRespItem := payloads.CartItemResponse{
			Id:          cart.Id,
			Quantity:    cart.Quantity,
			ReqQuantity: item.Quantity,
			Product:     product,
			Price:       price,
		}

		cartResp.Items = append(cartResp.Items, cartRespItem)
	}

	if len(errorMsg) > 0 {
		if len(req.Items)-len(errorMsg) > 0 {
			utils.JsonResponseWithError(cartResp, w, constants.SomeItemAddedToCart, http.StatusCreated, errorMsg)
			return
		} else {
			utils.JsonResponseWithError(cartResp, w, utils.CartItemAdditionError, http.StatusCreated, errorMsg)
			return
		}
	}
	utils.JsonResponse(cartResp, w, constants.ItemsAddedToCart, http.StatusCreated)
}

func (db *Service) UpdateCartQty(w http.ResponseWriter, r *http.Request) {
	token := utils.GetTokenFromRequestHeader(r)
	if token == "" {
		utils.JsonError(w, utils.MissingAuthorizationHeader, http.StatusUnauthorized, nil)
		return
	}
	//avoiding race condition
	cartMutex.Lock()
	defer cartMutex.Unlock()

	cartId, err := utils.GetIDFromPath(r)
	if err != nil {
		utils.JsonError(w, utils.InvalidCartRequest, http.StatusBadRequest, err)
		return
	}

	var req payloads.CartQtyUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JsonError(w, utils.InvalidCartRequest, http.StatusBadRequest, err)
		return
	}

	var existingCart models.Cart
	if err := existingCart.GetCartByCartId(db.DB, cartId); err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.CartItemNotFoundError, cartId), http.StatusNotFound, err)
		return
	}

	if req.Method != CartQuantityAddMethod && req.Method != CartQuantitySubMethod {
		utils.JsonError(w, utils.InvalidCartRequest, http.StatusBadRequest, fmt.Errorf(utils.InvalidCartRequest))
		return
	}
	var method string
	if req.Method == CartQuantityAddMethod {
		method = CartQuantitySubMethod
	} else if req.Method == CartQuantitySubMethod {
		method = CartQuantityAddMethod
	}

	if existingCart.Quantity == 0 && req.Method == CartQuantitySubMethod {
		utils.JsonError(w, utils.CartItemUpdateError, http.StatusInternalServerError, fmt.Errorf(utils.CartItemUpdateError))
		return
	}

	cartLength := 1
	cartLengthR, errs := updateProductQuantity(db.DB, token, method, []models.Cart{existingCart}, cartLength, req.Quantity, "update")
	if len(errs) > 0 {
		utils.JsonErrorWithExtra(w, utils.CartItemUpdateError, http.StatusInternalServerError, fmt.Errorf(utils.CartItemUpdateError), errs)
		return
	}
	if cartLengthR == 1 {
		utils.JsonError(w, utils.CartItemUpdateError, http.StatusInternalServerError, fmt.Errorf(utils.CartItemUpdateError))
		return
	}

	if req.Method == CartQuantityAddMethod {
		existingCart.Quantity += req.Quantity
	} else if req.Method == CartQuantitySubMethod {
		existingCart.Quantity -= req.Quantity
	}

	if err := existingCart.UpdateCart(db.DB); err != nil {
		utils.JsonError(w, utils.CartItemUpdateError, http.StatusInternalServerError, err)
		return
	}

	utils.JsonResponseWithExtra(existingCart, w, fmt.Sprintf(utils.CartItemUpdatedSuccessfully, cartId), http.StatusOK, "UpdateCartQty")
}

func (db *Service) DeleteCartByCartId(w http.ResponseWriter, r *http.Request) {
	token := utils.GetTokenFromRequestHeader(r)
	if token == "" {
		utils.JsonError(w, utils.MissingAuthorizationHeader, http.StatusUnauthorized, nil)
		return
	}

	//avoiding race condition
	cartMutex.Lock()
	defer cartMutex.Unlock()

	cartId, err := utils.GetIDFromPath(r)
	if err != nil {
		utils.JsonError(w, utils.InvalidCartRequest, http.StatusBadRequest, err)
		return
	}

	var req models.Cart
	if err := req.GetCartByCartId(db.DB, cartId); err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.CartItemNotFoundError, cartId), http.StatusNotFound, err)
		return
	}

	cartLength := 1
	cartLengthR, errs := updateProductQuantity(db.DB, token, CartQuantityAddMethod, []models.Cart{req}, cartLength, -1, "delete")
	if len(errs) > 0 {
		utils.JsonErrorWithExtra(w, utils.CartItemDeletionError, http.StatusInternalServerError, fmt.Errorf(utils.CartItemDeletionError), errs)
		return
	}
	if cartLengthR == 1 {
		utils.JsonError(w, utils.CartItemDeletionError, http.StatusInternalServerError, fmt.Errorf(utils.CartItemDeletionError))
		return
	}

	utils.JsonResponseWithExtra(req, w, fmt.Sprintf(utils.CartItemDeletedSuccessfully, cartId), http.StatusOK, "DeleteCartByCartId")
}

func (db *Service) Checkout(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		utils.JsonError(w, utils.MissingAuthorizationHeader, http.StatusUnauthorized, nil)
		return
	}
	token = strings.TrimPrefix(token, "Bearer ")
	//userId, ok := verifyUserUsingIdAndCtxId(r)
	//if !ok {
	//	utils.JsonError(w, utils.UnauthorizedError, http.StatusUnauthorized, fmt.Errorf(utils.UserNotFoundError, userId))
	//	return
	//}
	userId := utils.GetUserIdFromContext(r)

	var cart models.Cart
	carts, err := cart.GetCartByUserId(db.DB, userId)
	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.UserCartNotFoundError, userId), http.StatusNotFound, err)
		return
	}

	if len(carts) == 0 {
		utils.JsonError(w, "Cart is empty", http.StatusBadRequest, nil)
		return
	}

	err = db.ValidateCart(carts)
	if err != nil {
		utils.JsonError(w, "Cart validation failed", http.StatusBadRequest, err)
		return
	}
}

func (db *Service) ValidateCart(cartItems []models.Cart) error {
	// Loop through the cart items and validate each product
	for _, item := range cartItems {
		// Call Product Microservice to check product availability
		available, err := db.CheckProductStock(item.ProductId, item.Quantity)
		if err != nil {
			return fmt.Errorf("error checking stock for product %s: %v", item.ProductId, err)
		}
		if !available {
			return fmt.Errorf("product %s is out of stock", item.ProductId)
		}
	}

	// Optionally, validate other cart details, like pricing, discount codes, etc.
	return nil
}

// CheckProductStock makes an HTTP request to the Product Microservice to validate stock levels
func (db *Service) CheckProductStock(productId int, quantity int) (bool, error) {
	// Define the URL of the Product Microservice (for example)
	url := fmt.Sprintf(constants.ProductMicroserviceGetProductCall, productId)
	resp, err := http.Get(url)
	if err != nil {
		return false, fmt.Errorf("failed to contact Product Microservice: %v", err)
	}
	defer resp.Body.Close()

	// Check if the response is valid
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("received invalid response from Product Microservice: %v", resp.Status)
	}

	// Decode the response
	var product map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return false, fmt.Errorf("failed to decode response from Product Microservice: %v", err)
	}

	// Check if the product is in stock
	stock := int(product["quantity"].(float64))

	return stock >= quantity, nil
}

func (db *Service) GetCartByCartId(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		utils.JsonError(w, utils.MissingAuthorizationHeader, http.StatusUnauthorized, nil)
		return
	}
	token = strings.TrimPrefix(token, "Bearer ")
	userId := utils.GetUserIdFromContext(r)

	cartId, ok := getCartIdFromParams(r)
	if !ok {
		utils.JsonError(w, utils.CartIdNotProvided, http.StatusBadRequest, nil)
		return
	}
	log.Println("cartId:", cartId, "userId:", userId)

	var cart models.Cart
	if err := cart.GetCartItemByCartAndUserId(db.DB, userId, cartId); err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.CartItemNotFoundError, cartId), http.StatusNotFound, err)
		return
	}
	log.Println("GetCartByCartItemByCartId:", cart)
	utils.JsonResponse(cart, w, utils.CartFetchedSuccessfully, http.StatusOK)

}

func (db *Service) ClearCart(w http.ResponseWriter, r *http.Request) {
	token := utils.GetTokenFromRequestHeader(r)
	if token == "" {
		utils.JsonError(w, utils.MissingAuthorizationHeader, http.StatusUnauthorized, nil)
		return
	}
	userId := utils.GetUserIdFromContext(r)
	if userId == 0 {
		utils.JsonError(w, utils.UserIdNotFoundInCtx, http.StatusBadRequest, nil)
		return
	}
	var temp models.Cart
	carts, err := temp.GetCartByUserId(db.DB, userId)
	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.UserCartNotFoundError, userId), http.StatusNotFound, err)
		return
	}
	if len(carts) == 0 {
		utils.JsonError(w, "Cart is empty", http.StatusBadRequest, nil)
		return
	}

	//avoiding race condition
	cartMutex.Lock()
	defer cartMutex.Unlock()

	//update products
	cartLength := len(carts)
	cartLengthR, errs := updateProductQuantity(db.DB, token, CartQuantityAddMethod, carts, cartLength, -1, "clear")
	if len(errs) > 0 {
		if cartLength != cartLengthR || cartLength == cartLengthR {
			utils.JsonErrorWithExtra(
				w,
				"Failed to update product quantity", http.StatusInternalServerError, errors.New("failed to update product quantity"),
				map[string]interface{}{
					"active_cart_length": cartLengthR,
					"total_cart_length":  cartLength,
					"error_messages":     errs,
				},
			)
			return
		}
	}

	utils.JsonResponse(nil, w, utils.CartClearedSuccessfully, http.StatusOK)
}

func updateProductQuantity(db *gorm.DB, token, method string, carts []models.Cart, cartLength, quantity int, action string) (int, []error) {
	links := constants.MicroserviceLinks()
	productQtyUpdateLink := links["updateProductQuantityMSCallLink"]
	var errs []error
	if cartLength <= 0 {
		return cartLength, errs
	}

	for _, item := range carts {
		if quantity == -1 {
			quantity = item.Quantity
		}
		bodyToSend := map[string]interface{}{
			"method":   method,
			"quantity": quantity,
		}
		jsonData, err := json.Marshal(bodyToSend)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to marshal body: %w", err))
			return cartLength, errs
		}
		productMicroserviceCall := fmt.Sprintf(productQtyUpdateLink, item.ProductId)
		req, err := http.NewRequest("POST", productMicroserviceCall, bytes.NewBuffer(jsonData))
		if err != nil {
			errs = append(errs, err)
			return cartLength, errs
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", token)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			errs = append(errs, err)
			return cartLength, errs
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errs = append(errs, fmt.Errorf("unexpected status code: %d", resp.StatusCode))
			return cartLength, errs
		}

		var buf bytes.Buffer
		if _, err := io.Copy(&buf, resp.Body); err != nil {
			errs = append(errs, fmt.Errorf("failed to read response body: %w", err))
			return cartLength, errs
		}
		body := buf.Bytes()
		if strings.Contains(string(body), "error") {
			errs = append(errs, fmt.Errorf("failed to update product quantity: %s", resp.Status))
			return cartLength, errs
		}

		if strings.EqualFold(action, "clear") || strings.EqualFold(action, "delete") {
			if err := item.DeleteCartItem(db); err != nil {
				errs = append(errs, fmt.Errorf("failed to delete cart item: %w", err))
				return cartLength, errs
			}
		}
		cartLength--
	}
	return cartLength, errs
}
