package controller

import (
	"e-commerce-backend/products/internal/services"
	"e-commerce-backend/shared/utils"
	"errors"
	"net/http"
)

type ProductController struct {
	Service *services.Services
}

func NewProductController(svc *services.Services) *ProductController {
	return &ProductController{Service: svc}
}

func (h *ProductController) FetchProductByID(w http.ResponseWriter, r *http.Request) {
	pid := utils.GetIDFromPathString(r)
	if pid == "" {
		utils.ErrorResponseFunc(w, "Invalid product ID", http.StatusBadRequest, errors.New("invalid product ID"))
		return
	}

	product, err := h.Service.ProductRepo.GetByID(pid)
	if err != nil {
		utils.ErrorResponseFunc(w, "Product not found", http.StatusNotFound, err)
		return
	}

	utils.SuccessResponseFunc(w, utils.ProductFetchedSuccessfully, product, http.StatusOK)
}

func (h *ProductController) FetchProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.Service.ProductRepo.GetAll()
	if err != nil {
		utils.ErrorResponseFunc(w, "Products not found", http.StatusNotFound, err)
		return
	}

	utils.SuccessResponseFunc(w, utils.ProductsFetchedSuccessfully, products, http.StatusOK)
}
