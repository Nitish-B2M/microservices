package utils

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"reflect"
)

const UserIDKey string = "userID"

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
	productBaseUrl := "http://localhost:" + os.Getenv("CART_PORT") + "/user/cart"
	if extra != "" {
		productBaseUrl = productBaseUrl + extra
	}
	return productBaseUrl
}

func GetPaymentMicroserviceLink(extra string) string {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}
	paymentBaseUrl := "http://localhost:" + os.Getenv("PAYMENT_PORT") + "/order/%s/payment"
	if extra != "" {
		paymentBaseUrl = paymentBaseUrl + extra
	}
	return paymentBaseUrl
}

func GetUserMicroserviceLink(extra string) string {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal("Error loading .env file from common.go")
	}
	userBaseUrl := "http://localhost:" + os.Getenv("USER_PORT") + "/user"
	if extra != "" {
		userBaseUrl = userBaseUrl + extra
	}
	return userBaseUrl
}

func ErrorsToString(errs []error) string {
	errStr := ""
	for i, err := range errs {
		if i != len(errs)-1 {
			errStr += err.Error()
		} else {
			errStr += err.Error() + ", "
		}
	}
	return errStr
}

func GetUserIdFromContext(r *http.Request) int {
	userId, ok := r.Context().Value(UserIDKey).(int)
	if !ok {
		return 0
	}
	return userId
}

func GetUserFromGinCtx(c *gin.Context) (int, error) {
	ctxUserId, ok := c.Get(UserIDKey)
	if !ok {
		return 0, fmt.Errorf(UserIdNotFoundInCtx)
	}
	return ctxUserId.(int), nil
}
