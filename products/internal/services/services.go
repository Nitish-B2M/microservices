// Package services for product-related operations.
package services

import (
	"e-commerce-backend/products/internal/models"
	"e-commerce-backend/products/pkg/payloads"
	"e-commerce-backend/shared/utils"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"

	"gorm.io/gorm"
)

const ProductAddQuanMethod = "add"
const ProductSubQuanMethod = "subtract"

var productMutex sync.Mutex

type Service struct {
	DB *gorm.DB
}

func NewProduct(db *gorm.DB) *Service {
	return &Service{
		DB: db,
	}
}

type Product interface {
	// old
	GetProducts(w http.ResponseWriter, r *http.Request)
	GetProductById(w http.ResponseWriter, r *http.Request)
	FetchCategories(w http.ResponseWriter, r *http.Request)
	AddProduct(w http.ResponseWriter, r *http.Request)
	UpdateProduct(w http.ResponseWriter, r *http.Request)
	DeleteProduct(w http.ResponseWriter, r *http.Request)
	UploadProductImageHandler(w http.ResponseWriter, r *http.Request)
	UpdateProductQuantityHandler(w http.ResponseWriter, r *http.Request)
	GetProductByIdForCart(w http.ResponseWriter, r *http.Request)
}

func (db *Service) GetProducts(w http.ResponseWriter, r *http.Request) {
	products, err := models.GetProducts(db.DB)
	if err != nil {
		utils.ErrorResponseFunc(w, utils.ProductsNotFoundError, http.StatusNotFound, err)
	}

	utils.SuccessResponseFunc(w, utils.ProductsFetchedSuccessfully, products, http.StatusOK)
}

func (db *Service) GetProductsForSeller(w http.ResponseWriter, r *http.Request) {
	products, err := models.GetProductsForSeller(db.DB)
	if err != nil {
		utils.ErrorResponseFunc(w, utils.ProductNotFoundError, http.StatusNotFound, err)
	}

	utils.SuccessResponseFunc(w, utils.ProductsFetchedSuccessfully, products, http.StatusOK)
}

func (db *Service) GetProductByID(w http.ResponseWriter, r *http.Request) {
	if !utils.CheckRequestMethod(w, r, http.MethodGet) {
		return
	}

	PID := utils.GetIDFromPathString(r)
	if PID == "" {
		utils.ErrorResponseFunc(w, utils.InvalidProductIDError, http.StatusBadRequest, errors.New(utils.InvalidProductIDError))
		return
	}

	var product models.Product
	products, err := product.GetProductById(db.DB, PID)
	if err != nil {
		if strings.Contains(err.Error(), utils.ProductNotFoundError) {
			utils.ErrorResponseFunc(w, utils.ProductNotFoundError, http.StatusNotFound, err)
			return
		}
		utils.ErrorResponseFunc(w, utils.ProductRetrievalError, http.StatusInternalServerError, err)
		return
	}
	utils.SuccessResponseFunc(w, utils.ProductFetchedSuccessfully, products, http.StatusOK)
}

func (db *Service) FetchCategories(w http.ResponseWriter, r *http.Request) {
	var p models.Product
	categories, err := p.FetchProductCategories(db.DB)
	if err != nil {
		utils.ErrorResponseFunc(w, utils.CategoryNotFoundError, http.StatusNotFound, err)
		return
	}

	utils.SuccessResponseFunc(w, utils.CategoriesFetchedSuccessfully, categories, http.StatusOK)
}

