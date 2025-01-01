package utils

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"reflect"
)

func MapStructFields(src interface{}, dest interface{}) error {
	srcValue := reflect.ValueOf(src)
	destValue := reflect.ValueOf(dest)

	if srcValue.Kind() != reflect.Ptr || destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("src and dest must be pointers to structs")
	}

	srcValue = srcValue.Elem()
	destValue = destValue.Elem()

	for i := 0; i < srcValue.NumField(); i++ {
		srcField := srcValue.Field(i)
		destField := destValue.FieldByName(srcValue.Type().Field(i).Name)
		if destField.IsValid() && destField.CanSet() {
			if !srcField.IsZero() {
				destField.Set(srcField)
			}
		}
	}
	return nil
}

func GetProductMicroserviceLink(extra string) string {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	productBaseUrl := "http://localhost:" + os.Getenv("PRODUCT_PORT") + "/product"
	if extra != "" {
		productBaseUrl = productBaseUrl + extra
	}
	return productBaseUrl
}

func GetCartMicroserviceLink(extra string) string {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}
	productBaseUrl := "http://localhost:" + os.Getenv("CART_PORT") + "/user/%d/cart"
	if extra != "" {
		productBaseUrl = productBaseUrl + extra
	}
	return productBaseUrl
}

func GetPaymentMicroserviceLink(extra string) string {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}
	paymentBaseUrl := "http://localhost:" + os.Getenv("PAYMENT_PORT") + "/order/%d/payment"
	if extra != "" {
		paymentBaseUrl = paymentBaseUrl + extra
	}
	return paymentBaseUrl
}
