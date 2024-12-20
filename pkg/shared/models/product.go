package models

import (
	"errors"
	"fmt"
	"log"
	"microservices/pkg/shared/dbs"
	"microservices/pkg/shared/utils"

	"gorm.io/gorm"
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
	db := dbs.DB
	if err := db.AutoMigrate(&Product{}); err != nil {
		log.Fatalf(utils.DatabaseMigrationError, "Product", err)
	} else {
		log.Printf(utils.SchemaMigrationSuccess, "Product")
	}
}

func GetProducts() ([]Product, error) {
	var products []Product
	db := dbs.DB
	if err := db.Find(&products).Error; err != nil {
		utils.SimpleLog("error", utils.DatabaseConnectionError, err)
		return nil, fmt.Errorf(utils.DatabaseConnectionError)
	}

	utils.SimpleLog("info", utils.ProductsFetchedSuccessfully, len(products))
	return products, nil
}

func GetProductById(id int) (*Product, error) {
	var product Product
	db := dbs.DB
	if err := db.First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.SimpleLog("error", fmt.Sprintf(utils.ProductNotFoundError, id), err)
			return nil, fmt.Errorf(utils.ProductNotFoundError, id)
		}
		utils.SimpleLog("error", utils.UnexpectedDatabaseError, err)
		return nil, fmt.Errorf(utils.ProductUnexpectedFetchError)
	}

	utils.SimpleLog("info", fmt.Sprintf(utils.ProductFetchedSuccessfully, id), id)
	return &product, nil
}

func AddProduct(newProduct Product) (int, error) {
	newProduct.ID = utils.GenerateRandomID()
	if err := dbs.DB.Create(&newProduct).Error; err != nil {
		utils.SimpleLog("error", utils.ProductCreationError, newProduct.ID, err)
		return 0, fmt.Errorf(utils.ProductCreationError)
	}

	utils.SimpleLog("info", fmt.Sprintf(utils.ProductCreatedSuccessfully, newProduct.ID), newProduct.ID)
	return newProduct.ID, nil
}

func UpdateProduct(id int, newProduct Product) (*Product, error) {
	db := dbs.DB

	var oldProduct Product
	if err := db.First(&oldProduct, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.SimpleLog("error", fmt.Sprintf(utils.ProductNotFoundError, id), err)
			return nil, fmt.Errorf(utils.ProductNotFoundError, id)
		}
		utils.SimpleLog("error", utils.ProductUnexpectedUpdateError, err)
		return nil, err
	}

	updatedFields := map[string]interface{}{}
	if newProduct.PName != "" {
		oldProduct.PName = newProduct.PName
		updatedFields["PName"] = newProduct.PName
	}

	if newProduct.PDesc != "" {
		oldProduct.PDesc = newProduct.PDesc
		updatedFields["PDesc"] = newProduct.PDesc
	}

	if newProduct.Price != 0.0 {
		oldProduct.Price = newProduct.Price
		updatedFields["Price"] = newProduct.Price
	}

	if newProduct.Quantity != 0 {
		oldProduct.Quantity = newProduct.Quantity
		updatedFields["Quantity"] = newProduct.Quantity
	}

	if err := db.Save(&oldProduct).Error; err != nil {
		utils.SimpleLog("error", fmt.Sprintf(utils.ProductUpdateError, id), err)
		return nil, fmt.Errorf(utils.ProductUpdateError, id)
	}

	utils.SimpleLog("info", fmt.Sprintf(utils.ProductUpdatedSuccessfully, id), updatedFields)
	return &oldProduct, nil
}

func DeleteProduct(id int) (*Product, error) {
	db := dbs.DB
	var exisitngProduct Product
	if err := db.First(&exisitngProduct, id).Error; err != nil {
		utils.SimpleLog("error", fmt.Sprintf(utils.ProductNotFoundError, id), err)
		return nil, fmt.Errorf(utils.ProductNotFoundError, id)
	}

	if err := db.Delete(&exisitngProduct, id).Error; err != nil {
		utils.SimpleLog("error", fmt.Sprintf(utils.ProductDeletionError, id), err)
		return nil, fmt.Errorf(utils.ProductDeletionError, id)
	}

	utils.SimpleLog("info", fmt.Sprintf(utils.ProductDeletedSuccessfully, id))
	return &exisitngProduct, nil
}
