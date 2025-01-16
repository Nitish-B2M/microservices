package models

import (
	"e-commerce-backend/products/dbs"
	"e-commerce-backend/products/pkg/payloads"
	"e-commerce-backend/shared/utils"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log"
	"reflect"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID         int       `json:"id"`
	SellerID   uuid.UUID `json:"seller_id" gorm:"not null"`
	PName      string    `json:"name" gorm:"unique;not null;column:name"`
	PDesc      string    `json:"description" gorm:"column:description"`
	Price      float64   `json:"price" gorm:"not null"`
	Quantity   int       `json:"quantity" gorm:"not null"`
	IsDeleted  bool      `json:"is_deleted" gorm:"default:false"`
	InStock    bool      `json:"in_stock" gorm:"default:true"`
	Discount   float64   `json:"discount" gorm:"default:0"`
	Category   string    `json:"category" gorm:"not null"`
	IsFeatured bool      `json:"is_featured" gorm:"default:false"`
	TaxRate    float64   `json:"tax_rate" gorm:"default:0"`
	CreatedAt  time.Time `json:"created_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	Rating     float64   `json:"rating" gorm:"default:0"`
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
		return nil, []error{fmt.Errorf(utils.DatabaseConnectionError)}
	}

	if len(products) == 0 {
		return nil, []error{errors.New("no products found")}
	}

	for _, product := range products {
		var resProduct payloads.ProductResponse
		if product.Quantity <= 0 {
			product.InStock = false
			product.UpdateProduct(db, product.ID, map[string]interface{}{"in_stock": false})
		}
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

	if productResp.Quantity <= 0 {
		productResp.InStock = false
		p.UpdateProduct(db, productResp.ID, map[string]interface{}{"in_stock": false})
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

func (p *Product) FetchProductCategories(db *gorm.DB) ([]string, error) {
	var categories []string
	if err := db.Table("products").Select("DISTINCT category").Where("is_deleted = ?", false).Pluck("category", &categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
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
	if p.Quantity <= 0 {
		p.InStock = false
	}
	// update quantity and in_stock
	if err := db.Model(&p).Where("id =?", p.ID).Updates(map[string]interface{}{"quantity": p.Quantity, "in_stock": p.InStock}).Error; err != nil {
		return err
	}

	return nil
}

// FilterCriteria represents the criteria for filtering products
type FilterCriteria struct {
	PName       string   `json:"name"`
	MinPrice    float64  `json:"min_price"`
	MaxPrice    float64  `json:"max_price"`
	MinRating   float64  `json:"min_rating"`
	Category    string   `json:"category"`
	MinQuantity int      `json:"min_quantity"`
	MaxQuantity int      `json:"max_quantity"`
	Tags        []string `json:"tags"`
}

func FilterProduct(db *gorm.DB, criteria FilterCriteria) ([]payloads.ProductResponse, error) {
	var products []Product
	query := db.Model(&Product{}).Where("is_deleted = ?", false)

	// Dynamically apply filters to the query
	if criteria.PName != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(criteria.PName)+"%")
	}
	if criteria.MinPrice > 0 {
		query = query.Where("price >= ?", criteria.MinPrice)
	}
	if criteria.MaxPrice > 0 {
		query = query.Where("price <= ?", criteria.MaxPrice)
	}
	if criteria.MinRating > 0 {
		query = query.Where("rating >= ?", criteria.MinRating)
	}
	if criteria.Category != "" {
		query = query.Where("LOWER(category) LIKE ?", "%"+strings.ToLower(criteria.Category)+"%")
	}
	if criteria.MinQuantity > 0 {
		query = query.Where("quantity >= ?", criteria.MinQuantity)
	}
	if criteria.MaxQuantity > 0 {
		query = query.Where("quantity <= ?", criteria.MaxQuantity)
	}
	// Uncomment this if you have a Tags column or join for filtering
	// if criteria.Tags != "" {
	//     query = query.Where("tags LIKE ?", "%"+criteria.Tags+"%")
	// }

	// Execute the query
	if err := query.Find(&products).Error; err != nil {
		return nil, fmt.Errorf("error fetching products: %w", err)
	}

	// Map products to payloads
	filteredProducts := make([]payloads.ProductResponse, len(products))
	for i, product := range products {
		filteredProducts[i] = payloads.ProductResponse{
			ID:         product.ID,
			PName:      product.PName,
			PDesc:      product.PDesc,
			Price:      product.Price,
			Quantity:   product.Quantity,
			IsDeleted:  product.IsDeleted,
			Category:   product.Category,
			IsFeatured: product.IsFeatured,
			TaxRate:    product.TaxRate,
			Rating:     product.Rating,
		}
	}

	return filteredProducts, nil
}
