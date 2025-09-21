package services

import "e-commerce-backend/products/internal/repository"

type Services struct {
	ProductRepo *repository.ProductRepo
}

func NewService(productRepo *repository.ProductRepo) *Services {
	return &Services{
		ProductRepo: productRepo,
	}
}
