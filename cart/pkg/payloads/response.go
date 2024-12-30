package payloads

type CartResponse struct {
	Items []CartItemResponse `json:"items"`
}

type CartItemResponse struct {
	Id          int                    `json:"cart_id"`
	Quantity    int                    `json:"quantity"`
	ReqQuantity int                    `json:"-"`
	Price       float64                `json:"price"`
	Product     map[string]interface{} `json:"product"`
}
