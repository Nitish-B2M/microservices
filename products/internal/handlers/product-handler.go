package handlers

import (
	"e-commerce-backend/products/dbs"
	"e-commerce-backend/products/internal/services/product"
	"e-commerce-backend/shared/middlewares"
	"net/http"

	"github.com/gorilla/mux"
)

func ProductHandler(r *mux.Router) {
	productService := product.NewProduct(dbs.DB)

	// implementing mux router
	r.Handle("/product", http.HandlerFunc(productService.GetProducts)).Methods(http.MethodGet)
	r.Handle("/product/{id}", http.HandlerFunc(productService.GetProductById)).Methods(http.MethodGet)
	r.Handle("/product/{id}/cart", http.HandlerFunc(productService.GetProductByIdForCart)).Methods(http.MethodGet)
	r.Handle("/product/filter/search", http.HandlerFunc(productService.FilterProducts)).Methods(http.MethodGet)
	r.Handle("/product/add", middlewares.AuthMiddleware(http.HandlerFunc(productService.AddProduct))).Methods(http.MethodPost)
	r.Handle("/product/update/{id}", middlewares.AuthMiddleware(http.HandlerFunc(productService.UpdateProduct))).Methods(http.MethodPut)
	r.Handle("/product/delete/{id}", middlewares.AuthMiddleware(http.HandlerFunc(productService.DeleteProduct))).Methods(http.MethodDelete)
	//retrieve categories
	//add category api(admin only)
	//filter out category
	//media upload api
	r.Handle("/product/{id}/image-upload", http.HandlerFunc(productService.UploadProductImageHandler)).Methods(http.MethodPost)
	r.Handle("/product/{id}/update-quantity", middlewares.AuthMiddleware(http.HandlerFunc(productService.UpdateProductQuantityHandler))).Methods(http.MethodPost)
}
