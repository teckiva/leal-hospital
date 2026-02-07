package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/leal-hospital/server/config"
	"github.com/leal-hospital/server/di"
	"github.com/leal-hospital/server/handler/auth/api"
	"github.com/leal-hospital/server/logger"
	"github.com/leal-hospital/server/middleware"
	"github.com/leal-hospital/server/models/db"
	authPersistence "github.com/leal-hospital/server/persistence/auth"
	"github.com/leal-hospital/server/services/email"
	"github.com/leal-hospital/server/services/jwt"
	"github.com/leal-hospital/server/services/otp"
	"github.com/leal-hospital/server/services/password"
	utilsdb "github.com/leal-hospital/server/utils/db"
)

// AuthModule represents the authentication module
type AuthModule struct{}

// NewAuthModule creates a new auth module
func NewAuthModule() *AuthModule {
	return &AuthModule{}
}

// Name returns the module name
func (m *AuthModule) Name() string {
	return "auth"
}

// Configure registers module dependencies in the DI container
func (m *AuthModule) Configure(container *di.Container) {
	// Register Auth Persistence (using factory pattern like KB-PAY)
	container.RegisterFactory((*authPersistence.AuthPersistence)(nil), func(c *di.Container) interface{} {
		dbManager := c.Resolve((*utilsdb.DBManagerInterface)(nil)).(utilsdb.DBManagerInterface)
		return authPersistence.NewAuthPersistence(dbManager.GetDB())
	})

	// Register Email Service
	container.RegisterFactory((*email.EmailSvcDriver)(nil), func(c *di.Container) interface{} {
		cfg := c.Resolve((*config.AppConfig)(nil)).(*config.AppConfig)
		return email.NewEmailSvc(&cfg.Email)
	})

	// Register Password Service
	container.RegisterFactory((*password.PasswordSvcDriver)(nil), func(c *di.Container) interface{} {
		return password.NewPasswordSvc()
	})

	// Register OTP Service (depends on persistence and email service)
	container.RegisterFactory((*otp.OTPSvcDriver)(nil), func(c *di.Container) interface{} {
		dbManager := c.Resolve((*utilsdb.DBManagerInterface)(nil)).(utilsdb.DBManagerInterface)
		cfg := c.Resolve((*config.AppConfig)(nil)).(*config.AppConfig)
		emailSvc := c.Resolve((*email.EmailSvcDriver)(nil)).(email.EmailSvcDriver)
		queries := db.New(dbManager.GetDB())
		return otp.NewOTPSvc(queries, &cfg.Security, emailSvc)
	})

	// Register JWT Service
	container.RegisterFactory((*jwt.JWTSvcDriver)(nil), func(c *di.Container) interface{} {
		cfg := c.Resolve((*config.AppConfig)(nil)).(*config.AppConfig)
		return jwt.NewJWTSvc(&cfg.Security)
	})

	// Register Auth Handler (depends on persistence and services)
	container.RegisterFactory((*api.AuthHandler)(nil), func(c *di.Container) interface{} {
		authPers := c.Resolve((*authPersistence.AuthPersistence)(nil)).(authPersistence.AuthPersistence)
		otpSvc := c.Resolve((*otp.OTPSvcDriver)(nil)).(otp.OTPSvcDriver)
		passwordSvc := c.Resolve((*password.PasswordSvcDriver)(nil)).(password.PasswordSvcDriver)
		jwtSvc := c.Resolve((*jwt.JWTSvcDriver)(nil)).(jwt.JWTSvcDriver)
		return api.NewAuthHandler(authPers, otpSvc, passwordSvc, jwtSvc)
	})
}

// RegisterRoutes registers module HTTP routes
func (m *AuthModule) RegisterRoutes(router *gin.Engine, container *di.Container) {
	const functionName = "auth.Module.RegisterRoutes"
	logger.Info(functionName, "registering_routes")

	// Resolve dependencies
	authHandler := container.Resolve((*api.AuthHandler)(nil)).(api.AuthHandler)
	jwtSvc := container.Resolve((*jwt.JWTSvcDriver)(nil)).(jwt.JWTSvcDriver)

	// Register routes to multiple prefixes
	prefixes := []string{"/lael", "/api/v1/lael"}
	for _, prefix := range prefixes {
		apiRouter := router.Group(prefix)
		registerAuthRoutes(apiRouter, authHandler, jwtSvc)
	}
}

// registerAuthRoutes sets up all authentication routes for a given router group
func registerAuthRoutes(apiRouter *gin.RouterGroup, authHandler api.AuthHandler, jwtSvc jwt.JWTSvcDriver) {
	// Public auth routes (no authentication required)
	publicGroup := apiRouter.Group("/auth")
	{
		publicGroup.POST("/register", authHandler.Register())
		publicGroup.POST("/verify-otp", authHandler.VerifyOTP())
		publicGroup.POST("/login", authHandler.Login())
		publicGroup.POST("/forgot-password", authHandler.ForgotPassword())
		publicGroup.POST("/reset-password", authHandler.ResetPassword())
		publicGroup.POST("/refresh-token", authHandler.RefreshToken())
	}

	// Protected routes (requires authentication)
	protectedGroup := apiRouter.Group("/")
	protectedGroup.Use(middleware.AuthMiddleware(jwtSvc))
	{
		// Add protected routes here as needed
		// Example: protectedGroup.GET("/profile", profileHandler.GetProfile())
		// Example: protectedGroup.PUT("/profile", profileHandler.UpdateProfile())
	}

	// Admin routes (requires authentication and admin role)
	adminGroup := apiRouter.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware(jwtSvc))
	adminGroup.Use(middleware.AdminMiddleware())
	{
		// Add admin routes here as needed
		// Example: adminGroup.GET("/users", userHandler.ListUsers())
		// Example: adminGroup.POST("/users/:id/approve", userHandler.ApproveUser())
	}
}
