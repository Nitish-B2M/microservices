package utils

// ***********Database***********
// Error messages
const (
	DatabaseConnectionError = "failed to connect to the database"
	UnexpectedDatabaseError = "unexpected database error"
	SchemaMigrationSuccess  = "%s schema migrated successfully"
	DatabaseMigrationError  = "failed to migrate %s schema: %v"
)

// *********** General Errors ***********
const (
	InternalServerError  = "an unexpected error occurred"
	NotFoundError        = "resource not found"
	BadRequestError      = "bad request"
	UnauthorizedError    = "unauthorized access"
	ForbiddenError       = "access forbidden"
	FailedToSendResponse = "failed to send response"
	InvalidJSONBody      = "invalid JSON body"
)

// ***********Product***********
// Error messages
const (
	ProductNotFoundError          = "product with ID %d not found"
	ProductDeletionError          = "error occurred while deleting product with ID %d"
	ProductUpdateError            = "error occurred while updating product with ID %d"
	ProductCreationError          = "error occurred while creating product"
	InvalidProductIDError         = "invalid product ID provided"
	InvalidProductDataError       = "invalid data provided for product creation or update"
	InvalidProductRequest         = "invalid product request"
	ProductsFetchError            = "error occurred while fetching products"
	CategoryParamRequired         = "category parameter is required"
	ProductUnexpectedFetchError   = "unexpected error fetching product for update"
	ProductUnexpectedUpdateError  = "unexpected error while updating product"
	ProductTagUpdateError         = "failed to update tags for product with ID %d"
	ProductOutOfStockError        = "product with ID %d is out of stock"
	ProductCategoryError          = "invalid category for product with ID %d"
	ProductPriceError             = "invalid price for product with ID %d"
	ProductIDRequiredError        = "product Id is required"
	FailedToFetchProductDetails   = "failed to fetch product details"
	InSufficientProductStockError = "insufficient stock for product %s"
	ErrorProductMicroservices     = "error occurred while calling product microservice"
	ProductQuantityUpdated        = "product quantity updated successfully"
	CategoryNotFoundError         = "product category not found"
)

// Info messages
const (
	ProductsFetchedSuccessfully        = "all products fetched successfully"
	ProductFetchedSuccessfully         = "fetched product successfully with ID %d"
	ProductCreatedSuccessfully         = "product created successfully with ID %d"
	ProductUpdatedSuccessfully         = "product updated successfully with ID %d"
	ProductDeletedSuccessfully         = "product deleted successfully with ID %d"
	ProductNotModified                 = "no updates required for product with ID %d"
	ProductCategoryFetchedSuccessfully = "products in category '%s' fetched successfully"
	CategoriesFetchedSuccessfully      = "all categories fetched successfully"
)

// ************Tag*************
const (
	TagCreationFailed        = "failed to create tag: %v"
	TagExistError            = "error checking tag existence: %v"
	TagNotExist              = "tag does not exist, tag name: %v"
	TagFetchError            = "error occurred while fetching tag"
	TagAlreadyExistsError    = "tag with name %v already exists"
	TagUpdateFailed          = "failed to update tag: %v"
	FailedAddingTagToProduct = "failed to add tag to product"
	FailedToFetchTag         = "failed to fetch tag"
)

// Validation error messages
const (
	InvalidRequestMethod = "invalid request method"
	InvalidRequestPath   = "invalid request path"
	InvalidRequestBody   = "invalid request body"
)

// *********** User ***********
// Error messages
const (
	UserNotFoundError          = "user with ID %d not found"
	UserNotFoundWithEmailError = "user with Email '%v' not found"
	UserDeletionError          = "error occurred while deleting user with ID %d"
	UserUpdateError            = "failed to update user with ID %d"
	UserCreationError          = "failed to create user"
	UserNotModified            = "user with ID %d not modified"
	InvalidUserIDError         = "invalid user ID %d provided"
	InvalidUserDataError       = "invalid data provided for user creation or update"
	EmailRequiredError         = "email is required"
	PasswordRequiredError      = "password is required"
	PasswordLengthError        = "password must be at least 6 characters"
	EmailFormatError           = "invalid email format"
	FirstNameRequiredError     = "first name is required"
	LastNameRequiredError      = "last name is required"
	PasswordHashError          = "error hashing password"
	EmailAlreadyExistsError    = "email is already in use"
	InvalidPasswordError       = "password does not meet security requirements"
	UserSuspendedError         = "user is suspended and cannot perform this action"
	PasswordResetError         = "error occurred while resetting password"
	EmailVerificationFailed    = "failed to verify email for user with ID: %d"
	UserDeActivationFailed     = "failed to deactivate user with ID %d"
	UserReactivationFailed     = "failed to reactivate user with ID %d"
	EmailNotVerifiedError      = "email is not verified email: %s"
	UserIsNotVerifiedError     = "user is not verified"
)

