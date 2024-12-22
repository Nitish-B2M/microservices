package utils

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

func ValidateRequestPath(path string) ([]string, bool) {
	parts := strings.Split(path, "/")

	if len(parts) < 3 {
		return []string{}, false
	}

	if len(parts[2]) < 1 || parts[2] == "" {
		return []string{}, false
	}

	if len(parts) >= 4 && (len(parts[3]) < 1 || parts[3] == "") {
		return []string{}, false
	}

	return parts, true
}

func GetProductIdFromPath(r *http.Request) (int, error) {
	path := r.URL.Path
	parts, ok := ValidateRequestPath(path)
	if !ok {
		return 0, fmt.Errorf("invalid path")
	}
	id, err := strconv.Atoi(parts[len(parts)-1])
	return id, err
}

func GetUserIdFromPath(r *http.Request) (int, error) {
	path := r.URL.Path
	parts, ok := ValidateRequestPath(path)
	if !ok {
		return 0, fmt.Errorf("invalid path")
	}
	id, err := strconv.Atoi(parts[len(parts)-1])
	return id, err
}

func CheckRequestMethod(w http.ResponseWriter, r *http.Request, expectedMethod string) bool {
	if r.Method != expectedMethod {
		JsonError(w, InvalidRequestMethod, http.StatusMethodNotAllowed, nil)
		return false
	}
	return true
}

func CheckPasswordSecurity(password string) error {

	conditions := []struct {
		check   func(string) bool
		message string
	}{
		{check: minLength(6), message: "password must be at least 6 characters long"},
		{check: hasUppercase, message: "password must contain at least one uppercase letter"},
		{check: hasLowercase, message: "password must contain at least one lowercase letter"},
		{check: hasDigit, message: "password must contain at least one number"},
		{check: hasSpecialChar, message: "password must contain at least one special character"},
	}

	for _, condition := range conditions {
		if !condition.check(password) {
			return errors.New(condition.message)
		}
	}
	return nil
}

// implementation of higher order function or example of higher order function
func minLength(length int) func(string) bool {
	return func(s string) bool {
		return len(s) >= length
	}
}

func hasUppercase(s string) bool {
	re := regexp.MustCompile(`[A-Z]`)
	return re.MatchString(s)
}

func hasLowercase(s string) bool {
	re := regexp.MustCompile(`[a-z]`)
	return re.MatchString(s)
}

func hasDigit(s string) bool {
	re := regexp.MustCompile(`[0-9]`)
	return re.MatchString(s)
}

func hasSpecialChar(s string) bool {
	re := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`)
	return re.MatchString(s)
}

// For Email Validation
// CheckEmailSecurity validates the email by applying a series of conditions using CheckSecurity.
func CheckEmailSecurity(email string) error {
	if email == "" {
		return errors.New("email cannot be empty")
	}

	conditions := []struct {
		check   func(string) bool
		message string
	}{
		{check: isValidEmailFormat, message: "invalid email format"},
		{check: isEmailLengthValid, message: "email must be between 5 and 100 characters long"},
		{check: isDomainValid, message: "email domain must be valid"},
	}

	return CheckSecurity(email, conditions)
}

func CheckSecurity(input string, conditions []struct {
	check   func(string) bool
	message string
}) error {
	if input == "" {
		return errors.New("input cannot be empty")
	}

	// Check if conditions are nil or empty
	if len(conditions) == 0 {
		return errors.New("no validation conditions provided")
	}

	for _, condition := range conditions {
		if !condition.check(input) {
			return errors.New(condition.message)
		}
	}
	return nil
}

// isValidEmailFormat checks if the email format is valid using regex.
func isValidEmailFormat(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

// isEmailLengthValid checks if the email length is within valid bounds.
func isEmailLengthValid(email string) bool {
	return len(email) >= 5 && len(email) <= 100
}

// isDomainValid checks if the domain part of the email has at least one dot.
func isDomainValid(email string) bool {
	parts := strings.Split(email, "@")
	if len(parts) < 2 {
		return false
	}
	domain := parts[1]
	return len(domain) > 3 && strings.Contains(domain, ".")
}
