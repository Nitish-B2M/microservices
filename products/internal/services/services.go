package services

import (
	"e-commerce-backend/products/internal/models"
	"e-commerce-backend/products/pkg/payloads"
	"e-commerce-backend/shared/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type FilterCriteria struct {
	PName       string   `json:"p_name"`
	MinPrice    float64  `json:"min_price"`
	MaxPrice    float64  `json:"max_price"`
	MinRating   float64  `json:"min_rating"`
	Category    string   `json:"category"`
	IsDeleted   bool     `json:"is_deleted"`
	MinQuantity int      `json:"min_quantity"`
	MaxQuantity int      `json:"max_quantity"`
	Tags        []string `json:"tags"`
}

const ProductAddQuanMethod = "add"
const ProductSubQuanMethod = "subtract"

type Service struct {
	DB *gorm.DB
}

func NewProduct(db *gorm.DB) *Service {
	return &Service{
		DB: db,
	}
}

type Product interface {
	GetProducts(w http.ResponseWriter, r *http.Request)
	GetProductById(w http.ResponseWriter, r *http.Request)
	CreateProduct(w http.ResponseWriter, r *http.Request)
	UpdateProduct(w http.ResponseWriter, r *http.Request)
	DeleteProduct(w http.ResponseWriter, r *http.Request)
	FilterProducts(w http.ResponseWriter, r *http.Request)
	UploadProductImageHandler(w http.ResponseWriter, r *http.Request)
}

func trackUpdatedProductFields(oldData models.Product, newData payloads.ProductRequest) map[string]interface{} {
	updatedFields := make(map[string]interface{})
	v := reflect.ValueOf(&newData).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		fieldName := field.Name
		fieldValue := v.Field(i)
		if fieldName == "Tags" {
			continue
		}
		if fieldValue.IsZero() || fieldName == "ID" {
			continue
		}

		oldFieldValue := reflect.ValueOf(&oldData).Elem().FieldByName(fieldName)

		if !reflect.DeepEqual(fieldValue.Interface(), oldFieldValue.Interface()) {
			updatedFields[fieldName] = fieldValue.Interface()

			reflect.ValueOf(&oldData).Elem().FieldByName(fieldName).Set(fieldValue)
		}
	}

	return updatedFields
}

func (db *Service) GetProducts(w http.ResponseWriter, r *http.Request) {
	products, err := models.GetProducts()
	if err != nil {
		utils.JsonError(w, utils.ProductNotFoundError, http.StatusNotFound, err[0])
	}

	utils.JsonResponse(products, w, utils.ProductsFetchedSuccessfully, http.StatusOK)
}

func (db *Service) GetProductById(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetIDFromPath(r)
	if err != nil {
		utils.JsonError(w, utils.InvalidProductIDError, http.StatusBadRequest, err)
		return
	}

	var product models.Product
	productResp, err := product.FetchProductResp(db.DB, id)
	if err != nil {
		utils.JsonError(w, utils.ProductNotFoundError, http.StatusNotFound, err)
		return
	}

	tags, errs := models.FetchProductTagsName(db.DB, product.ID)
	if len(errs) > 0 {
		errStr := utils.ErrorsToString(errs)
		utils.JsonError(w, utils.FailedToFetchTag, http.StatusNotFound, errors.New(errStr))
	}

	if tags == nil {
		productResp.Tags = []string{}
	} else {
		productResp.Tags = tags
	}

	utils.JsonResponse(productResp, w, fmt.Sprintf(utils.ProductFetchedSuccessfully, id), http.StatusOK)
}

