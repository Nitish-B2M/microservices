package utils

import (
	"encoding/json"
	"net/http"
)

func JsonResponse(data interface{}, w http.ResponseWriter, message string, status int) {
	if status == 0 {
		status = http.StatusOK
	}

	res, err := json.Marshal(data)
	if err != nil {
		LogError("Error marshaling data", map[string]interface{}{"error": err})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	LogInfo("Sending response", map[string]interface{}{"response": string(res)})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := map[string]interface{}{
		"message": message,
		"data":    data,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		LogError("Error encoding final response", map[string]interface{}{"error": err})
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
	}
}

func JsonError(w http.ResponseWriter, message string, status int, err error) {
	LogError(message, map[string]interface{}{"error": err})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
