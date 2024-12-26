package product

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

	product, err := models.FetchProductById(id)
	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.ProductNotFoundError, id), http.StatusNotFound, err)
		return
	}

	utils.JsonResponse(product, w, fmt.Sprintf(utils.ProductFetchedSuccessfully, id), http.StatusOK)
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

	id, err := models.AddProduct(newProduct)
	if err != nil {
		utils.JsonError(w, utils.ProductCreationError, http.StatusInternalServerError, err)
		return
	}

	createdProduct, err := models.FetchProductById(id)
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

	oldProductDB, ok := models.CheckProductExistsById(id)
	if !ok {
		utils.JsonError(w, fmt.Sprintf(utils.ProductNotFoundError, id), http.StatusNotFound, nil)
		return
	}

	updatedFields := trackUpdatedProductFields(*oldProductDB, newProduct)
	if len(updatedFields) == 0 {
		utils.JsonResponse(*oldProductDB, w, fmt.Sprintf(utils.UserNotModified, id), http.StatusNotModified)
		return
	}

	err = models.UpdateProduct(id, newProduct, updatedFields)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.JsonError(w, fmt.Sprintf(utils.ProductNotFoundError, id), http.StatusInternalServerError, err)
			return
		}
		utils.JsonError(w, fmt.Sprintf(utils.ProductUpdateError, id), http.StatusInternalServerError, err)
		return
	}

	updatedProduct, err := models.FetchProductById(id)
	if err != nil {
		utils.JsonError(w, utils.ProductNotFoundError, http.StatusNotFound, nil)
		return
	}

	utils.JsonResponse(updatedProduct, w, fmt.Sprintf(utils.ProductUpdatedSuccessfully, updatedProduct.ID), http.StatusOK)
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

	err = models.DeleteProduct(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.JsonError(w, fmt.Sprintf(utils.ProductNotFoundError, id), http.StatusInternalServerError, err)
			return
		}
		utils.JsonError(w, fmt.Sprintf(utils.ProductDeletionError, id), http.StatusInternalServerError, err)
		return
	}

	deletedProduct, err := models.FetchProductById(id)
	if err != nil {
		utils.JsonError(w, utils.ProductNotFoundError, http.StatusNotFound, nil)
		return
	}

	utils.JsonResponse(deletedProduct, w, fmt.Sprintf(utils.ProductDeletedSuccessfully, deletedProduct.ID), http.StatusOK)
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
		http.Error(w, utils.InvalidProductIDError, http.StatusBadRequest)
		return
	}
	log.Println(productID)

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	err := r.ParseMultipartForm(10 << 20) // 10MB max memory
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, utils.FileRetrieveFailed, http.StatusBadRequest)
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
		http.Error(w, utils.UnableToSaveFile, http.StatusInternalServerError)
		return
	}
	log.Println(filePath, *destFile)
	defer destFile.Close()

	_, err = io.Copy(destFile, file)
	if err != nil {
		http.Error(w, utils.ErrorSavingFile, http.StatusInternalServerError)
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

func (db *Service) AddToCart(w http.ResponseWriter, r *http.Request) {
	productID := getProductID(r.URL.Path)
	if productID == 0 {
		utils.JsonError(w, utils.InvalidProductIDError, http.StatusBadRequest, errors.New("product ID is required"))
		return
	}

	cartUrl := "http://localhost:8082/cart/add"
	_ = cartUrl
	//req, err := http.NewRequest("POST", cartUrl, )
}
