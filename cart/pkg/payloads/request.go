package payloads

type CartRequest struct {
	Items []CartItem `json:"items"`
}

type CartItem struct {
	Id        int `json:"cart_id"`
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type CartRemoveRequest struct {
	Items []CartRemoveItem `json:"items"`
}

type CartRemoveItem struct {
	Id        int `json:"cart_id"`
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}