// Info messages
const (
	UsersFetchedSuccessfully     = "all users fetched successfully"
	UserFetchedSuccessfully      = "fetched user successfully with ID %d"
	UserCreatedSuccessfully      = "user created successfully with ID %d"
	UserUpdatedSuccessfully      = "user updated successfully with ID %d"
	UserDeletedSuccessfully      = "user deleted successfully with ID %d"
	UserLoggedInSuccessfully     = "user logged in successfully with ID %d"
	NewPasswordSetSuccessfully   = "new password set successfully"
	EmailVerifiedSuccessfully    = "email is verified"
	EmailAlreadyVerified         = "email is already verified"
	UserAlreadyActivated         = "user with ID %d has already been activated"
	UserAlreadyDeactivated       = "user with ID %d has already been deactivated"
	UserDeActivationSuccessfully = "user account de-activated successfully"
	UserReactivationSuccessfully = "user account re-activated successfully"
	PleaseVerifyEmail            = "please verify your email"
	RequestUserIsDeactivated     = "request user is deactivated"
)

// *************JWT*************
// General JWT Errors
const (
	TokenGenerationError  = "token generation failed due to an internal error"
	TokenExpiredError     = "token has expired"
	InvalidTokenError     = "invalid or expired token"
	TokenSignatureError   = "token signature verification failed"
	MissingTokenError     = "authorization token is missing"
	InvalidTokenClaims    = "invalid token claims"
	TokenBlacklistedError = "token is blacklisted"
)

// JWT Info
const (
	ResetPasswordTokenSent     = "password reset token sent"
	EmailVerificationTokenSent = "email verification token sent"
	ResetTokenValue            = "your reset token: %s"
)

// ************* Middleware *************
// Errors
const (
	MissingAuthorizationHeader = "authorization header is missing"
	InvalidAuthorizationHeader = "invalid authorization header"
	UserIdNotFoundInToken      = "user ID not found in token"
	UserIdNotFoundInCtx        = "user ID not found in context"
	UserIdNotFoundInParam      = "user ID not found in param"
)

// ************* Cart **************
// Errors
const (
	CartIdNotFoundError          = "cart with Id %d not found"
	UserCartNotFoundError        = "cart not found for user with userId %d"
	CartItemNotFoundError        = "cart item with cartId %d not found in cart"
	CartItemAdditionError        = "failed to add item to cart"
	CartItemUpdateError          = "error occurred while updating item in cart"
	CartItemDeletionError        = "error occurred while deleting item from cart"
	CartOutOfStockError          = "item with ID %d is out of stock"
	InvalidCartRequest           = "invalid cart request"
	CartInvalidProductError      = "invalid product for item with ID %d"
	CartUnexpectedFetchError     = "unexpected error fetching cart data"
	CartUnexpectedUpdateError    = "unexpected error updating cart data"
	ErrorCallingCartMicroservice = "error occurred while calling cart microservice"
	CartIdNotProvided            = "cart ID not provided"
)

// Info messages
const (
	CartClearedSuccessfully     = "cart cleared successfully for user with ID %d"
	CartFetchedSuccessfully     = "cart fetched successfully for user with ID %d"
	CartItemAddedSuccessfully   = "item with ID %d added to cart successfully"
	CartItemUpdatedSuccessfully = "item with ID %d updated in cart successfully"
	CartItemDeletedSuccessfully = "item with ID %d removed from cart successfully"
	CartCheckedOutSuccessfully  = "cart checked out successfully for user with ID %d"
)

// *************** Templates and Files ********************
// Errors
const (
	TemplateParsingFailed = "failed to parse template: %v"
	TemplateExecuteFailed = "failed to execute template: %v"
	FileRetrieveFailed    = "error retrieving file"
	UnableToSaveFile      = "unable to save file"
	ErrorSavingFile       = "error saving file"
)

// InsufficientPermissionsError Permission related Errors
const (
	InsufficientPermissionsError = "insufficient permissions to access this resource"
)

// *************** Order *****************
const (
	OrderCreatedSuccessfully  = "order successfully created"
	OrderFetchSuccess         = "order with orderId %s fetch successfully"
	OrderIdInvalid            = "order id %s is invalid"
	OrderIdRequired           = "order id is required"
	OrdersFetchedSuccessfully = "orders fetched successfully"
)

const (
	OrderStatusPending    = "Pending"
	OrderStatusProcessing = "Processing"
	OrderStatusShipped    = "Shipped"
	OrderStatusDelivered  = "Delivered"
	OrderStatusCancelled  = "Cancelled"
	OrderStatusFailed     = "Failed"
)

const (
	PaymentFailed           = "payment failed"
	PaymentCancelled        = "payment cancelled"
	PaymentStatusPaid       = "paid"
	PaymentStatusRejected   = "rejected"
	PaymentStatusCanceled   = "canceled"
	PaymentStatusPending    = "pending"
	PaymentStatusFailed     = "failed"
	InvalidPaymentRequest   = "invalid payment request"
	PaymentValidationFailed = "payment validation failed"
)

const (
	PaymentSuccessful = "payment successfully created"
)
