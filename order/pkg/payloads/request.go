package payloads

type RequestCart struct {
	Carts []RequestCartItem `json:"carts"`
}
type RequestCartItem struct {
	CartID int `json:"cart_id"`
}

type OrderRequest struct {
	Carts             []RequestCartItem `json:"carts"`
	TotalPrice        float64           `json:"total_price"`
	BillingAddressId  string            `json:"billing_address_id"`
	BillingAddress    string            `json:"billing_address"`
	BillingUserId     int               `json:"billing_user_id"`
	BillingContact    string            `json:"billing_contact"`
	ShippingCost      float64           `json:"shipping_cost"`
	ShippingAddressId string            `json:"shipping_address_id"`
	ShippingAddress   string            `json:"shipping_address"`
	ShippingContact   string            `json:"shipping_contact"`
	PromoCode         string            `json:"promo_code,omitempty"`
	DiscountAmount    float64           `json:"discount_amount,omitempty"`
	TaxAmount         float64           `json:"tax_amount,omitempty"`
	TotalAmount       float64           `json:"total_amount,omitempty"`
}

//
//example request
//"carts": [
//	{
//		"cart_id": 1
//	},
//	{
//		"cart_id": 2
//	}
//],
//"total_price": 100,
//"billing_address": "123 Main St",
//"billing_user_id": 1,
//"billing_contact": "John Doe",
//"shipping_cost": 5,
//"shipping_address": "456 Elm St",
//"shipping_contact": "Jane Doe",
//"promo_code": "ABC123",
//"discount_amount": 10,
//"tax_amount": 2,
//"total_amount": 98
