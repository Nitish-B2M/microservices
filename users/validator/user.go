// Package validator provides validation rules for user-related data.
package validator

type UserLoginUsingEmailValidator struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UserLoginUsingUsernameValidator struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}
