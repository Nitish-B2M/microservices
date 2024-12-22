package utils

import (
	"encoding/json"
	"net/http"
	"strings"
)

func JsonResponse(data interface{}, w http.ResponseWriter, message string, status int) {
	if status == 0 {
		status = http.StatusOK
	}

	res, err := json.Marshal(data)
	if err != nil {
		LogError("error marshaling data", map[string]interface{}{"error": err})
		http.Error(w, InternalServerError, http.StatusInternalServerError)
		return
	}

	res1 := ""
	if len(string(res)) > 100 {
		res1 = strings.TrimSpace(string(res[:100])) + "...."
	} else {
		res1 = string(res)
	}
	LogInfo("sending response", map[string]interface{}{"response": res1})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := map[string]interface{}{
		"message": message,
		"data":    data,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		LogError("error while encoding final response", map[string]interface{}{"error": err})
		http.Error(w, FailedToSendResponse, http.StatusInternalServerError)
	}
}

func JsonError(w http.ResponseWriter, message string, status int, err error) {
	LogError(message, map[string]interface{}{"error": err})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
