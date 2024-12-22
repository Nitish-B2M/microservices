package product

import (
	"encoding/json"
	"fmt"
	"microservices/pkg/shared/models"
	"microservices/pkg/shared/payloads"
	"microservices/pkg/shared/utils"
	"net/http"
	"strings"
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

func ProductHandler() {
	http.HandleFunc("/products", getProducts)
	http.HandleFunc("/products/{id}", getProductById)
	http.HandleFunc("/products/add", addProduct)
	http.HandleFunc("/products/update/{id}", updateProduct)
	http.HandleFunc("/products/delete/{id}", deleteProduct)
	http.HandleFunc("/products/filter", filterProducts)
}

func getProducts(w http.ResponseWriter, r *http.Request) {
	products, err := models.GetProducts()
	if err != nil {
		utils.JsonError(w, utils.ProductNotFoundError, http.StatusNotFound, err[0])
	}

	utils.JsonResponse(products, w, utils.ProductsFetchedSuccessfully, http.StatusOK)
}

func getProductById(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetProductIdFromPath(r)
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

func addProduct(w http.ResponseWriter, r *http.Request) {
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

func updateProduct(w http.ResponseWriter, r *http.Request) {
	if !utils.CheckRequestMethod(w, r, http.MethodPut) {
		return
	}

	id, err := utils.GetProductIdFromPath(r)
	if err != nil {
		utils.JsonError(w, utils.InvalidProductIDError, http.StatusBadRequest, err)
		return
	}

	var oldProduct payloads.ProductRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&oldProduct); err != nil {
		utils.JsonError(w, utils.InvalidRequestBody, http.StatusBadRequest, err)
		return
	}

	err = models.UpdateProduct(id, oldProduct)
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

func deleteProduct(w http.ResponseWriter, r *http.Request) {
	if !utils.CheckRequestMethod(w, r, http.MethodDelete) {
		return
	}

	id, err := utils.GetProductIdFromPath(r)
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

func filterProducts(w http.ResponseWriter, r *http.Request) {
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
