# Lael Hospital Backend Setup Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Set up complete backend directory structure and core infrastructure for Lael Hospital management system

**Architecture:** Following clean architecture with layered design (handler/services/persistence/domain), SQLC for type-safe queries, Gin for HTTP framework, and dependency injection pattern

**Tech Stack:** Go 1.23, Gin, MySQL, SQLC, Redis, Viper, Zerolog

---

## Task 1: Initialize Go Project and Directory Structure

**Files:**
- Create: `server/go.mod`

**Step 1: Initialize Go module**

Run: `cd /Users/abhay.sahani/Documents/MyRepos/leal-hospital/server && go mod init github.com/leal-hospital/server`
Expected: Success - go.mod created

**Step 2: Create base directory structure**

Run:
```bash
cd /Users/abhay.sahani/Documents/MyRepos/leal-hospital/server && mkdir -p cmd app handler services persistence domain models/db models/dbConf utils/db utils/logger config middleware gateway di medierror telemetry deployments/docker deployments/k8s
```
Expected: All directories created

**Step 3: Verify directory structure**

Run: `cd /Users/abhay.sahani/Documents/MyRepos/leal-hospital/server && ls -la`
Expected: See all created directories including medierror, config

---

## Task 2: Install Core Dependencies

**Files:**
- Modify: `server/go.mod`

**Step 1: Install HTTP framework and core dependencies**

Run:
```bash
cd /Users/abhay.sahani/Documents/MyRepos/leal-hospital/server && go get github.com/gin-gonic/gin && go get github.com/gin-contrib/cors
```
Expected: Dependencies added to go.mod

**Step 2: Install configuration management**

Run:
```bash
cd /Users/abhay.sahani/Documents/MyRepos/leal-hospital/server && go get github.com/spf13/viper
```
Expected: Viper added to dependencies

**Step 3: Install database dependencies**

Run:
```bash
cd /Users/abhay.sahani/Documents/MyRepos/leal-hospital/server && go get github.com/go-sql-driver/mysql
```
Expected: MySQL driver added

**Step 4: Install Redis client**

Run:
```bash
cd /Users/abhay.sahani/Documents/MyRepos/leal-hospital/server && go get github.com/redis/go-redis/v9
```
Expected: Redis client added

**Step 5: Install logging dependencies**

Run:
```bash
cd /Users/abhay.sahani/Documents/MyRepos/leal-hospital/server && go get github.com/rs/zerolog && go get gopkg.in/natefinch/lumberjack.v2
```
Expected: Logging libraries added

**Step 6: Install utility libraries**

Run:
```bash
cd /Users/abhay.sahani/Documents/MyRepos/leal-hospital/server && go get github.com/shopspring/decimal && go get github.com/spf13/cast && go get gopkg.in/yaml.v3
```
Expected: Utility libraries added (yaml for errors.yaml parsing)

**Step 7: Verify all dependencies**

Run: `cd /Users/abhay.sahani/Documents/MyRepos/leal-hospital/server && go mod tidy`
Expected: go.mod and go.sum updated

---

## Task 3: Create Configuration Files

**Files:**
- Create: `server/config/config.go`
- Create: `server/config/errors.yaml`
- Create: `server/config/.env`

**Step 1: Write configuration structure**

Create `server/config/config.go`:
```go
package config

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/leal-hospital/server/logger"
	"github.com/spf13/viper"
)

// AppConfig defines application configuration
type AppConfig struct {
	Environment string
	Server      ServerConfig
	Database    DBConfig
	Redis       RedisConfig
	Security    SecurityConfig
	Logging     LoggingConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port         int
	ReadTimeout  int
	WriteTimeout int
}

// DBConfig holds database configuration
type DBConfig struct {
	Type     string // "mysql", "memory", etc.
	Host     string
	Port     string
	Username string
	Password string
	Database string
	DSN      string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Type     string // "redis", "memory", etc.
	Host     string
	Port     string
	UserName string
	Password string
	Db       string
	URL      string
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	JWTSecret            string
	OTPExpiry            int // minutes
	SessionExpiry        int // hours
	MaxOTPRetries        int
	PasswordMinLength    int
	SessionInactivityMax int // hours
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string
	FilePath   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
}

// LoadConfig initializes the configuration for the service
func LoadConfig(env string) *AppConfig {
	// If env is empty, use environment variable or default
	if env == "" {
		env = viper.GetString("APP_ENV")
		if env == "" {
			env = "development"
		}
	}

	var configPath string
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()

	viper.AutomaticEnv()

	// Load from config file
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		configPath = os.Getenv("CONFIG_PATH")
		if configPath != "" {
			viper.AddConfigPath(configPath)
		}
		_, b, _, _ := runtime.Caller(0)
		basePath := filepath.Dir(b)

		logger.Info("BasePath", basePath)
		viper.AddConfigPath(filepath.Join(basePath)) // Look for .env in config directory
		viper.AddConfigPath(".")
		viper.SetConfigName(".env")
		viper.SetConfigType("env")
	}

	// Set defaults
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		logger.Error("Config file not found, using environment variables only: %v", err)
	}

	// Create config instance
	config := &AppConfig{
		Environment: env,
		Server: ServerConfig{
			Port:         viper.GetInt("SERVER_PORT"),
			ReadTimeout:  viper.GetInt("SERVER_READ_TIMEOUT"),
			WriteTimeout: viper.GetInt("SERVER_WRITE_TIMEOUT"),
		},
		Database: DBConfig{
			Type:     viper.GetString("DATABASE_TYPE"),
			Host:     viper.GetString("DATABASE_HOST"),
			Port:     viper.GetString("DATABASE_PORT"),
			Username: viper.GetString("DATABASE_USERNAME"),
			Password: viper.GetString("DATABASE_PASSWORD"),
			Database: viper.GetString("DATABASE_NAME"),
		},
		Redis: RedisConfig{
			Type:     viper.GetString("REDIS_TYPE"),
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetString("REDIS_PORT"),
			UserName: viper.GetString("REDIS_USERNAME"),
			Password: viper.GetString("REDIS_PASSWORD"),
			Db:       viper.GetString("REDIS_DB"),
		},
		Security: SecurityConfig{
			JWTSecret:            viper.GetString("JWT_SECRET"),
			OTPExpiry:            viper.GetInt("OTP_EXPIRY"),
			SessionExpiry:        viper.GetInt("SESSION_EXPIRY"),
			MaxOTPRetries:        viper.GetInt("MAX_OTP_RETRIES"),
			PasswordMinLength:    viper.GetInt("PASSWORD_MIN_LENGTH"),
			SessionInactivityMax: viper.GetInt("SESSION_INACTIVITY_MAX"),
		},
		Logging: LoggingConfig{
			Level:      viper.GetString("LOG_LEVEL"),
			FilePath:   viper.GetString("LOG_FILE_PATH"),
			MaxSize:    viper.GetInt("LOG_MAX_SIZE"),
			MaxBackups: viper.GetInt("LOG_MAX_BACKUPS"),
			MaxAge:     viper.GetInt("LOG_MAX_AGE"),
		},
	}

	// Build connection strings
	if config.Database.Type == "mysql" {
		config.Database.DSN = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
			config.Database.Username,
			config.Database.Password,
			config.Database.Host,
			config.Database.Port,
			config.Database.Database,
		)
	}

	if config.Redis.Type == "redis" {
		config.Redis.URL = fmt.Sprintf("redis://%s:%s",
			config.Redis.Host,
			config.Redis.Port,
		)
	}

	return config
}

func setDefaults() {
	// Environment
	viper.SetDefault("APP_ENV", "development")

	// Server
	viper.SetDefault("SERVER_PORT", 8080)
	viper.SetDefault("SERVER_READ_TIMEOUT", 30)
	viper.SetDefault("SERVER_WRITE_TIMEOUT", 30)

	// Database
	viper.SetDefault("DATABASE_TYPE", "mysql")
	viper.SetDefault("DATABASE_HOST", "localhost")
	viper.SetDefault("DATABASE_PORT", "3306")

	// Redis
	viper.SetDefault("REDIS_TYPE", "redis")
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_DB", "0")

	// Logging
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_FILE_PATH", "logs/app.log")
	viper.SetDefault("LOG_MAX_SIZE", 100)
	viper.SetDefault("LOG_MAX_BACKUPS", 3)
	viper.SetDefault("LOG_MAX_AGE", 28)

	// Security
	viper.SetDefault("OTP_EXPIRY", 1)           // 1 minute
	viper.SetDefault("SESSION_EXPIRY", 1)       // 1 hour
	viper.SetDefault("MAX_OTP_RETRIES", 3)
	viper.SetDefault("PASSWORD_MIN_LENGTH", 8)
	viper.SetDefault("SESSION_INACTIVITY_MAX", 1) // 1 hour
}
```

