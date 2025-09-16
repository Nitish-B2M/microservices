package services

import (
	"net/http"

	"errors"

	"gorm.io/gorm"

	"e-commerce-backend/products/internal/models"
	"e-commerce-backend/products/pkg/payloads"
	"e-commerce-backend/shared/utils"
)

type ProductService struct {
	db *gorm.DB
}

func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{
		db: db,
	}
}

func (db *ProductService) GetProduct(w http.ResponseWriter, r *http.Request) {
	var products []models.Product
	if err := db.db.
		Preload("Category.Entity").
		Preload("Brand.Entity").
		Preload("Tags.Entity").
		Preload("Images").
		Preload("Variants").
		Preload("Attributes").
		Where("is_deleted = ?", false).
		Find(&products).Error; err != nil {
		utils.ErrorResponseFunc(w, utils.ProductsNotFoundError, http.StatusNotFound, err)
	}

	var response []payloads.ProductResponse
	for _, product := range products {
		if len(product.Images) > 0 {
			for i := range product.Images {
				product.Images[i].URL = models.CreateImageURL(product.Images[i].URL)
			}
		}
		res := models.CopyProductToProductResponse(product)
		response = append(response, res)
	}
	utils.SuccessResponseFunc(w, utils.ProductsFetchedSuccessfully, response, http.StatusOK)
}

func (db *ProductService) GetProductByID(w http.ResponseWriter, r *http.Request) {
	if !utils.CheckRequestMethod(w, r, http.MethodGet) {
		return
	}

	id := utils.GetIDFromPathString(r)
	if id == "" {
		utils.ErrorResponseFunc(w, utils.InvalidProductIDError, http.StatusBadRequest, errors.New(utils.InvalidProductIDError))
		return
	}

	var product models.Product
	if err := db.db.
		Preload("Category.Entity").
		Preload("Brand.Entity").
		Preload("Tags.Entity").
		Preload("Images").
		Preload("Variants").
		Preload("Attributes").
		First(&product, "id = ?", id).Error; err != nil {
		utils.ErrorResponseFunc(w, utils.ProductNotFoundError, http.StatusNotFound, err)
	}

	if len(product.Images) > 0 {
		for i := range product.Images {
			product.Images[i].URL = models.CreateImageURL(product.Images[i].URL)
		}
	}
	utils.SuccessResponseFunc(w, utils.ProductFetchedSuccessfully, product, http.StatusOK)
}

func (db *ProductService) GetProductByIDForCart(w http.ResponseWriter, r *http.Request, id string) {
	var product models.Product
	if err := db.db.
		Preload("Category.Entity").
		Preload("Brand.Entity").
		Preload("Tags.Entity").
		Preload("Images").
		Preload("Variants").
		Preload("Attributes").
		First(&product, "id = ?", id).Error; err != nil {
		utils.ErrorResponseFunc(w, utils.ProductNotFoundError, http.StatusNotFound, err)
	}

	if len(product.Images) > 0 {
		for i := range product.Images {
			product.Images[i].URL = models.CreateImageURL(product.Images[i].URL)
		}
	}
	utils.SuccessResponseFunc(w, utils.ProductFetchedSuccessfully, product, http.StatusOK)
}
