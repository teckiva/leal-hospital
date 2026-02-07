package jwt

import (
	"github.com/leal-hospital/server/config"
	"github.com/leal-hospital/server/domain"
)

// JWTSvcDriver defines the interface for JWT operations
type JWTSvcDriver interface {
	GenerateAccessToken(userID int64, email string, isAdmin bool) (string, error)
	GenerateRefreshToken(userID int64) (string, error)
	ValidateToken(tokenString string) (*domain.TokenClaims, error)
}

// JWTSvc implements JWTSvcDriver interface
type JWTSvc struct {
	SecurityConfig *config.SecurityConfig
}

// NewJWTSvc creates a new JWT service instance
func NewJWTSvc(securityConfig *config.SecurityConfig) JWTSvcDriver {
	return &JWTSvc{
		SecurityConfig: securityConfig,
	}
}