**Step 2: Create errors.yaml file**

Create `server/config/errors.yaml`:
```yaml
errors:
  INTERNAL:
    "400":
      type: "BAD_REQUEST"
      message: "Bad request"
      display_message: "Invalid request. Please check your input."
      decline_type: "BD"
      source: "INTERNAL"

    "401":
      type: "AUTH_ERROR"
      message: "Unauthorized access"
      display_message: "You need to be logged in to access this resource."
      decline_type: "BD"
      source: "INTERNAL"

    "403":
      type: "AUTH_ERROR"
      message: "Forbidden access"
      display_message: "You don't have permission to access this resource."
      decline_type: "BD"
      source: "INTERNAL"

    "404":
      type: "NOT_FOUND"
      message: "Resource not found"
      display_message: "The requested resource was not found."
      decline_type: "BD"
      source: "INTERNAL"

    "500":
      type: "SYSTEM_ERROR"
      message: "Internal server error"
      display_message: "Something went wrong on our end. Please try again later."
      decline_type: "TD"
      source: "INTERNAL"

    "503":
      type: "DATABASE_ERROR"
      message: "Database operation failed"
      display_message: "We're experiencing technical difficulties. Please try again."
      decline_type: "TD"
      source: "INTERNAL"

  AUTH_SERVICE:
    "AUTH001":
      type: "AUTH_ERROR"
      message: "Invalid credentials"
      display_message: "Mobile number or password is incorrect."
      decline_type: "BD"
      source: "AUTH_SERVICE"

    "AUTH002":
      type: "AUTH_ERROR"
      message: "OTP expired"
      display_message: "OTP has expired. Please request a new one."
      decline_type: "BD"
      source: "AUTH_SERVICE"

    "AUTH003":
      type: "AUTH_ERROR"
      message: "Invalid OTP"
      display_message: "The OTP you entered is incorrect."
      decline_type: "BD"
      source: "AUTH_SERVICE"

    "AUTH004":
      type: "AUTH_ERROR"
      message: "Session expired"
      display_message: "Your session has expired. Please login again."
      decline_type: "BD"
      source: "AUTH_SERVICE"

    "AUTH005":
      type: "BUSINESS_ERROR"
      message: "Maximum OTP attempts exceeded"
      display_message: "You have exceeded maximum OTP attempts. Please try again later."
      decline_type: "BD"
      source: "AUTH_SERVICE"

    "AUTH006":
      type: "BUSINESS_ERROR"
      message: "Staff not approved"
      display_message: "Your registration is pending approval from admin."
      decline_type: "BD"
      source: "AUTH_SERVICE"

  USER_SERVICE:
    "USR001":
      type: "NOT_FOUND"
      message: "User not found"
      display_message: "User not found."
      decline_type: "BD"
      source: "USER_SERVICE"

    "USR002":
      type: "BUSINESS_ERROR"
      message: "User already exists"
      display_message: "A user with this mobile number already exists."
      decline_type: "BD"
      source: "USER_SERVICE"

    "USR003":
      type: "BAD_REQUEST"
      message: "Invalid mobile number"
      display_message: "Please enter a valid 10-digit mobile number."
      decline_type: "BD"
      source: "USER_SERVICE"

    "USR004":
      type: "BAD_REQUEST"
      message: "Invalid password format"
      display_message: "Password must be at least 8 characters with letters, numbers and special characters."
      decline_type: "BD"
      source: "USER_SERVICE"

  PATIENT_SERVICE:
    "PAT001":
      type: "NOT_FOUND"
      message: "Patient not found"
      display_message: "Patient record not found."
      decline_type: "BD"
      source: "PATIENT_SERVICE"

    "PAT002":
      type: "BAD_REQUEST"
      message: "Invalid patient data"
      display_message: "Please provide valid patient information."
      decline_type: "BD"
      source: "PATIENT_SERVICE"

    "PAT003":
      type: "BUSINESS_ERROR"
      message: "Patient already registered today"
      display_message: "This patient has already been registered today."
      decline_type: "BD"
      source: "PATIENT_SERVICE"

  OPD_SERVICE:
    "OPD001":
      type: "NOT_FOUND"
      message: "OPD record not found"
      display_message: "OPD record not found."
      decline_type: "BD"
      source: "OPD_SERVICE"

    "OPD002":
      type: "BAD_REQUEST"
      message: "Invalid OPD data"
      display_message: "Please provide valid OPD information."
      decline_type: "BD"
      source: "OPD_SERVICE"

    "OPD003":
      type: "BUSINESS_ERROR"
      message: "OPD already exists for this visit"
      display_message: "OPD record already exists for this patient visit."
      decline_type: "BD"
      source: "OPD_SERVICE"
```

**Step 3: Create .env file**

Create `server/config/.env`:
```env
# Application
APP_ENV=development

# Server
SERVER_PORT=8080
SERVER_READ_TIMEOUT=30
SERVER_WRITE_TIMEOUT=30

# Database
DATABASE_TYPE=mysql
DATABASE_HOST=localhost
DATABASE_PORT=3306
DATABASE_USERNAME=root
DATABASE_PASSWORD=your_password_here
DATABASE_NAME=lael_hospital

# Redis
REDIS_TYPE=redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_USERNAME=
REDIS_PASSWORD=
REDIS_DB=0

# Logging
LOG_LEVEL=debug
LOG_FILE_PATH=logs/app.log
LOG_MAX_SIZE=100
LOG_MAX_BACKUPS=3
LOG_MAX_AGE=28

# Security
JWT_SECRET=your-secret-key-change-in-production
OTP_EXPIRY=1
SESSION_EXPIRY=1
MAX_OTP_RETRIES=3
PASSWORD_MIN_LENGTH=8
SESSION_INACTIVITY_MAX=1
```

**Step 4: Create .gitignore**

Create `server/.gitignore`:
```
# Environment files
.env
.env.local
.env.*.local

# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/
dist/

# Go
vendor/
*.test
*.out
coverage.txt

# Logs
logs/
*.log

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db
```

