package utils

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
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

	res1 := shortResponseData(res)
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

	res1 := shortResponseData(res)
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

func shortResponseData(message []byte) string {
	res1 := ""
	res := string(message)
	if len(res) > 100 {
		res1 = strings.TrimSpace(res[:100]) + "...."
	} else {
		res1 = res
	}
	return res1
}

func GinError(c *gin.Context, message string, status int, err error) {
	LogError(message, map[string]interface{}{"error": err, "status": status})
	c.JSON(status, gin.H{
		"message": message,
	})
}

func GinErrorWithExtra(c *gin.Context, message string, status int, err error, filename string) {
	LogErrorWithFilename(filename, message, map[string]interface{}{"error": err})
	c.JSON(status, gin.H{
		"message":     message,
		"errorDetail": err.Error(),
	})
}

func GinResponse(data interface{}, c *gin.Context, message string, status int) {
	res, err := json.Marshal(data)
	if err != nil {
		LogError("Error marshalling response data", map[string]interface{}{"error": err})
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to process response data"})
		return
	}

	res1 := shortResponseData(res)

	LogInfo("sending response", map[string]interface{}{"response": res1})
	if status == 0 {
		status = http.StatusOK
	}

	c.JSON(status, gin.H{
		"message": message,
		"data":    data,
	})
}

func ParseJSON(data []byte, v interface{}) error {
	err := json.Unmarshal(data, v)
	if err != nil {
		log.Println("Error unmarshalling JSON:", err)
		return err
	}
	return nil
}
