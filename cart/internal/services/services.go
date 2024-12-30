package services

import (
	"e-commerce-backend/cart/internal/models"
	"e-commerce-backend/cart/pkg/constants"
	"e-commerce-backend/cart/pkg/payloads"
	"e-commerce-backend/shared/utils"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

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
	RemoveFromCart(w http.ResponseWriter, r *http.Request)
	Checkout(w http.ResponseWriter, r *http.Request)
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
	userId, ok := verifyUserUsingIdAndCtxId(r)
	if !ok {
		utils.JsonError(w, utils.UnauthorizedError, http.StatusUnauthorized, fmt.Errorf(utils.UserNotFoundError, userId))
		return
	}

	var cart models.Cart
	cartItems, err := cart.GetCartByUserId(db.DB, userId)
	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.CartNotFoundError, userId), http.StatusNotFound, err)
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
	token := r.Header.Get("Authorization")
	if token == "" {
		utils.JsonError(w, utils.MissingAuthorizationHeader, http.StatusUnauthorized, nil)
		return
	}
	token = strings.TrimPrefix(token, "Bearer ")

	userId, ok := verifyUserUsingIdAndCtxId(r)
	if !ok {
		utils.JsonError(w, utils.UnauthorizedError, http.StatusUnauthorized, fmt.Errorf(utils.UserNotFoundError, userId))
		return
	}

	var req payloads.CartRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utils.JsonError(w, utils.InvalidProductRequest, http.StatusBadRequest, err)
		return
	}

	//response
	var cartResp payloads.CartResponse
	var errorMsg []error

	// go routine
	go StartProductUpdateQuantityWorker(token)
	db.mu.Lock()
	defer db.mu.Unlock()

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
			return
		}
		log.Printf(constants.AddingProductToCart, item.ProductID, item.Quantity)

		//insert into cart table logic int 4
		cartData := models.Cart{
			UserId:    userId,
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
		}

		cart, err := cartData.AddToCart(db.DB)
		if err != nil {
			utils.JsonError(w, utils.CartItemAdditionError, http.StatusInternalServerError, err)
			return
		}

		//cart response
		price := product["price"].(float64) * float64(cart.Quantity)
		product["quantity"] = int(product["quantity"].(float64)) - item.Quantity
		cartRespItem := payloads.CartItemResponse{
			Id:          cart.Id,
			Quantity:    cart.Quantity,
			ReqQuantity: item.Quantity,
			Product:     product,
			Price:       price,
		}
		// initialize channel with task variable(this will update quantity of product in product table)
		task := UpdateTask{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Method:    "Subtract",
		}

		UpdateChannel <- task
		log.Printf(constants.TaskForProductPushedToChannel, item.ProductID)
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

func (db *Service) RemoveFromCart(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		utils.JsonError(w, utils.MissingAuthorizationHeader, http.StatusUnauthorized, nil)
		return
	}
	token = strings.TrimPrefix(token, "Bearer ")

	userId, ok := verifyUserUsingIdAndCtxId(r)
	if !ok {
		utils.JsonError(w, utils.UnauthorizedError, http.StatusUnauthorized, fmt.Errorf(utils.UserNotFoundError, userId))
		return
	}

	var req payloads.CartRemoveRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, utils.InvalidProductRequest, http.StatusBadRequest)
		return
	}

	//response
	var cartResp payloads.CartResponse

	go StartProductUpdateQuantityWorker(token)

	for _, item := range req.Items {
		var cart models.Cart
		if err := cart.GetCartByCartId(db.DB, item.Id); err != nil {
			utils.JsonErrorWithExtra(w, utils.CartItemDeletionError, http.StatusInternalServerError, err)
			return
		}

		product, err := fetchProductUsingMicroservices(item.ProductID)
		if err != nil {
			return
		}

		//insert into cart table logic
		if err := cart.RemoveItemFromCart(db.DB, item.Id, item.Quantity); err != nil {
			utils.JsonErrorWithExtra(w, utils.CartItemDeletionError, http.StatusInternalServerError, err)
			return
		}

		//cart response
		product["quantity"] = int(product["quantity"].(float64)) + item.Quantity
		price := product["price"].(float64) * float64(cart.Quantity)
		cartRespItem := payloads.CartItemResponse{
			Id:       item.Id,
			Quantity: cart.Quantity,
			Product:  nil,
			Price:    price,
		}

		// initialize channel with task variable(this will update quantity of product in product table)
		task := UpdateTask{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Method:    "Add",
		}
		UpdateChannel <- task
		log.Printf(constants.TaskForProductPushedToChannel, item.ProductID)
		cartResp.Items = append(cartResp.Items, cartRespItem)
	}
	utils.JsonResponse(cartResp, w, constants.ItemsRemoveFromCart, http.StatusCreated)
}

func (db *Service) Checkout(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		utils.JsonError(w, utils.MissingAuthorizationHeader, http.StatusUnauthorized, nil)
		return
	}
	token = strings.TrimPrefix(token, "Bearer ")
	userId, ok := verifyUserUsingIdAndCtxId(r)
	if !ok {
		utils.JsonError(w, utils.UnauthorizedError, http.StatusUnauthorized, fmt.Errorf(utils.UserNotFoundError, userId))
		return
	}

}
