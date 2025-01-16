package handlers

import (
	"e-commerce-backend/products/dbs"
	"e-commerce-backend/products/internal/services"
	"e-commerce-backend/shared/middlewares"
	"net/http"

	"github.com/gorilla/mux"
)

func ProductHandler(r *mux.Router) {
	productService := services.NewProduct(dbs.DB)

	// implementing mux router
	r.Handle("/products", http.HandlerFunc(productService.GetProducts)).Methods(http.MethodGet)
	r.Handle("/product/{id}", http.HandlerFunc(productService.GetProductById)).Methods(http.MethodGet)
	r.Handle("/product/{id}/cart", http.HandlerFunc(productService.GetProductByIdForCart)).Methods(http.MethodGet)
	r.Handle("/products/filter", http.HandlerFunc(productService.FilterProducts)).Methods(http.MethodGet)
	r.Handle("/product/add", middlewares.AuthMiddleware(http.HandlerFunc(productService.AddProduct))).Methods(http.MethodPost)
	r.Handle("/product/update/{id}", middlewares.AuthMiddleware(http.HandlerFunc(productService.UpdateProduct))).Methods(http.MethodPut)
	r.Handle("/product/delete/{id}", middlewares.AuthMiddleware(http.HandlerFunc(productService.DeleteProduct))).Methods(http.MethodDelete)
	r.HandleFunc("/products/categories", productService.FetchCategories).Methods(http.MethodGet)
	//retrieve categories
	//add category api(admin only)
	//filter out category
	//media upload api
	r.Handle("/product/{id}/image-upload", http.HandlerFunc(productService.UploadProductImageHandler)).Methods(http.MethodPost)
	r.Handle("/product/{id}/update-quantity", middlewares.AuthMiddleware(http.HandlerFunc(productService.UpdateProductQuantityHandler))).Methods(http.MethodPost)

	//more filters
	r.Handle("/products/deals", http.HandlerFunc(productService.GetDeals)).Methods(http.MethodGet)
	r.Handle("/products/offers", http.HandlerFunc(productService.GetOffers)).Methods(http.MethodGet)
	r.Handle("/products/featured", http.HandlerFunc(productService.GetFeatured)).Methods(http.MethodGet)
	//r.Handle("/products/bestsellers", http.HandlerFunc(productService.GetBestSellers)).Methods(http.MethodGet)
	//r.Handle("/products/new-arrivals", http.HandlerFunc(productService.GetNewArrivals)).Methods(http.MethodGet)
	//r.Handle("/products/{id}/reviews", http.HandlerFunc(productService.GetProductReviews)).Methods(http.MethodGet)
}
