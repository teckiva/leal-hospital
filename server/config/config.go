package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
)

// AppConfig holds all configuration for the application
type AppConfig struct {
	Environment string
	Server      ServerConfig
	Database    DBConfig
	Security    SecurityConfig
	Logging     LoggingConfig
	Email       EmailConfig
}

// ServerConfig contains HTTP server configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DBConfig contains database connection configuration
type DBConfig struct {
	Type            string
	Host            string
	Port            string
	Username        string
	Password        string
	Database        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// GetDSN returns the database connection string for DBConfig
func (db *DBConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		db.Username,
		db.Password,
		db.Host,
		db.Port,
		db.Database,
	)
}

// SecurityConfig contains security-related configuration
type SecurityConfig struct {
	JWTSecret         string
	JWTExpiration     time.Duration
	RefreshExpiration time.Duration
	OTPLength         int
	OTPExpiration     time.Duration
	SessionTimeout    time.Duration
	MaxLoginAttempts  int
	LockoutDuration   time.Duration
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level         string
	FilePath      string
	MaxSize       int // megabytes
	MaxBackups    int
	MaxAge        int // days
	Compress      bool
	ConsoleOutput bool
}

// EmailConfig contains SMTP email configuration
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

// Global config instance
var Config *AppConfig

// LoadConfig loads configuration from environment variables and .env file
func LoadConfig() (*AppConfig, error) {
	// Set config file path
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath("./config")    // For running from server/
	viper.AddConfigPath("../config")   // For running from server/cmd/
	viper.AddConfigPath(".")

	// Read from environment variables
	viper.AutomaticEnv()

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		log.Println("No .env file found, using environment variables and defaults")
	}

	// Set defaults
	setDefaults()

	// Build configuration
	config := &AppConfig{
		Environment: viper.GetString("APP_ENV"),
		Server: ServerConfig{
			Port:         viper.GetString("SERVER_PORT"),
			ReadTimeout:  viper.GetDuration("SERVER_READ_TIMEOUT") * time.Second,
			WriteTimeout: viper.GetDuration("SERVER_WRITE_TIMEOUT") * time.Second,
			IdleTimeout:  viper.GetDuration("SERVER_IDLE_TIMEOUT") * time.Second,
		},
		Database: DBConfig{
			Type:            "mysql", // Fixed to mysql
			Host:            viper.GetString("LAEL_MYSQL_DB_HOST"),
			Port:            viper.GetString("LAEL_MYSQL_DB_PORT"),
			Username:        viper.GetString("LAEL_MYSQL_DB_USERNAME"),
			Password:        viper.GetString("LAEL_MYSQL_DB_PASSWORD"),
			Database:        viper.GetString("LAEL_MYSQL_DB_SCHEMA"),
			MaxOpenConns:    25, // Default value
			MaxIdleConns:    5,  // Default value
			ConnMaxLifetime: 5 * time.Minute,
		},
		Security: SecurityConfig{
			JWTSecret:         viper.GetString("JWT_SECRET"),
			JWTExpiration:     viper.GetDuration("JWT_EXPIRATION") * time.Hour,
			RefreshExpiration: viper.GetDuration("REFRESH_EXPIRATION") * time.Hour,
			OTPLength:         viper.GetInt("OTP_LENGTH"),
			OTPExpiration:     viper.GetDuration("OTP_EXPIRATION") * time.Minute,
			SessionTimeout:    viper.GetDuration("SESSION_TIMEOUT") * time.Hour,
			MaxLoginAttempts:  viper.GetInt("MAX_LOGIN_ATTEMPTS"),
			LockoutDuration:   viper.GetDuration("LOCKOUT_DURATION") * time.Minute,
		},
		Logging: LoggingConfig{
			Level:         viper.GetString("LOG_LEVEL"),
			FilePath:      viper.GetString("LOG_FILE_PATH"),
			MaxSize:       viper.GetInt("LOG_MAX_SIZE"),
			MaxBackups:    viper.GetInt("LOG_MAX_BACKUPS"),
			MaxAge:        viper.GetInt("LOG_MAX_AGE"),
			Compress:      viper.GetBool("LOG_COMPRESS"),
			ConsoleOutput: viper.GetBool("LOG_CONSOLE_OUTPUT"),
		},
		Email: EmailConfig{
			SMTPHost:     viper.GetString("SMTP_HOST"),
			SMTPPort:     viper.GetInt("SMTP_PORT"),
			SMTPUsername: viper.GetString("SMTP_USERNAME"),
			SMTPPassword: viper.GetString("SMTP_PASSWORD"),
			FromEmail:    viper.GetString("SMTP_FROM_EMAIL"),
			FromName:     viper.GetString("SMTP_FROM_NAME"),
		},
	}

	// Validate required fields
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Set global config
	Config = config

	return config, nil
}

