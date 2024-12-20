package models

import (
	"log"
	"math/rand"
	"microservices/pkg/shared/dbs"
)

type Product struct {
	ID       int     `json:"id"`
	PName    string  `json:"product_name"`
	PDesc    string  `json:"product_desc"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

func NewProduct() *Product {
	return &Product{}
}

func InitProductSchema() {
	if err := dbs.DB.AutoMigrate(&Product{}); err != nil {
		log.Fatalf("failed to migrate Product schema: %v", err)
	}
}

func GetProducts() ([]Product, error) {
	var products []Product
	db := dbs.DB
	if err := db.Find(&products).Error; err != nil {
		log.Printf("Error fetching products: %v", err)
		return nil, err
	}
	return products, nil
}

func AddProduct(newProduct Product) (int, error) {
	r := rand.New(rand.NewSource(99))
	newProduct.ID = int(r.Int31())
	if err := dbs.DB.Create(&newProduct).Error; err != nil {
		log.Printf("Error fetching products: %v", err)
		return 0, err
	}
	return newProduct.ID, nil
}
