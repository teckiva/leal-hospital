package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/leal-hospital/server/domain"
	"github.com/leal-hospital/server/logger"
	"github.com/leal-hospital/server/medierror"
	"github.com/spf13/cast"
)

// ValidationType defines different user validation levels
type ValidationType int

const (
	// NoValidation - No user validation required (public endpoints)
	NoValidation ValidationType = iota
	// AuthenticatedUser - Requires valid JWT token
	AuthenticatedUser
	// ApprovedUser - Requires authenticated and approved user
	ApprovedUser
	// AdminUser - Requires authenticated admin user
	AdminUser
)

// ValidateRequest validates the request based on validation type
func ValidateRequest(
	c *gin.Context,
	functionName string,
	registry *medierror.ErrorRegistry,
	validationType ValidationType,
) (*domain.User, *medierror.AppError) {

	// No validation needed
	if validationType == NoValidation {
		return nil, nil
	}

	// Get user from context (set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Error(functionName, "user_id not found in context")
		return nil, medierror.NewFromRegistry(medierror.ErrUnauthorized, nil)
	}

	email, exists := c.Get("email")
	if !exists {
		logger.Error(functionName, "email not found in context")
		return nil, medierror.NewFromRegistry(medierror.ErrUnauthorized, nil)
	}

	isAdmin, _ := c.Get("is_admin")

	user := &domain.User{
		ID:      cast.ToInt64(userID),
		Email:   cast.ToString(email),
		IsAdmin: cast.ToBool(isAdmin),
	}

	// Validate based on type
	switch validationType {
	case AuthenticatedUser:
		// Basic authentication check already done
		return user, nil

	case ApprovedUser:
		// Check if user is approved (would need to fetch from DB)
		// For now, just return user
		return user, nil

	case AdminUser:
		if !user.IsAdmin {
			logger.Error(functionName, "admin access required")
			return nil, medierror.NewFromRegistry(medierror.ErrForbidden, nil)
		}
		return user, nil
	}

	return user, nil
}

// FormatValidationError formats validation errors from binding into AppError
func FormatValidationError(target interface{}, err error) *medierror.AppError {
	return medierror.ErrBadRequestWithMsg(
		"Validation failed",
		"Invalid request data. Please check your input.",
		err,
	)
}
