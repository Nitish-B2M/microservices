package utils

import "log"

func LogInfo(message string) {
	log.Printf("[INFO] %s", message)
}

func LogError(message string) {
	log.Printf("[ERROR] %s", message)
}