**Step 5: Verify config files**

Run: `ls -la /Users/abhay.sahani/Documents/MyRepos/leal-hospital/server/config/`
Expected: See config.go, errors.yaml, .env files

---

## Task 4: Create Database Schema

**Files:**
- Create: `server/models/dbConf/schema.sql`

**Step 1: Write database schema**

```sql
-- Lael Hospital Database Schema

-- Users table (Admin and Staff)
CREATE TABLE `lael_users` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `mobile` varchar(15) NOT NULL,
  `email` varchar(255) DEFAULT NULL,
  `designation` enum('doctor','nurse','staff') NOT NULL,
  `status` enum('active','inactive','temporary_inactive') NOT NULL DEFAULT 'active',
  `is_admin` tinyint(1) NOT NULL DEFAULT '0',
  `is_approved` tinyint(1) NOT NULL DEFAULT '0',
  `approved_by` bigint DEFAULT NULL,
  `password_hash` varchar(255) DEFAULT NULL,
  `created_on` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_on` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `last_login_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `mobile` (`mobile`),
  KEY `idx_mobile` (`mobile`),
  KEY `idx_email` (`email`),
  KEY `idx_is_admin` (`is_admin`),
  KEY `approved_by` (`approved_by`),
  CONSTRAINT `lael_users_ibfk_1` FOREIGN KEY (`approved_by`) REFERENCES `lael_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- OTP table
