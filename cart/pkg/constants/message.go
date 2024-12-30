package constants

const (
	ProductMicroserviceGetProductCall       = "http://localhost:8081/product/%d/cart"
	ProductMicroserviceUpdateQuantityCall   = "http://localhost:8081/product/%d/update-quantity"
	FailedToUpdateProductQuantity           = "failed to update product %d quantity: %v\n"
	FailedToUpdateProductQuantityWithStatus = "failed to update product %d quantity: Status Code %d\n"
	SuccessfullyUpdatedProductQuantity      = "successfully updated quantity for product %d\n"
	FailedToFetchProductDetails             = "Failed to fetch product %d details"
	ProductNotFound                         = "Product %d not found"
	ErrorDecodingProductDetails             = "Error decoding product %d details"
	ProductQuantityOutOfStock               = "Product quantity out of stock"
	ProductDetailsNotEnough                 = "Product details not enough"
	AddingProductToCart                     = "Adding product %d (quantity: %d) to cart\n"
	TaskForProductPushedToChannel           = "Task for product %d pushed to the channel\n"
	ItemsAddedToCart                        = "items added to cart"
	SomeItemAddedToCart                     = "Some item added to cart"
	ItemsRemoveFromCart                     = "items removed from cart"
	ProcessingProductQuantityUpdate         = "Processing update for ProductID: %d, Quantity: %d\n"
	FailedToCreateRequest                   = "failed to create request for product %d and error: %v\n"
	UserMicroserviceGetUserDataCall         = "http://localhost:8080/user/%d"
)
