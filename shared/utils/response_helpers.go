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

	w.Header().Set("Content-Type", "application/json")
	cw := NewCustomResponseWriter(w)
	cw.WriteHeader(status)

	res, err := json.Marshal(data)
	if err != nil {
		LogError("error marshaling data", map[string]interface{}{"error": err})
		http.Error(cw, InternalServerError, http.StatusInternalServerError)
		return
	}

	res1 := ""
	if len(string(res)) > 100 {
		res1 = strings.TrimSpace(string(res[:100])) + "...."
	} else {
		res1 = string(res)
	}
	LogInfo("sending response", map[string]interface{}{"response": res1})

	resp := map[string]interface{}{
		"message": message,
		"data":    data,
	}

	if status != 304 {
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			LogError("error while encoding final response", map[string]interface{}{"error": err})
			http.Error(cw, FailedToSendResponse, http.StatusInternalServerError)
		}
	}
}

func JsonResponseWithError(data interface{}, w http.ResponseWriter, message string, status int, errors []error) {
	if status == 0 {
		status = http.StatusOK
	}

	w.Header().Set("Content-Type", "application/json")
	cw := NewCustomResponseWriter(w)
	cw.WriteHeader(status)

	res, err := json.Marshal(data)
	if err != nil {
		LogError("error marshaling data", map[string]interface{}{"error": err})
		http.Error(cw, InternalServerError, http.StatusInternalServerError)
		return
	}

	res1 := ""
	if len(string(res)) > 100 {
		res1 = strings.TrimSpace(string(res[:100])) + "...."
	} else {
		res1 = string(res)
	}
	LogInfo("sending response2", map[string]interface{}{"response": res1, "errors": errors})

	var errList []string
	if len(errors) > 0 {
		for _, err := range errors {
			errList = append(errList, err.Error())
		}
	}

	resp := map[string]interface{}{
		"message": message,
		"data":    data,
		"error":   errList,
	}

	if status != 304 {
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			LogError("error while encoding final response", map[string]interface{}{"error": err})
			http.Error(cw, FailedToSendResponse, http.StatusInternalServerError)
		}
	}
}

func JsonError(w http.ResponseWriter, message string, status int, err error) {
	LogError(message, map[string]interface{}{"error": err})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func JsonErrorWithExtra(w http.ResponseWriter, message string, status int, err error) {
	LogError(message, map[string]interface{}{"error": err})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{"message": message, "errorDetail": err.Error()})
}
