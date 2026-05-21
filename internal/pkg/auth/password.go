package auth

import (
	"errors"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func ComparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func ValidatePassword(password string) error {
	trimmed := strings.TrimSpace(password)
	if len(trimmed) < 6 || len(trimmed) > 64 {
		return errors.New("密码长度必须在 6-64 位之间")
	}
	return nil
}

func ValidateUsername(username string) error {
	if !usernamePattern.MatchString(username) {
		return errors.New("用户名只能包含字母、数字、下划线，长度 3-32")
	}
	return nil
}
