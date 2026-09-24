package service

import (
	"fmt"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// PasswordPolicy validates password rules.
type PasswordPolicy struct{}

// NewPasswordPolicy creates a new PasswordPolicy.
func NewPasswordPolicy() *PasswordPolicy {
	return &PasswordPolicy{}
}

// Validate checks if a password meets the policy requirements.
// Rules: >= 8 chars, >= 1 uppercase, >= 1 lowercase, >= 1 digit
func (p *PasswordPolicy) Validate(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	var hasUpper, hasLower, hasDigit bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}

	return nil
}

// HashPassword hashes a password using bcrypt.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// ComparePassword compares a plaintext password with a bcrypt hash.
func ComparePassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}