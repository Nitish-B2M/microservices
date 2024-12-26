package models

import (
	"e-commerce-backend/products/dbs"
	"e-commerce-backend/products/pkg/payloads"
	"e-commerce-backend/shared/utils"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID        int       `json:"id"`
	PName     string    `json:"product_name" gorm:"unique;not null"`
	PDesc     string    `json:"product_desc"`
	Price     float64   `json:"price" gorm:"not null"`
	Quantity  int       `json:"quantity" gorm:"not null"`
	IsDeleted bool      `json:"is_deleted" gorm:"default:false"`
	Discount  float64   `json:"discount" gorm:"default:0"`
	Category  string    `json:"category" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `json:"updated_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	Rating    float64   `json:"rating" gorm:"default:0"`
}

type ProductTag struct {
	ProductID int `json:"product_id"`
	TagID     int `json:"tag_id"`
}

func InitProductSchema() {
	db := dbs.DB
	if err := db.AutoMigrate(&Product{}, &ProductTag{}); err != nil {
		log.Fatalf(utils.DatabaseMigrationError, "Product/ProductTag", err)
	} else {
		log.Printf(utils.SchemaMigrationSuccess, "Product/ProductTag")
	}
}

func CheckProductExistsById(id int) (*Product, bool) {
	var product Product
	db := dbs.DB
	if err := db.Where("id = ? AND is_deleted = ?", id, false).First(&product, id).Error; err != nil {
		return nil, false
	}
	return &product, true
}

func GetProducts() ([]payloads.ProductResponse, []error) {
	var products []Product
	var resProducts []payloads.ProductResponse
	var errs []error
	db := dbs.DB
	if err := db.Where("is_deleted =?", false).Find(&products).Error; err != nil {
		utils.SimpleLog("error", utils.DatabaseConnectionError, err)
		return nil, []error{fmt.Errorf(utils.DatabaseConnectionError)}
	}

	for _, product := range products {
		var resProduct payloads.ProductResponse
		if err := CopyStructIntoStruct(&product, &resProduct); err != nil {
			utils.SimpleLog("error", err.Error())
			return nil, []error{err}
		}
		resProducts = append(resProducts, resProduct)
	}

	for i, product := range resProducts {
		tags, err := FetchProductTagsName(db, product.ID)
		if err != nil {
			errs = append(errs, err...)
		}

		if tags == nil {
			resProducts[i].Tags = []string{}
		} else {
			resProducts[i].Tags = tags
		}
	}

	if len(errs) > 0 {
		utils.SimpleLog("error", utils.TagFetchError, errs)
		return nil, errs
	}

	utils.SimpleLog("info", utils.ProductsFetchedSuccessfully, len(resProducts))
	return resProducts, nil
}

func FetchProductById(id int) (*payloads.ProductResponse, error) {
	db := dbs.DB
	var resProduct payloads.ProductResponse

	product, ok := CheckProductExistsById(id)
	if !ok {
		utils.SimpleLog("error", fmt.Sprintf(utils.ProductNotFoundError, id))
		return nil, fmt.Errorf(utils.ProductNotFoundError, id)
	}

	if err := CopyStructIntoStruct(product, &resProduct); err != nil {
		utils.SimpleLog("error", err.Error())
		return nil, err
	}

	tags, err := FetchProductTagsName(db, product.ID)
	if err != nil {
		utils.SimpleLog("error", "error while fetching tags", err)
	}

	if tags == nil {
		resProduct.Tags = []string{}
	} else {
		resProduct.Tags = tags
	}

	utils.SimpleLog("info", fmt.Sprintf(utils.ProductFetchedSuccessfully, id), id)
	return &resProduct, nil
}

