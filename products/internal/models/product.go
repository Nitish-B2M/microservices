package models

import (
	"e-commerce-backend/products/pkg/payloads"
	"e-commerce-backend/shared/utils"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func migrateDB(db *gorm.DB, models ...interface{}) {
	for _, model := range models {
		if !db.Migrator().HasTable(model) {
			if err := db.Migrator().CreateTable(model); err != nil {
				log.Printf("Error creating table %v: %v", model, err)
				return
			}
		}
	}

	log.Printf(utils.SchemaMigrationSuccess, "Migrated product-related tables")
}

func InitNewProductSchema(db *gorm.DB) {
	if db == nil {
		log.Fatal("Database connection is nil")
		return
	}

	// Drop existing tables in reverse order to avoid foreign key constraints
	//tables := []interface{}{
	//	&ProductAttribute{},
	//	&ProductImage{},
	//	&ProductVariant{},
	//	&Entities{},
	//	&ProductEntity{},
	//	//&Product{},
	//}
	//
	////log.Println("tables", tables)
	////// Drop tables safely
	//for _, table := range tables {
	//	if db.Migrator().HasTable(table) {
	//		if err := db.Migrator().DropTable(table); err != nil {
	//			log.Printf("Error dropping table %v: %v", table, err)
	//			// Continue even if drop fails
	//		}
	//	}
	//}

	//log.Printf(utils.SchemaMigrationSuccess, "Dropped existing product-related tables")

	// Create tables in correct order
	migrations := []interface{}{
		&Entities{},
		&Product{},
		&ProductImage{},
		&ProductVariant{},
		&ProductAttribute{},
		&Entities{},
	}

	// Migrate all tables
	migrateDB(db, migrations...)
}

func NewVariant() *ProductVariant {
	return &ProductVariant{
		ID:       "",
		PID:      "",
		Name:     "",
		SKU:      "",
		Price:    0,
		Quantity: 0,
		Options:  []VariantOption{},
	}
}

var productMutex sync.Mutex

type ProductInterface interface {
	CheckProductExistsById(db *gorm.DB, id string) error
	FetchProductResp(db *gorm.DB, id string) (*payloads.ProductResponse, error)
	AddProduct(db *gorm.DB) (*Product, error)
	UpdateProduct(db *gorm.DB, id string, changedFieldValue payloads.ProductRequest) error
	DeleteProduct(db *gorm.DB, id string) error
	UpdateProductQuantity(db *gorm.DB) error
	GetProductById(db *gorm.DB, id string) (*Product, error)
}

func CreateImageURL(url string) string {
	if !strings.HasPrefix(url, "http") {
		return "http://localhost:8081/" + url
	}
	return url
}

func RemoveImageURLDomain(url string) string {
	if strings.HasPrefix(url, "http") {
		if strings.HasPrefix(url, "http://localhost:8081") {
			return strings.TrimPrefix(url, "http://localhost:8081/")
		}
		return strings.TrimPrefix(url, "http://")
	}
	return url
}

func (product *Product) CheckProductExistsById(db *gorm.DB, id string) error {
	if err := db.
		Preload("Category.Entity").
		Preload("Brand.Entity").
		Preload("Tags.Entity").
		Preload("Images").
		Preload("Variants").
		Preload("Attributes").
		Where("is_deleted = ?", false).
		First(&product, "id = ?", id).Error; err != nil {
		return err
	}

	// if len(product.Images) > 0 {
	// 	for i := range product.Images {
	// 		product.Images[i].URL = CreateImageURL(product.Images[i].URL)
	// 	}
	// }
	return nil
}

func GetProducts(db *gorm.DB) ([]payloads.ProductResponse, error) {
	var products []Product
	if err := db.
		Preload("Entities").
		Where("is_deleted = ?", false).
		Find(&products).Error; err != nil {
		return nil, err
	}

	var response []payloads.ProductResponse
	for _, product := range products {
		// if len(product.Images) > 0 {
		// 	for i := range product.Images {
		// 		product.Images[i].URL = CreateImageURL(product.Images[i].URL)
		// 	}
		// }
		res := CopyProductToProductResponse(product)
		response = append(response, res)
	}
	return response, nil
}

func GetProductsForSeller(db *gorm.DB) ([]Product, error) {
	var products []Product
	if err := db.
		Where("is_deleted = ?", false).
		Find(&products).Error; err != nil {
		return nil, err
	}

	// for i := range products {
	// 	if len(products[i].Images) > 0 {
	// 		for j := range products[i].Images {
	// 			products[i].Images[j].URL = CreateImageURL(products[i].Images[j].URL)
	// 		}
	// 	}
	// }
	return products, nil
}

func CheckProductExistsBySKU(db *gorm.DB, sku string, table string) error {
	if table == "" {
		table = "products"
	}
	if table == "products" {
		return db.Where("sku = ?", sku).First(&Product{}).Error
	} else if table == "product_variants" {
		return db.Table("product_variants").Where("sku = ?", sku).First(&ProductVariant{}).Error
	}
	return nil
}

func (product *Product) FetchProductResp(db *gorm.DB, id string) (*payloads.ProductResponse, error) {
	if err := product.CheckProductExistsById(db, id); err != nil {
		return nil, err
	}
	// avoiding race condition
	productMutex.Lock()
	defer productMutex.Unlock()

	//if product.Quantity <= 0 {
	//	productResp.InStock = false
	//	//product.UpdateProduct(db, productResp.ID, map[string]interface{}{"in_stock": false})
	//}

	productResp := CopyProductToProductResponse(*product)
	return &productResp, nil
}

func (product *Product) AddProduct(db *gorm.DB, category, brand string, tags []string) (*payloads.ProductResponse, error) {
	//avoiding race condition
	productMutex.Lock()
	defer productMutex.Unlock()

	//First check if product exists by SKU
	if err := CheckProductExistsBySKU(db, product.SKU, "products"); err == nil {
		return nil, fmt.Errorf("product with SKU %s already exists", product.SKU)
	}

	// Start transaction
	tx := db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	//Handle product variants
	// if len(product.Variants) > 0 {
	// 	for _, variant := range product.Variants {
	// 		if err := CheckProductExistsBySKU(tx, variant.SKU, "product_variants"); err == nil {
	// 			tx.Rollback()
	// 			return nil, fmt.Errorf("product variant with SKU %s already exists", variant.SKU)
	// 		}
	// 	}
	// }

	// Create the product
	if err := tx.Create(&product).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error creating product: %v", err)
	}

	// Create product variants
	// if len(product.Variants) > 0 {
	// 	for i := range product.Variants {
	// 		product.Variants[i].PID = product.ID
	// 		if err := tx.Create(&product.Variants[i]).Error; err != nil {
	// 			tx.Rollback()
	// 			return nil, fmt.Errorf("error creating product variants: %v", err)
	// 		}
	// 	}
	// }

	// Create product images
	// if len(product.Images) > 0 {
	// 	for i := range product.Images {
	// 		product.Images[i].PID = product.ID
	// 		if err := tx.Create(&product.Images[i]).Error; err != nil {
	// 			tx.Rollback()
	// 			return nil, fmt.Errorf("error creating product images: %v", err)
	// 		}
	// 	}
	// }

	// Create product attributes
	// if len(product.Attributes) > 0 {
	// 	for i := range product.Attributes {
	// 		product.Attributes[i].PID = product.ID
	// 		if err := tx.Create(&product.Attributes[i]).Error; err != nil {
	// 			tx.Rollback()
	// 			return nil, fmt.Errorf("error creating product attributes: %v", err)
	// 		}
	// 	}
	// }

	// Handle tags, categories, and brands within the same transaction
	// Tags
	// if len(tags) > 0 {
	// 	entityLabelResp, err := CheckAndCreateTags(tx, tags, product.ID)
	// 	if err != nil {
	// 		tx.Rollback()
	// 		return nil, fmt.Errorf("error creating tags: %v", err)
	// 	}
	// 	product.Tags = entityLabelResp
	// }

	// category
	// if len(category) > 0 {
	// 	entityProductCategoryResp, err := CheckAndCreateCategoryOrBrand(tx, category, "category", product.ID)
	// 	if err != nil {
	// 		tx.Rollback()
	// 		return nil, fmt.Errorf("error creating category: %v", err)
	// 	}
	// 	product.Category = *entityProductCategoryResp
	// }

	// // brand
	// if len(brand) > 0 {
	// 	entityProductBrandResp, err := CheckAndCreateCategoryOrBrand(tx, brand, "brand", product.ID)
	// 	if err != nil {
	// 		tx.Rollback()
	// 		return nil, fmt.Errorf("error creating brand: %v", err)
	// 	}
	// 	product.Brand = *entityProductBrandResp
	// }

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	// Fetch the complete product with all associations
	if err := db.Preload("Tags.Entity").Preload("Category.Entity").Preload("Brand.Entity").Preload("Images").Preload("Variants").Preload("Attributes").First(&product, "id = ?", product.ID).Error; err != nil {
		return nil, err
	}

	createdProduct := CopyProductToProductResponse(*product)

	return &createdProduct, nil
}

