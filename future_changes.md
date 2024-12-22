### make change in product models, add these below field according to requirements
```type Product struct {
    ImageURL      string     `json:"image_url"`      // Default ""
    SKU            string     `json:"sku"`            // Default ""
    Manufacturer  string     `json:"manufacturer"`  // Default ""
    Brand          string     `json:"brand"`          // Default ""
    ReviewCount   int         `json:"review_count"`   // Default 0
    Weight        float64    `json:"weight"`         // Default 0.0
    Dimensions    string     `json:"dimensions"`     // Default ""
} 
```


#### this function use to update the table each row with default value for newly added column
``` func UpdateOldProductData() {
	// Query for all products (or products with NULL/missing data)
	var products []Product
	if err := dbs.DB.Find(&products).Error; err != nil {
		log.Printf("Error fetching products: %v", err)
		return
	}

	// Iterate over each product and check for fields that need to be set to default values
	for _, product := range products {
		updated := false

		// Simplified check for IsDeleted
		if !product.IsDeleted { // default value is false
			product.IsDeleted = false
			updated = true
		}

		// Check and update Discount if it's 0
		if product.Discount == 0.0 {
			product.Discount = 0.0
			updated = true
		}

		// Check and update Category if it's empty
		if product.Category == "" {
			product.Category = "" // Default is empty string
			updated = true
		}

		// Check and update Tags if it's nil or empty
		if len(product.Tags) == 0 {
			product.Tags = []string{} // Default is empty slice
			updated = true
		}

		// Check and update Rating if it's 0.0
		if product.Rating == 0.0 {
			product.Rating = 0.0 // Default is 0.0
			updated = true
		}

		if product.CreatedAt.IsZero() {
			product.CreatedAt = time.Now() // Set current time
			updated = true
		}

		// Check and set UpdatedAt if it is zero time (invalid date '0000-00-00')
		if product.UpdatedAt.IsZero() {
			product.UpdatedAt = time.Now() // Set current time
			updated = true
		}

		// If the product was updated, save it
		if updated {
			if err := dbs.DB.Save(&product).Error; err != nil {
				log.Printf("Error updating product ID %d: %v", product.ID, err)
			} else {
				log.Printf("Product ID %d updated with default values.", product.ID)
			}
		}
	}
}

```

## interesting topic
- if we want to automatically insert into 3rd table entries of table1 and table2 data then use 'gorm:"many2many:product_tags;"' into struct
- more details step
	+ type Product struct {
	+	Tags      []Tag     `json:"tags" gorm:"many2many:product_tags;"`
	+ }
	+ here we have 3 table, table1 = product, table2 = tags, and table3 = product_tags
	+ so when we have table1 product_id=123 and table2 tag_id=23 then it automatically add it into table3 product_tags like (product_id, tag_id)(123, 23)



temp

models/product.go
package models

import (
	"errors"
	"fmt"
	"log"
	"microservices/pkg/shared/dbs"
	"microservices/pkg/shared/payloads"
	"microservices/pkg/shared/utils"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID        int       `json:"id"`
	PName     string    `json:"product_name"`
	PDesc     string    `json:"product_desc"`
	Price     float64   `json:"price"`
	Quantity  int       `json:"quantity"`
	IsDeleted bool      `json:"is_deleted"`
	Discount  float64   `json:"discount"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `json:"updated_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	TagIds    string    `json:"tag_ids"`
	Rating    float64   `json:"rating"`
}

// Define the join table explicitly if needed:
type ProductTag struct {
	ProductID int `json:"product_id"`
	TagID     int `json:"tag_id"`
}

func NewProduct() *Product {
	return &Product{
		IsDeleted: false,
		Discount:  0.0,
		Category:  "",
		Rating:    0,
	}
}