func (db *Service) AddProduct(w http.ResponseWriter, r *http.Request) {
	if !utils.CheckRequestMethod(w, r, http.MethodPost) {
		return
	}

	// Extract seller ID from context/token
	sellerID := utils.GetStringUserIDFromContext(r)
	if sellerID == "" {
		utils.ErrorResponseFunc(w, "Seller ID is required", http.StatusBadRequest, errors.New("missing seller id"))
		return
	}

	var newProduct payloads.ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&newProduct); err != nil {
		utils.ErrorResponseFunc(w, utils.InvalidRequestBody, http.StatusBadRequest, err)
		return
	}

	// Validate the product data
	if ok, err := utils.ValidateStructUsingValidators(newProduct); !ok {
		errStr := strings.Join(err, ", ")
		utils.ErrorResponseFunc(w, utils.InvalidProductDataError, http.StatusBadRequest, errors.New(errStr))
		return
	}

	// Generate UUID for product
	if newProduct.ID == "" {
		newProduct.ID = uuid.New().String()
	}
	newProduct.SellerID = sellerID

	// Generate slug URL if not provided
	if newProduct.SlugURL == "" {
		newProduct.SlugURL = utils.GenerateSlug(newProduct.Name)
	}

	// Generate SKU if not provided
	if newProduct.SKU == "" {
		newProduct.SKU = utils.GenerateSKU(newProduct.Name)
	}

	if len(newProduct.Images) > 0 {
		for i := range newProduct.Images {
			if i == 0 {
				newProduct.Images[i].IsMain = true
			}
			if newProduct.Images[i].ID == "" {
				newProduct.Images[i].ID = uuid.New().String()
			}
			newProduct.Images[i].PID = newProduct.ID
			newProduct.Images[i].SortOrder = i + 1
			if newProduct.Images[i].URL == "" {
				filePath, err := models.UploadBase64Image(newProduct.Images[i].Image, newProduct.Images[i].PID, newProduct.Images[i].ID)
				if err != nil {
					utils.ErrorResponseFunc(w, utils.ProductCreationError, http.StatusInternalServerError, err)
					return
				}
				newProduct.Images[i].URL = filePath
			}
		}
	}

	category := ""
	if newProduct.Category != "" {
		category = newProduct.Category
	}

	brand := ""
	if newProduct.Brand != "" {
		brand = newProduct.Brand
	}

	product := models.CopyProductRequestToProduct(newProduct)
	// Save the new product
	createdProduct, err := product.AddProduct(db.DB, category, brand, newProduct.Tags)
	if err != nil {
		utils.ErrorResponseFunc(w, utils.ProductCreationError, http.StatusInternalServerError, err)
		return
	}

	utils.SuccessResponseFunc(w, fmt.Sprintf(utils.ProductCreatedSuccessfully, createdProduct.ID), createdProduct, http.StatusCreated)
}

func (db *Service) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	if !utils.CheckRequestMethod(w, r, http.MethodPut) {
		return
	}

	id := utils.GetIDFromPathString(r)
	if id == "" {
		utils.ErrorResponseFunc(w, utils.InvalidProductIDError, http.StatusBadRequest, errors.New(utils.InvalidProductIDError))
		return
	}

	var updatedProduct payloads.ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&updatedProduct); err != nil {
		utils.ErrorResponseFunc(w, utils.InvalidRequestBody, http.StatusBadRequest, err)
		return
	}

	if ok, err := utils.ValidateStructUsingValidators(updatedProduct); !ok {
		errStr := strings.Join(err, ", ")
		utils.ErrorResponseFunc(w, utils.InvalidProductDataError, http.StatusBadRequest, errors.New(errStr))
		return
	}

	var existingP models.Product
	if err := existingP.CheckProductExistsById(db.DB, id); err != nil {
		utils.ErrorResponseFunc(w, fmt.Sprintf(utils.ProductNotFoundError, id), http.StatusNotFound, err)
		return
	}

	if err := existingP.UpdateProduct(db.DB, id, updatedProduct); err != nil {
		utils.ErrorResponseFunc(w, fmt.Sprintf(utils.ProductUpdateError, id), http.StatusInternalServerError, err)
		return
	}

	productResp, err := existingP.FetchProductResp(db.DB, id)
	if err != nil {
		utils.ErrorResponseFunc(w, utils.ProductNotFoundError, http.StatusNotFound, err)
		return
	}

	utils.SuccessResponseFunc(w, fmt.Sprintf(utils.ProductUpdatedSuccessfully, id), productResp, http.StatusOK)
}

func (db *Service) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	if !utils.CheckRequestMethod(w, r, http.MethodDelete) {
		return
	}

	id := utils.GetIDFromPathString(r)
	if id == "" || strings.Trim(id, " ") == "" {
		utils.ErrorResponseFunc(w, utils.InvalidProductIDError, http.StatusBadRequest, errors.New(utils.InvalidProductIDError))
		return
	}

	var product models.Product
	err := product.DeleteProduct(db.DB, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.ErrorResponseFunc(w, fmt.Sprintf(utils.ProductNotFoundError, id), http.StatusBadRequest, err)
			return
		}
		utils.ErrorResponseFunc(w, fmt.Sprintf(utils.ProductDeletionError, id), http.StatusInternalServerError, err)
		return
	}

	utils.SuccessResponseFunc(w, fmt.Sprintf(utils.ProductDeletedSuccessfully, id), nil, http.StatusOK)
}

