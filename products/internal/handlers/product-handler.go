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
	r.Handle("/products", http.HandlerFunc(productService.GetProducts)).Methods(http.MethodGet)
	r.Handle("/products/{id}", http.HandlerFunc(productService.GetProductById)).Methods(http.MethodGet)
	r.Handle("/products/filter/search", http.HandlerFunc(productService.FilterProducts)).Methods(http.MethodGet)
	r.Handle("/products/add", middlewares.AuthMiddleware(http.HandlerFunc(productService.AddProduct))).Methods(http.MethodPost)
	r.Handle("/products/update/{id}", middlewares.AuthMiddleware(http.HandlerFunc(productService.UpdateProduct))).Methods(http.MethodPut)
	r.Handle("/products/delete/{id}", middlewares.AuthMiddleware(http.HandlerFunc(productService.DeleteProduct))).Methods(http.MethodDelete)
	//retrieve categories
	//add category api(admin only)
	//filter out category
	//media upload api
	r.Handle("/products/{id}/image-upload", http.HandlerFunc(productService.UploadProductImageHandler)).Methods(http.MethodPost)
	r.Handle("product/{id}/cart", middlewares.AuthMiddleware(http.HandlerFunc(productService.AddToCart))).Methods(http.MethodPost)
}
