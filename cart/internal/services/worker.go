package services

import (
	"bytes"
	"e-commerce-backend/cart/pkg/constants"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type UpdateTask struct {
	ProductID int
	Quantity  int
	Method    string
}

var UpdateChannel = make(chan UpdateTask, 100)

func StartProductUpdateQuantityWorker(token string) {
	for task := range UpdateChannel {
		log.Printf(constants.ProcessingProductQuantityUpdate, task.ProductID, task.Quantity)

		callbackURL := fmt.Sprintf(constants.ProductMicroserviceUpdateQuantityCall, task.ProductID)
		payload := map[string]interface{}{"quantity": task.Quantity, "method": task.Method}
		jsonPayload, _ := json.Marshal(payload)
		req, err := http.NewRequest("POST", callbackURL, bytes.NewBuffer(jsonPayload))
		if err != nil {
			log.Printf(constants.FailedToCreateRequest, task.ProductID, err)
			continue
		}
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf(constants.FailedToUpdateProductQuantity, task.ProductID, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf(constants.FailedToUpdateProductQuantityWithStatus, task.ProductID, resp.StatusCode)
			continue
		}

		log.Printf(constants.SuccessfullyUpdatedProductQuantity, task.ProductID)
	}
}
