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
)

type contextKey string

var typeRegistry = map[string]reflect.Type{
	"UserLoginUsingEmailValidator":    reflect.TypeOf(UserLoginUsingEmailValidator{}),
	"UserLoginUsingUsernameValidator": reflect.TypeOf(UserLoginUsingUsernameValidator{}),
}

func Validator(schema string) func(http.Handler) http.Handler {
	schema = strings.TrimSpace(schema)
	if schema == "" {
		return nil
	}

	schemaType, exists := typeRegistry[schema]
	if !exists {
		return nil
	}
	fmt.Println(schemaType)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bodyBytes, err := io.ReadAll(r.Body)
			fmt.Printf("%+v", r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}
			defer r.Body.Close()

			if len(bodyBytes) == 0 {
				http.Error(w, "Empty request body", http.StatusBadRequest)
				return
			}

			schemaInstance := reflect.New(schemaType).Interface()
			if err := json.Unmarshal(bodyBytes, schemaInstance); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			// Replace body so the next handler can read it again
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			ctx := context.WithValue(r.Context(), contextKey(schema), schemaInstance)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
