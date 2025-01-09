package constants

import "e-commerce-backend/shared/utils"

const (
	UpdateProductQuantityMSCall = "/%d/update-quantity"
)

func MicroserviceLinks() map[string]string {
	links := map[string]string{}

	updateProductQuantityLink := utils.GetProductMicroserviceLink(UpdateProductQuantityMSCall)
	links["updateProductQuantityMSCallLink"] = updateProductQuantityLink

	return links
}