func processImages(existingImages []ProductImage, updatedImages []payloads.ProductImageRequest, productID string) (imageToBeDeleted, imageToBeAdded []ProductImage, err error) {
	log.Println("processing images", existingImages, updatedImages)
	if len(existingImages) > 0 {
		if len(updatedImages) > 0 {
			newImageMap := make(map[string]ProductImage)
			for _, image := range existingImages {
				newImageMap[image.ID] = image
			}

			for i := range updatedImages {
				if _, ok := newImageMap[updatedImages[i].ID]; ok {
					delete(newImageMap, updatedImages[i].ID)
					continue
				}
				if updatedImages[i].ID == "" {
					updatedImages[i].ID = uuid.New().String()
				}
				updatedImages[i].PID = productID
				if updatedImages[i].URL == "" {
					filePath, uploadErr := UploadBase64Image(updatedImages[i].Image, updatedImages[i].PID, updatedImages[i].ID)
					if uploadErr != nil {
						return nil, nil, fmt.Errorf("failed to upload base64 image: %v", uploadErr)
					}
					updatedImages[i].URL = filePath
				}
				imageToBeAdded = append(imageToBeAdded, CopyProductImgReqToProductImg(updatedImages[i]))
			}

			for _, image := range newImageMap {
				imageToBeDeleted = append(imageToBeDeleted, image)
			}
		} else {
			imageToBeDeleted = existingImages
		}
	} else {
		for i := range updatedImages {
			if updatedImages[i].ID == "" {
				updatedImages[i].ID = uuid.New().String()
			}
			updatedImages[i].PID = productID
			if updatedImages[i].URL == "" {
				filePath, uploadErr := UploadBase64Image(updatedImages[i].Image, updatedImages[i].PID, updatedImages[i].ID)
				if uploadErr != nil {
					return nil, nil, fmt.Errorf("failed to upload base64 image: %v", uploadErr)
				}
				updatedImages[i].URL = filePath
			}
			imageToBeAdded = append(imageToBeAdded, CopyProductImgReqToProductImg(updatedImages[i]))
		}
	}

	return imageToBeDeleted, imageToBeAdded, nil
}