// setDefaults sets default values for configuration
func setDefaults() {
	// Application defaults
	viper.SetDefault("APP_ENV", "development")

	// Server defaults
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("SERVER_READ_TIMEOUT", 15)
	viper.SetDefault("SERVER_WRITE_TIMEOUT", 15)
	viper.SetDefault("SERVER_IDLE_TIMEOUT", 60)

	// Database defaults
	viper.SetDefault("LAEL_MYSQL_DB_HOST", "localhost")
	viper.SetDefault("LAEL_MYSQL_DB_PORT", "3306")
	viper.SetDefault("LAEL_MYSQL_DB_USERNAME", "root")
	viper.SetDefault("LAEL_MYSQL_DB_PASSWORD", "")
	viper.SetDefault("LAEL_MYSQL_DB_SCHEMA", "lael")

	// Security defaults
	viper.SetDefault("JWT_SECRET", "change-this-secret-in-production")
	viper.SetDefault("JWT_EXPIRATION", 24)
	viper.SetDefault("REFRESH_EXPIRATION", 168)
	viper.SetDefault("OTP_LENGTH", 6)
	viper.SetDefault("OTP_EXPIRATION", 5)
	viper.SetDefault("SESSION_TIMEOUT", 24)
	viper.SetDefault("MAX_LOGIN_ATTEMPTS", 5)
	viper.SetDefault("LOCKOUT_DURATION", 15)

	// Logging defaults
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_FILE_PATH", "logs/app.log")
	viper.SetDefault("LOG_MAX_SIZE", 100)
	viper.SetDefault("LOG_MAX_BACKUPS", 3)
	viper.SetDefault("LOG_MAX_AGE", 28)
	viper.SetDefault("LOG_COMPRESS", true)
	viper.SetDefault("LOG_CONSOLE_OUTPUT", true)

	// Email defaults
	viper.SetDefault("SMTP_HOST", "smtp.gmail.com")
	viper.SetDefault("SMTP_PORT", 587)
	viper.SetDefault("SMTP_USERNAME", "")
	viper.SetDefault("SMTP_PASSWORD", "")
	viper.SetDefault("SMTP_FROM_EMAIL", "noreply@laelhospital.com")
	viper.SetDefault("SMTP_FROM_NAME", "Lael Hospital")
}

// validateConfig validates required configuration fields
func validateConfig(config *AppConfig) error {
	// Validate server config
	if config.Server.Port == "" {
		return fmt.Errorf("SERVER_PORT is required")
	}

	// Validate database config
	if config.Database.Host == "" {
		return fmt.Errorf("LAEL_MYSQL_DB_HOST is required")
	}
	if config.Database.Port == "" {
		return fmt.Errorf("LAEL_MYSQL_DB_PORT is required")
	}
	if config.Database.Database == "" {
		return fmt.Errorf("LAEL_MYSQL_DB_SCHEMA is required")
	}

	// Validate security config
	if config.Security.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if config.Environment == "production" && config.Security.JWTSecret == "change-this-secret-in-production" {
		return fmt.Errorf("JWT_SECRET must be changed in production environment")
	}

	// Validate logging config
	if config.Logging.Level == "" {
		return fmt.Errorf("LOG_LEVEL is required")
	}

	return nil
}

// GetDatabaseDSN returns the database connection string
func (c *AppConfig) GetDatabaseDSN() string {
	switch c.Database.Type {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			c.Database.Username,
			c.Database.Password,
			c.Database.Host,
			c.Database.Port,
			c.Database.Database,
		)
	case "postgres":
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			c.Database.Host,
			c.Database.Port,
			c.Database.Username,
			c.Database.Password,
			c.Database.Database,
		)
	default:
		return ""
	}
}

// IsProduction returns true if running in production environment
func (c *AppConfig) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment returns true if running in development environment
func (c *AppConfig) IsDevelopment() bool {
	return c.Environment == "development"
}

// GetLogFilePath returns the full path to the log file
func (c *AppConfig) GetLogFilePath() string {
	if c.Logging.FilePath == "" {
		return "logs/app.log"
	}
	return c.Logging.FilePath
}

// EnsureLogDirectory creates the log directory if it doesn't exist
func (c *AppConfig) EnsureLogDirectory() error {
	logDir := "logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
	}
	return nil
}
