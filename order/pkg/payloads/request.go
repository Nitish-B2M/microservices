package payloads

type RequestCart struct {
	Carts []map[string]interface{} `json:"carts"`
}

type OrderRequest struct {
	Id                 int                      `json:"id"`
	UserId             int                      `json:"user_id"`
	Carts              []map[string]interface{} `json:"carts"`
	TotalPrice         float64                  `json:"total_price"`
	IsPay              bool                     `json:"is_pay"`
	ShippingAddress    string                   `json:"shipping_address"`
	ShippingMethod     string                   `json:"shipping_method"`
	BillingAddress     string                   `json:"billing_address"`
	PaymentMethod      string                   `json:"payment_method"`
	PaymentTransaction string                   `json:"payment_transaction"`
	OrderDate          string                   `json:"order_date"`
	EstimatedDelivery  string                   `json:"estimated_delivery"`
	OrderStatus        string                   `json:"order_status"`
	PromoCode          string                   `json:"promo_code,omitempty"`
	Discount           float64                  `json:"discount,omitempty"`
	Tax                float64                  `json:"tax"`
}