func (db *Service) AddProduct(w http.ResponseWriter, r *http.Request) {
	if !utils.CheckRequestMethod(w, r, http.MethodPost) {
		return
	}

	var newProduct payloads.ProductRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&newProduct); err != nil {
		utils.JsonError(w, utils.InvalidRequestBody, http.StatusBadRequest, err)
		return
	}

	if newProduct.PName == "" || newProduct.Price == 0 {
		utils.JsonError(w, utils.InvalidProductDataError, http.StatusBadRequest, nil)
		return
	}

	var product models.Product
	if err := models.CopyStructIntoStruct(&newProduct, &product); err != nil {
		utils.JsonError(w, utils.InvalidProductDataError, http.StatusBadRequest, err)
		return
	}

	id, err := product.AddProduct(db.DB)
	if err != nil {
		utils.JsonError(w, utils.ProductCreationError, http.StatusInternalServerError, err)
		return
	}
	newProduct.ID = id

	tagIds, tagErrors := models.CheckAndCreateProductTags(db.DB, newProduct.Tags, 0)
	if len(tagErrors) > 0 {
		utils.JsonError(w, utils.TagCreationFailed, http.StatusBadRequest, nil)
	}

	if errs := models.AddTagToProduct(db.DB, tagIds, id); len(errs) > 0 {
		errStr := utils.ErrorsToString(errs)
		utils.JsonError(w, utils.FailedAddingTagToProduct, http.StatusNotFound, errors.New(errStr))
	}

	createdProduct, err := product.FetchProductResp(db.DB, id)
	if err != nil {
		utils.JsonError(w, utils.ProductNotFoundError, http.StatusNotFound, nil)
		return
	}

	utils.JsonResponse(createdProduct, w, fmt.Sprintf(utils.ProductCreatedSuccessfully, id), http.StatusCreated)
}

func (db *Service) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	if !utils.CheckRequestMethod(w, r, http.MethodPut) {
		return
	}

	id, err := utils.GetIDFromPath(r)
	if err != nil {
		utils.JsonError(w, utils.InvalidProductIDError, http.StatusBadRequest, err)
		return
	}

	var newProduct payloads.ProductRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&newProduct); err != nil {
		utils.JsonError(w, utils.InvalidRequestBody, http.StatusBadRequest, err)
		return
	}

	var existingP models.Product
	if err := existingP.CheckProductExistsById(db.DB, id); err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.ProductNotFoundError, id), http.StatusNotFound, err)
		return
	}

	updatedFields := trackUpdatedProductFields(existingP, newProduct)
	if len(updatedFields) == 0 {
		utils.JsonResponse(existingP, w, fmt.Sprintf(utils.UserNotModified, id), http.StatusNotModified)
		return
	}

	if err := existingP.UpdateProduct(db.DB, id, updatedFields); err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.JsonError(w, fmt.Sprintf(utils.ProductNotFoundError, id), http.StatusInternalServerError, err)
			return
		}
		utils.JsonError(w, fmt.Sprintf(utils.ProductUpdateError, id), http.StatusInternalServerError, err)
		return
	}

	tagIds, tagErrors := models.CheckAndCreateProductTags(db.DB, newProduct.Tags, id)
	if len(tagErrors) > 0 {
		utils.JsonError(w, utils.TagCreationFailed, http.StatusBadRequest, nil)
		return
	}

	if errs := models.UpdateTagToProduct(db.DB, tagIds, id); len(errs) > 0 {
		errStr := utils.ErrorsToString(errs)
		utils.JsonError(w, fmt.Sprintf(utils.ProductTagUpdateError, id), http.StatusBadRequest, errors.New(errStr))
		return
	}

	productResp, err := existingP.FetchProductResp(db.DB, id)
	if err != nil {
		utils.JsonError(w, utils.ProductNotFoundError, http.StatusNotFound, nil)
		return
	}

	utils.JsonResponse(productResp, w, fmt.Sprintf(utils.ProductUpdatedSuccessfully, id), http.StatusOK)
}

func (db *Service) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	if !utils.CheckRequestMethod(w, r, http.MethodDelete) {
		return
	}

	id, err := utils.GetIDFromPath(r)
	if err != nil {
		utils.JsonError(w, utils.InvalidProductIDError, http.StatusBadRequest, err)
		return
	}

	var product models.Product
	err = product.DeleteProduct(db.DB, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.JsonError(w, fmt.Sprintf(utils.ProductNotFoundError, id), http.StatusBadRequest, err)
			return
		}
		utils.JsonError(w, fmt.Sprintf(utils.ProductDeletionError, id), http.StatusInternalServerError, err)
		return
	}

	utils.JsonResponse(nil, w, fmt.Sprintf(utils.ProductDeletedSuccessfully, id), http.StatusOK)
}

func (db *Service) FilterProducts(w http.ResponseWriter, r *http.Request) {
	criteria := FilterCriteria{}

	query := r.URL.Query()
	fmt.Println("query:", query)

	FilterKey := []string{
		"p_name", "min_price", "max_price", "min_rating", "category",
		"is_deleted", "min_quantity", "max_quantity", "tags",
	}

	for k, v := range query {
		if HasKey(FilterKey, k) {
			if err := SetField(&criteria, k, v[0]); err != nil {
				utils.SimpleLog("error", fmt.Sprintf("Error setting field: %v", err))
			}
		}
	}

	resp := FilterOutProducts(criteria)
	utils.JsonResponse(resp, w, "filtered data", 0)
}