func InitProductSchema() {
	db := dbs.DB
	if err := db.AutoMigrate(&Product{}, &ProductTag{}); err != nil {
		log.Fatalf(utils.DatabaseMigrationError, "Product/ProductTag", err)
	} else {
		log.Printf(utils.SchemaMigrationSuccess, "Product/ProductTag")
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

func GetProductById(id int) (*payloads.ProductResponse, error) {
	var product Product
	var resProduct payloads.ProductResponse

	db := dbs.DB
	if err := db.First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.SimpleLog("error", fmt.Sprintf(utils.ProductNotFoundError, id), err)
			return nil, fmt.Errorf(utils.ProductNotFoundError, id)
		}
		utils.SimpleLog("error", utils.UnexpectedDatabaseError, err)
		return nil, fmt.Errorf(utils.ProductUnexpectedFetchError)
	}

	if err := CopyStructIntoStruct(&product, &resProduct); err != nil {
		utils.SimpleLog("error", err.Error())
		return nil, err
	}

	var tags []Tag
	tag_ids := strings.Split(product.TagIds, ",")
	for _, tag_id := range tag_ids {
		id, _ := strconv.Atoi(strings.TrimSpace(tag_id))
		tag, _ := FetchTag(db, id)
		tags = append(tags, *tag)
	}
	resProduct.Tags = tags

	utils.SimpleLog("info", fmt.Sprintf(utils.ProductFetchedSuccessfully, id), id)
	return &resProduct, nil
}

func AddProduct(newProduct payloads.ProductRequest) (int, error) {
	var productDB Product
	db := dbs.DB
	newProduct.ID = utils.GenerateRandomID()

	tagIds, tagErrors := CheckAndCreateProductTags(db, newProduct)
	if len(tagErrors) > 0 {
		utils.SimpleLog("error", "creating tag error", tagErrors)
	}

	if err := CopyStructIntoStruct(&newProduct, &productDB); err != nil {
		log.Println("errrrrrrrr", err)
	}

	productDB.TagIds = strings.Join(tagIds, ", ")
	if err := db.Create(&productDB).Error; err != nil {
		utils.SimpleLog("error", utils.ProductCreationError, productDB.ID, err)
		return 0, fmt.Errorf(utils.ProductCreationError)
	}

	errs := AddProductTag(db, tagIds, newProduct.ID)
	if len(errs) > 0 {
		for _, err := range errs {
			utils.SimpleLog("error", err.Error())
		}
	}

	utils.SimpleLog("info", fmt.Sprintf(utils.ProductCreatedSuccessfully, newProduct.ID), newProduct.ID)
	return newProduct.ID, nil
}

func AddProductTag(db *gorm.DB, tagIds []string, productID int) []error {
	var errs []error
	for _, tagID := range tagIds {
		tagID, err := strconv.Atoi(tagID)
		if err != nil {

		}
		productTag := ProductTag{
			ProductID: productID,
			TagID:     tagID,
		}
		if err := db.Create(&productTag).Error; err != nil {
			errs = append(errs, fmt.Errorf("failed to create product[%d]-tag[%d] association: %v", tagID, productID, err))
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
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

	updatedFields := make(map[string]interface{})
	v := reflect.ValueOf(&newProduct).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		fieldName := field.Name
		fieldValue := v.Field(i)

		if fieldValue.IsZero() {
			continue
		}

		if fieldName == "Tags" {
			updatedFields[fieldName] = fieldValue.Interface()
			reflect.ValueOf(&oldProduct).Elem().FieldByName(fieldName).Set(fieldValue)
		} else if fieldName != "ID" && !reflect.DeepEqual(fieldValue.Interface(), reflect.ValueOf(&oldProduct).Elem().FieldByName(fieldName).Interface()) {
			updatedFields[fieldName] = fieldValue.Interface()
			reflect.ValueOf(&oldProduct).Elem().FieldByName(fieldName).Set(fieldValue)
		}
	}

	if len(updatedFields) > 0 {
		if err := db.Save(&oldProduct).Error; err != nil {
			utils.SimpleLog("error", fmt.Sprintf(utils.ProductUpdateError, id), err)
			return nil, fmt.Errorf(utils.ProductUpdateError, id)
		}
		utils.SimpleLog("info", fmt.Sprintf(utils.ProductUpdatedSuccessfully, id), updatedFields)
	} else {
		utils.SimpleLog("info", fmt.Sprintf(utils.ProductNotModified, id), nil)
	}
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


product/services.go
package product

import (
	"encoding/json"
	"fmt"
	"microservices/pkg/shared/models"
	"microservices/pkg/shared/payloads"
	"microservices/pkg/shared/utils"
	"net/http"
	"strings"
)

func HandleProductRequest() {
	http.HandleFunc("/products", getProducts)
	http.HandleFunc("/products/{id}", getProductById)
	http.HandleFunc("/products/add", addProduct)
	http.HandleFunc("/products/update/{id}", updateProduct)
	http.HandleFunc("/products/delete/{id}", deleteProduct)
}

func getProducts(w http.ResponseWriter, r *http.Request) {
	products, err := models.GetProducts()
	if err != nil {
		utils.JsonError(w, utils.ProductNotFoundError, http.StatusNotFound, err)
	}

	utils.JsonResponse(products, w, utils.ProductsFetchedSuccessfully, http.StatusOK)
}

func getProductById(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetProductIdFromPath(r)
	if err != nil {
		utils.JsonError(w, utils.InvalidProductIDError, http.StatusBadRequest, err)
		return
	}

	product, err := models.GetProductById(id)
	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.ProductNotFoundError, id), http.StatusNotFound, err)
		return
	}

	utils.JsonResponse(product, w, fmt.Sprintf(utils.ProductFetchedSuccessfully, id), http.StatusOK)
}

