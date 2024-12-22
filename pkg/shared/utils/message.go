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
)

// ***********Product***********
// Error messages
const (
	ProductNotFoundError         = "product with ID %d not found"
	ProductDeletionError         = "error occurred while deleting product with ID %d"
	ProductUpdateError           = "error occurred while updating product with ID %d"
	ProductCreationError         = "error occurred while creating product"
	InvalidProductIDError        = "invalid product ID provided"
	InvalidProductDataError      = "invalid data provided for product creation or update"
	InvalidProductRequest        = "invalid product request"
	ProductUnexpectedFetchError  = "unexpected error fetching product for update"
	ProductUnexpectedUpdateError = "unexpected error while updating product"
	ProductTagUpdateError        = "failed to update tags for product with ID %d"
	ProductOutOfStockError       = "product with ID %d is out of stock"
	ProductCategoryError         = "invalid category for product with ID %d"
	ProductPriceError            = "invalid price for product with ID %d"
)

// Info messages
const (
	ProductsFetchedSuccessfully = "all products fetched successfully"
	ProductFetchedSuccessfully  = "fetched product successfully with ID %d"
	ProductCreatedSuccessfully  = "product created successfully with ID %d"
	ProductUpdatedSuccessfully  = "product updated successfully with ID %d"
	ProductDeletedSuccessfully  = "product deleted successfully with ID %d"
	ProductNotModified          = "no updates required for product with ID %d"
)

// ************Tag*************
const (
	TagCreationFailed     = "failed to create tag: %v"
	TagExistError         = "error checking tag existence: %v"
	TagNotExist           = "tag does not exist, tag name: %v"
	TagFetchError         = "error occurred while fetching tag"
	TagAlreadyExistsError = "tag with name %v already exists"
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
	UserNotFoundError       = "user with ID %d not found"
	UserDeletionError       = "error occurred while deleting user with ID %d"
	UserUpdateError         = "failed to update user with ID %d"
	UserCreationError       = "failed to create user"
	UserNotModified         = "user with ID %d not modified"
	InvalidUserIDError      = "invalid user ID %d provided"
	InvalidUserDataError    = "invalid data provided for user creation or update"
	EmailRequiredError      = "email is required"
	PasswordRequiredError   = "password is required"
	PasswordLengthError     = "password must be at least 6 characters"
	EmailFormatError        = "invalid email format"
	FirstNameRequiredError  = "first name is required"
	LastNameRequiredError   = "last name is required"
	PasswordHashError       = "error hashing password"
	EmailAlreadyExistsError = "email is already in use"
	InvalidPasswordError    = "password does not meet security requirements"
	UserSuspendedError      = "user is suspended and cannot perform this action"
)

// Info messages
const (
	UsersFetchedSuccessfully = "all users fetched successfully"
	UserFetchedSuccessfully  = "fetched user successfully with ID %d"
	UserCreatedSuccessfully  = "user created successfully with ID %d"
	UserUpdatedSuccessfully  = "user updated successfully with ID %d"
	UserDeletedSuccessfully  = "user deleted successfully with ID %d"
	UserLoggedInSuccessfully = "user logged in successfully with ID %d"
)

// *************JWT*************
// General JWT Errors
const (
	TokenGenerationError  = "token generation failed due to an internal error"
	TokenExpiredError     = "token has expired"
	InvalidTokenError     = "token is invalid or malformed"
	TokenRevokedError     = "token has been revoked"
	InvalidClaimsError    = "invalid claims in the token"
	TokenParseError       = "error parsing token"
	TokenSignatureError   = "token signature verification failed"
	MissingTokenError     = "authorization token is missing"
	TokenFormatError      = "token is in an invalid format"
	TokenBlacklistedError = "token is blacklisted"

	// Permission-related Errors
	InsufficientPermissionsError = "insufficient permissions to access this resource"
)
