package constants

import "e-commerce-backend/shared/utils"

const (
	CartMicroserviceCallById    = "/%d"
	ProductMicroserviceCallById = "/%d/cart"
	PaymentMicroserviceCallById = "/initiate"
)

func MicroserviceLinks() map[string]string {
	links := map[string]string{}

	productCallByIdLink := utils.GetProductMicroserviceLink(ProductMicroserviceCallById)
	links["productMSCallByIdLink"] = productCallByIdLink

	cartCallByIdLink := utils.GetCartMicroserviceLink(CartMicroserviceCallById)
	links["cartMSCallByIdLink"] = cartCallByIdLink

	paymentCallByIdLink := utils.GetPaymentMicroserviceLink(PaymentMicroserviceCallById)
	links["paymentMSInitiateCallLink"] = paymentCallByIdLink
	return links
}