func UpdateProductImages(db *gorm.DB, pid string, changedFieldValue []payloads.ProductImageRequest) error {
	var existingImages []ProductImage
	if err := db.Where("p_id = ?", pid).Find(&existingImages).Error; err != nil {
		return fmt.Errorf("failed to fetch existing images: %v", err)
	}
	imageToBeDeleted, imageToBeAdded, err := processImages(existingImages, changedFieldValue, pid)
	if err != nil {
		return err
	}
	log.Println(imageToBeDeleted, imageToBeAdded)
	if len(imageToBeDeleted) > 0 {
		if err := db.Delete(&imageToBeDeleted).Error; err != nil {
			return fmt.Errorf("failed to delete images: %v", err)
		}
	}
	if len(imageToBeAdded) > 0 {
		if err := db.Create(&imageToBeAdded).Error; err != nil {
			return fmt.Errorf("failed to create images: %v", err)
		}
	}
	return nil
}

func UpdateProductVariants(db *gorm.DB, pid string, changedFieldValue []payloads.ProductVariantRequest) error {
	var existingVariants []ProductVariant
	if err := db.Preload("Options").Where("p_id = ?", pid).Find(&existingVariants).Error; err != nil {
		return fmt.Errorf("failed to fetch existing variants: %v", err)
	}

	// Map of existing variants by ID
	existingVariantMap := make(map[string]ProductVariant)
	for _, ev := range existingVariants {
		existingVariantMap[ev.ID] = ev
	}

	// Map of incoming variants by ID
	newVariantMap := make(map[string]payloads.ProductVariantRequest)
	for _, nv := range changedFieldValue {
		newVariantMap[nv.ID] = nv
	}

	var variantsToCreate []ProductVariant
	var variantsToUpdate []ProductVariant
	var variantsToDelete []ProductVariant

	// Step 1: Update or mark for deletion
	for _, existing := range existingVariants {
		if updated, exists := newVariantMap[existing.ID]; exists {
			isUpdated := false

			if existing.Name != updated.Name {
				existing.Name = updated.Name
				isUpdated = true
			}
			if existing.Price != updated.Price {
				existing.Price = updated.Price
				isUpdated = true
			}
			if existing.Quantity != updated.Quantity {
				existing.Quantity = updated.Quantity
				isUpdated = true
			}
			if existing.SKU != updated.SKU {
				existing.SKU = updated.SKU
				isUpdated = true
			}

			// ⚠️ Update Options: Replace old with new
			if err := db.Where("variant_id = ?", existing.ID).Delete(&VariantOption{}).Error; err != nil {
				return fmt.Errorf("failed to clear old options: %v", err)
			}

			var newOptions []VariantOption
			for _, opt := range updated.Options {
				newOptions = append(newOptions, VariantOption{
					ID:          uuid.NewString(),
					VariantID:   existing.ID,
					OptionName:  opt.OptionName,
					OptionValue: opt.OptionValue,
				})
			}
			existing.Options = newOptions
			isUpdated = true

			if isUpdated {
				variantsToUpdate = append(variantsToUpdate, existing)
			}
		} else {
			// Not found in updated list, mark for deletion
			variantsToDelete = append(variantsToDelete, existing)
		}
	}

	// Step 2: Identify new variants to create
	for _, nv := range changedFieldValue {
		if _, exists := existingVariantMap[nv.ID]; !exists {
			newID := uuid.New().String()
			variant := ProductVariant{
				ID:       newID,
				PID:      pid,
				Name:     nv.Name,
				SKU:      nv.SKU,
				Price:    nv.Price,
				Quantity: nv.Quantity,
			}

			for _, opt := range nv.Options {
				variant.Options = append(variant.Options, VariantOption{
					ID:          uuid.NewString(),
					VariantID:   newID,
					OptionName:  opt.OptionName,
					OptionValue: opt.OptionValue,
				})
			}
			variantsToCreate = append(variantsToCreate, variant)
		}
	}

	// Step 3: DB operations
	if len(variantsToCreate) > 0 {
		if err := db.Create(&variantsToCreate).Error; err != nil {
			return fmt.Errorf("failed to create variants: %v", err)
		}
	}

	if len(variantsToUpdate) > 0 {
		for _, updatedVariant := range variantsToUpdate {
			if err := db.Session(&gorm.Session{FullSaveAssociations: true}).Save(&updatedVariant).Error; err != nil {
				return fmt.Errorf("failed to update variant: %v", err)
			}
		}
	}

	if len(variantsToDelete) > 0 {
		var idsToDelete []string
		for _, v := range variantsToDelete {
			idsToDelete = append(idsToDelete, v.ID)
		}
		if err := db.Where("id IN ?", idsToDelete).Delete(&ProductVariant{}).Error; err != nil {
			return fmt.Errorf("failed to delete variants: %v", err)
		}
		// Also delete options for these variants
		if err := db.Where("variant_id IN ?", idsToDelete).Delete(&VariantOption{}).Error; err != nil {
			return fmt.Errorf("failed to delete variant options: %v", err)
		}
	}

	return nil
}

