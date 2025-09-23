// Package validator provides validation rules for user-related data.
package validator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Custom context key type to avoid collisions
type contextKey string

// Create a shared validator instance
var validate = validator.New()

func makeContextKey(schema string) contextKey {
	return contextKey(schema)
}

// Schema type registry for dynamic binding
var typeRegistry = map[string]reflect.Type{
	"UserLoginUsingEmailValidator":    reflect.TypeOf(UserLoginUsingEmailValidator{}),
	"UserLoginUsingUsernameValidator": reflect.TypeOf(UserLoginUsingUsernameValidator{}),
	"CreateUserValidator":             reflect.TypeOf(CreateUserValidator{}),
	"SendVerificationEmailValidator":  reflect.TypeOf(SendVerificationEmailValidator{}),
	"EmailVerificationValidator":      reflect.TypeOf(EmailVerificationValidator{}),
	"FetchUserProfileValidator":       reflect.TypeOf(FetchUserProfileValidator{}),
	"LoginUserValidator":              reflect.TypeOf(LoginUserValidator{}),
	"SwitchRoleValidator":             reflect.TypeOf(SwitchRoleValidator{}),
}

// Validator function that binds and validates JSON request bodies
func Validator(schema string) func(http.Handler) http.Handler {
	schema = strings.TrimSpace(schema)
	if schema == "" {
		return nil
	}

	schemaType, exists := typeRegistry[schema]
	if !exists {
		return nil
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}
			defer r.Body.Close()

			if len(bodyBytes) == 0 {
				http.Error(w, "Empty request body", http.StatusBadRequest)
				return
			}

			// Create a new instance of the schema type
			schemaInstance := reflect.New(schemaType).Interface()
			// Unmarshal request JSON into the instance
			if err := json.Unmarshal(bodyBytes, schemaInstance); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			// ✅ Perform validation
			if err := validate.Struct(schemaInstance); err != nil {
				validationErrors := err.(validator.ValidationErrors)
				http.Error(w, formatValidationErrors(validationErrors), http.StatusBadRequest)
				return
			}

			// Replace the body so the next handler can read it
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			// Attach validated struct to context
			ctx := context.WithValue(r.Context(), contextKey(schema), schemaInstance)
			r = r.WithContext(ctx)

			// Proceed to the next handler
			next.ServeHTTP(w, r)
		})
	}
}

func formatValidationErrors(verrs validator.ValidationErrors) string {
	var sb strings.Builder
	for _, err := range verrs {
		sb.WriteString(fmt.Sprintf("Field '%s' failed on the '%s' rule\n", err.Field(), err.Tag()))
	}
	return sb.String()
}

func GetValidatedStruct[T any](ctx context.Context) (*T, error) {
	var typeName string
	var t T
	typeName = reflect.TypeOf(t).Name()

	val := ctx.Value(makeContextKey(typeName))
	if val == nil {
		return nil, fmt.Errorf("validated data for type %s not found in context", typeName)
	}

	typedVal, ok := val.(*T)
	if !ok {
		return nil, fmt.Errorf("context value is not of expected type *%s", typeName)
	}

	return typedVal, nil
}
