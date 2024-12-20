package utils

import (
	"fmt"
	"log"
)

func LogInfo(message string, context map[string]interface{}) {
	log.Printf("[INFO] %s | Context: %+v", message, context)
}

func LogInfoString(message string, context string) {
	log.Printf("[INFO] %s | Context: %+v", message, context)
}

func LogError(message string, context map[string]interface{}) {
	log.Printf("[ERROR] %s | Context: %+v", message, context)
}

func LogErrorString(message string, context string) {
	log.Printf("[ERROR] %s | Context: %+v", message, context)
}

func SimpleLog(level, message string, context ...interface{}) {
	var logMessage string
	if len(context) > 0 {
		logMessage = fmt.Sprintf("%s | Context: %+v", message, context)
	} else {
		logMessage = message
	}

	// Log based on the level
	switch level {
	case "info":
		log.Printf("[INFO] %s", logMessage)
	case "error":
		log.Printf("[ERROR] %s", logMessage)
	}
}
