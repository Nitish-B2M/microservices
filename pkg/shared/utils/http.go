package utils

import (
	"encoding/json"
	"net/http"
)

func JsonResponse(data interface{}, w http.ResponseWriter) {
	res, _ := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}