func (db *Service) UploadProductImageHandler(w http.ResponseWriter, r *http.Request) {
	productID := utils.GetIDFromPathString(r)
	if productID == "" {
		utils.ErrorResponseFunc(w, utils.InvalidProductIDError, http.StatusBadRequest, errors.New(utils.InvalidProductIDError))
		return
	}
	log.Println(productID)

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	err := r.ParseMultipartForm(10 << 20) // 10MB max memory
	if err != nil {
		utils.ErrorResponseFunc(w, "Unable to parse form", http.StatusBadRequest, err)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		utils.ErrorResponseFunc(w, utils.FileRetrieveFailed, http.StatusBadRequest, err)
		return
	}
	defer file.Close()

	uploadDir := "./uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, 0755)
	}

	filePath := filepath.Join(uploadDir, fmt.Sprintf("%s-%s", productID, handler.Filename))
	destFile, err := os.Create(filePath)
	if err != nil {
		utils.ErrorResponseFunc(w, utils.UnableToSaveFile, http.StatusInternalServerError, err)
		return
	}
	log.Println(filePath, *destFile)
	defer destFile.Close()

	_, err = io.Copy(destFile, file)
	if err != nil {
		utils.ErrorResponseFunc(w, utils.ErrorSavingFile, http.StatusInternalServerError, err)
		return
	}

}

func (db *Service) UpdateProductQuantityHandler(w http.ResponseWriter, r *http.Request) {
	id := utils.GetIDFromPathString(r)
	if id == "" {
		utils.ErrorResponseFunc(w, utils.InvalidProductIDError, http.StatusBadRequest, errors.New(utils.InvalidProductIDError))
		return
	}

	// Parse request body
	var req struct {
		Quantity int    `json:"quantity"`
		Method   string `json:"method"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponseFunc(w, "Invalid request body", http.StatusBadRequest, err)
		return
	}

	// Update product quantity in the database
	var productResp payloads.ProductQtyUpdateResponse
	var err error
	if req.Method == ProductAddQuanMethod {
		productResp, err = AddQuantity(db.DB, id, req.Quantity)
	} else if req.Method == ProductSubQuanMethod {
		productResp, err = SubtractQuantity(db.DB, id, req.Quantity)
	} else {
		utils.ErrorResponseFunc(w, "Invalid method", http.StatusBadRequest, errors.New(utils.InvalidRequestMethod))
		return
	}

	if err != nil {
		utils.ErrorResponseFunc(w, err.Error(), http.StatusInternalServerError, err)
		return
	}
	utils.SuccessResponseFunc(w, utils.ProductQuantityUpdated, productResp, http.StatusOK)
}

func SubtractQuantity(db *gorm.DB, productID string, quantity int) (payloads.ProductQtyUpdateResponse, error) {
	var product models.Product
	var productResp payloads.ProductQtyUpdateResponse

	if err := product.CheckProductExistsById(db, productID); err != nil {
		return productResp, err
	}

	if product.Quantity < quantity {
		return productResp, fmt.Errorf(utils.ProductOutOfStockError, product.ID)
	}

	product.Quantity -= quantity
	// Save the updated product back to the database
	if err := product.UpdateProductQuantity(db); err != nil {
		return productResp, err
	}

	if err := models.CopyStructIntoStruct(&product, &productResp); err != nil {
		return productResp, err
	}
	log.Println(productResp, product, "from subtractQuantity")

	return productResp, nil
}

func AddQuantity(db *gorm.DB, productID string, quantity int) (payloads.ProductQtyUpdateResponse, error) {
	var product models.Product
	var productResp payloads.ProductQtyUpdateResponse

	if err := product.CheckProductExistsById(db, productID); err != nil {
		return productResp, err
	}
	product.Quantity += quantity

	// Save the updated product back to the database
	if err := product.UpdateProductQuantity(db); err != nil {
		return productResp, err
	}

	if err := models.CopyStructIntoStruct(&product, &productResp); err != nil {
		return productResp, err
	}

	log.Println(productResp, product, "from addQuantity")
	return productResp, nil
}

func (db *Service) GetProductByIDForCart(w http.ResponseWriter, r *http.Request) {
	id := utils.GetIDFromPathString(r)
	if id == "" {
		utils.ErrorResponseFunc(w, utils.InvalidProductIDError, http.StatusBadRequest, errors.New(utils.InvalidProductIDError))
		return
	}

	var product models.Product
	productResp, err := product.FetchProductResp(db.DB, id)
	if err != nil {
		utils.ErrorResponseFunc(w, fmt.Sprintf(utils.ProductNotFoundError, id), http.StatusNotFound, err)
		return
	}

	response := map[string]interface{}{
		"id":          productResp.ID,
		"name":        productResp.Name,
		"sku":         productResp.SKU,
		"short_desc":  productResp.ShortDesc,
		"description": productResp.Description,
		"price":       productResp.Price,
		"quantity":    productResp.Quantity,
		"in_stock":    productResp.InStock,
		"discount":    productResp.Discount,
		"tax_rate":    productResp.Tax,
	}

	utils.SuccessResponseFunc(w, fmt.Sprintf(utils.ProductFetchedSuccessfully, id), response, http.StatusOK)
}
