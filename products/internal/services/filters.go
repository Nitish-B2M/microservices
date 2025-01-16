package services

import (
	"e-commerce-backend/products/internal/models"
	"e-commerce-backend/shared/utils"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

func (db *Service) FilterProducts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters into a FilterCriteria object
	criteria := models.FilterCriteria{}
	query := r.URL.Query()
	log.Println(query)
	//remove ? from query
	cleanedQuery := make(map[string][]string)

	for k, v := range query {
		if strings.HasPrefix(k, "?") {
			k = k[1:]
		}
		cleanedQuery[k] = v
	}

	// List of allowed filter keys
	FilterKey := []string{
		"name", "min_price", "max_price", "min_rating", "category",
		"min_quantity", "max_quantity", "tags",
	}

	// Iterate over query params and set them if they are valid
	for k, v := range cleanedQuery {
		if isValidFilterKey(FilterKey, k) && v[0] != "" {
			if err := setField(&criteria, k, v[0]); err != nil {
				utils.SimpleLog("error", fmt.Sprintf("Error setting field %s: %v", k, err))
			}
		}
	}
	log.Printf("%+v", criteria)

	// Fetch filtered products and return response
	resp, _ := models.FilterProduct(db.DB, criteria)
	utils.JsonResponse(resp, w, "filtered data", http.StatusOK)
}

// Check if a key is in the valid filter keys
func isValidFilterKey(validKeys []string, key string) bool {
	for _, validKey := range validKeys {
		if validKey == key {
			return true
		}
	}
	return false
}

// Set the field value in the FilterCriteria struct based on its type
func setField(obj interface{}, fieldName string, value string) error {
	v := reflect.ValueOf(obj).Elem()

	// Iterate over struct fields to find a match
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		jsonTag := field.Tag.Get("json")

		// Check if the field matches by json tag or camelCase
		if jsonTag == fieldName || toLowerCamelCase(field.Name) == fieldName {
			fieldValue := v.Field(i)
			if !fieldValue.CanSet() {
				continue
			}

			// Set the field value based on its type
			switch fieldValue.Kind() {
			case reflect.String:
				fieldValue.SetString(value)
			case reflect.Float64:
				if val, err := strconv.ParseFloat(value, 64); err == nil {
					fieldValue.SetFloat(val)
				} else {
					return fmt.Errorf("invalid float value for field %s", fieldName)
				}
			case reflect.Int:
				if val, err := strconv.Atoi(value); err == nil {
					fieldValue.SetInt(int64(val))
				} else {
					return fmt.Errorf("invalid int value for field %s", fieldName)
				}
			case reflect.Bool:
				if val, err := strconv.ParseBool(value); err == nil {
					fieldValue.SetBool(val)
				} else {
					return fmt.Errorf("invalid bool value for field %s", fieldName)
				}
			case reflect.Slice:
				if fieldName == "tags" {
					fieldValue.Set(reflect.ValueOf(strings.Split(value, ",")))
				}
			default:
				return fmt.Errorf("unsupported field type: %s", fieldName)
			}
			return nil
		}
	}
	return fmt.Errorf("no such field: %s", fieldName)
}

//func FilterProducts(products []payloads.ProductResponse, criteria FilterCriteria) []payloads.ProductResponse {
//	var filteredProducts []payloads.ProductResponse
//
//	filterTagSet := make(map[string]struct{})
//	for _, tag := range criteria.Tags {
//		filterTagSet[tag] = struct{}{}
//	}
//
//	for _, product := range products {
//		if criteria.PName != "" && !containsSubStr(product.PName, criteria.PName) {
//			continue
//		}
//
//		if product.Price < criteria.MinPrice || product.Price > criteria.MaxPrice {
//			continue
//		}
//
//		if product.Rating < criteria.MinRating {
//			continue
//		}
//
//		if criteria.Category != "" && product.Category != criteria.Category {
//			continue
//		}
//
//		if product.IsDeleted != criteria.IsDeleted {
//			continue
//		}
//
//		if product.Quantity < criteria.MinQuantity || product.Quantity > criteria.MaxQuantity {
//			continue
//		}
//
//		if len(criteria.Tags) > 0 && !hasMatchingTags(product.Tags.([]string), filterTagSet) {
//			continue
//		}
//
//		filteredProducts = append(filteredProducts, product)
//	}
//	log.Println("filteredProducts:", filteredProducts)
//
//	return filteredProducts
//}

