package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leal-hospital/server/config"
	"github.com/leal-hospital/server/di"
	"github.com/leal-hospital/server/logger"
	"github.com/leal-hospital/server/medierror"
	"github.com/leal-hospital/server/middleware"
	"github.com/leal-hospital/server/utils/db"
)

// App represents the application
type App struct {
	Config    *config.AppConfig
	Router    *gin.Engine
	Container *di.Container
	Server    *http.Server
	Modules   []Module
}

// NewApp creates a new application instance
func NewApp(cfg *config.AppConfig, modules []Module) *App {
	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	return &App{
		Config:    cfg,
		Router:    gin.New(),
		Container: di.NewContainer(),
		Modules:   modules,
	}
}

// Bootstrap initializes the application
func (a *App) Bootstrap() error {
	// Initialize logger
	if err := logger.InitLogger(
		a.Config.Logging.Level,
		a.Config.Logging.FilePath,
		a.Config.Logging.MaxSize,
		a.Config.Logging.MaxBackups,
		a.Config.Logging.MaxAge,
	); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	logger.Info("Logger initialized")

	// Initialize error registry
	if err := medierror.InitErrorRegistry(a.Router); err != nil {
		return fmt.Errorf("failed to initialize error registry: %w", err)
	}

	logger.Info("Error registry initialized")

	// Initialize database
	dbManager := db.NewDBManager(&a.Config.Database)
	if err := dbManager.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	logger.Info("Database initialized")

	// Register core dependencies
	a.Container.Register((*db.DBManagerInterface)(nil), dbManager)
	a.Container.Register((*config.AppConfig)(nil), a.Config)

	// Setup middleware
	a.setupMiddleware()

	// Configure and register modules
	for _, module := range a.Modules {
		logger.Info("Configuring module: %s", module.Name())
		module.Configure(a.Container)
		module.RegisterRoutes(a.Router, a.Container)
	}

	// Health check endpoint
	a.Router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Create HTTP server
	a.Server = &http.Server{
		Addr:         fmt.Sprintf(":%s", a.Config.Server.Port),
		Handler:      a.Router,
		ReadTimeout:  a.Config.Server.ReadTimeout,
		WriteTimeout: a.Config.Server.WriteTimeout,
	}

	return nil
}

// setupMiddleware configures HTTP middleware
func (a *App) setupMiddleware() {
	a.Router.Use(middleware.RecoveryMiddleware())
	a.Router.Use(middleware.RequestLoggerMiddleware())
	a.Router.Use(middleware.CORSMiddleware())
}

// Run starts the HTTP server
func (a *App) Run() error {
	// Start server in goroutine
	go func() {
		logger.Info("Starting HTTP server on port %s (environment: %s)",
			a.Config.Server.Port, a.Config.Environment)

		if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	logger.Info("Server exited")
	return nil
}
