// Package handlers provides HTTP handler functions for product-related endpoints.
package handlers

import (
	"e-commerce-backend/products/dbs"
	"e-commerce-backend/products/internal/controller"
	"e-commerce-backend/products/internal/repository"
	"e-commerce-backend/products/internal/services"
	mw "e-commerce-backend/shared/middlewares"
	"net/http"

	"github.com/gorilla/mux"
)

func ProductHandler(r *mux.Router) {
	// implementing mux router
	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("../uploads")))).Methods(http.MethodGet)
	//more filters
	//r.Handle("/products/bestsellers", http.HandlerFunc(productService.GetBestSellers)).Methods(http.MethodGet)
	//r.Handle("/products/new-arrivals", http.HandlerFunc(productService.GetNewArrivals)).Methods(http.MethodGet)
	//r.Handle("/products/{id}/reviews", http.HandlerFunc(productService.GetProductReviews)).Methods(http.MethodGet)
}

func ProductRoutes(r *mux.Router) {
	productRepo := repository.NewProductRepo(dbs.DB)
	svc := services.NewService(productRepo)
	productController := controller.NewProductController(svc)

	productSvc := services.NewProduct(dbs.DB)
	filterService := services.NewFilterService(dbs.DB)

	mw.HandleGet(r, "/list", "", productController.FetchProducts)
	mw.HandleGet(r, "/categories", "", productSvc.FetchCategories)
	mw.HandleGet(r, "/filter", "", filterService.FilterProducts)
	mw.HandleGet(r, "/deals", "", filterService.GetDeals)
	mw.HandleGet(r, "/offers", "", filterService.GetOffers)
	mw.HandleGet(r, "/featured", "", filterService.GetFeatured)

	mw.HandleGet(r, "/item/{id}", "", productController.FetchProductByID)
	mw.HandleGet(r, "/item/{id}/cart", "", productSvc.GetProductByIDForCart, mw.AuthMiddleware)
}

func SellerRoutes(r *mux.Router) {
	productSvc := services.NewProduct(dbs.DB)

	mw.HandlePost(r, "/add", "", productSvc.AddProduct, mw.AuthMiddleware, mw.RoleMiddleware(dbs.DB, "seller"))
	mw.HandlePost(r, "/update/{id}", "", productSvc.UpdateProduct, mw.AuthMiddleware, mw.RoleMiddleware(dbs.DB, "seller"))
	mw.HandleGet(r, "/manage", "", productSvc.GetProductsForSeller, mw.AuthMiddleware, mw.RoleMiddleware(dbs.DB, "seller"))
	mw.HandleDelete(r, "/delete/{id}", "", productSvc.DeleteProduct, mw.AuthMiddleware, mw.RoleMiddleware(dbs.DB, "seller"))
	mw.HandlePost(r, "/{id}/image-upload", "", productSvc.UploadProductImageHandler, mw.AuthMiddleware, mw.RoleMiddleware(dbs.DB, "seller"))
	mw.HandlePost(r, "/{id}/update-quantity", "", productSvc.UpdateProductQuantityHandler, mw.AuthMiddleware, mw.RoleMiddleware(dbs.DB, "seller"))
}
