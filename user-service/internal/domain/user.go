package domain

import (
	"errors"
	"net/mail"
	"strings"
	"time"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailTaken         = errors.New("email already taken")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
	ErrEmptyName          = errors.New("full_name is required")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrTokenInvalid       = errors.New("token invalid")
	ErrTokenRevoked       = errors.New("token revoked")
	ErrAlreadyVerified    = errors.New("already verified")
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	FullName     string
	Verified     bool
	CreatedAt    time.Time
}

func NormaliseEmail(e string) (string, error) {
	e = strings.ToLower(strings.TrimSpace(e))
	if _, err := mail.ParseAddress(e); err != nil {
		return "", ErrInvalidEmail
	}
	return e, nil
}

func ValidatePassword(p string) error {
	if len(p) < 8 {
		return ErrWeakPassword
	}
	return nil
}

func ValidateName(n string) error {
	if strings.TrimSpace(n) == "" {
		return ErrEmptyName
	}
	return nil
}