func UpdateProductAttributes(db *gorm.DB, pid string, changedFieldValue []payloads.ProductAttributeRequest) error {
	var existingAttributes []ProductAttribute
	if err := db.Where("p_id = ?", pid).Find(&existingAttributes).Error; err != nil {
		return fmt.Errorf("failed to fetch existing attributes: %v", err)
	}

	// Step 2: Create maps for quick lookups
	existingAttributeMap := make(map[string]ProductAttribute)
	for _, existingAttribute := range existingAttributes {
		existingAttributeMap[existingAttribute.ID] = existingAttribute
	}

	newAttributeMap := make(map[string]payloads.ProductAttributeRequest)
	for _, newAttribute := range changedFieldValue {
		newAttributeMap[newAttribute.ID] = newAttribute
	}

	// Step 3: Variables to accumulate changes
	var attributesToCreate []ProductAttribute
	var attributesToUpdate []ProductAttribute
	var attributesToDelete []ProductAttribute

	// Step 4: Identify updates and deletions
	for _, existingAttribute := range existingAttributes {
		if newAttribute, exists := newAttributeMap[existingAttribute.ID]; exists {
			// Attribute exists, check if it needs to be updated
			isUpdated := false

			// Update Name if changed
			if existingAttribute.Name != newAttribute.Name {
				existingAttribute.Name = newAttribute.Name
				isUpdated = true
			}

			// Update Value if changed
			if existingAttribute.Value != newAttribute.Value {
				existingAttribute.Value = newAttribute.Value
				isUpdated = true
			}

			// If any field was updated, add it to the update list
			if isUpdated {
				attributesToUpdate = append(attributesToUpdate, existingAttribute)
			}
		} else {
			// If the attribute does not exist in the new data, add it to the delete list
			attributesToDelete = append(attributesToDelete, existingAttribute)
		}
	}

	// Step 5: Identify new attributes (to create)
	for _, newAttribute := range changedFieldValue {
		if _, exists := existingAttributeMap[newAttribute.ID]; !exists {
			// If the attribute is not in the existing attributes, add it to the creation list
			newAttribute.ID = uuid.New().String()
			newAttribute.PID = pid

			attribute := ProductAttribute{
				ID:    newAttribute.ID,
				PID:   newAttribute.PID,
				Name:  newAttribute.Name,
				Value: newAttribute.Value,
			}
			attributesToCreate = append(attributesToCreate, attribute)
		}
	}

	// Step 6: Perform batch database operations
	// Create new attributes
	if len(attributesToCreate) > 0 {
		if err := db.Create(&attributesToCreate).Error; err != nil {
			return fmt.Errorf("failed to create new attributes: %v", err)
		}
		log.Println("Created new attributes:", attributesToCreate)
	}

	// Update existing attributes
	if len(attributesToUpdate) > 0 {
		if err := db.Save(&attributesToUpdate).Error; err != nil {
			return fmt.Errorf("failed to update attributes: %v", err)
		}
		log.Println("Updated existing attributes:", attributesToUpdate)
	}

	// Delete removed attributes
	if len(attributesToDelete) > 0 {
		if err := db.Delete(&attributesToDelete).Error; err != nil {
			return fmt.Errorf("failed to delete attributes: %v", err)
		}
		log.Println("Deleted attributes:", attributesToDelete)
	}
	return nil
}