func addProduct(w http.ResponseWriter, r *http.Request) {
	if !utils.CheckRequestMethod(w, r, http.MethodPost) {
		return
	}

	var newProduct payloads.ProductRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&newProduct); err != nil {
		utils.JsonError(w, utils.InvalidRequestBody, http.StatusBadRequest, err)
		return
	}

	if newProduct.PName == "" || newProduct.Price == 0 {
		utils.JsonError(w, utils.InvalidProductDataError, http.StatusBadRequest, nil)
		return
	}

	id, err := models.AddProduct(newProduct)
	if err != nil {
		utils.JsonError(w, utils.ProductCreationError, http.StatusInternalServerError, err)
		return
	}

	createdProduct, err := models.GetProductById(id)
	if err != nil {
		utils.JsonError(w, utils.ProductNotFoundError, http.StatusNotFound, nil)
		return
	}

	utils.JsonResponse(createdProduct, w, fmt.Sprintf(utils.ProductCreatedSuccessfully, id), http.StatusCreated)
}

func updateProduct(w http.ResponseWriter, r *http.Request) {
	if !utils.CheckRequestMethod(w, r, http.MethodPut) {
		return
	}

	id, err := utils.GetProductIdFromPath(r)
	if err != nil {
		utils.JsonError(w, utils.InvalidProductIDError, http.StatusBadRequest, err)
		return
	}

	var oldProduct models.Product
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&oldProduct); err != nil {
		utils.JsonError(w, utils.InvalidRequestBody, http.StatusBadRequest, err)
		return
	}

	updatedProduct, err := models.UpdateProduct(id, oldProduct)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.JsonError(w, fmt.Sprintf(utils.ProductNotFoundError, id), http.StatusInternalServerError, err)
			return
		}
		utils.JsonError(w, fmt.Sprintf(utils.ProductUpdateError, id), http.StatusInternalServerError, err)
		return
	}

	utils.JsonResponse(updatedProduct, w, fmt.Sprintf(utils.ProductUpdatedSuccessfully, updatedProduct.ID), http.StatusOK)
}

func deleteProduct(w http.ResponseWriter, r *http.Request) {
	if !utils.CheckRequestMethod(w, r, http.MethodDelete) {
		return
	}

	id, err := utils.GetProductIdFromPath(r)
	if err != nil {
		utils.JsonError(w, utils.InvalidProductIDError, http.StatusBadRequest, err)
		return
	}

	deletedProduct, err := models.DeleteProduct(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.JsonError(w, fmt.Sprintf(utils.ProductNotFoundError, id), http.StatusInternalServerError, err)
			return
		}
		utils.JsonError(w, fmt.Sprintf(utils.ProductDeletionError, id), http.StatusInternalServerError, err)
		return
	}

	utils.JsonResponse(deletedProduct, w, fmt.Sprintf(utils.ProductDeletedSuccessfully, deletedProduct.ID), http.StatusOK)
}
