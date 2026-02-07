package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/leal-hospital/server/logger"
	"github.com/leal-hospital/server/services/jwt"
)

// AuthMiddleware validates JWT tokens and attaches user info to context
func AuthMiddleware(jwtSvc jwt.JWTSvcDriver) gin.HandlerFunc {
	return func(c *gin.Context) {
		functionName := "AuthMiddleware"

		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Info(functionName, "Missing Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Check if header has Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Info(functionName, "Invalid Authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization format. Use: Bearer <token>",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := jwtSvc.ValidateToken(tokenString)
		if err != nil {
			logger.Error(functionName, "Token validation failed:", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Check if token is access token (not refresh token)
		if claims.Type != "access" {
			logger.Info(functionName, "Invalid token type:", claims.Type)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Access token required",
			})
			c.Abort()
			return
		}

		// Attach user info to context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("is_admin", claims.IsAdmin)

		logger.Info(functionName, "User authenticated:", claims.UserID)

		c.Next()
	}
}

// AdminMiddleware ensures the user is an admin
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		functionName := "AdminMiddleware"

		// Get is_admin from context (set by AuthMiddleware)
		isAdmin, exists := c.Get("is_admin")
		if !exists {
			logger.Error(functionName, "is_admin not found in context")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Authentication context not found",
			})
			c.Abort()
			return
		}

		// Check if user is admin
		if !isAdmin.(bool) {
			userID, _ := c.Get("user_id")
			logger.Info(functionName, "Access denied for non-admin user:", userID)
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