func (product *Product) UpdateProduct(db *gorm.DB, id string, changedFieldValue payloads.ProductRequest) error {
	//avoiding race condition
	productMutex.Lock()
	defer productMutex.Unlock()

	// Start transaction
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // Ensures rollback in case of panic
		}
	}()

	// Fetch the existing product by ID
	var existingP Product
	if err := tx.Where("id = ?", id).First(&existingP).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("product not found with ID: %v", id)
	}

	// Update product fields
	existingP.SKU = utils.UpdateField(existingP.SKU, changedFieldValue.SKU).(string)
	existingP.Name = utils.UpdateField(existingP.Name, changedFieldValue.Name).(string)
	existingP.Description = utils.UpdateField(existingP.Description, changedFieldValue.Description).(string)
	existingP.ShortDesc = utils.UpdateField(existingP.ShortDesc, changedFieldValue.ShortDesc).(string)
	existingP.Price = utils.UpdateField(existingP.Price, changedFieldValue.Price).(float64)
	existingP.Quantity = utils.UpdateField(existingP.Quantity, changedFieldValue.Quantity).(int)
	existingP.IsFeatured = utils.UpdateField(existingP.IsFeatured, changedFieldValue.IsFeatured).(bool)
	existingP.Rating = utils.UpdateField(existingP.Rating, changedFieldValue.Rating).(float64)
	existingP.Tax = utils.UpdateField(existingP.Tax, changedFieldValue.Tax).(float64)
	existingP.Discount = utils.UpdateField(existingP.Discount, changedFieldValue.Discount).(float64)
	existingP.DiscountType = utils.UpdateField(existingP.DiscountType, changedFieldValue.DiscountType).(string)
	existingP.Weight = utils.UpdateField(existingP.Weight, changedFieldValue.Weight).(float64)
	existingP.Length = utils.UpdateField(existingP.Length, changedFieldValue.Length).(float64)
	existingP.Width = utils.UpdateField(existingP.Width, changedFieldValue.Width).(float64)
	existingP.DimensionUnit = utils.UpdateField(existingP.DimensionUnit, changedFieldValue.DimensionUnit).(string)
	existingP.MetaTitle = utils.UpdateField(existingP.MetaTitle, changedFieldValue.MetaTitle).(string)
	existingP.MetaDesc = utils.UpdateField(existingP.MetaDesc, changedFieldValue.MetaDesc).(string)
	existingP.MainImage = utils.UpdateField(existingP.MainImage, changedFieldValue.MainImage).(string)

	// Update images (if any)
	if err := UpdateProductImages(db, existingP.ID, changedFieldValue.Images); err != nil {
		tx.Rollback()
		return err
	}

	// Update variants (if any)
	if err := UpdateProductVariants(db, existingP.ID, changedFieldValue.Variants); err != nil {
		tx.Rollback()
		return err
	}

	// Update attributes (if any)
	if err := UpdateProductAttributes(db, existingP.ID, changedFieldValue.Attributes); err != nil {
		tx.Rollback()
		return err
	}

	// Finally, update the main product record
	if err := tx.Save(&existingP).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update product: %v", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func (product *Product) DeleteProduct(db *gorm.DB, id string) error {
	if err := product.CheckProductExistsById(db, id); err != nil {
		return fmt.Errorf(utils.ProductNotFoundError, id)
	}

	if err := db.Model(&Product{}).Where("id = ? and is_deleted = ?", id, false).Update("is_deleted", true).Error; err != nil {
		return fmt.Errorf(utils.InternalServerError)
	}

	return nil
}

