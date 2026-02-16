# KB-Pay Architecture Documentation

> A comprehensive guide to replicate the kb-pay architecture in new projects

## Table of Contents
1. [Overview](#overview)
2. [Project Structure](#project-structure)
3. [Layered Architecture](#layered-architecture)
4. [HTTP Framework & Routing](#http-framework--routing)
5. [Middleware Architecture](#middleware-architecture)
6. [Dependency Injection](#dependency-injection)
7. [Service Layer Patterns](#service-layer-patterns)
8. [Database Architecture](#database-architecture)
9. [Error Management System](#error-management-system)
10. [Configuration Management](#configuration-management)
11. [External Integrations](#external-integrations)
12. [Setup Guide](#setup-guide)

---

## Overview

kb-pay follows a **clean, layered architecture** with clear separation of concerns, modular design, and dependency injection. The architecture is cloud-native, supports containerization, and is designed for scalability.

### Core Architectural Principles
- **Modular Design**: Feature-based organization across layers
- **Clean Architecture**: Clear separation between layers
- **Dependency Injection**: Loose coupling and testability
- **Interface-Driven Design**: Abstractions for flexibility
- **Cloud-Native**: AWS integration and containerization support

### Technology Stack
- **Language**: Go
- **HTTP Framework**: Gin
- **Database**: MySQL with SQLC for type-safe queries
- **Caching**: Redis
- **Configuration**: Viper
- **Logging**: Zerolog with Lumberjack rotation
- **Tracing**: OpenTelemetry
- **Cloud**: AWS (Lambda, Secrets Manager, S3)
- **Containerization**: Docker, Kubernetes

---

## Project Structure

### High-Level Directory Organization

```
project-root/
├── cmd/                    # Application entry points
│   └── main.go
├── app/                    # Application bootstrap and initialization
├── handler/                # HTTP request handlers (Presentation Layer)
├── services/               # Business logic (Service Layer)
├── persistence/            # Data access layer (Repository Pattern)
├── domain/                 # Core domain models and interfaces
├── models/                 # Database models and SQLC configuration
│   ├── db/                # Generated SQLC code
│   ├── dbConf/            # SQL schema and queries
│   └── ypDbConf/          # Secondary database configuration
├── utils/                  # Utility functions and helpers
├── config/                 # Configuration management
├── middleware/             # HTTP middleware components
├── gateway/                # External service integrations
├── cloud/                  # Cloud-specific utilities (AWS)
├── di/                     # Dependency injection container
├── payerror/              # Custom error management
├── telemetry/             # Monitoring and logging
├── deployments/           # Deployment configurations
│   ├── docker/           # Docker configurations
│   ├── k8s/              # Kubernetes manifests
│   └── terraform/        # Infrastructure as code
└── vendor/                # Go module dependencies
```

### Feature-Based Modular Organization

Each major feature follows this pattern across layers:

```
handler/feature_name/      # HTTP handlers for the feature
services/feature_name/     # Business logic for the feature
persistence/feature_name/  # Data access for the feature
domain/feature_name/       # Domain models for the feature
```

**Example**: Payment feature
```
handler/payment/
services/payment/
persistence/payment/
domain/payment/
```

### Naming Conventions

- **Directories**: Lowercase, singular names (e.g., `handler`, `service`)
- **Files**: Snake_case (e.g., `payment_handler.go`, `payment_service.go`)
- **Consistency**: Same naming pattern across layers for related features

---

## Layered Architecture

### 1. Presentation Layer (`/handler`)
**Responsibility**: Handle HTTP requests and responses

- Receive and validate HTTP requests
- Coordinate service calls
- Format responses
- Handle routing registration

**Example Handler Structure**:
```go
type PaymentHandler struct {
    paymentService services.PaymentService
}

func (h *PaymentHandler) HandlePayment(c *gin.Context) {
    // 1. Parse and validate request
    // 2. Call service layer
    // 3. Format and return response
}
```

### 2. Service Layer (`/services`)
**Responsibility**: Implement business logic

- Core business rules and workflows
- Orchestrate multiple repositories
- Handle business validations
- Coordinate external service calls

**Example Service Structure**:
```go
type PaymentService interface {
    ProcessPayment(ctx context.Context, req PaymentRequest) (*PaymentResponse, error)
}

type paymentService struct {
    repo        persistence.PaymentRepository
    gateway     gateway.PaymentGateway
    redis       redis.RedisClientInterface
}
```

### 3. Data Access Layer (`/persistence`)
**Responsibility**: Manage database operations

- Execute database queries using SQLC
- Handle transactions
- Map database models to domain models
- Implement repository pattern

**Example Repository Structure**:
```go
type PaymentRepository interface {
    Create(ctx context.Context, payment *domain.Payment) error
    GetByID(ctx context.Context, id int64) (*domain.Payment, error)
}

type paymentRepository struct {
    db *db.Queries
}
```

### 4. Domain Layer (`/domain`)
**Responsibility**: Define core domain models and interfaces

- Business entities
- Domain interfaces
- Business rules and validations
- Independent of implementation details

**Example Domain Model**:
```go
type Payment struct {
    ID          int64
    Amount      decimal.Decimal
    Status      PaymentStatus
    CreatedAt   time.Time
}
```

---

## HTTP Framework & Routing

### Framework: Gin

Gin is used for its lightweight, high-performance characteristics and excellent middleware support.

### Module-Based Routing Pattern

**Module Interface**:
```go
type Module interface {
    Name() string
    Configure(container *di.Container)
    RegisterRoutes(router *gin.Engine, container *di.Container)
}
```

**Module Implementation Example**:
```go
type PaymentModule struct{}

func (m *PaymentModule) Name() string {
    return "payment"
}

func (m *PaymentModule) Configure(container *di.Container) {
    // Register dependencies
    container.Register((*services.PaymentService)(nil), services.NewPaymentService)
}

func (m *PaymentModule) RegisterRoutes(router *gin.Engine, container *di.Container) {
    handler := container.Get((*handlers.PaymentHandler)(nil)).(handlers.PaymentHandler)

    v1 := router.Group("/v1/payments")
    {
        v1.POST("", handler.CreatePayment)
        v1.GET("/:id", handler.GetPayment)
    }
}
```

### Application Bootstrap

**In `app/app.go`**:
```go
type App struct {
    Router    *gin.Engine
    Container *di.Container
    Modules   []Module
}

func (a *App) Bootstrap() {
    // Setup middleware
    a.setupMiddleware()

    // Register all modules
    for _, module := range a.Modules {
        module.Configure(a.Container)
        module.RegisterRoutes(a.Router, a.Container)
    }
}

func (a *App) setupMiddleware() {
    a.Router.Use(cors.New(corsConfig))
    a.Router.Use(otelgin.Middleware(serviceName))
    a.Router.Use(middleware.TracingMiddleware())
    a.Router.Use(middleware.PanicRecoveryMiddleware())
}
```

---

## Middleware Architecture

### Standard Middleware Chain

1. **CORS Middleware**: Handle cross-origin requests
2. **OpenTelemetry Middleware**: Distributed tracing
3. **Custom Tracing Middleware**: Additional trace context
4. **Panic Recovery Middleware**: Graceful error handling

### Authentication Middleware

**Location**: `middleware/authorizationMW.go`

**Authorization Rules**:
```go
const (
    CustomerAuthOnly              = "customerAuth"
    CustomerAuthWithBasicDevice   = "customerAuthDeviceBasic"
    CustomerAuthWithStrictDevice  = "customerAuthDeviceStrict"
    BasicAuthWithIPWhitelist      = "basicAuthIPWhiteList"
)
```

**Implementation Pattern**:
```go
func AuthorizationMiddleware(ruleName string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Extract token from header
        // 2. Validate token
        // 3. Check session expiry
        // 4. Validate device information (if required)
        // 5. Set user context
        // 6. Call next handler
        c.Next()
    }
}
```

### Custom Middleware Structure

```go
func CustomMiddleware(dependencies ...interface{}) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Before request processing

        c.Next()

        // After request processing
    }
}
```

---

## Dependency Injection

### Custom DI Container

**Location**: `di/dependencyInjection.go`

**Features**:
- Thread-safe
- Circular dependency detection
- Factory-based registration
- Lazy initialization

### Container Interface

```go
type Container struct {
    services map[reflect.Type]interface{}
    mu       sync.RWMutex
}

func NewContainer() *Container {
    return &Container{
        services: make(map[reflect.Type]interface{}),
    }
}
```

### Registration Patterns

**1. Direct Registration**:
```go
// Register concrete implementation
container.Register((*ServiceInterface)(nil), concreteImplementation)
```

**2. Factory Registration**:
```go
// Register factory for lazy initialization
container.RegisterFactory((*ServiceInterface)(nil), func(c *Container) interface{} {
    // Resolve dependencies
    repo := c.Get((*RepositoryInterface)(nil)).(RepositoryInterface)

    // Create and return service
    return NewService(repo)
})
```

**3. Retrieval**:
```go
service := container.Get((*ServiceInterface)(nil)).(ServiceInterface)
```

### Service Construction Pattern

**Constructor-Based Injection**:
```go
type PaymentService struct {
    repo    persistence.PaymentRepository
    gateway gateway.PaymentGateway
    redis   redis.RedisClientInterface
    logger  logger.Logger
}

func NewPaymentService(
    repo persistence.PaymentRepository,
    gateway gateway.PaymentGateway,
    redis redis.RedisClientInterface,
    logger logger.Logger,
) PaymentService {
    return &PaymentService{
        repo:    repo,
        gateway: gateway,
        redis:   redis,
        logger:  logger,
    }
}
```

---

## Service Layer Patterns

### Interface-Driven Design

**Define Interfaces**:
```go
type PaymentService interface {
    ProcessPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error)
    GetPaymentStatus(ctx context.Context, paymentID string) (*PaymentStatus, error)
}
```

**Implement Services**:
```go
type paymentService struct {
    repo          persistence.PaymentRepository
    gateway       gateway.PaymentGateway
    redis         redis.RedisClientInterface
    notification  services.NotificationService
}

func (s *paymentService) ProcessPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
    // 1. Validate request
    // 2. Check business rules
    // 3. Call gateway
    // 4. Save to database
    // 5. Send notification
    // 6. Return response
}
```

### Rule-Based Logic Pattern

For complex authorization or validation logic:

```go
type AuthenticationDetails struct {
    Ctx           context.Context
    UPIDeviceInfo string
    BearerToken   string
    ClientIP      string
    Svc           *AuthenticationSvc
}

var authRules = map[string]func(*AuthenticationDetails) domain.AuthenticationResponse{
    "customerAuth":            (*AuthenticationDetails).ValidateCustomerAuth,
    "customerAuthDeviceBasic": (*AuthenticationDetails).ValidateCustomerAuthDeviceBasic,
    "customerAuthDeviceStrict": (*AuthenticationDetails).ValidateCustomerAuthDeviceStrict,
}

func (s *AuthenticationSvc) Authorize(ctx context.Context, ruleName string, params ...string) (*domain.AuthenticationResponse, error) {
    details := &AuthenticationDetails{
        Ctx: ctx,
        Svc: s,
        // ... populate other fields
    }

    if ruleFunc, exists := authRules[ruleName]; exists {
        response := ruleFunc(details)
        return &response, nil
    }

    return nil, errors.New("rule not found")
}
```

### Service Interaction with Repositories

```go
func (s *paymentService) ProcessPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
    // Begin transaction
    txn, err := s.repo.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    defer txn.Rollback(ctx)

    // Create payment record
    payment := &domain.Payment{
        Amount: req.Amount,
        Status: domain.StatusPending,
    }

    err = s.repo.Create(ctx, txn, payment)
    if err != nil {
        return nil, err
    }

    // Call external gateway
    gatewayResp, err := s.gateway.ProcessPayment(ctx, payment)
    if err != nil {
        return nil, err
    }

    // Update payment status
    payment.Status = domain.StatusCompleted
    err = s.repo.Update(ctx, txn, payment)
    if err != nil {
        return nil, err
    }

    // Commit transaction
    err = txn.Commit(ctx)
    if err != nil {
        return nil, err
    }

    return &PaymentResponse{
        PaymentID: payment.ID,
        Status:    payment.Status,
    }, nil
}
```

### Caching Pattern with Redis

```go
func (s *paymentService) GetPaymentStatus(ctx context.Context, paymentID string) (*PaymentStatus, error) {
    // Try cache first
    cacheKey := fmt.Sprintf("payment:%s", paymentID)
    cached, err := s.redis.Get(ctx, cacheKey)
    if err == nil && cached != "" {
        var status PaymentStatus
        json.Unmarshal([]byte(cached), &status)
        return &status, nil
    }

    // Cache miss - fetch from database
    payment, err := s.repo.GetByID(ctx, paymentID)
    if err != nil {
        return nil, err
    }

    status := &PaymentStatus{
        Status:    payment.Status,
        UpdatedAt: payment.UpdatedAt,
    }

    // Update cache
    data, _ := json.Marshal(status)
    s.redis.Set(ctx, cacheKey, string(data), 5*time.Minute)

    return status, nil
}
```

---

## Database Architecture

### SQLC Configuration

**Location**: `models/dbConf/sqlc.yaml`

```yaml
version: "2"
sql:
  - engine: "mysql"
    queries:
      - "upi_banks.sql"
      - "upi_collect_requests.sql"
      - "upi_transactions.sql"
      # Add all your query files
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
        overrides:
          - db_type: "datetime"
            go_type: "time.Time"
            nullable: false
          - db_type: "decimal"
            go_type: "github.com/shopspring/decimal.Decimal"
```

### SQL Organization

```
models/
├── dbConf/
│   ├── sqlc.yaml              # SQLC configuration
│   ├── schema.sql             # Database schema
│   ├── upi_banks.sql          # Domain-specific queries
│   ├── upi_transactions.sql
│   ├── users.sql
│   └── ... (other query files)
├── db/                        # Generated SQLC code
│   ├── db.go
│   ├── models.go
│   ├── querier.go
│   └── upi_banks.sql.go
└── ypDbConf/                  # Secondary database (if needed)
```

### Query File Example

**`models/dbConf/upi_transactions.sql`**:
```sql
-- name: CreateTransaction :execresult
INSERT INTO upi_transactions (
    user_id,
    amount,
    status,
    created_at
) VALUES (?, ?, ?, ?);

-- name: GetTransactionByID :one
SELECT * FROM upi_transactions
WHERE id = ?;

-- name: ListTransactionsByUserID :many
SELECT * FROM upi_transactions
WHERE user_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateTransactionStatus :exec
UPDATE upi_transactions
SET status = ?, updated_at = ?
WHERE id = ?;
```

### Database Manager

**Location**: `utils/db/dbManager.go`

**Connection Pool Configuration**:
```go
const (
    defaultWriteMaxOpenConns = 25
    defaultWriteMaxIdleConns = 10
    defaultReadMaxOpenConns  = 50
    defaultReadMaxIdleConns  = 15
    defaultConnMaxLifetime   = 30 * time.Minute
)
```

**DBManager Interface**:
```go
type DBManagerInterface interface {
    Initialize() error
    GetWriteDBConn(ctx context.Context) (DBConnInterface, error)
    GetReadDBConn(ctx context.Context) (DBConnInterface, error)
    Close() error
}

type DBConnInterface interface {
    GetConn(ctx context.Context) *sql.Conn
    Close(ctx context.Context) error
    BeginTx(ctx context.Context, opts *sql.TxOptions) (TxnInterface, error)
}

type TxnInterface interface {
    Commit(ctx context.Context) error
    Rollback(ctx context.Context) error
}
```

**Singleton Implementation**:
```go
var (
    instance DBManagerInterface
    once     sync.Once
)

func GetDBManager() DBManagerInterface {
    once.Do(func() {
        instance = &DBManager{}
    })
    return instance
}
```

**Initialization**:
```go
func (dm *DBManager) Initialize() error {
    // Read configuration
    writeMaxOpen := viper.GetInt("DB_WRITE_MAX_OPEN_CONNS")
    if writeMaxOpen == 0 {
        writeMaxOpen = defaultWriteMaxOpenConns
    }

    readMaxOpen := viper.GetInt("DB_READ_MAX_OPEN_CONNS")
    if readMaxOpen == 0 {
        readMaxOpen = defaultReadMaxOpenConns
    }

    // Establish write connection
    writeDB, err := sql.Open("mysql", writeDSN)
    if err != nil {
        return fmt.Errorf("failed to open write database: %w", err)
    }

    writeDB.SetMaxOpenConns(writeMaxOpen)
    writeDB.SetMaxIdleConns(defaultWriteMaxIdleConns)
    writeDB.SetConnMaxLifetime(defaultConnMaxLifetime)

    // Establish read connection (replica)
    readDB, err := sql.Open("mysql", readDSN)
    if err != nil {
        log.Warn("Failed to open read database, using write connection for reads")
        readDB = writeDB
    } else {
        readDB.SetMaxOpenConns(readMaxOpen)
        readDB.SetMaxIdleConns(defaultReadMaxIdleConns)
        readDB.SetConnMaxLifetime(defaultConnMaxLifetime)
    }

    dm.writeDB = writeDB
    dm.readDB = readDB

    return nil
}
```

### Transaction Management Pattern

```go
func (r *paymentRepository) CreatePayment(ctx context.Context, payment *domain.Payment) error {
    // Get connection
    conn, err := r.dbManager.GetWriteDBConn(ctx)
    if err != nil {
        return err
    }
    defer conn.Close(ctx)

    // Begin transaction
    txn, err := conn.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer txn.Rollback(ctx)

    // Execute queries
    queries := db.New(conn.GetConn(ctx))
    result, err := queries.WithTx(txn).CreatePayment(ctx, db.CreatePaymentParams{
        Amount: payment.Amount,
        Status: payment.Status,
    })
    if err != nil {
        return err
    }

    // Commit transaction
    return txn.Commit(ctx)
}
```

### Repository Pattern Implementation

```go
type PaymentRepository interface {
    Create(ctx context.Context, payment *domain.Payment) error
    GetByID(ctx context.Context, id int64) (*domain.Payment, error)
    Update(ctx context.Context, payment *domain.Payment) error
    List(ctx context.Context, limit, offset int) ([]*domain.Payment, error)
}

type paymentRepository struct {
    dbManager db.DBManagerInterface
}

func NewPaymentRepository(dbManager db.DBManagerInterface) PaymentRepository {
    return &paymentRepository{
        dbManager: dbManager,
    }
}

func (r *paymentRepository) GetByID(ctx context.Context, id int64) (*domain.Payment, error) {
    // Use read connection for read operations
    conn, err := r.dbManager.GetReadDBConn(ctx)
    if err != nil {
        return nil, err
    }
    defer conn.Close(ctx)

    queries := db.New(conn.GetConn(ctx))
    dbPayment, err := queries.GetPaymentByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // Map database model to domain model
    return &domain.Payment{
        ID:        dbPayment.ID,
        Amount:    dbPayment.Amount,
        Status:    domain.PaymentStatus(dbPayment.Status),
        CreatedAt: dbPayment.CreatedAt,
    }, nil
}
```

---

## Error Management System

### Error Structure

**Location**: `payerror/error.go`

**Core Error Type**:
```go
type AppError struct {
    code           ErrorCode       // Numeric error code
    errorType      ErrorType       // Error category
    message        string          // Technical error message
    displayMessage string          // User-friendly message
    details        []ErrorDetail   // Additional context
    stack          []string        // Stack trace
    cause          error           // Original error
}

type ErrorCode int

const (
    ErrValidation     ErrorCode = 400
    ErrInternal       ErrorCode = 500
    ErrGateway        ErrorCode = 453
    ErrDatabase       ErrorCode = 503
    ErrBusiness       ErrorCode = 600
    ErrNotFound       ErrorCode = 404
)

type ErrorType string

const (
    ValidationError ErrorType = "BAD_REQUEST"
    SystemError     ErrorType = "SYSTEM_ERROR"
    GatewayErr      ErrorType = "GATEWAY_ERROR"
    BusinessError   ErrorType = "BUSINESS_ERROR"
    DatabaseError   ErrorType = "DATABASE_ERROR"
)
```

### Error Creation Functions

```go
// Create a new error
func Newf(errType ErrorType, code ErrorCode, format string, args ...interface{}) error {
    return &AppError{
        errorType: errType,
        code:      code,
        message:   fmt.Sprintf(format, args...),
        stack:     captureStack(),
    }
}

// Wrap an existing error
func Wrapf(err error, errType ErrorType, code ErrorCode, format string, args ...interface{}) error {
    return &AppError{
        errorType: errType,
        code:      code,
        message:   fmt.Sprintf(format, args...),
        cause:     err,
        stack:     captureStack(),
    }
}

// Add display message for users
func (e *AppError) WithDisplayMessage(msg string) *AppError {
    e.displayMessage = msg
    return e
}

// Add structured error details
func (e *AppError) WithDetails(details ...ErrorDetail) *AppError {
    e.details = append(e.details, details...)
    return e
}
```

### Error Response Format

```go
type ErrorResponse struct {
    Code  string      `json:"code"`
    Msg   string      `json:"msg"`
    Model *ErrorModel `json:"model"`
}

type ErrorModel struct {
    ErrorCode      string        `json:"errorCode"`
    Message        string        `json:"message"`
    DisplayMessage string        `json:"displayMessage"`
    Details        []ErrorDetail `json:"details,omitempty"`
}

type ErrorDetail struct {
    Field   string `json:"field,omitempty"`
    Message string `json:"message"`
}
```

### Error Registry Pattern

**Location**: `handler/*/dto/error_registry.go`

```go
var errorRegistry = map[string]map[string]string{
    "fetchAccount": {
        "INVALID_USER_ID":     "User ID is invalid or missing",
        "ACCOUNT_NOT_FOUND":   "Account not found for the given user",
        "DATABASE_ERROR":      "Unable to fetch account details",
    },
    "generateOtp": {
        "INVALID_MOBILE":      "Mobile number is invalid",
        "OTP_GENERATION_FAILED": "Failed to generate OTP",
        "MAX_ATTEMPTS_EXCEEDED": "Maximum OTP attempts exceeded",
    },
}

func GetErrorMessage(module, errorCode string) string {
    if moduleErrors, exists := errorRegistry[module]; exists {
        if msg, found := moduleErrors[errorCode]; found {
            return msg
        }
    }
    return "An unexpected error occurred"
}
```

### Error Handling in Handlers

```go
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
    var req PaymentRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        appErr := payerror.Newf(
            payerror.ValidationError,
            payerror.ErrValidation,
            "Invalid request payload",
        ).WithDisplayMessage("Please check your input and try again")

        c.JSON(http.StatusBadRequest, appErr.ToResponse())
        return
    }

    payment, err := h.service.ProcessPayment(c.Request.Context(), &req)
    if err != nil {
        // Check if it's already an AppError
        if appErr, ok := err.(*payerror.AppError); ok {
            c.JSON(appErr.HTTPStatus(), appErr.ToResponse())
            return
        }

        // Wrap unknown errors
        appErr := payerror.Wrapf(
            err,
            payerror.SystemError,
            payerror.ErrInternal,
            "Failed to process payment",
        ).WithDisplayMessage("We're experiencing technical difficulties. Please try again later.")

        c.JSON(http.StatusInternalServerError, appErr.ToResponse())
        return
    }

    c.JSON(http.StatusOK, payment)
}
```

### Error Handling in Services

```go
func (s *paymentService) ProcessPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
    // Validation
    if req.Amount <= 0 {
        return nil, payerror.Newf(
            payerror.ValidationError,
            payerror.ErrValidation,
            "Amount must be greater than zero",
        ).WithDisplayMessage("Please enter a valid amount")
    }

    // Database operation
    payment, err := s.repo.Create(ctx, req)
    if err != nil {
        return nil, payerror.Wrapf(
            err,
            payerror.DatabaseError,
            payerror.ErrDatabase,
            "Failed to create payment record",
        )
    }

    // External gateway call
    gatewayResp, err := s.gateway.ProcessPayment(ctx, payment)
    if err != nil {
        return nil, payerror.Wrapf(
            err,
            payerror.GatewayErr,
            payerror.ErrGateway,
            "Payment gateway failed",
        ).WithDisplayMessage("Payment processing failed. Please try again.")
    }

    return &PaymentResponse{
        PaymentID: payment.ID,
        Status:    gatewayResp.Status,
    }, nil
}
```

### Global Error Handling Middleware

```go
func ErrorHandlingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        // Check if there are any errors
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err

            if appErr, ok := err.(*payerror.AppError); ok {
                c.JSON(appErr.HTTPStatus(), appErr.ToResponse())
            } else {
                // Handle unexpected errors
                appErr := payerror.Wrapf(
                    err,
                    payerror.SystemError,
                    payerror.ErrInternal,
                    "Unexpected error occurred",
                )
                c.JSON(http.StatusInternalServerError, appErr.ToResponse())
            }

            return
        }
    }
}
```

---

## Configuration Management

### Configuration Loading with Viper

**Location**: `config/config.go`

**Multi-Source Configuration Priority**:
1. Command-line flags
2. AWS Secrets Manager (for Lambda)
3. Environment variables
4. Configuration files
5. Default values

### Configuration Structure

```go
type AppConfig struct {
    Environment string
    Server      ServerConfig
    Database    DBConfig
    YPDatabase  YPDBConfig
    Redis       RedisConfig
    AWS         AWSConfig
    Logging     LoggingConfig
}

type ServerConfig struct {
    Port         int
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
}

type DBConfig struct {
    Host            string
    Port            int
    Username        string
    Password        string
    Database        string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
}

type RedisConfig struct {
    Host     string
    Port     int
    Password string
    DB       int
}

type AWSConfig struct {
    Region          string
    SecretName      string
    S3Bucket        string
}

type LoggingConfig struct {
    Level      string
    FilePath   string
    MaxSize    int
    MaxBackups int
    MaxAge     int
}
```

### Configuration Loading Function

```go
func LoadConfig() (*AppConfig, error) {
    // Set defaults
    viper.SetDefault("APP_ENV", "production")
    viper.SetDefault("SERVER_PORT", 8080)
    viper.SetDefault("DB_MAX_OPEN_CONNS", 25)

    // Read from environment
    viper.AutomaticEnv()

    // Read config file if specified
    configFile := flag.String("config", "", "Path to config file")
    flag.Parse()

    if *configFile != "" {
        viper.SetConfigFile(*configFile)
        if err := viper.ReadInConfig(); err != nil {
            return nil, fmt.Errorf("failed to read config file: %w", err)
        }
    }

    // Load secrets from AWS Secrets Manager for Lambda
    if isLambdaEnvironment() {
        if err := loadSecretsFromAWS(); err != nil {
            return nil, fmt.Errorf("failed to load AWS secrets: %w", err)
        }
    }

    // Unmarshal config
    var config AppConfig
    if err := viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }

    return &config, nil
}
```

### AWS Secrets Manager Integration

```go
func loadSecretsFromAWS() error {
    secretName := viper.GetString("AWS_SECRET_NAME")
    region := viper.GetString("AWS_REGION")

    sess, err := session.NewSession(&aws.Config{
        Region: aws.String(region),
    })
    if err != nil {
        return err
    }

    svc := secretsmanager.New(sess)
    input := &secretsmanager.GetSecretValueInput{
        SecretId: aws.String(secretName),
    }

    result, err := svc.GetSecretValue(input)
    if err != nil {
        return err
    }

    var secrets map[string]string
    if err := json.Unmarshal([]byte(*result.SecretString), &secrets); err != nil {
        return err
    }

    // Set secrets as environment variables or Viper values
    for key, value := range secrets {
        viper.Set(key, value)
    }

    return nil
}
```

### Logging Configuration

```go
func SetupLogging(config LoggingConfig) {
    // Parse log level
    level, err := zerolog.ParseLevel(config.Level)
    if err != nil {
        level = zerolog.InfoLevel
    }
    zerolog.SetGlobalLevel(level)

    // Configure log rotation
    fileWriter := &lumberjack.Logger{
        Filename:   config.FilePath,
        MaxSize:    config.MaxSize,    // megabytes
        MaxBackups: config.MaxBackups,
        MaxAge:     config.MaxAge,     // days
        Compress:   true,
    }

    // Multi-writer for console and file
    multi := zerolog.MultiLevelWriter(
        zerolog.ConsoleWriter{Out: os.Stdout},
        fileWriter,
    )

    log.Logger = zerolog.New(multi).With().
        Timestamp().
        Caller().
        Logger()
}
```

### Environment-Specific Configuration

**`.env.development`**:
```env
APP_ENV=development
SERVER_PORT=8080
DB_HOST=localhost
DB_PORT=3306
DB_USERNAME=devuser
DB_PASSWORD=devpass
DB_DATABASE=kbpay_dev
REDIS_HOST=localhost
REDIS_PORT=6379
LOG_LEVEL=debug
```

**`.env.production`**:
```env
APP_ENV=production
SERVER_PORT=8080
AWS_REGION=ap-south-1
AWS_SECRET_NAME=kbpay-prod-secrets
LOG_LEVEL=info
```

---

## External Integrations

### Gateway Pattern

**Location**: `gateway/`

**Gateway Interface**:
```go
type PaymentGateway interface {
    ProcessPayment(ctx context.Context, payment *domain.Payment) (*GatewayResponse, error)
    CheckStatus(ctx context.Context, transactionID string) (*StatusResponse, error)
    RefundPayment(ctx context.Context, transactionID string, amount decimal.Decimal) error
}
```

**Gateway Implementation**:
```go
type paymentGateway struct {
    httpClient *http.Client
    baseURL    string
    apiKey     string
    logger     logger.Logger
}

func NewPaymentGateway(baseURL, apiKey string, logger logger.Logger) PaymentGateway {
    return &paymentGateway{
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
        baseURL: baseURL,
        apiKey:  apiKey,
        logger:  logger,
    }
}

func (g *paymentGateway) ProcessPayment(ctx context.Context, payment *domain.Payment) (*GatewayResponse, error) {
    // Prepare request
    reqBody := map[string]interface{}{
        "amount":   payment.Amount,
        "currency": "INR",
        "metadata": payment.Metadata,
    }

    jsonBody, _ := json.Marshal(reqBody)

    req, err := http.NewRequestWithContext(ctx, "POST", g.baseURL+"/process", bytes.NewBuffer(jsonBody))
    if err != nil {
        return nil, payerror.Wrapf(err, payerror.GatewayErr, payerror.ErrGateway, "Failed to create request")
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+g.apiKey)

    // Execute request
    resp, err := g.httpClient.Do(req)
    if err != nil {
        return nil, payerror.Wrapf(err, payerror.GatewayErr, payerror.ErrGateway, "Gateway request failed")
    }
    defer resp.Body.Close()

    // Parse response
    var gatewayResp GatewayResponse
    if err := json.NewDecoder(resp.Body).Decode(&gatewayResp); err != nil {
        return nil, payerror.Wrapf(err, payerror.GatewayErr, payerror.ErrGateway, "Failed to parse gateway response")
    }

    if resp.StatusCode != http.StatusOK {
        return nil, payerror.Newf(
            payerror.GatewayErr,
            payerror.ErrGateway,
            "Gateway returned error: %s", gatewayResp.Error,
        )
    }

    return &gatewayResp, nil
}
```

### Redis Integration

```go
type RedisClientInterface interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key string, value string, expiration time.Duration) error
    Delete(ctx context.Context, keys ...string) error
    Exists(ctx context.Context, keys ...string) (int64, error)
}

type redisClient struct {
    client *redis.Client
}

func NewRedisClient(config RedisConfig) RedisClientInterface {
    client := redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
        Password: config.Password,
        DB:       config.DB,
    })

    return &redisClient{client: client}
}
```

### AWS S3 Integration

```go
type S3Client interface {
    UploadFile(ctx context.Context, key string, data []byte) error
    DownloadFile(ctx context.Context, key string) ([]byte, error)
    DeleteFile(ctx context.Context, key string) error
}

type s3Client struct {
    client *s3.S3
    bucket string
}

func NewS3Client(bucket, region string) S3Client {
    sess := session.Must(session.NewSession(&aws.Config{
        Region: aws.String(region),
    }))

    return &s3Client{
        client: s3.New(sess),
        bucket: bucket,
    }
}
```

---

## Setup Guide

### Prerequisites

- Go 1.21+
- MySQL 8.0+
- Redis 6.0+
- Docker (optional, for containerization)
- SQLC CLI (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)

### Project Initialization

**1. Create Project Structure**:
```bash
mkdir -p my-project/{cmd,app,handler,services,persistence,domain,models/{db,dbConf},utils,config,middleware,gateway,di,payerror}
cd my-project
go mod init github.com/yourusername/my-project
```

**2. Install Dependencies**:
```bash
# HTTP framework
go get github.com/gin-gonic/gin

# Configuration
go get github.com/spf13/viper

# Database
go get github.com/go-sql-driver/mysql
go get github.com/sqlc-dev/sqlc/cmd/sqlc

# Redis
go get github.com/redis/go-redis/v9

# Logging
go get github.com/rs/zerolog
go get gopkg.in/natefinch/lumberjack.v2

# AWS SDK
go get github.com/aws/aws-sdk-go

# Tracing
go get go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin

# Utilities
go get github.com/spf13/cast
go get github.com/shopspring/decimal
```

**3. Setup Database**:

Create `models/dbConf/schema.sql`:
```sql
CREATE TABLE IF NOT EXISTS users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

Create `models/dbConf/users.sql`:
```sql
-- name: CreateUser :execresult
INSERT INTO users (email) VALUES (?);

-- name: GetUserByID :one
SELECT * FROM users WHERE id = ?;
```

Create `models/dbConf/sqlc.yaml`:
```yaml
version: "2"
sql:
  - engine: "mysql"
    queries:
      - "users.sql"
    schema: "schema.sql"
    gen:
      go:
        package: "db"
        out: "../db"
        emit_json_tags: true
        emit_interface: true
```

Generate code:
```bash
cd models/dbConf
sqlc generate
```

**4. Create Configuration Files**:

`.env.development`:
```env
APP_ENV=development
SERVER_PORT=8080
DB_HOST=localhost
DB_PORT=3306
DB_USERNAME=root
DB_PASSWORD=password
DB_DATABASE=myapp_dev
REDIS_HOST=localhost
REDIS_PORT=6379
LOG_LEVEL=debug
```

**5. Implement Core Components**:

Follow the patterns documented in this guide:
- Dependency Injection container (`di/`)
- Database manager (`utils/db/`)
- Error management (`payerror/`)
- Configuration loading (`config/`)
- Logging setup (`config/`)

**6. Create Application Bootstrap**:

`cmd/main.go`:
```go
package main

import (
    "log"
    "my-project/app"
    "my-project/config"
)

func main() {
    // Load configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Setup logging
    config.SetupLogging(cfg.Logging)

    // Initialize and start application
    application := app.NewApp(cfg)
    if err := application.Bootstrap(); err != nil {
        log.Fatalf("Failed to bootstrap application: %v", err)
    }

    if err := application.Run(); err != nil {
        log.Fatalf("Failed to run application: %v", err)
    }
}
```

**7. Run Application**:
```bash
go run cmd/main.go
```

### Development Workflow

1. **Define Domain Models** in `domain/`
2. **Create SQL Queries** in `models/dbConf/`
3. **Generate SQLC Code**: `cd models/dbConf && sqlc generate`
4. **Implement Repository** in `persistence/`
5. **Implement Service** in `services/`
6. **Implement Handler** in `handler/`
7. **Register Routes** in module's `RegisterRoutes`
8. **Test** each layer

### Deployment

**Docker**:
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```

**Kubernetes**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - name: my-app
        image: my-app:latest
        ports:
        - containerPort: 8080
        env:
        - name: APP_ENV
          value: "production"
```

---

## Best Practices

### Code Organization
- Keep layers independent
- Use interfaces for abstractions
- Follow naming conventions consistently
- Group related functionality

### Error Handling
- Always use custom error types
- Provide user-friendly error messages
- Log errors with context
- Don't expose internal errors to users

### Database
- Use transactions for multi-step operations
- Use read replicas for read-heavy operations
- Implement connection pooling
- Use SQLC for type-safe queries

### Performance
- Implement caching strategically
- Use connection pooling
- Optimize database queries
- Use context for request cancellation

### Security
- Validate all inputs
- Use environment variables for secrets
- Implement proper authentication/authorization
- Use HTTPS in production
- Sanitize error messages

### Testing
- Write unit tests for services
- Write integration tests for repositories
- Mock external dependencies
- Use test containers for database tests

---

## Summary

This architecture provides:
- **Scalability**: Modular design, connection pooling, caching
- **Maintainability**: Clear separation of concerns, consistent patterns
- **Testability**: Dependency injection, interfaces
- **Reliability**: Comprehensive error handling, transaction management
- **Performance**: Optimized database access, caching strategies
- **Security**: Input validation, secrets management, error sanitization

Follow these patterns and structures to build robust, scalable Go applications with clean architecture principles.