func AddProduct(newProduct payloads.ProductRequest) (int, error) {
	var productDB Product
	db := dbs.DB
	newProduct.ID = utils.GenerateRandomID()

	tagIds, tagErrors := CheckAndCreateProductTags(db, newProduct.Tags)
	if len(tagErrors) > 0 {
		utils.SimpleLog("error", "creating tag error", tagErrors)
		return 0, nil
	}

	if err := CopyStructIntoStruct(&newProduct, &productDB); err != nil {
		utils.SimpleLog("error", err.Error())
		return 0, err
	}

	errs := AddTagToProduct(db, tagIds, newProduct.ID)
	if len(errs) > 0 {
		for _, err := range errs {
			utils.SimpleLog("error", err.Error())
		}
	}

	if err := db.Create(&productDB).Error; err != nil {
		utils.SimpleLog("error", "creating product error", err)
		return 0, err
	}

	utils.SimpleLog("info", fmt.Sprintf(utils.ProductCreatedSuccessfully, newProduct.ID), newProduct.ID)
	return newProduct.ID, nil
}

func UpdateProduct(id int, newProduct payloads.ProductRequest, updatedFields map[string]interface{}) error {
	db := dbs.DB

	tagIds, tagErrors := CheckAndCreateProductTags(db, newProduct.Tags)
	if len(tagErrors) > 0 {
		utils.SimpleLog("error", "creating tag error", tagErrors)
		return nil
	}

	errs := UpdateTagToProduct(db, tagIds, id)
	if len(errs) > 0 {
		utils.SimpleLog("error", "while updating product tag", errs)
		return fmt.Errorf(utils.ProductTagUpdateError, id)
	}

	var productDB Product
	if err := db.Model(&productDB).Where("id = ?", id).Updates(updatedFields).Error; err != nil {
		return err
	}

	return nil
}

func DeleteProduct(id int) error {
	db := dbs.DB

	_, ok := CheckProductExistsById(id)
	if !ok {
		utils.SimpleLog("error", fmt.Sprintf(utils.ProductNotFoundError, id))
		return fmt.Errorf(utils.ProductNotFoundError, id)
	}

	if err := db.Model(&Product{}).Where("id = ? and is_deleted = ?", id, false).Update("is_deleted", true).Error; err != nil {
		utils.SimpleLog("error", fmt.Sprintf(utils.InternalServerError), err)
		return fmt.Errorf(utils.InternalServerError)
	}

	utils.SimpleLog("info", fmt.Sprintf(utils.ProductDeletedSuccessfully, id))
	return nil
}

func CopyStructIntoStruct(struct1, struct2 interface{}) error {
	srcValue := reflect.ValueOf(struct1)
	destValue := reflect.ValueOf(struct2)

	if srcValue.Kind() != reflect.Ptr || destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("both src and dest must be pointers")
	}

	srcValue = srcValue.Elem()
	destValue = destValue.Elem()

	if srcValue.Kind() != reflect.Struct || destValue.Kind() != reflect.Struct {
		return fmt.Errorf("both src and dest must be structs")
	}

	for i := 0; i < srcValue.NumField(); i++ {
		srcField := srcValue.Field(i)
		fieldName := srcValue.Type().Field(i).Name

		destField := destValue.FieldByName(fieldName)

		if destField.IsValid() && destField.CanSet() {
			if !srcField.IsZero() {
				destField.Set(srcField)
			}
		}
	}
	return nil
}

func FetchProductTagsName(db *gorm.DB, productID int) ([]string, []error) {
	var productTags []ProductTag
	var tagWithName []string
	var errs []error
	var tags []Tag

	if err := db.Find(&productTags, productID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
		} else {
			return tagWithName, []error{err}
		}
	}

	var tagIDs []int
	for _, pt := range productTags {
		tagIDs = append(tagIDs, pt.TagID)
	}

	if len(tagIDs) > 0 {
		if err := db.Where("id IN ?", tagIDs).Find(&tags).Error; err != nil {
			errs = append(errs, err)
		}

		for _, t := range tags {
			tagWithName = append(tagWithName, t.Name)
		}
	}

	if len(errs) > 0 {
		return tagWithName, errs
	}

	return tagWithName, nil
}