func (product *Product) FetchProductCategories(db *gorm.DB) ([]Entities, error) {
	var entities []Entities
	if err := db.Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
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

//func FetchProductTagsName(db *gorm.DB, productID string) ([]string, []error) {
//	var productTags []ProductTag
//	var tagWithName []string
//	var errs []error
//	var tags []Tag
//
//	if err := db.Find(&productTags, productID).Error; err != nil {
//		if errors.Is(err, gorm.ErrRecordNotFound) {
//		} else {
//			return tagWithName, []error{err}
//		}
//	}
//
//	var tagIDs []int
//	for _, pt := range productTags {
//		tagIDs = append(tagIDs, pt.TagID)
//	}
//
//	if len(tagIDs) > 0 {
//		if err := db.Where("id IN ?", tagIDs).Find(&tags).Error; err != nil {
//			errs = append(errs, err)
//		}
//
//		for _, t := range tags {
//			tagWithName = append(tagWithName, t.Name)
//		}
//	}
//
//	if len(errs) > 0 {
//		return tagWithName, errs
//	}
//
//	return tagWithName, nil
//}

func (product *Product) UpdateProductQuantity(db *gorm.DB) error {
	// if product.Quantity <= 0 {
	// 	product.InStock = false
	// } else {
	// 	product.InStock = true
	// }
	// update quantity and in_stock
	if err := db.Model(&product).Where("id =?", product.ID).Updates(map[string]interface{}{"quantity": product.Quantity}).Error; err != nil {
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
	Brand       string   `json:"brand"`
	MinQuantity int      `json:"min_quantity"`
	MaxQuantity int      `json:"max_quantity"`
	Tags        []string `json:"tags"`
}

func FilterProduct(db *gorm.DB, criteria FilterCriteria) ([]payloads.ProductResponse, error) {
	var products []Product
	query := db.
		Preload("Category.Entity").
		Preload("Brand.Entity").
		Preload("Tags.Entity").
		Preload("Images").
		Preload("Variants").
		Preload("Attributes").Model(&Product{}).Where("is_deleted = ?", false)

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
	if criteria.MinQuantity > 0 {
		query = query.Where("quantity >= ?", criteria.MinQuantity)
	}
	if criteria.MaxQuantity > 0 {
		query = query.Where("quantity <= ?", criteria.MaxQuantity)
	}
	if criteria.Category != "" || criteria.Brand != "" {
		query = query.Joins("JOIN product_entities AS pe ON pe.product_id = products.id").
			Joins("JOIN entities AS e ON e.e_id = pe.entity_id")
		if criteria.Category != "" {
			categories := strings.Split(criteria.Category, ",")
			categoryConditions := []string{}
			var args []interface{}

			for _, category := range categories {
				categoryConditions = append(categoryConditions, "LOWER(e.entity_name) LIKE ?")
				args = append(args, "%"+strings.ToLower(category)+"%")
			}
			query = query.Where("("+strings.Join(categoryConditions, " OR ")+")", args...)
			query = query.Where("e.entity_type = ?", "category")
		}

		if criteria.Brand != "" {
			brands := strings.Split(criteria.Brand, ",")
			brandConditions := []string{}
			var args []interface{}

			for _, brand := range brands {
				brandConditions = append(brandConditions, "LOWER(e.entity_name) LIKE ?")
				args = append(args, "%"+strings.ToLower(brand)+"%")
			}

			query = query.Where("("+strings.Join(brandConditions, " OR ")+")", args...)
			query = query.Where("e.entity_type1 = ?", "brand")
		}
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
			ID:          product.ID,
			Name:        product.Name,
			SKU:         product.SKU,
			ShortDesc:   product.ShortDesc,
			Description: product.Description,
			Price:       product.Price,
			IsFeatured:  product.IsFeatured,
			Tax:         product.Tax,
			Rating:      product.Rating,
		}
	}

	return filteredProducts, nil
}

func (product *Product) GetProductById(db *gorm.DB, id string) (*Product, error) {
	var p Product
	if err := db.
		Preload("Category.Entity").
		Preload("Brand.Entity").
		Preload("Tags.Entity").
		Preload("Images").
		Preload("Variants").
		Preload("Attributes").
		First(&p, "id = ?", id).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf(utils.ProductNotFoundError, id)
		}
		return nil, err
	}

	// if len(p.Images) > 0 {
	// 	for i := range p.Images {
	// 		p.Images[i].URL = CreateImageURL(p.Images[i].URL)
	// 	}
	// }
	return &p, nil
}

