// Package validator provides validation rules for user-related data.
package validator

type UserLoginUsingEmailValidator struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginUserValidator struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UserLoginUsingUsernameValidator struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type CreateUserValidator struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"omitempty"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required"`
}

type EmailVerificationValidator struct {
	Token string `json:"token" validate:"required"`
}

type FetchUserProfileValidator struct {
	ID int `json:"id" validate:"required"`
}