CREATE TABLE `lael_otp` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `mobile` varchar(15) NOT NULL,
  `otp` varchar(6) NOT NULL,
  `expiry` datetime NOT NULL,
  `is_validated` tinyint(1) NOT NULL DEFAULT '0',
  `otp_type` enum('registration','login','forgot_password') NOT NULL,
  `retry_count` int NOT NULL DEFAULT '0',
  `created_on` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_on` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_mobile_type` (`mobile`,`otp_type`),
  KEY `idx_expiry` (`expiry`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Patients table
CREATE TABLE `lael_patients` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `mobile` varchar(15) NOT NULL,
  `opd_id` varchar(50) NOT NULL,
  `age` int NOT NULL,
  `sex` enum('male','female','other') NOT NULL,
  `address_locality` varchar(255) DEFAULT NULL,
  `address_city` varchar(100) DEFAULT NULL,
  `address_state` varchar(100) DEFAULT NULL,
  `address_pincode` varchar(10) DEFAULT NULL,
  `visit_number` int NOT NULL DEFAULT '1',
  `created_on` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_on` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `opd_id` (`opd_id`),
  KEY `idx_mobile` (`mobile`),
  KEY `idx_opd_id` (`opd_id`),
  KEY `idx_created_on` (`created_on`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Patient OPD records
CREATE TABLE `patient_opd` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `patient_id` bigint NOT NULL,
  `doctor_id` bigint NOT NULL,
  `symptoms` json DEFAULT NULL,
  `prescription` json DEFAULT NULL,
  `medicines` json DEFAULT NULL,
  `future_suggestion` json DEFAULT NULL,
  `template_version` int NOT NULL DEFAULT '1',
  `created_on` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_on` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_patient_id` (`patient_id`),
  KEY `idx_doctor_id` (`doctor_id`),
  KEY `idx_created_on` (`created_on`),
  CONSTRAINT `patient_opd_ibfk_1` FOREIGN KEY (`patient_id`) REFERENCES `lael_patients` (`id`),
  CONSTRAINT `patient_opd_ibfk_2` FOREIGN KEY (`doctor_id`) REFERENCES `lael_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Sessions table
CREATE TABLE `lael_sessions` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL,
  `session_token` varchar(255) NOT NULL,
  `session_expiry` datetime NOT NULL,
  `last_active_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `device_info` varchar(500) DEFAULT NULL,
  `ip_address` varchar(45) DEFAULT NULL,
  `created_on` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `session_token` (`session_token`),
  KEY `idx_session_token` (`session_token`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_session_expiry` (`session_expiry`),
  CONSTRAINT `lael_sessions_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `lael_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Activity logs table
CREATE TABLE `lael_activity_logs` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint DEFAULT NULL,
  `activity_type` enum('login','logout','patient_creation','staff_approval','opd_generation') NOT NULL,
  `description` text,
  `ip_address` varchar(45) DEFAULT NULL,
  `created_on` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_activity_type` (`activity_type`),
  KEY `idx_created_on` (`created_on`),
  CONSTRAINT `lael_activity_logs_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `lael_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
```

**Step 2: Verify schema file**

Run: `cat /Users/abhay.sahani/Documents/MyRepos/leal-hospital/server/models/dbConf/schema.sql | head -20`
Expected: See schema content without IF NOT EXISTS

---

## Task 5: Create SQLC Query Files

**Files:**
- Create: `server/models/dbConf/users.sql`
- Create: `server/models/dbConf/otp.sql`
- Create: `server/models/dbConf/patients.sql`
- Create: `server/models/dbConf/patient_opd.sql`
- Create: `server/models/dbConf/sessions.sql`

**Step 1: Write users queries**

Create `server/models/dbConf/users.sql`:
```sql
-- name: CreateUser :execresult
INSERT INTO lael_users (
    name, mobile, email, designation, is_admin, is_approved, password_hash
) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetUserByID :one
SELECT * FROM lael_users WHERE id = ?;

-- name: GetUserByMobile :one
SELECT * FROM lael_users WHERE mobile = ?;

-- name: GetUserByEmail :one
SELECT * FROM lael_users WHERE email = ?;

-- name: ListPendingStaffApprovals :many
SELECT * FROM lael_users
WHERE is_admin = FALSE AND is_approved = FALSE
ORDER BY created_on DESC;

-- name: ListStaffByStatus :many
SELECT * FROM lael_users
WHERE is_admin = FALSE AND status = ?
ORDER BY name ASC
LIMIT ? OFFSET ?;

-- name: UpdateUserStatus :exec
UPDATE lael_users
SET status = ?, updated_on = NOW()
WHERE id = ?;

-- name: ApproveStaff :exec
UPDATE lael_users
SET is_approved = TRUE, approved_by = ?, updated_on = NOW()
WHERE id = ?;

-- name: UpdateUserPassword :exec
UPDATE lael_users
SET password_hash = ?, updated_on = NOW()
WHERE id = ?;

-- name: UpdateLastLogin :exec
UPDATE lael_users
SET last_login_at = NOW(), updated_on = NOW()
WHERE id = ?;
```

**Step 2: Write OTP queries**

Create `server/models/dbConf/otp.sql`:
```sql
-- name: CreateOTP :execresult
INSERT INTO lael_otp (
    mobile, otp, expiry, otp_type, retry_count
) VALUES (?, ?, ?, ?, 0);

-- name: GetLatestOTP :one
SELECT * FROM lael_otp
WHERE mobile = ? AND otp_type = ? AND is_validated = FALSE
ORDER BY created_on DESC
LIMIT 1;

-- name: ValidateOTP :exec
UPDATE lael_otp
SET is_validated = TRUE, updated_on = NOW()
WHERE id = ?;

-- name: IncrementRetryCount :exec
UPDATE lael_otp
SET retry_count = retry_count + 1, updated_on = NOW()
WHERE id = ?;

-- name: DeleteExpiredOTP :exec
DELETE FROM lael_otp
WHERE expiry < NOW();
```

**Step 3: Write patients queries**

Create `server/models/dbConf/patients.sql`:
```sql
-- name: CreatePatient :execresult
INSERT INTO lael_patients (
    name, mobile, opd_id, age, sex,
    address_locality, address_city, address_state, address_pincode,
    visit_number
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetPatientByID :one
SELECT * FROM lael_patients WHERE id = ?;

-- name: GetPatientByMobile :one
SELECT * FROM lael_patients
WHERE mobile = ?
ORDER BY created_on DESC
LIMIT 1;

-- name: GetPatientByOPDID :one
SELECT * FROM lael_patients WHERE opd_id = ?;

-- name: SearchPatientsByMobile :many
SELECT * FROM lael_patients
WHERE mobile LIKE ?
ORDER BY created_on DESC
LIMIT ? OFFSET ?;

-- name: SearchPatientsByName :many
SELECT * FROM lael_patients
WHERE name LIKE ?
ORDER BY created_on DESC
LIMIT ? OFFSET ?;

-- name: GetTodayPatients :many
SELECT * FROM lael_patients
WHERE DATE(created_on) = CURDATE()
ORDER BY created_on DESC;

-- name: GetPatientsByDateRange :many
SELECT * FROM lael_patients
WHERE created_on BETWEEN ? AND ?
ORDER BY created_on DESC
LIMIT ? OFFSET ?;

-- name: UpdatePatientVisitNumber :exec
UPDATE lael_patients
SET visit_number = visit_number + 1, updated_on = NOW()
WHERE mobile = ?;
```

**Step 4: Write patient OPD queries**

Create `server/models/dbConf/patient_opd.sql`:
```sql
-- name: CreatePatientOPD :execresult
INSERT INTO patient_opd (
    patient_id, doctor_id, symptoms, prescription,
    medicines, future_suggestion, template_version
) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetPatientOPDByID :one
SELECT * FROM patient_opd WHERE id = ?;

-- name: GetPatientOPDsByPatientID :many
SELECT * FROM patient_opd
WHERE patient_id = ?
ORDER BY created_on DESC;

-- name: GetLatestPatientOPD :one
SELECT * FROM patient_opd
WHERE patient_id = ?
ORDER BY created_on DESC
LIMIT 1;

-- name: GetPatientOPDsByDoctor :many
SELECT po.*, lp.name as patient_name, lp.mobile as patient_mobile
FROM patient_opd po
JOIN lael_patients lp ON po.patient_id = lp.id
WHERE po.doctor_id = ?
ORDER BY po.created_on DESC
LIMIT ? OFFSET ?;
```

**Step 5: Write sessions queries**

Create `server/models/dbConf/sessions.sql`:
```sql
-- name: CreateSession :execresult
INSERT INTO lael_sessions (
    user_id, session_token, session_expiry, device_info, ip_address
) VALUES (?, ?, ?, ?, ?);

-- name: GetSessionByToken :one
SELECT * FROM lael_sessions WHERE session_token = ?;

-- name: UpdateSessionActivity :exec
UPDATE lael_sessions
SET last_active_at = NOW()
WHERE session_token = ?;

-- name: DeleteSession :exec
DELETE FROM lael_sessions WHERE session_token = ?;

-- name: DeleteExpiredSessions :exec
DELETE FROM lael_sessions WHERE session_expiry < NOW();

-- name: GetUserSessions :many
SELECT * FROM lael_sessions
WHERE user_id = ?
ORDER BY last_active_at DESC;
```

**Step 6: Verify query files**

Run: `ls -la /Users/abhay.sahani/Documents/MyRepos/leal-hospital/server/models/dbConf/*.sql`
Expected: See all SQL files

---

## Task 6: Configure SQLC

**Files:**
- Create: `server/models/dbConf/sqlc.yaml`

**Step 1: Write SQLC configuration**

```yaml
version: "2"
sql:
  - engine: "mysql"
    queries:
      - "users.sql"
      - "otp.sql"
      - "patients.sql"
      - "patient_opd.sql"
      - "sessions.sql"
    schema: "schema.sql"
    gen:
      go:
        package: "db"
        sql_package: "database/sql"
        sql_driver: "github.com/go-sql-driver/mysql"
        out: "../db"
        emit_json_tags: true
        emit_prepared_queries: false
        emit_interface: true
        emit_exact_table_names: false
        emit_empty_slices: true
        overrides:
          - db_type: "datetime"
            go_type: "time.Time"
            nullable: false
```

**Step 2: Install SQLC**

Run: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
Expected: SQLC installed globally

**Step 3: Generate SQLC code**

Run: `cd /Users/abhay.sahani/Documents/MyRepos/leal-hospital/server/models/dbConf && sqlc generate`
Expected: Generated files in ../db/

**Step 4: Verify generated code**

Run: `ls -la /Users/abhay.sahani/Documents/MyRepos/leal-hospital/server/models/db/`
Expected: See db.go, models.go, querier.go, and query files

---

## Task 7: Create Error Management System with Registry

**Files:**
- Create: `server/medierror/error.go`
- Create: `server/medierror/loaderrorcode.go`

**Step 1: Write error types and structures**

Create `server/medierror/error.go`:
```go
package medierror

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

// ErrorType represents the category of error
type ErrorType string

const (
	ValidationError ErrorType = "BAD_REQUEST"
	SystemError     ErrorType = "SYSTEM_ERROR"
	DatabaseError   ErrorType = "DATABASE_ERROR"
	BusinessError   ErrorType = "BUSINESS_ERROR"
	AuthError       ErrorType = "AUTH_ERROR"
	NotFoundError   ErrorType = "NOT_FOUND"
	GatewayErr      ErrorType = "GATEWAY_ERROR"
	Failure         ErrorType = "FAILED"
)

// ErrorCode represents numeric error codes
type ErrorCode int

const (
	ErrValidation   ErrorCode = 400
	ErrUnauthorized ErrorCode = 401
	ErrForbidden    ErrorCode = 403
	ErrNotFound     ErrorCode = 404
	ErrInternal     ErrorCode = 500
	ErrDatabase     ErrorCode = 503
	ErrBusiness     ErrorCode = 600
	ErrGateway      ErrorCode = 453
)

// FieldViolation represents a validation error on a specific field
type FieldViolation struct {
	Field       string `json:"field"`
	Description string `json:"description"`
}

// ErrorDetail provides structured error context
type ErrorDetail struct {
	FieldViolations []FieldViolation `json:"fieldViolations,omitempty"`
}

// AppError represents application-specific errors
type AppError struct {
	code           ErrorCode
	errorType      ErrorType
	message        string
	displayMessage string
	details        []ErrorDetail
	stack          []string
	cause          error
	declineType    string
	source         string
}

// Error implements error interface
func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.message, e.cause)
	}
	return e.message
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.cause
}

// HTTPStatus returns appropriate HTTP status code
func (e *AppError) HTTPStatus() int {
	return int(e.code)
}

// Newf creates a new AppError
func Newf(errType ErrorType, code ErrorCode, format string, args ...interface{}) *AppError {
	return &AppError{
		errorType: errType,
		code:      code,
		message:   fmt.Sprintf(format, args...),
		stack:     captureStack(),
	}
}

// Wrapf wraps an existing error
func Wrapf(err error, errType ErrorType, code ErrorCode, format string, args ...interface{}) *AppError {
	if err == nil {
		return nil
	}

	// If already an AppError, preserve it
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}

	return &AppError{
		errorType: errType,
		code:      code,
		message:   fmt.Sprintf(format, args...),
		cause:     err,
		stack:     captureStack(),
	}
}

// WithDisplayMessage adds user-friendly message
func (e *AppError) WithDisplayMessage(msg string) *AppError {
	e.displayMessage = msg
	return e
}

// WithDetails adds structured error details
func (e *AppError) WithDetails(details []ErrorDetail) *AppError {
	e.details = append(e.details, details...)
	return e
}

// ToResponse converts error to API response format
func (e *AppError) ToResponse() ErrorResponse {
	httpCode := e.HTTPStatus()
	responseCode := fmt.Sprintf("%d", httpCode)
	responseMsg := "FAILED"

	if httpCode >= 200 && httpCode < 300 {
		responseMsg = "SUCCESS"
	}

	return ErrorResponse{
		Code: responseCode,
		Msg:  responseMsg,
		Model: &ErrorModel{
			ErrorCode:      string(e.errorType),
			Message:        e.message,
			DisplayMessage: e.getDisplayMessage(),
			Details:        e.details,
		},
	}
}

func (e *AppError) getDisplayMessage() string {
	if e.displayMessage != "" {
		return e.displayMessage
	}
	return e.message
}

// ErrorResponse is the API response format
type ErrorResponse struct {
	Code  string      `json:"code"`
	Msg   string      `json:"msg"`
	Model *ErrorModel `json:"model"`
}

// ErrorModel contains detailed error information
type ErrorModel struct {
	ErrorCode      string        `json:"errorCode"`
	Message        string        `json:"message"`
	DisplayMessage string        `json:"displayMessage,omitempty"`
	Details        []ErrorDetail `json:"details,omitempty"`
}

// captureStack captures the current stack trace
func captureStack() []string {
	var stack []string
	for i := 2; i < 10; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			break
		}
		// Skip runtime frames
		if strings.HasPrefix(fn.Name(), "runtime.") {
			continue
		}
		stack = append(stack, fmt.Sprintf("%s:%d %s", file, line, fn.Name()))
	}
	return stack
}

// Common error constructors

// ErrInvalidCredentials creates authentication error
func ErrInvalidCredentials() *AppError {
	return Newf(AuthError, ErrUnauthorized, "Invalid credentials").
		WithDisplayMessage("Mobile number or password is incorrect")
}

// ErrOTPExpired creates OTP expiry error
func ErrOTPExpired() *AppError {
	return Newf(AuthError, ErrUnauthorized, "OTP has expired").
		WithDisplayMessage("OTP has expired. Please request a new one")
}

// ErrOTPInvalid creates invalid OTP error
func ErrOTPInvalid() *AppError {
	return Newf(AuthError, ErrUnauthorized, "Invalid OTP").
		WithDisplayMessage("The OTP you entered is incorrect")
}

// ErrSessionExpired creates session expiry error
func ErrSessionExpired() *AppError {
	return Newf(AuthError, ErrUnauthorized, "Session has expired").
		WithDisplayMessage("Your session has expired. Please login again")
}

// ErrUnauthorizedAccess creates unauthorized error
func ErrUnauthorizedAccess() *AppError {
	return Newf(AuthError, ErrForbidden, "Unauthorized access").
		WithDisplayMessage("You don't have permission to access this resource")
}

// ErrInvalidMobile creates invalid mobile error
func ErrInvalidMobile() *AppError {
	return Newf(ValidationError, ErrValidation, "Invalid mobile number").
		WithDisplayMessage("Please enter a valid 10-digit mobile number")
}

// ErrInvalidPassword creates invalid password error
func ErrInvalidPassword() *AppError {
	return Newf(ValidationError, ErrValidation, "Invalid password format").
		WithDisplayMessage("Password must be at least 8 characters with letters, numbers and special characters")
}

// ErrMissingField creates missing field error
func ErrMissingField(field string) *AppError {
	return Newf(ValidationError, ErrValidation, "Missing required field: %s", field).
		WithDisplayMessage(fmt.Sprintf("%s is required", field))
}

// ErrUserAlreadyExists creates user exists error
func ErrUserAlreadyExists() *AppError {
	return Newf(BusinessError, ErrBusiness, "User already exists").
		WithDisplayMessage("A user with this mobile number already exists")
}

// ErrUserNotFound creates user not found error
func ErrUserNotFound() *AppError {
	return Newf(NotFoundError, ErrNotFound, "User not found").
		WithDisplayMessage("User not found")
}

// ErrPatientNotFound creates patient not found error
func ErrPatientNotFound() *AppError {
	return Newf(NotFoundError, ErrNotFound, "Patient not found").
		WithDisplayMessage("Patient record not found")
}

// ErrStaffNotApproved creates staff not approved error
func ErrStaffNotApproved() *AppError {
	return Newf(BusinessError, ErrBusiness, "Staff not approved").
		WithDisplayMessage("Your registration is pending approval from admin")
}

// ErrOTPMaxRetries creates max retries error
func ErrOTPMaxRetries() *AppError {
	return Newf(BusinessError, ErrBusiness, "Maximum OTP attempts exceeded").
		WithDisplayMessage("You have exceeded maximum OTP attempts. Please try again later")
}

// ErrDatabaseOperation creates database error
func ErrDatabaseOperation(operation string) *AppError {
	return Newf(DatabaseError, ErrDatabase, "Database operation failed: %s", operation).
		WithDisplayMessage("We're experiencing technical difficulties. Please try again")
}

// ErrInternalServer creates internal server error
func ErrInternalServer() *AppError {
	return Newf(SystemError, ErrInternal, "Internal server error").
		WithDisplayMessage("Something went wrong. Please try again later")
}
```

**Step 2: Write error registry loader**

Create `server/medierror/loaderrorcode.go`:
```go
package medierror

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

const ErrorRegistryKey = "error_registry"

// ErrorRegistry holds all errors loaded from errors.yaml
type ErrorRegistry struct {
	errors map[string]*AppError
	mu     sync.RWMutex
}

// ErrorConfig represents the structure of errors.yaml
type ErrorConfig struct {
	Errors map[string]map[string]ErrorConfigItem `yaml:"errors"`
}

// ErrorConfigItem represents individual error configuration
type ErrorConfigItem struct {
	Type           string `yaml:"type"`
	Message        string `yaml:"message"`
	DisplayMessage string `yaml:"display_message"`
	DeclineType    string `yaml:"decline_type"`
	Source         string `yaml:"source"`
}

var (
	globalRegistry *ErrorRegistry
	once           sync.Once
)

// InitErrorRegistry initializes the error registry from errors.yaml
func InitErrorRegistry(router *gin.Engine) error {
	var err error
	once.Do(func() {
		globalRegistry, err = loadErrorRegistry()
		if err != nil {
			return
		}

		// Register middleware to inject registry into context
		router.Use(func(c *gin.Context) {
			c.Set(ErrorRegistryKey, globalRegistry)
			c.Next()
		})
	})
	return err
}

// InitializeErrorRegistry creates a new error registry (for testing)
func InitializeErrorRegistry() (*ErrorRegistry, error) {
	return loadErrorRegistry()
}

// loadErrorRegistry loads errors from errors.yaml
func loadErrorRegistry() (*ErrorRegistry, error) {
	registry := &ErrorRegistry{
		errors: make(map[string]*AppError),
	}

	// Find errors.yaml file
	_, b, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(b)
	configPath := filepath.Join(basePath, "..", "config", "errors.yaml")

	// Try multiple paths
	possiblePaths := []string{
		configPath,
		"config/errors.yaml",
		"../config/errors.yaml",
		"./config/errors.yaml",
	}

	var yamlFile []byte
	var err error
	var foundPath string

	for _, path := range possiblePaths {
		yamlFile, err = os.ReadFile(path)
		if err == nil {
			foundPath = path
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read errors.yaml from any path: %w", err)
	}

	fmt.Printf("Loaded errors.yaml from: %s\n", foundPath)

	// Parse YAML
	var config ErrorConfig
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		return nil, fmt.Errorf("failed to parse errors.yaml: %w", err)
	}

	// Build error registry
	for source, errors := range config.Errors {
		for code, errorItem := range errors {
			key := fmt.Sprintf("%s:%s", source, code)

			// Map error type string to ErrorType
			var errType ErrorType
			switch errorItem.Type {
			case "BAD_REQUEST":
				errType = ValidationError
			case "AUTH_ERROR":
				errType = AuthError
			case "NOT_FOUND":
				errType = NotFoundError
			case "BUSINESS_ERROR":
				errType = BusinessError
			case "DATABASE_ERROR":
				errType = DatabaseError
			case "SYSTEM_ERROR":
				errType = SystemError
			case "GATEWAY_ERROR":
				errType = GatewayErr
			default:
				errType = SystemError
			}

			// Map error code
			var errCode ErrorCode
			switch code {
			case "400":
				errCode = ErrValidation
			case "401":
				errCode = ErrUnauthorized
			case "403":
				errCode = ErrForbidden
			case "404":
				errCode = ErrNotFound
			case "500":
				errCode = ErrInternal
			case "503":
				errCode = ErrDatabase
			default:
				// Determine code based on type
				switch errType {
				case ValidationError:
					errCode = ErrValidation
				case AuthError:
					errCode = ErrUnauthorized
				case NotFoundError:
					errCode = ErrNotFound
				case DatabaseError:
					errCode = ErrDatabase
				case BusinessError:
					errCode = ErrBusiness
				case GatewayErr:
					errCode = ErrGateway
				default:
					errCode = ErrInternal
				}
			}

			appError := &AppError{
				code:           errCode,
				errorType:      errType,
				message:        errorItem.Message,
				displayMessage: errorItem.DisplayMessage,
				declineType:    errorItem.DeclineType,
				source:         errorItem.Source,
			}

			registry.errors[key] = appError
		}
	}

	fmt.Printf("Loaded %d errors into registry\n", len(registry.errors))
	return registry, nil
}

// GetError retrieves an error by code and source
func (r *ErrorRegistry) GetError(code, source string) *AppError {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", source, code)
	if err, ok := r.errors[key]; ok {
		// Return a copy to avoid modification of cached error
		return &AppError{
			code:           err.code,
			errorType:      err.errorType,
			message:        err.message,
			displayMessage: err.displayMessage,
			declineType:    err.declineType,
			source:         err.source,
			stack:          captureStack(),
		}
	}

	// Return generic error if not found
	return &AppError{
		code:           ErrInternal,
		errorType:      SystemError,
		message:        fmt.Sprintf("Error code mapping not found: %s:%s", source, code),
		displayMessage: "An unexpected error occurred",
		stack:          captureStack(),
	}
}

// GetErrorRegistry retrieves the error registry from gin context
func GetErrorRegistry(c *gin.Context) *ErrorRegistry {
	if registry, exists := c.Get(ErrorRegistryKey); exists {
		if r, ok := registry.(*ErrorRegistry); ok {
			return r
		}
	}
	return globalRegistry
}

// GetGlobalRegistry returns the global registry (for non-HTTP contexts)
func GetGlobalRegistry() *ErrorRegistry {
	return globalRegistry
}
```

**Step 3: Verify error files**

Run: `ls -la /Users/abhay.sahani/Documents/MyRepos/leal-hospital/server/medierror/`
Expected: See error.go and loaderrorcode.go

---

## Task 8: Create Database Manager

**Files:**
- Create: `server/utils/db/db_manager.go`
- Create: `server/utils/db/interfaces.go`

**Step 1: Write database interfaces**

Create `server/utils/db/interfaces.go`:
```go
package db

import (
	"context"
	"database/sql"
)

// DBManagerInterface manages database connections
type DBManagerInterface interface {
	Initialize() error
	GetWriteDBConn(ctx context.Context) (DBConnInterface, error)
	GetReadDBConn(ctx context.Context) (DBConnInterface, error)
	Close() error
}

// DBConnInterface wraps a database connection
type DBConnInterface interface {
	GetConn(ctx context.Context) *sql.Conn
	Close(ctx context.Context) error
	BeginTx(ctx context.Context, opts *sql.TxOptions) (TxnInterface, error)
}

// TxnInterface wraps a database transaction
type TxnInterface interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	GetTx() *sql.Tx
}
```

**Step 2: Write database manager implementation**

Create `server/utils/db/db_manager.go`:
```go
package db

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/leal-hospital/server/config"
)

const (
	defaultWriteMaxOpenConns = 25
	defaultWriteMaxIdleConns = 10
	defaultReadMaxOpenConns  = 50
	defaultReadMaxIdleConns  = 15
	defaultConnMaxLifetime   = 30 // minutes
)

var (
	instance DBManagerInterface
	once     sync.Once
)

// DBManager implements DBManagerInterface
type DBManager struct {
	writeDB *sql.DB
	readDB  *sql.DB
	config  *config.DBConfig
	mu      sync.RWMutex
}

// DBConn implements DBConnInterface
type DBConn struct {
	conn *sql.Conn
}

// Txn implements TxnInterface
type Txn struct {
	tx *sql.Tx
}

// GetDBManager returns singleton instance
func GetDBManager() DBManagerInterface {
	once.Do(func() {
		instance = &DBManager{}
	})
	return instance
}

// NewDBManager creates a new DBManager with config
func NewDBManager(cfg *config.DBConfig) DBManagerInterface {
	return &DBManager{
		config: cfg,
	}
}

// Initialize sets up database connections
func (dm *DBManager) Initialize() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if dm.config == nil {
		return fmt.Errorf("database configuration is nil")
	}

	// Open write connection
	writeDB, err := sql.Open("mysql", dm.config.DSN)
	if err != nil {
		return fmt.Errorf("failed to open write database: %w", err)
	}

	// Configure write connection pool
	writeDB.SetMaxOpenConns(defaultWriteMaxOpenConns)
	writeDB.SetMaxIdleConns(defaultWriteMaxIdleConns)

	// Ping to verify connection
	if err := writeDB.PingContext(context.Background()); err != nil {
		return fmt.Errorf("failed to ping write database: %w", err)
	}

	dm.writeDB = writeDB
	dm.readDB = writeDB // Use same connection for reads initially

	return nil
}

// GetWriteDBConn returns a write database connection
func (dm *DBManager) GetWriteDBConn(ctx context.Context) (DBConnInterface, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if dm.writeDB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	conn, err := dm.writeDB.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get write connection: %w", err)
	}

	return &DBConn{conn: conn}, nil
}

// GetReadDBConn returns a read database connection
func (dm *DBManager) GetReadDBConn(ctx context.Context) (DBConnInterface, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if dm.readDB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	conn, err := dm.readDB.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get read connection: %w", err)
	}

	return &DBConn{conn: conn}, nil
}

// Close closes all database connections
func (dm *DBManager) Close() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	var err error
	if dm.writeDB != nil {
		err = dm.writeDB.Close()
	}
	if dm.readDB != nil && dm.readDB != dm.writeDB {
		if closeErr := dm.readDB.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	return err
}

// GetConn returns the underlying connection
func (dc *DBConn) GetConn(ctx context.Context) *sql.Conn {
	return dc.conn
}

// Close closes the connection
func (dc *DBConn) Close(ctx context.Context) error {
	return dc.conn.Close()
}

// BeginTx starts a transaction
func (dc *DBConn) BeginTx(ctx context.Context, opts *sql.TxOptions) (TxnInterface, error) {
	tx, err := dc.conn.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Txn{tx: tx}, nil
}

// Commit commits the transaction
func (t *Txn) Commit(ctx context.Context) error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *Txn) Rollback(ctx context.Context) error {
	return t.tx.Rollback()
}

// GetTx returns the underlying transaction
func (t *Txn) GetTx() *sql.Tx {
	return t.tx
}
```

**Step 4: Verify database manager files**

Run: `ls -la /Users/abhay.sahani/Documents/MyRepos/leal-hospital/server/utils/db/`
Expected: See interfaces.go and db_manager.go

---

## Task 9: Create Logging Utility

**Files:**
- Create: `server/utils/logger/logger.go`

**Step 1: Write logger implementation (fixing import issue)**

Create a temporary file first to define logger package:

Create `server/logger/logger.go`:
```go
package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log zerolog.Logger

// InitLogger initializes the global logger
func InitLogger(level, logPath string, maxSize, maxBackups, maxAge int) error {
	// Parse log level
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(logLevel)

	// Create log directory if it doesn't exist
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// Configure log rotation
	fileWriter := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    maxSize,    // megabytes
		MaxBackups: maxBackups, // number of backups
		MaxAge:     maxAge,     // days
		Compress:   true,
	}

	// Console writer for development
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	// Multi-writer for console and file
	multi := zerolog.MultiLevelWriter(consoleWriter, fileWriter)

	// Create logger
	Log = zerolog.New(multi).
		With().
		Timestamp().
		Caller().
		Logger()

	return nil
}

// GetLogger returns the global logger
func GetLogger() *zerolog.Logger {
	return &Log
}

// Info logs info message
func Info(msg string, args ...interface{}) {
	if len(args) > 0 {
		Log.Info().Msgf(msg, args...)
	} else {
		Log.Info().Msg(msg)
	}
}

// Error logs error message
func Error(msg string, args ...interface{}) {
	if len(args) > 0 {
		Log.Error().Msgf(msg, args...)
	} else {
		Log.Error().Msg(msg)
	}
}

// Debug logs debug message
func Debug(msg string, args ...interface{}) {
	if len(args) > 0 {
		Log.Debug().Msgf(msg, args...)
	} else {
		Log.Debug().Msg(msg)
	}
}

// Warn logs warning message
func Warn(msg string, args ...interface{}) {
	if len(args) > 0 {
		Log.Warn().Msgf(msg, args...)
	} else {
		Log.Warn().Msg(msg)
	}
}

// Fatal logs fatal message and exits
func Fatal(msg string, args ...interface{}) {
	if len(args) > 0 {
		Log.Fatal().Msgf(msg, args...)
	} else {
		Log.Fatal().Msg(msg)
	}
}

// WithContext creates a child logger with context fields
func WithContext(fields map[string]interface{}) *zerolog.Logger {
	logger := Log.With()
	for k, v := range fields {
		logger = logger.Interface(k, v)
	}
	l := logger.Logger()
	return &l
}
```

**Step 2: Verify logger**

Run: `cat /Users/abhay.sahani/Documents/MyRepos/leal-hospital/server/logger/logger.go | head -30`
Expected: See logger implementation

---

## Task 10: Create Core Middleware

**Files:**
- Create: `server/middleware/cors.go`
- Create: `server/middleware/recovery.go`
- Create: `server/middleware/logger.go`

**Step 1: Write CORS middleware**

Create `server/middleware/cors.go`:
```go
package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware returns CORS middleware configuration
func CORSMiddleware() gin.HandlerFunc {
	config := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60, // 12 hours
	}

	return cors.New(config)
}
```

**Step 2: Write recovery middleware**

Create `server/middleware/recovery.go`:
```go
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
```

**Step 3: Write request logger middleware**

Create `server/middleware/logger.go`:
```go
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
```

**Step 4: Verify middleware files**

Run: `ls -la /Users/abhay.sahani/Documents/MyRepos/leal-hospital/server/middleware/`
Expected: See cors.go, recovery.go, logger.go

---

## Task 11: Create Dependency Injection Container

**Files:**
- Create: `server/di/container.go`

**Step 1: Write DI container implementation**

```go
package di

import (
	"fmt"
	"reflect"
	"sync"
)

// Container manages dependency injection
type Container struct {
	services  map[reflect.Type]interface{}
	factories map[reflect.Type]interface{}
	mu        sync.RWMutex
	building  map[reflect.Type]bool // Track circular dependencies
}

// NewContainer creates a new DI container
func NewContainer() *Container {
	return &Container{
		services:  make(map[reflect.Type]interface{}),
		factories: make(map[reflect.Type]interface{}),
		building:  make(map[reflect.Type]bool),
	}
}

// Register registers a service instance
func (c *Container) Register(interfacePtr interface{}, implementation interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	t := reflect.TypeOf(interfacePtr).Elem()
	c.services[t] = implementation
}

// RegisterFactory registers a factory function for lazy initialization
func (c *Container) RegisterFactory(interfacePtr interface{}, factory interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	t := reflect.TypeOf(interfacePtr).Elem()
	c.factories[t] = factory
}

// Get retrieves a service from the container
func (c *Container) Get(interfacePtr interface{}) interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()

	t := reflect.TypeOf(interfacePtr).Elem()

	// Check if already instantiated
	if service, ok := c.services[t]; ok {
		return service
	}

	// Check for factory
	if factory, ok := c.factories[t]; ok {
		// Check for circular dependency
		if c.building[t] {
			panic(fmt.Sprintf("circular dependency detected for type: %v", t))
		}

		// Mark as building
		c.building[t] = true
		defer delete(c.building, t)

		// Call factory
		factoryValue := reflect.ValueOf(factory)
		if factoryValue.Kind() != reflect.Func {
			panic(fmt.Sprintf("factory for type %v is not a function", t))
		}

		// Invoke factory with container as argument
		results := factoryValue.Call([]reflect.Value{reflect.ValueOf(c)})
		if len(results) != 1 {
			panic(fmt.Sprintf("factory for type %v must return exactly one value", t))
		}

		service := results[0].Interface()
		c.services[t] = service
		return service
	}

	panic(fmt.Sprintf("no service or factory registered for type: %v", t))
}

// Has checks if a service is registered
func (c *Container) Has(interfacePtr interface{}) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	t := reflect.TypeOf(interfacePtr).Elem()
	_, hasService := c.services[t]
	_, hasFactory := c.factories[t]
	return hasService || hasFactory
}

// Clear removes all registered services
func (c *Container) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services = make(map[reflect.Type]interface{})
	c.factories = make(map[reflect.Type]interface{})
	c.building = make(map[reflect.Type]bool)
}
```

**Step 2: Verify DI container**

Run: `cat /Users/abhay.sahani/Documents/MyRepos/leal-hospital/server/di/container.go | head -40`
Expected: See container implementation

---

## Task 12: Create Application Bootstrap

**Files:**
- Create: `server/app/app.go`
- Create: `server/app/module.go`

**Step 1: Write module interface**

Create `server/app/module.go`:
```go
package app

import (
	"github.com/gin-gonic/gin"
	"github.com/leal-hospital/server/di"
)

// Module represents a feature module
type Module interface {
	Name() string
	Configure(container *di.Container)
	RegisterRoutes(router *gin.Engine, container *di.Container)
}
```

**Step 2: Write application bootstrap**

Create `server/app/app.go`:
```go
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
		Addr:         fmt.Sprintf(":%d", a.Config.Server.Port),
		Handler:      a.Router,
		ReadTimeout:  time.Duration(a.Config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(a.Config.Server.WriteTimeout) * time.Second,
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
		logger.Info("Starting HTTP server on port %d (environment: %s)",
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
```

**Step 3: Verify app files**

Run: `ls -la /Users/abhay.sahani/Documents/MyRepos/leal-hospital/server/app/`
Expected: See module.go and app.go

---

## Task 13: Create Main Entry Point

**Files:**
- Create: `server/cmd/main.go`

**Step 1: Write main.go**

```go
package main

import (
	"log"

	"github.com/leal-hospital/server/app"
	"github.com/leal-hospital/server/config"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig("")
	if cfg == nil {
		log.Fatal("Failed to load configuration")
	}

	// Initialize modules
	modules := []app.Module{
		// Modules will be added here as we implement them
		// Example: &auth.AuthModule{},
	}

	// Create and bootstrap application
	application := app.NewApp(cfg, modules)
	if err := application.Bootstrap(); err != nil {
		log.Fatalf("Failed to bootstrap application: %v", err)
	}

	// Run application
	if err := application.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
```

**Step 2: Update go.mod with correct module path**

Run: `cd /Users/abhay.sahani/Documents/MyRepos/leal-hospital/server && go mod tidy`
Expected: Dependencies resolved

**Step 3: Build the application**

Run: `cd /Users/abhay.sahani/Documents/MyRepos/leal-hospital/server && go build -o bin/server cmd/main.go`
Expected: Binary created in bin/server (errors are expected since logger package needs fixing)

---

## Task 14: Update README Documentation

**Files:**
- Modify: `server/readme.md`

**Step 1: Write comprehensive README**

```markdown
# Lael Hospital - Backend

Backend server for Lael Hospital internal management system.

## Tech Stack

- **Language:** Go 1.23+
- **HTTP Framework:** Gin
- **Database:** MySQL with SQLC
- **Caching:** Redis
- **Configuration:** Viper
- **Logging:** Zerolog with Lumberjack rotation
- **Error Handling:** YAML-based error registry

## Architecture

This project follows clean architecture principles with clear separation of concerns:

```
server/
├── cmd/                    # Application entry point
├── app/                    # Application bootstrap
├── handler/                # HTTP request handlers
├── services/               # Business logic
├── persistence/            # Data access layer
├── domain/                 # Domain models and interfaces
├── models/                 # SQLC configuration and generated code
├── config/                 # Configuration and errors.yaml
├── utils/                  # Utility functions
├── middleware/             # HTTP middleware
├── di/                     # Dependency injection
└── medierror/             # Custom error handling with registry
```

## Setup

### Prerequisites

- Go 1.23 or higher
- MySQL 8.0 or higher
- Redis 6.0 or higher
- SQLC CLI: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd leal-hospital/server
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Setup database**
   ```bash
   mysql -u root -p
   CREATE DATABASE lael_hospital;
   USE lael_hospital;
   source models/dbConf/schema.sql;
   ```

4. **Configure environment**
   ```bash
   cp config/.env config/.env.local
   # Edit config/.env.local with your configuration
   ```

5. **Generate SQLC code**
   ```bash
   cd models/dbConf
   sqlc generate
   ```

### Running the Application

**Development mode:**
```bash
go run cmd/main.go
```

**Build and run:**
```bash
go build -o bin/server cmd/main.go
./bin/server
```

**With custom config:**
```bash
./bin/server --config=/path/to/.env
```

## Error Handling

This project uses a YAML-based error registry for centralized error management.

### Adding New Errors

Edit `config/errors.yaml`:
```yaml
errors:
  YOUR_SERVICE:
    "ERR001":
      type: "BUSINESS_ERROR"
      message: "Internal technical message"
      display_message: "User-friendly message"
      decline_type: "BD"
      source: "YOUR_SERVICE"
```

### Using Errors in Code

```go
// Get error from registry
registry := medierror.GetErrorRegistry(c)
return registry.GetError("ERR001", "YOUR_SERVICE")

// Or use predefined errors
return medierror.ErrUserNotFound()
```

## Development

### Project Structure

- **cmd/**: Application entry point
- **app/**: Bootstrap and module registration
- **handler/**: HTTP handlers (presentation layer)
- **services/**: Business logic (service layer)
- **persistence/**: Database operations (data layer)
- **domain/**: Core domain models
- **models/**: SQLC schema and queries
- **config/**: Configuration files (config.go, errors.yaml, .env)
- **medierror/**: Error registry and handling

### Adding a New Feature

1. Define domain models in `domain/`
2. Create SQL queries in `models/dbConf/`
3. Generate SQLC code: `cd models/dbConf && sqlc generate`
4. Implement repository in `persistence/`
5. Implement service in `services/`
6. Implement handler in `handler/`
7. Create module and register routes
8. Add errors to `config/errors.yaml`

### Database Migrations

SQL schema is in `models/dbConf/schema.sql`. Run migrations:
```bash
mysql -u root -p lael_hospital < models/dbConf/schema.sql
```

## API Documentation

API endpoints will be documented here as modules are implemented.

### Health Check

```
GET /health
Response: {"status": "healthy", "time": "2026-01-23T..."}
```

## Environment Variables

See `config/.env` for all available configuration options.

### Required Variables

- `DATABASE_USERNAME`: Database username
- `DATABASE_PASSWORD`: Database password
- `DATABASE_NAME`: Database name
- `JWT_SECRET`: Secret key for JWT tokens

## Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./services/...
```

## License

Proprietary - Lael Hospital
```

**Step 2: Verify README**

Run: `cat /Users/abhay.sahani/Documents/MyRepos/leal-hospital/server/readme.md | head -50`
Expected: See updated README content

---

## Summary

This plan sets up the complete backend infrastructure for Lael Hospital management system following clean architecture principles. The implementation includes:

**Core Infrastructure:**
- Go module initialization with all dependencies
- Configuration management (config.go, errors.yaml, .env)
- SQLC-based type-safe database layer (without IF NOT EXISTS)
- YAML-based error registry system (medierror package)
- Database connection manager with pooling
- Dependency injection container
- Structured logging with rotation

**HTTP Layer:**
- Gin-based HTTP framework
- CORS, recovery, and logging middleware
- Module-based route registration
- Graceful shutdown handling

**Next Steps:**
After this setup is complete, you can proceed with:
1. Implementing authentication module (OTP, sessions)
2. Implementing user management (admin/staff)
3. Implementing patient management
4. Implementing OPD generation
5. Adding Redis caching layer
6. Writing tests for each layer

---