// CopyProductToProductResponse converts a Product to a ProductResponse
func CopyProductToProductResponse(product Product) payloads.ProductResponse {
	// Create a ProductResponse instance
	productResp := payloads.ProductResponse{
		ID:            product.ID,
		SellerID:      product.SellerID,
		SKU:           product.SKU,
		Name:          product.Name,
		Description:   product.Description,
		ShortDesc:     product.ShortDesc,
		Price:         product.Price,
		ComparePrice:  product.ComparePrice,
		CostPrice:     product.CostPrice,
		Quantity:      product.Quantity,
		IsFeatured:    product.IsFeatured,
		Rating:        product.Rating,
		Tax:           product.Tax,
		Discount:      product.Discount,
		DiscountType:  product.DiscountType,
		Weight:        product.Weight,
		Length:        product.Length,
		Width:         product.Width,
		Height:        product.Height,
		DimensionUnit: product.DimensionUnit,
		MetaTitle:     product.MetaTitle,
		MetaDesc:      product.MetaDesc,
		SlugURL:       product.SlugURL,
		MainImage:     product.MainImage,
		CreatedAt:     product.CreatedAt,
		UpdatedAt:     product.UpdatedAt,
	}

	// Copy Images
	// for _, img := range product.Images {
	// 	productResp.Images = append(productResp.Images, payloads.ProductImageResp{
	// 		ID:        img.ID,
	// 		PID:       img.PID,
	// 		URL:       img.URL,
	// 		IsMain:    img.IsMain,
	// 		SortOrder: img.SortOrder,
	// 	})
	// }

	// // Copy Variants
	// for _, variant := range product.Variants {
	// 	productResp.Variants = append(productResp.Variants, payloads.ProductVariantResp{
	// 		ID:       variant.ID,
	// 		PID:      variant.PID,
	// 		SKU:      variant.SKU,
	// 		Name:     variant.Name,
	// 		Price:    variant.Price,
	// 		Quantity: variant.Quantity,
	// 		Options: []payloads.VariantOptionRequest{
	// 			{
	// 				OptionName:  variant.Options[0].OptionName,
	// 				OptionValue: variant.Options[0].OptionValue,
	// 			},
	// 		},
	// 	})
	// }

	// // Copy Attributes
	// for _, attribute := range product.Attributes {
	// 	productResp.Attributes = append(productResp.Attributes, payloads.ProductAttributeResp{
	// 		ID:    attribute.ID,
	// 		PID:   attribute.PID,
	// 		Name:  attribute.Name,
	// 		Value: attribute.Value,
	// 	})
	// }

	return productResp
}

