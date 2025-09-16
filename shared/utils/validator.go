// Package utils validator
package utils

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

func ValidatorMiddleware(schema interface{}) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			validate := validator.New()

			if err := validate.Struct(schema); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "Invalid request: %v", err)
				ErrorResponseFunc(w, "Invalid request", http.StatusBadRequest, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
