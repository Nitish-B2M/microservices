package constants

const (
	ProductMicroserviceGetProductCall       = "http://localhost:8081/product/%s/cart"
	FailedToUpdateProductQuantity           = "failed to update product %d quantity: %v\n"
	FailedToUpdateProductQuantityWithStatus = "failed to update product %d quantity: Status Code %d\n"
	SuccessfullyUpdatedProductQuantity      = "successfully updated quantity for product %d\n"
	FailedToFetchProductDetails             = "failed to fetch product %s details"
	ProductNotFound                         = "product %s not found"
	ErrorDecodingProductDetails             = "error decoding product %s details"
	ProductQuantityOutOfStock               = "product quantity out of stock"
	ProductDetailsNotEnough                 = "product details not enough"
	AddingProductToCart                     = "adding product %d (quantity: %d) to cart\n"
	TaskForProductPushedToChannel           = "task for product %d pushed to the channel\n"
	ItemsAddedToCart                        = "items added to cart"
	SomeItemAddedToCart                     = "some item added to cart"
	ItemsRemoveFromCart                     = "items removed from cart"
	ProcessingProductQuantityUpdate         = "processing update for ProductID: %d, Quantity: %d\n"
	FailedToCreateRequest                   = "failed to create request for product %d and error: %v\n"
	UserMicroserviceGetUserDataCall         = "http://localhost:8080/user/%d"
)