func (db *Service) UploadProductImageHandler(w http.ResponseWriter, r *http.Request) {
	productID := getProductID(r.URL.Path)
	if productID == 0 {
		utils.JsonError(w, utils.InvalidProductIDError, http.StatusBadRequest, nil)
		return
	}
	log.Println(productID)

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	err := r.ParseMultipartForm(10 << 20) // 10MB max memory
	if err != nil {
		utils.JsonError(w, "Unable to parse form", http.StatusBadRequest, err)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		utils.JsonError(w, utils.FileRetrieveFailed, http.StatusBadRequest, err)
		return
	}
	defer file.Close()

	uploadDir := "./uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, 0755)
	}

	filePath := filepath.Join(uploadDir, fmt.Sprintf("%d-%s", productID, handler.Filename))
	destFile, err := os.Create(filePath)
	if err != nil {
		utils.JsonError(w, utils.UnableToSaveFile, http.StatusInternalServerError, err)
		return
	}
	log.Println(filePath, *destFile)
	defer destFile.Close()

	_, err = io.Copy(destFile, file)
	if err != nil {
		utils.JsonError(w, utils.ErrorSavingFile, http.StatusInternalServerError, err)
		return
	}

}

func getProductID(path string) int {
	segments := splitPath(path)
	if len(segments) >= 2 && segments[0] == "products" {
		id, err := strconv.Atoi(segments[1])
		if err == nil {
			return id
		}
	}
	return 0
}

func splitPath(path string) []string {
	segments := strings.Split(path, "/")
	var cleanedSegments []string
	for _, segment := range segments {
		if segment != "" {
			cleanedSegments = append(cleanedSegments, segment)
		}
	}
	return cleanedSegments
}

func (db *Service) UpdateProductQuantityHandler(w http.ResponseWriter, r *http.Request) {
	productID := mux.Vars(r)["id"]
	if productID == "" {
		utils.JsonError(w, utils.ProductIDRequiredError, http.StatusBadRequest, nil)
		return
	}

	productIDInt, err := strconv.Atoi(productID)
	if err != nil {
		utils.JsonError(w, utils.InvalidProductIDError, http.StatusBadRequest, err)
		return
	}

	// Parse request body
	var req struct {
		Quantity int    `json:"quantity"`
		Method   string `json:"method"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JsonError(w, "Invalid request body", http.StatusBadRequest, err)
		return
	}

	// Update product quantity in the database
	var productResp payloads.ProductResponse
	if req.Method == ProductAddQuanMethod {
		productResp, err = AddQuantity(db.DB, productIDInt, req.Quantity)
	} else if req.Method == ProductSubQuanMethod {
		productResp, err = SubtractQuantity(db.DB, productIDInt, req.Quantity)
	} else {
		utils.JsonError(w, "Invalid method", http.StatusBadRequest, nil)
		return
	}

	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.ProductUnexpectedUpdateError), http.StatusInternalServerError, err)
		return
	}
	utils.JsonResponse(productResp, w, utils.ProductQuantityUpdated, http.StatusOK)
}

func SubtractQuantity(db *gorm.DB, productID int, quantity int) (payloads.ProductResponse, error) {
	var product models.Product
	var productResp payloads.ProductResponse

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

	return productResp, nil
}

func AddQuantity(db *gorm.DB, productID int, quantity int) (payloads.ProductResponse, error) {
	var product models.Product
	var productResp payloads.ProductResponse

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

	return productResp, nil
}

func (db *Service) GetProductByIdForCart(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetIDFromPath(r)
	if err != nil {
		utils.JsonError(w, utils.InvalidProductIDError, http.StatusBadRequest, err)
		return
	}

	var product models.Product
	productResp, err := product.FetchProductResp(db.DB, id)
	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.ProductNotFoundError, id), http.StatusNotFound, err)
		return
	}

	response := map[string]interface{}{
		"id":           productResp.ID,
		"product_name": productResp.PName,
		"description":  productResp.PDesc,
		"price":        productResp.Price,
		"quantity":     productResp.Quantity,
		"discount":     productResp.Discount,
	}

	utils.JsonResponse(response, w, fmt.Sprintf(utils.ProductFetchedSuccessfully, id), http.StatusOK)
}
