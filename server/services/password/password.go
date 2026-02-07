package password

import (
	"fmt"

	"github.com/leal-hospital/server/logger"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a plain text password using bcrypt
func (p *PasswordSvc) HashPassword(password string) (string, error) {
	functionName := "PasswordSvc.HashPassword"

	logger.Info(functionName, "Hashing password")

	// Generate bcrypt hash with default cost
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error(functionName, "Failed to hash password:", err)
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	logger.Info(functionName, "Password hashed successfully")
	return string(hashedBytes), nil
}

// VerifyPassword verifies a plain text password against a hashed password
func (p *PasswordSvc) VerifyPassword(hashedPassword, password string) error {
	functionName := "PasswordSvc.VerifyPassword"

	logger.Info(functionName, "Verifying password")

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			logger.Info(functionName, "Password mismatch")
			return fmt.Errorf("invalid password")
		}
		logger.Error(functionName, "Failed to verify password:", err)
		return fmt.Errorf("failed to verify password: %w", err)
	}

	logger.Info(functionName, "Password verified successfully")
	return nil
}
