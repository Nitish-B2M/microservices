package utils

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func HashedPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func CompareHashedPassword(originalPass, loginPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(originalPass), []byte(loginPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return errors.New("incorrect password")
		}
		return err // Some other error
	}
	return nil
}
