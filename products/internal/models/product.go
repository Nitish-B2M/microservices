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

func (p *Product) CheckProductExistsById(db *gorm.DB, id int) error {
	return db.Where("id = ? AND is_deleted = ?", id, false).First(&p).Error
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

func (p *Product) FetchProductResp(db *gorm.DB, id int) (*payloads.ProductResponse, error) {
	var productResp payloads.ProductResponse
	if err := p.CheckProductExistsById(db, id); err != nil {
		return nil, err
	}

	if err := CopyStructIntoStruct(p, &productResp); err != nil {
		return nil, err
	}
	return &productResp, nil
}

func (p *Product) AddProduct(db *gorm.DB) (int, error) {
	p.ID = utils.GenerateRandomID()
	if err := db.Create(&p).Error; err != nil {
		return 0, err
	}
	return p.ID, nil
}

func (p *Product) UpdateProduct(db *gorm.DB, id int, updatedFields map[string]interface{}) error {
	if err := db.Model(&p).Where("id = ?", id).Updates(updatedFields).Error; err != nil {
		return err
	}
	return nil
}

func (p *Product) DeleteProduct(db *gorm.DB, id int) error {
	if err := p.CheckProductExistsById(db, id); err != nil {
		return fmt.Errorf(utils.ProductNotFoundError, id)
	}

	if err := db.Model(&Product{}).Where("id = ? and is_deleted = ?", id, false).Update("is_deleted", true).Error; err != nil {
		return fmt.Errorf(utils.InternalServerError)
	}

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

func (p *Product) UpdateProductQuantity(db *gorm.DB) error {
	if err := db.Model(&p).Where("id =?", p.ID).Update("quantity", p.Quantity).Error; err != nil {
		return err
	}
	return nil
}
