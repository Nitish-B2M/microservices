package utils

import (
	"errors"
	"fmt"

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
	fmt.Println(err)
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, errors.New("incorrect password")
		}
		return false, err // Some other error
	}
	return true, nil
}
