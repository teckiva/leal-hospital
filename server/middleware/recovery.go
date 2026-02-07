package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leal-hospital/server/logger"
	"github.com/leal-hospital/server/medierror"
)

// RecoveryMiddleware recovers from panics and returns proper error response
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				logger.Error("Panic recovered: %v", err)

				// Return error response
				appErr := medierror.ErrInternalServer()
				c.JSON(http.StatusInternalServerError, appErr.ToResponse())
				c.Abort()
			}
		}()

		c.Next()
	}
}
