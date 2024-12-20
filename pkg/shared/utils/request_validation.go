package utils

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func ValidateRequestPath(path string) ([]string, bool) {
	parts := strings.Split(path, "/")

	if len(parts) < 3 {
		return []string{}, false
	}

	if len(parts[2]) < 2 || parts[2] == "" {
		return []string{}, false
	}

	if len(parts) >= 4 && (len(parts[3]) < 2 || parts[3] == "") {
		return []string{}, false
	}

	return parts, true
}

func GetProductIdFromPath(r *http.Request) (int, error) {
	path := r.URL.Path
	parts, ok := ValidateRequestPath(path)
	if !ok {
		return 0, fmt.Errorf("invalid path")
	}
	id, err := strconv.Atoi(parts[len(parts)-1])
	return id, err
}

func CheckRequestMethod(w http.ResponseWriter, r *http.Request, expectedMethod string) bool {
	if r.Method != expectedMethod {
		JsonError(w, InvalidRequestMethod, http.StatusMethodNotAllowed, nil)
		return false
	}
	return true
}
