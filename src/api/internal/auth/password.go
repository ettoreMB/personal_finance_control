package auth

import (
	"errors"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

var ErrWeakPassword = errors.New("password must be more than 8 characters and contain at least one special character")

const bcryptCost = 10

func ValidatePassword(password string) error {
	if len(password) <= 8 {
		return ErrWeakPassword
	}

	hasSpecial := false
	for _, r := range password {
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			hasSpecial = true
			break
		}
	}

	if !hasSpecial {
		return ErrWeakPassword
	}

	return nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func ComparePassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
