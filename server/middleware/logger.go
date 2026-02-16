package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leal-hospital/server/logger"
)

// RequestLoggerMiddleware logs HTTP requests
func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		logger.Info("HTTP %s %s - Status: %d - Latency: %v - IP: %s",
			method, path, statusCode, latency, clientIP)
	}
}