func CopyProductRequestToProduct(productReq payloads.ProductRequest) Product {
	// Create a ProductResponse instance
	product := Product{
		ID:            productReq.ID,
		SellerID:      productReq.SellerID,
		SKU:           productReq.SKU,
		Name:          productReq.Name,
		Description:   productReq.Description,
		ShortDesc:     productReq.ShortDesc,
		Price:         productReq.Price,
		ComparePrice:  productReq.ComparePrice,
		CostPrice:     productReq.CostPrice,
		Quantity:      productReq.Quantity,
		IsFeatured:    productReq.IsFeatured,
		Rating:        productReq.Rating,
		Tax:           productReq.Tax,
		Discount:      productReq.Discount,
		DiscountType:  productReq.DiscountType,
		Weight:        productReq.Weight,
		Length:        productReq.Length,
		Width:         productReq.Width,
		Height:        productReq.Height,
		DimensionUnit: productReq.DimensionUnit,
		MetaTitle:     productReq.MetaTitle,
		MetaDesc:      productReq.MetaDesc,
		SlugURL:       productReq.SlugURL,
		MainImage:     productReq.MainImage,
	}

	// Images
	// for _, img := range productReq.Images {
	// 	product.Images = append(product.Images, ProductImage{
	// 		ID:        img.ID,
	// 		PID:       img.PID,
	// 		URL:       img.URL,
	// 		IsMain:    img.IsMain,
	// 		SortOrder: img.SortOrder,
	// 	})
	// }

	// // Variants
	// for _, variant := range productReq.Variants {
	// 	product.Variants = append(product.Variants, ProductVariant{
	// 		ID:       variant.ID,
	// 		PID:      variant.PID,
	// 		SKU:      variant.SKU,
	// 		Name:     variant.Name,
	// 		Price:    variant.Price,
	// 		Quantity: variant.Quantity,
	// 		Options: []VariantOption{
	// 			{
	// 				OptionName:  variant.Options[0].OptionName,
	// 				OptionValue: variant.Options[0].OptionValue,
	// 			},
	// 		},
	// 	})
	// }

	// // Attributes
	// for _, attribute := range productReq.Attributes {
	// 	product.Attributes = append(product.Attributes, ProductAttribute{
	// 		ID:    attribute.ID,
	// 		PID:   attribute.PID,
	// 		Name:  attribute.Name,
	// 		Value: attribute.Value,
	// 	})
	// }

	return product
}

func CopyProductImgReqToProductImg(productImgReq payloads.ProductImageRequest) ProductImage {
	return ProductImage{
		ID:        productImgReq.ID,
		PID:       productImgReq.PID,
		URL:       productImgReq.URL,
		IsMain:    productImgReq.IsMain,
		SortOrder: productImgReq.SortOrder,
	}
}

func CopyEntityToEntityResp(entity Entities) payloads.ProductEntityResponse {
	return payloads.ProductEntityResponse{
		PidEid:     entity.ID,
		Entity:     entity.EntityName,
		EntityType: entity.EntityType,
		CreatedAt:  entity.CreatedAt,
	}
}

func UploadBase64Image(imageBase64, id, imgId string) (string, error) {
	if imageBase64 == "" {
		return "", fmt.Errorf("image data is missing in the request")
	}

	var imageType string
	if strings.HasPrefix(imageBase64, "data:image/jpeg;base64,") {
		imageBase64 = strings.TrimPrefix(imageBase64, "data:image/jpeg;base64,")
		imageType = "jpeg"
	} else if strings.HasPrefix(imageBase64, "data:image/png;base64,") {
		imageBase64 = strings.TrimPrefix(imageBase64, "data:image/png;base64,")
		imageType = "png"
	} else if strings.HasPrefix(imageBase64, "data:image/jpg;base64,") {
		imageBase64 = strings.TrimPrefix(imageBase64, "data:image/jpg;base64,")
		imageType = "jpg"
	} else {
		return "", fmt.Errorf("unsupported image format")
	}
	decodedData, err := base64.StdEncoding.DecodeString(imageBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %v", err)
	}

	uploadDir := "../uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create upload directory: %v", err)
		}
	}

	filePath := filepath.Join(uploadDir, fmt.Sprintf("image-%s-%s.%s", id, imgId, imageType))
	err = os.WriteFile(filePath, decodedData, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to save image file: %v", err)
	}

	return filePath, nil
}
