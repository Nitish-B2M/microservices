package utils

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gopkg.in/go-playground/validator.v9"
)

const UserIDKey string = "userID"
const UserNameKey string = "userName"
const ActiveRoleIDKey string = "activeRoleID"
const ActiveRoleNameKey string = "activeRoleName"

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

	productBaseURL := "http://localhost:" + os.Getenv("PRODUCT_PORT") + "/product"
	if extra != "" {
		productBaseURL = productBaseURL + extra
	}
	return productBaseURL
}

func GetCartMicroserviceLink(extra string) string {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}
	productBaseURL := "http://localhost:" + os.Getenv("CART_PORT") + "/user/cart"
	if extra != "" {
		productBaseURL = productBaseURL + extra
	}
	return productBaseURL
}

func GetPaymentMicroserviceLink(extra string) string {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}
	paymentBaseURL := "http://localhost:" + os.Getenv("PAYMENT_PORT") + "/order/%s/payment"
	if extra != "" {
		paymentBaseURL = paymentBaseURL + extra
	}
	return paymentBaseURL
}

func GetUserMicroserviceLink(extra string) string {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal("Error loading .env file from common.go")
	}
	userBaseURL := "http://localhost:" + os.Getenv("USER_PORT") + "/user"
	if extra != "" {
		userBaseURL = userBaseURL + extra
	}
	return userBaseURL
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

func RegisterSubRoutes(
	prefix string,
	r *mux.Router,
	register func(*mux.Router),
) {
	sub := r.PathPrefix(prefix).Subrouter()
	register(sub)
}

func GetUserIDFromContext(r *http.Request) int {
	userID, ok := r.Context().Value(UserIDKey).(int)
	if !ok {
		return 0
	}
	return userID
}

func GetStringUserIDFromContext(r *http.Request) string {
	userID, ok := r.Context().Value(UserIDKey).(int)
	if !ok {
		return ""
	}
	return strconv.Itoa(userID)
}

func GetUserFromGinCtx(c *gin.Context) (int, error) {
	ctxUserID, ok := c.Get(UserIDKey)
	if !ok {
		return 0, fmt.Errorf(UserIdNotFoundInCtx)
	}
	return ctxUserID.(int), nil
}

func GetUserNameIDFromContext(r *http.Request) string {
	val := r.Context().Value(UserNameKey)
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}

func ValidateStructUsingValidators(data interface{}) (bool, []string) {
	validate := validator.New()
	var validationErrors []string

	if err := validate.Struct(data); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, fmt.Sprintf("Field: '%s', Error: '%s'", e.Field(), e.Tag()))
		}
		return false, validationErrors
	}
	return true, nil
}

func GenerateSlug(text string) string {
	text = strings.ToLower(text)
	text = strings.ReplaceAll(text, " ", "-")
	reg := regexp.MustCompile("[^a-zA-Z0-9-]")
	text = reg.ReplaceAllString(text, "")
	reg = regexp.MustCompile("-+")
	text = reg.ReplaceAllString(text, "-")
	text = strings.Trim(text, "-")
	randomStr := rand.New(rand.NewSource(time.Now().UnixNano()))
	text = text + "-" + strconv.Itoa(randomStr.Intn(1000))

	if text == "" {
		return "slug-" + strconv.Itoa(randomStr.Intn(1000))
	}
	if len(text) > 100 {
		text = text[:100]
	}
	return text
}

func GenerateSKU(text string) string {
	text = strings.ToLower(text)
	text = strings.ReplaceAll(text, " ", "-")
	reg := regexp.MustCompile("[^a-zA-Z0-9-]")
	text = reg.ReplaceAllString(text, "")
	reg = regexp.MustCompile("-+")
	text = reg.ReplaceAllString(text, "-")
	text = strings.Trim(text, "-")
	return text
}

func UpdateField(existingValue interface{}, newValue interface{}) interface{} {
	// Check if the values are different
	if existingValue != newValue {
		return newValue
	}
	return existingValue
}
