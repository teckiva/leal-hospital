package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/leal-hospital/server/domain"
	"github.com/leal-hospital/server/logger"
)

// jwtClaims represents JWT claims with standard JWT fields (internal use only)
type jwtClaims struct {
	UserID  int64  `json:"user_id"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"is_admin"`
	Type    string `json:"type"`
	jwt.RegisteredClaims
}

// GenerateAccessToken generates a new access token
func (j *JWTSvc) GenerateAccessToken(userID int64, email string, isAdmin bool) (string, error) {
	functionName := "JWTSvc.GenerateAccessToken"

	logger.Info(functionName, "Generating access token for user:", userID)

	claims := jwtClaims{
		UserID:  userID,
		Email:   email,
		IsAdmin: isAdmin,
		Type:    "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.SecurityConfig.JWTExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "lael-hospital",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.SecurityConfig.JWTSecret))
	if err != nil {
		logger.Error(functionName, "Failed to generate access token:", err)
		return "", fmt.Errorf("failed to generate access token: %w", err)
	}

	logger.Info(functionName, "Access token generated successfully for user:", userID)
	return tokenString, nil
}

// GenerateRefreshToken generates a new refresh token
func (j *JWTSvc) GenerateRefreshToken(userID int64) (string, error) {
	functionName := "JWTSvc.GenerateRefreshToken"

	logger.Info(functionName, "Generating refresh token for user:", userID)

	claims := jwtClaims{
		UserID: userID,
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.SecurityConfig.RefreshExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "lael-hospital",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.SecurityConfig.JWTSecret))
	if err != nil {
		logger.Error(functionName, "Failed to generate refresh token:", err)
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	logger.Info(functionName, "Refresh token generated successfully for user:", userID)
	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTSvc) ValidateToken(tokenString string) (*domain.TokenClaims, error) {
	functionName := "JWTSvc.ValidateToken"

	logger.Info(functionName, "Validating token")

	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.SecurityConfig.JWTSecret), nil
	})

	if err != nil {
		logger.Error(functionName, "Failed to parse token:", err)
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(*jwtClaims); ok && token.Valid {
		logger.Info(functionName, "Token validated successfully for user:", claims.UserID)
		return &domain.TokenClaims{
			UserID:  claims.UserID,
			Email:   claims.Email,
			IsAdmin: claims.IsAdmin,
			Type:    claims.Type,
		}, nil
	}

	logger.Error(functionName, "Token validation failed: invalid claims")
	return nil, fmt.Errorf("invalid token claims")
}
