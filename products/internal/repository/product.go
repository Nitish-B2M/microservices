// Package repository provides database operations for product-related operations.
package repository

import (
	"e-commerce-backend/products/internal/models"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type ProductRepo struct {
	DB *gorm.DB
}

func NewProductRepo(db *gorm.DB) *ProductRepo {
	return &ProductRepo{DB: db}
}

func (repo *ProductRepo) GetByID(id string) (*models.Product, error) {
	var product models.Product

	err := repo.DB.
		Where("id = ? AND is_deleted = ?", id, false).
		First(&product).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}

	if err != nil {
		return nil, err
	}

	fmt.Printf("product: %+v\n", product)
	return &product, nil
}

func (repo *ProductRepo) GetAll() ([]models.Product, error) {
	var products []models.Product

	err := repo.DB.
		Preload("Entities").
		Where("is_deleted = ?", false).
		Find(&products).Error

	if err != nil {
		return nil, err
	}

	return products, nil
}
