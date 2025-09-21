// Package utils provides utility functions for user-related operations.
package utils

import (
	"bytes"
	"errors"
	"html/template"
)

type UserCreationTD struct {
	Email    string
	FullName string
}

type VerificationMailTD struct {
	Email           string
	FullName        string
	VerificationURL string
}

// Define custom error types
var (
	ErrTokenAlreadyVerified = errors.New("token already verified")
	ErrTokenNotFound        = errors.New("invalid or expired token")
)

func GetUserCreatedTemplate(data UserCreationTD) (string, error) {
	t, err := template.New("welcome").Parse(UserCreatedTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func SendVerificationEmail(data VerificationMailTD) (string, error) {
	t, err := template.New("verification_mail").Parse(VerificationMailTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
