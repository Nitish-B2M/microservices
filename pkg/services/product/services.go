package product

import (
	"encoding/json"
	"microservices/pkg/shared/models"
	"microservices/pkg/shared/utils"
	"net/http"
)

func HandleProductRequest() {
	http.HandleFunc("/products", getProduct)
	// http.HandleFunc("/products/{id}", getProductById)
	http.HandleFunc("/products/add", addProduct)
	// http.HandleFunc("/products/update/{id}", updateProduct)
	// http.HandleFunc("/products/delete/{id}", deleteProduct)
}

func getProduct(w http.ResponseWriter, r *http.Request) {
	products := models.NewProduct()
	utils.JsonResponse(products, w)
}

func addProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var newProduct models.Product
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&newProduct); err != nil {
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return
	}
	id, err := models.AddProduct(newProduct)
	if err != nil {
		http.Error(w, "Failed to add product", http.StatusInternalServerError)
		return
	}
	newProduct.ID = id
	utils.JsonResponse(newProduct, w)
}