func containsSubStr(str, substr string) bool {
	return len(str) >= len(substr) && str[:len(substr)] == substr
}

func hasMatchingTags(productTags []string, filterTagSet map[string]struct{}) bool {
	for _, productTag := range productTags {
		if _, found := filterTagSet[productTag]; found {
			return true
		}
	}
	return false
}

func HasKey(validKeys []string, key string) bool {
	for _, s := range validKeys {
		if s == key {
			return true
		}
	}
	return false
}

func SetField(obj interface{}, fieldName string, value string) error {
	v := reflect.ValueOf(obj).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		structFieldName := field.Name
		jsonTag := field.Tag.Get("json")

		if jsonTag == fieldName || toLowerCamelCase(structFieldName) == fieldName {
			fieldValue := v.Field(i)
			if !fieldValue.CanSet() {
				continue
			}

			switch fieldValue.Kind() {
			case reflect.String:
				fieldValue.SetString(value)
			case reflect.Float64:
				if val, err := strconv.ParseFloat(value, 64); err == nil {
					fieldValue.SetFloat(val)
				} else {
					return fmt.Errorf("invalid float value for field %s", fieldName)
				}
			case reflect.Int:
				if val, err := strconv.Atoi(value); err == nil {
					fieldValue.SetInt(int64(val))
				} else {
					return fmt.Errorf("invalid int value for field %s", fieldName)
				}
			case reflect.Bool:
				if val, err := strconv.ParseBool(value); err == nil {
					fieldValue.SetBool(val)
				} else {
					return fmt.Errorf("invalid bool value for field %s", fieldName)
				}
			case reflect.Slice:
				if fieldName == "tags" {
					fieldValue.Set(reflect.ValueOf(strings.Split(value, ",")))
				}
			default:
				return fmt.Errorf("unsupported field type: %s", fieldName)
			}
			return nil
		}
	}
	return fmt.Errorf("no such field: %s", fieldName)
}

func toLowerCamelCase(s string) string {
	parts := strings.Split(s, "_")
	if len(parts) == 0 {
		return ""
	}
	for i := 0; i < len(parts); i++ {
		if i > 0 {
			parts[i] = strings.ToUpper(string(parts[i][0])) + parts[i][1:]
		} else {
			parts[i] = strings.ToLower(parts[i])
		}
	}
	return strings.Join(parts, "")
}

func (db *Service) GetCategories(w http.ResponseWriter, r *http.Request) {
	var products []models.Product
	category := r.URL.Query().Get("category")

	// Validate category input
	if category == "" {
		utils.JsonError(w, utils.CategoryParamRequired, http.StatusBadRequest, nil)
		return
	}

	err := db.DB.Where("category = ?", category).Find(&products).Error
	if err != nil {
		utils.JsonError(w, utils.ProductsFetchError, http.StatusInternalServerError, err)
		return
	}

	// Return products belonging to the category
	utils.JsonResponseWithExtra(products, w, fmt.Sprintf(utils.ProductCategoryFetchedSuccessfully, category), http.StatusOK, "getCategories")
}

func (db *Service) GetFeatured(w http.ResponseWriter, r *http.Request) {
	var products []models.Product

	// Fetch featured products
	err := db.DB.Where("is_featured = ?", true).Find(&products).Error
	if err != nil {
		utils.JsonError(w, utils.ProductsFetchError, http.StatusInternalServerError, err)
		return
	}

	utils.JsonResponseWithExtra(products, w, utils.ProductsFetchedSuccessfully, http.StatusOK, "getFeatured")
}

func (db *Service) GetOffers(w http.ResponseWriter, r *http.Request) {
	var products []models.Product

	// Fetch products with offers (discount > 0)
	err := db.DB.Where("discount > ?", 0).Find(&products).Error
	if err != nil {
		utils.JsonError(w, utils.ProductsFetchError, http.StatusInternalServerError, err)
		return
	}

	utils.JsonResponseWithExtra(products, w, "Products with offers fetched successfully", http.StatusOK, "getOffers")
}

func (db *Service) GetDeals(w http.ResponseWriter, r *http.Request) {
	var products []models.Product

	// Fetch products with discounts (deals)
	err := db.DB.Where("discount >= ?", 20).Find(&products).Error
	if err != nil {
		utils.JsonError(w, utils.ProductsFetchError, http.StatusInternalServerError, err)
		return
	}

	utils.JsonResponseWithExtra(products, w, "Deals fetched successfully", http.StatusOK, "getDeals")
}
