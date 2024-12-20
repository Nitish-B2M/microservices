package middleware

import (
	"log"
	"net/http"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request from %s for %s", r.RemoteAddr, r.URL)
		next.ServeHTTP(w, r)
	})
}
