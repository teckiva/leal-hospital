# Lael Hospital - Backend Server

Backend server for Lael Hospital internal management system.

## Tech Stack

- **Language:** Go 1.23+
- **HTTP Framework:** Gin
- **Database:** MySQL with SQLC for type-safe queries
- **Configuration:** Viper (loads from .env files)
- **Logging:** Zerolog with Lumberjack rotation
- **Error Handling:** YAML-based error registry (medierror package)
- **Dependency Injection:** Custom DI container

## Architecture

This project follows clean architecture principles with clear separation of concerns:

```
server/
├── cmd/                    # Application entry point (main.go)
├── app/                    # Application bootstrap and module system
├── handler/                # HTTP request handlers (presentation layer)
├── services/               # Business logic (service layer)
├── persistence/            # Data access layer
├── domain/                 # Domain models and interfaces
├── models/                 # SQLC configuration and generated code
│   ├── dbConf/            # SQL schema and query files
│   └── db/                # Generated Go code from SQLC
├── config/                 # Configuration (config.go, errors.yaml, .env)
├── utils/                  # Utility packages
│   ├── db/                # Database connection manager
│   └── logger/            # Logging utilities
├── middleware/             # HTTP middleware (CORS, recovery, logging)
├── di/                     # Dependency injection container
├── medierror/             # Error management with YAML registry
├── logger/                 # Structured logging with rotation
└── gateway/                # External service integrations
```

## Prerequisites

- **Go 1.23** or higher
- **MySQL 8.0** or higher
- **SQLC CLI**: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`

## Setup

### 1. Clone the Repository

```bash
git clone <repository-url>
cd leal-hospital/server
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Configure Environment

Copy the example .env file and update with your settings:

```bash
cp config/.env config/.env.local
```

Edit `config/.env` with your configuration:

```env
# Database Configuration
LAEL_MYSQL_DB_USERNAME=root
LAEL_MYSQL_DB_PASSWORD=your_password
LAEL_MYSQL_DB_HOST=localhost
LAEL_MYSQL_DB_PORT=3306
LAEL_MYSQL_DB_SCHEMA=lael_hospital

# Server Configuration
SERVER_PORT=8080
SERVER_READ_TIMEOUT=15
SERVER_WRITE_TIMEOUT=15

# Logging Configuration
LOG_LEVEL=info
LOG_FILE_PATH=logs/app.log

# Security Configuration
JWT_SECRET=your-secret-key-change-in-production
```

### 4. Setup Database

Create the database and run the schema:

```bash
mysql -u root -p
```

```sql
CREATE DATABASE lael_hospital;
USE lael_hospital;
SOURCE models/dbConf/schema.sql;
```

### 5. Generate SQLC Code (if needed)

If you modify SQL queries, regenerate the Go code:

```bash
cd models/dbConf
sqlc generate
```

## Running the Application

### Development Mode

```bash
go run cmd/main.go
```

### Production Build

```bash
go build -o bin/server cmd/main.go
./bin/server
```

The server will start on the configured port (default: 8080).

## Project Structure

### Core Components

#### Configuration System (`config/`)
- **config.go**: Application configuration with Viper
- **errors.yaml**: YAML-based error registry (numeric codes: 1000-5999)
- **.env**: Environment variables

#### Error Management (`medierror/`)
- YAML-based error registry loaded at startup
- Consistent error responses across all APIs
- Error codes organized by service:
  - 1000-1999: Internal/system errors
  - 2000-2999: Authentication service errors
  - 3000-3999: User service errors
  - 4000-4999: Patient service errors
  - 5000-5999: OPD service errors

#### Database Layer (`models/`, `utils/db/`)
- **SQLC**: Type-safe SQL query generation
- **Schema**: Direct CREATE TABLE statements (no IF NOT EXISTS)
- **Connection Manager**: Pooled connections with read/write separation support
- **Transactions**: Full transaction support via interfaces

#### Logging (`logger/`)
- Structured logging with Zerolog
- Automatic log rotation with Lumberjack
- Console and file output
- Configurable log levels

#### Middleware (`middleware/`)
- **Recovery**: Catches panics and returns proper error responses
- **Logger**: Logs all HTTP requests with timing
- **CORS**: Handles cross-origin requests (dev ports: 3000, 5173)

#### Dependency Injection (`di/`)
- Centralized dependency management
- Factory pattern for lazy initialization
- Circular dependency detection
- Thread-safe with sync.Map

### Database Schema

The application uses 6 main tables:

- **lael_users**: Admin and staff users with approval workflow
- **lael_otp**: OTP management for authentication
- **lael_patients**: Patient registration with OPD ID
- **patient_opd**: OPD records with JSON fields (symptoms, prescriptions, medicines)
- **lael_sessions**: User session management
- **lael_activity_logs**: Audit trail for user activities

## API Endpoints

### Health Check

```
GET /health
```

Response:
```json
{
  "status": "healthy",
  "time": "2026-01-23T..."
}
```

### Future Endpoints

Additional endpoints will be documented as modules are implemented:
- Authentication (OTP-based login)
- User management (admin/staff)
- Patient management
- OPD generation
- Activity logs

## Error Response Format

All errors follow a consistent format:

```json
{
  "code": "1004",
  "msg": "FAILED",
  "model": {
    "errorCode": "SYSTEM_ERROR",
    "message": "Internal server error",
    "displayMessage": "Something went wrong. Please try again later.",
    "details": []
  }
}
```

## Development Workflow

### Adding a New Feature

1. **Define domain models** in `domain/`
2. **Create SQL queries** in `models/dbConf/`
3. **Generate SQLC code**: `cd models/dbConf && sqlc generate`
4. **Implement repository** in `persistence/`
5. **Implement service** in `services/`
6. **Implement handler** in `handler/`
7. **Create module** and register routes in `app/`
8. **Add errors** to `config/errors.yaml`

### Adding a New Module

Create a module struct that implements the `app.Module` interface:

```go
package mymodule

import (
    "github.com/gin-gonic/gin"
    "github.com/leal-hospital/server/app"
    "github.com/leal-hospital/server/di"
)

type MyModule struct{}

func (m *MyModule) Name() string {
    return "mymodule"
}

func (m *MyModule) Configure(container *di.Container) {
    // Register dependencies
    container.RegisterFactory((*MyService)(nil), func(c *di.Container) any {
        // Resolve dependencies and create service
        return NewMyService(...)
    })
}

func (m *MyModule) RegisterRoutes(router *gin.Engine, container *di.Container) {
    // Register HTTP routes
    handler := NewMyHandler(container)
    router.POST("/api/myresource", handler.Create)
    router.GET("/api/myresource/:id", handler.Get)
}
```

Then register the module in `cmd/main.go`:

```go
modules := []app.Module{
    &mymodule.MyModule{},
}
```

## Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./services/...
```

## Logging

Logs are written to both console and file:
- **Console**: Pretty-formatted for development
- **File**: JSON-formatted at `logs/app.log` (configurable)
- **Rotation**: Automatic based on size/age

Log levels: `debug`, `info`, `warn`, `error`, `fatal`

## Graceful Shutdown

The application handles graceful shutdown on SIGINT and SIGTERM:
1. Stops accepting new requests
2. Waits for active requests to complete (5-second timeout)
3. Closes database connections
4. Exits cleanly

## Security Notes

- Change `JWT_SECRET` in production
- Use environment-specific .env files
- Database credentials should never be committed
- Use `.gitignore` to protect sensitive files

## License

Proprietary - Lael Hospital
