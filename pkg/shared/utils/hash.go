package utils

import (
	"golang.org/x/crypto/bcrypt"
)

func HashedPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func CompareHashedPassword(originalPass, loginPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(originalPass), []byte(loginPassword))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil // Passwords do not match
		}
		return false, err // Some other error
	}
	return true, nil
}
