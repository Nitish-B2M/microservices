// Package utils ImageUpload
package utils

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func UploadBase64Image(imageBase64, id, imgID string) (string, error) {
	if imageBase64 == "" {
		return "", fmt.Errorf("image data is missing in the request")
	}

	base64Data := strings.TrimPrefix(imageBase64, "data:image/jpeg;base64,")
	decodedData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %v", err)
	}

	uploadDir := "./uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create upload directory: %v", err)
		}
	}

	filePath := filepath.Join(uploadDir, fmt.Sprintf("image-%s-%s.jpg", id, imgID))
	err = os.WriteFile(filePath, decodedData, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to save image file: %v", err)
	}

	return filePath, nil
}
