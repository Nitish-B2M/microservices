package services

import (
	"e-commerce-backend/products/internal/models"
	"e-commerce-backend/products/pkg/payloads"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func FilterOutProducts(criteria FilterCriteria) []payloads.ProductResponse {
	products, _ := models.GetProducts()
	filteredProducts := FilterProducts(products, criteria)
	return filteredProducts
}

func FilterProducts(products []payloads.ProductResponse, criteria FilterCriteria) []payloads.ProductResponse {
	var filteredProducts []payloads.ProductResponse

	filterTagSet := make(map[string]struct{})
	for _, tag := range criteria.Tags {
		filterTagSet[tag] = struct{}{}
	}

	for _, product := range products {
		if criteria.PName != "" && !containsSubStr(product.PName, criteria.PName) {
			continue
		}

		if product.Price < criteria.MinPrice || product.Price > criteria.MaxPrice {
			continue
		}

		if product.Rating < criteria.MinRating {
			continue
		}

		if criteria.Category != "" && product.Category != criteria.Category {
			continue
		}

		if product.IsDeleted != criteria.IsDeleted {
			continue
		}

		if product.Quantity < criteria.MinQuantity || product.Quantity > criteria.MaxQuantity {
			continue
		}

		if len(criteria.Tags) > 0 && !hasMatchingTags(product.Tags.([]string), filterTagSet) {
			continue
		}

		filteredProducts = append(filteredProducts, product)
	}

	return filteredProducts
}

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

func HasKey(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
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
