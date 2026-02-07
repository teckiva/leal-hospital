# Dependency Injection Architecture

## Overview

This document describes the Dependency Injection (DI) system that provides centralized service management, loose coupling, and testability across the application. The DI container manages the lifecycle of dependencies and automatically resolves complex dependency graphs.

## What is Dependency Injection?

Dependency Injection is a design pattern where objects receive their dependencies from external sources rather than creating them internally. This promotes:

- **Loose Coupling**: Components depend on interfaces, not concrete implementations
- **Testability**: Easy to mock dependencies in unit tests
- **Flexibility**: Swap implementations without changing dependent code
- **Maintainability**: Centralized configuration of dependencies

### Without DI (Tightly Coupled)

```go
type UserService struct {
    db *sql.DB  // Direct dependency on concrete type
}

func NewUserService() *UserService {
    // Creating dependencies inside the service
    db, _ := sql.Open("mysql", "connection_string")
    return &UserService{db: db}
}
```

### With DI (Loosely Coupled)

```go
type UserService struct {
    repo UserRepository  // Depends on interface
}

func NewUserService(repo UserRepository) *UserService {
    // Dependencies injected from outside
    return &UserService{repo: repo}
}
```

---

## Architecture Components

### 1. Container

The `Container` is the central registry that manages all dependencies.

**Structure:**

```go
type Container struct {
    services  sync.Map  // Stores singleton instances
    factories sync.Map  // Stores factory functions
    resolving sync.Map  // Tracks dependencies being resolved (circular detection)
}
```

**Key Features:**
- Thread-safe using `sync.Map`
- Singleton pattern for registered services
- Factory pattern for lazy initialization
- Circular dependency detection
- Automatic dependency resolution

### 2. Registration Methods

#### Register (Singleton)

Registers a concrete instance that's created immediately and reused for all requests.

```go
func (c *Container) Register(interfaceType interface{}, implementation interface{})
```

**Use when:**
- Instance is ready at startup
- Same instance should be shared across all requests
- No complex initialization needed

#### RegisterFactory (Lazy Initialization)

Registers a factory function that creates instances on first use.

```go
func (c *Container) RegisterFactory(interfaceType interface{}, factory Factory)

type Factory func(*Container) interface{}
```

**Use when:**
- Instance requires complex initialization
- Need to resolve other dependencies first
- Expensive resource creation that should be deferred

### 3. Resolution

Resolves dependencies by interface type, creating instances using factories if needed.

```go
func (c *Container) Resolve(interfaceType interface{}) interface{}
```

**Resolution Flow:**

```
Request Dependency
    │
    ├─> Check if already resolved (singleton)
    │    └─> Return cached instance
    │
    ├─> Check for circular dependency
    │    └─> Detect infinite loops
    │
    ├─> Find factory function
    │    ├─> Execute factory
    │    ├─> Resolve nested dependencies
    │    ├─> Cache result
    │    └─> Return instance
    │
    └─> Fail if not registered
```

---

## Application Setup

### Step 1: Create Container

Initialize the DI container during application startup.

**Example: `main.go`**

```go
package main

import (
    "github.com/your-service/di"
    "github.com/your-service/app"
)

func main() {
    // Create DI container
    container := di.NewContainer()

    // Initialize application with container
    application := app.NewBaseApp("user-service", appConfig)
    application.Container = container

    // Initialize and bootstrap
    application.Init("user-service")
    application.Bootstrap()

    // Start server
    application.StartHTTPHandler(":8080")
}
```

### Step 2: Register Core Dependencies

Register infrastructure components like database connections, HTTP clients, and shared resources.

**Example: `app/app.go`**

```go
func (a *BaseApp) Init(serviceName string) {
    // Register database manager
    dbManager := db.GetDBManager()
    dbManager.Initialize()
    a.Container.Register((*db.DBManagerInterface)(nil), dbManager)

    // Register connections
    a.Container.Register((*utils.Connections)(nil), &a.Connections)

    // Register HTTP clients
    a.registerHttpClients()

    // Register encryption service with factory (lazy init)
    a.Container.RegisterFactory((*utils.EncryptionInterface)(nil), func(c *di.Container) interface{} {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        encryption, err := utils.NewEncryption(ctx)
        if err != nil {
            logger.Error("Failed to initialize encryption", err)
            return nil
        }
        return encryption
    })
}

func (a *BaseApp) registerHttpClients() {
    httpClient := http.Client{
        Timeout: 30 * time.Second,
    }

    // Register HTTP clients for various services
    a.Container.Register((*gateway.AuthClient)(nil), &gateway.AuthClient{
        HttpClient: httpClient,
        BaseURL:    "https://auth.example.com",
    })

    a.Container.Register((*gateway.NotificationClient)(nil), &gateway.NotificationClient{
        HttpClient: httpClient,
        BaseURL:    "https://notifications.example.com",
    })
}
```

### Step 3: Module-Based Configuration

Use modules to organize dependency configuration by feature or domain.

**Example: Module Structure**

```go
package user

import (
    "github.com/your-service/di"
)

type Module struct{}

func NewModule() *Module {
    return &Module{}
}

func (m *Module) Name() string {
    return "user"
}

func (m *Module) Configure(container *di.Container) {
    // Register module-specific dependencies
    m.registerPersistence(container)
    m.registerServices(container)
}

func (m *Module) RegisterRoutes(router *gin.Engine, container *di.Container) {
    // Setup HTTP routes
    handler := NewUserHandler(container)
    router.POST("/users", handler.CreateUser)
    router.GET("/users/:id", handler.GetUser)
}
```

---

## Complete Example: User Management System

### Domain Interfaces

Define interfaces for loose coupling.

**File: `domain/user.go`**

```go
package domain

import "context"

type User struct {
    ID        int64
    Name      string
    Email     string
    CreatedAt time.Time
}

type UserRepository interface {
    CreateUser(ctx context.Context, user *User) error
    GetUserByID(ctx context.Context, id int64) (*User, error)
    UpdateUser(ctx context.Context, user *User) error
    DeleteUser(ctx context.Context, id int64) error
}

type NotificationGateway interface {
    SendWelcomeEmail(ctx context.Context, email string, name string) error
    SendPasswordReset(ctx context.Context, email string, token string) error
}

type EncryptionService interface {
    Encrypt(plaintext string) (string, error)
    Decrypt(ciphertext string) (string, error)
}
```

### Persistence Layer

**File: `persistence/user/user_persistence.go`**

```go
package persistence

import (
    "context"
    "github.com/your-service/domain"
    "github.com/your-service/utils/db"
)

type UserPersistence struct {
    DBManager db.DBManagerInterface
}

func NewUserPersistence(dbManager db.DBManagerInterface) *UserPersistence {
    return &UserPersistence{
        DBManager: dbManager,
    }
}

func (p *UserPersistence) CreateUser(ctx context.Context, user *domain.User) error {
    conn, err := p.DBManager.GetWriteDBConn(ctx)
    if err != nil {
        return err
    }
    defer conn.Close(ctx)

    sqlConn := conn.GetConn(ctx)
    result, err := sqlConn.ExecContext(ctx,
        "INSERT INTO users (name, email, created_at) VALUES (?, ?, NOW())",
        user.Name, user.Email,
    )
    if err != nil {
        return err
    }

    userID, _ := result.LastInsertId()
    user.ID = userID
    return nil
}

func (p *UserPersistence) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
    conn, err := p.DBManager.GetReadDBConn(ctx)
    if err != nil {
        return nil, err
    }
    defer conn.Close(ctx)

    sqlConn := conn.GetConn(ctx)
    row := sqlConn.QueryRowContext(ctx,
        "SELECT id, name, email, created_at FROM users WHERE id = ?",
        id,
    )

    var user domain.User
    err = row.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
    if err != nil {
        return nil, err
    }

    return &user, nil
}

// Implement other methods...
```

### Service Layer

**File: `services/user/user_service.go`**

```go
package services

import (
    "context"
    "github.com/your-service/domain"
)

type UserService struct {
    UserRepo     domain.UserRepository
    Notifications domain.NotificationGateway
    Encryption   domain.EncryptionService
}

func NewUserService(
    userRepo domain.UserRepository,
    notifications domain.NotificationGateway,
    encryption domain.EncryptionService,
) *UserService {
    return &UserService{
        UserRepo:      userRepo,
        Notifications: notifications,
        Encryption:    encryption,
    }
}

func (s *UserService) CreateUser(ctx context.Context, user *domain.User) error {
    // Encrypt sensitive data
    encryptedEmail, err := s.Encryption.Encrypt(user.Email)
    if err != nil {
        return err
    }
    user.Email = encryptedEmail

    // Save to database
    if err := s.UserRepo.CreateUser(ctx, user); err != nil {
        return err
    }

    // Send welcome notification
    decryptedEmail, _ := s.Encryption.Decrypt(user.Email)
    if err := s.Notifications.SendWelcomeEmail(ctx, decryptedEmail, user.Name); err != nil {
        // Log error but don't fail the operation
        logger.Warn(ctx, "Failed to send welcome email", err)
    }

    return nil
}

func (s *UserService) GetUser(ctx context.Context, userID int64) (*domain.User, error) {
    user, err := s.UserRepo.GetUserByID(ctx, userID)
    if err != nil {
        return nil, err
    }

    // Decrypt email for display
    decryptedEmail, err := s.Encryption.Decrypt(user.Email)
    if err != nil {
        return nil, err
    }
    user.Email = decryptedEmail

    return user, nil
}
```

### Module Configuration

**File: `handler/user/module.go`**

```go
package user

import (
    "github.com/gin-gonic/gin"
    "github.com/your-service/di"
    "github.com/your-service/domain"
    "github.com/your-service/gateway"
    userPersistence "github.com/your-service/persistence/user"
    userService "github.com/your-service/services/user"
    "github.com/your-service/utils"
    "github.com/your-service/utils/db"
)

type Module struct{}

func NewModule() *Module {
    return &Module{}
}

func (m *Module) Name() string {
    return "user"
}

func (m *Module) Configure(container *di.Container) {
    // Register persistence layer
    container.RegisterFactory((*domain.UserRepository)(nil), func(c *di.Container) interface{} {
        dbManager := c.Resolve((*db.DBManagerInterface)(nil)).(db.DBManagerInterface)
        return userPersistence.NewUserPersistence(dbManager)
    })

    // Register notification gateway
    container.RegisterFactory((*domain.NotificationGateway)(nil), func(c *di.Container) interface{} {
        notificationClient := c.Resolve((*gateway.NotificationClient)(nil)).(*gateway.NotificationClient)
        return gateway.NewNotificationGateway(notificationClient)
    })

    // Register user service
    container.RegisterFactory((*userService.UserService)(nil), func(c *di.Container) interface{} {
        userRepo := c.Resolve((*domain.UserRepository)(nil)).(domain.UserRepository)
        notifications := c.Resolve((*domain.NotificationGateway)(nil)).(domain.NotificationGateway)
        encryption := c.Resolve((*domain.EncryptionService)(nil)).(domain.EncryptionService)

        return userService.NewUserService(userRepo, notifications, encryption)
    })
}

func (m *Module) RegisterRoutes(router *gin.Engine, container *di.Container) {
    handler := NewUserHandler(container)

    userGroup := router.Group("/api/v1/users")
    {
        userGroup.POST("", handler.CreateUser)
        userGroup.GET("/:id", handler.GetUser)
        userGroup.PUT("/:id", handler.UpdateUser)
        userGroup.DELETE("/:id", handler.DeleteUser)
    }
}
```

### HTTP Handler

**File: `handler/user/handler.go`**

```go
package user

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/your-service/di"
    "github.com/your-service/domain"
    userService "github.com/your-service/services/user"
)

type UserHandler struct {
    container *di.Container
}

func NewUserHandler(container *di.Container) *UserHandler {
    return &UserHandler{
        container: container,
    }
}

func (h *UserHandler) CreateUser(c *gin.Context) {
    var request struct {
        Name  string `json:"name" binding:"required"`
        Email string `json:"email" binding:"required,email"`
    }

    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Resolve service from container
    service := h.container.Resolve((*userService.UserService)(nil)).(*userService.UserService)

    user := &domain.User{
        Name:  request.Name,
        Email: request.Email,
    }

    if err := service.CreateUser(c.Request.Context(), user); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "code": "201",
        "msg":  "SUCCESS",
        "model": user,
    })
}

func (h *UserHandler) GetUser(c *gin.Context) {
    userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    service := h.container.Resolve((*userService.UserService)(nil)).(*userService.UserService)

    user, err := service.GetUser(c.Request.Context(), userID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "code": "200",
        "msg":  "SUCCESS",
        "model": user,
    })
}
```

---

## Dependency Resolution Flow

### Example: Resolving UserService

```
Handler calls: container.Resolve((*UserService)(nil))
    │
    ├─> Container checks if UserService is already resolved
    │    └─> Not found, need to create
    │
    ├─> Container finds UserService factory
    │
    ├─> Factory needs UserRepository
    │    ├─> Container.Resolve((*UserRepository)(nil))
    │    ├─> UserRepository factory needs DBManager
    │    │    └─> Container.Resolve((*DBManagerInterface)(nil))
    │    │         └─> DBManager already registered (return cached)
    │    └─> Create UserPersistence with DBManager
    │
    ├─> Factory needs NotificationGateway
    │    ├─> Container.Resolve((*NotificationGateway)(nil))
    │    ├─> Gateway factory needs NotificationClient
    │    │    └─> Container.Resolve((*NotificationClient)(nil))
    │    │         └─> Client already registered (return cached)
    │    └─> Create NotificationGateway with Client
    │
    ├─> Factory needs EncryptionService
    │    └─> Container.Resolve((*EncryptionService)(nil))
    │         └─> Already registered (return cached)
    │
    ├─> Create UserService with all dependencies
    │
    ├─> Cache UserService instance
    │
    └─> Return UserService to handler
```

---

## Best Practices

### 1. Register Interfaces, Not Concrete Types

```go
// Good: Register by interface
container.Register((*domain.UserRepository)(nil), userPersistence)

// Avoid: Register by concrete type
container.Register((*userPersistence.UserPersistence)(nil), userPersistence)
```

**Why:** Allows swapping implementations without changing dependent code.

### 2. Use Factories for Complex Dependencies

```go
// Good: Factory resolves nested dependencies
container.RegisterFactory((*UserService)(nil), func(c *di.Container) interface{} {
    repo := c.Resolve((*UserRepository)(nil)).(UserRepository)
    return NewUserService(repo)
})

// Avoid: Manual construction with hard dependencies
service := NewUserService(&UserPersistence{db: db})
container.Register((*UserService)(nil), service)
```

### 3. Register Core Dependencies First

Order matters for dependencies:

```go
func (a *BaseApp) Init() {
    // 1. Infrastructure (database, cache)
    a.Container.Register((*db.DBManager)(nil), dbManager)

    // 2. HTTP clients and external gateways
    a.registerHttpClients()

    // 3. Shared services (encryption, logging)
    a.registerSharedServices()

    // 4. Module-specific dependencies (in module.Configure())
}
```

### 4. Keep Module Configuration Isolated

Each module should register only its own dependencies:

```go
func (m *UserModule) Configure(container *di.Container) {
    // Only user-related dependencies
    container.RegisterFactory((*UserRepository)(nil), ...)
    container.RegisterFactory((*UserService)(nil), ...)
}

func (m *OrderModule) Configure(container *di.Container) {
    // Only order-related dependencies
    container.RegisterFactory((*OrderRepository)(nil), ...)
    container.RegisterFactory((*OrderService)(nil), ...)
}
```

### 5. Avoid Circular Dependencies

```go
// Bad: Circular dependency
type ServiceA struct {
    serviceB ServiceB
}

type ServiceB struct {
    serviceA ServiceA  // ServiceB depends on ServiceA
}
```

**Solution:** Introduce an interface or event system:

```go
type ServiceA struct {
    eventBus EventBus  // Both services depend on event bus
}

type ServiceB struct {
    eventBus EventBus
}
```

### 6. Resolve Dependencies Late (In Handlers)

```go
// Good: Resolve in handler method
func (h *UserHandler) CreateUser(c *gin.Context) {
    service := h.container.Resolve((*UserService)(nil)).(*UserService)
    // Use service...
}

// Avoid: Resolve in handler constructor
func NewUserHandler(container *di.Container) *UserHandler {
    service := container.Resolve((*UserService)(nil)).(*UserService)
    return &UserHandler{service: service}  // Early binding
}
```

**Why:** Allows hot-swapping dependencies and better testing.

### 7. Use Type Assertions Safely

```go
// Add error handling for type assertions
service, ok := c.Resolve((*UserService)(nil)).(*UserService)
if !ok || service == nil {
    logger.Error("Failed to resolve UserService")
    return err
}
```

---

## Testing with DI

### Unit Testing with Mocks

```go
package services_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/your-service/di"
    "github.com/your-service/domain"
    "github.com/your-service/services/user"
)

// Mock repository
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *domain.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*domain.User), args.Error(1)
}

// Mock notification gateway
type MockNotificationGateway struct {
    mock.Mock
}

func (m *MockNotificationGateway) SendWelcomeEmail(ctx context.Context, email, name string) error {
    args := m.Called(ctx, email, name)
    return args.Error(0)
}

// Test
func TestUserService_CreateUser(t *testing.T) {
    // Create mocks
    mockRepo := new(MockUserRepository)
    mockNotifications := new(MockNotificationGateway)
    mockEncryption := new(MockEncryption)

    // Setup expectations
    mockEncryption.On("Encrypt", "user@example.com").Return("encrypted_email", nil)
    mockRepo.On("CreateUser", mock.Anything, mock.Anything).Return(nil)
    mockEncryption.On("Decrypt", "encrypted_email").Return("user@example.com", nil)
    mockNotifications.On("SendWelcomeEmail", mock.Anything, "user@example.com", "John").Return(nil)

    // Create service with mocks
    service := user.NewUserService(mockRepo, mockNotifications, mockEncryption)

    // Execute test
    testUser := &domain.User{
        Name:  "John",
        Email: "user@example.com",
    }

    err := service.CreateUser(context.Background(), testUser)

    // Assertions
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
    mockNotifications.AssertExpectations(t)
}
```

### Integration Testing with Test Container

```go
func TestUserHandler_Integration(t *testing.T) {
    // Create test container
    container := di.NewContainer()

    // Register test dependencies
    testDB := setupTestDatabase()
    container.Register((*db.DBManagerInterface)(nil), testDB)

    // Register real implementations
    container.RegisterFactory((*domain.UserRepository)(nil), func(c *di.Container) interface{} {
        dbManager := c.Resolve((*db.DBManagerInterface)(nil)).(db.DBManagerInterface)
        return persistence.NewUserPersistence(dbManager)
    })

    container.RegisterFactory((*user.UserService)(nil), func(c *di.Container) interface{} {
        repo := c.Resolve((*domain.UserRepository)(nil)).(domain.UserRepository)
        // Use mock for external services
        return user.NewUserService(repo, mockNotifications, mockEncryption)
    })

    // Test handler
    handler := NewUserHandler(container)

    // Create test HTTP request
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    // ... test logic
}
```

---

## Circular Dependency Detection

The container automatically detects circular dependencies:

```go
// Example circular dependency scenario:
// ServiceA needs ServiceB
// ServiceB needs ServiceA

container.RegisterFactory((*ServiceA)(nil), func(c *di.Container) interface{} {
    serviceB := c.Resolve((*ServiceB)(nil)).(ServiceB)  // Tries to resolve ServiceB
    return NewServiceA(serviceB)
})

container.RegisterFactory((*ServiceB)(nil), func(c *di.Container) interface{} {
    serviceA := c.Resolve((*ServiceA)(nil)).(ServiceA)  // Tries to resolve ServiceA (circular!)
    return NewServiceB(serviceA)
})

// When resolving:
service := container.Resolve((*ServiceA)(nil))
// Output: "DI: Circular dependency detected for ServiceA"
// Returns: nil
```

**Prevention:** Use events, callbacks, or interfaces to break circular dependencies.

---

## Common Patterns

### Pattern 1: Repository + Service

```go
// Repository: Data access
container.RegisterFactory((*UserRepository)(nil), func(c *di.Container) interface{} {
    db := c.Resolve((*db.DBManager)(nil)).(db.DBManager)
    return persistence.NewUserRepository(db)
})

// Service: Business logic
container.RegisterFactory((*UserService)(nil), func(c *di.Container) interface{} {
    repo := c.Resolve((*UserRepository)(nil)).(UserRepository)
    return services.NewUserService(repo)
})
```

### Pattern 2: Gateway with HTTP Client

```go
// HTTP Client (singleton)
container.Register((*http.Client)(nil), &http.Client{Timeout: 30 * time.Second})

// Gateway (factory)
container.RegisterFactory((*PaymentGateway)(nil), func(c *di.Container) interface{} {
    client := c.Resolve((*http.Client)(nil)).(*http.Client)
    config := loadPaymentConfig()
    return gateway.NewPaymentGateway(client, config)
})
```

### Pattern 3: Multi-Implementation Selection

```go
// Register multiple implementations
container.Register((*EmailProvider)(nil), &SMTPProvider{})
container.Register((*SMSProvider)(nil), &TwilioProvider{})

// Service uses both
container.RegisterFactory((*NotificationService)(nil), func(c *di.Container) interface{} {
    email := c.Resolve((*EmailProvider)(nil)).(EmailProvider)
    sms := c.Resolve((*SMSProvider)(nil)).(SMSProvider)
    return services.NewNotificationService(email, sms)
})
```

---

## Troubleshooting

### Issue 1: Dependency Not Found

**Symptom:** `DI: Failed to resolve <InterfaceName>`

**Solution:**
- Verify the interface is registered in a module's `Configure` method
- Check if the module is registered with the application
- Ensure registration happens before resolution

```go
// Verify registration
if !container.IsRegistered((*UserService)(nil)) {
    logger.Error("UserService not registered!")
}
```

### Issue 2: Nil Pointer Panic

**Symptom:** `panic: runtime error: invalid memory address or nil pointer dereference`

**Solution:**
- Check if `Resolve` returns `nil`
- Add nil checks after type assertion
- Verify factory function returns non-nil value

```go
service := container.Resolve((*UserService)(nil))
if service == nil {
    logger.Error("Failed to resolve UserService")
    return
}
userService := service.(*UserService)
```

### Issue 3: Wrong Type Assertion

**Symptom:** `panic: interface conversion: interface {} is X, not Y`

**Solution:**
- Use safe type assertion with ok check
- Verify the registered type matches the resolved type

```go
// Safe type assertion
if service, ok := container.Resolve((*UserService)(nil)).(*UserService); ok {
    // Use service
} else {
    logger.Error("Type assertion failed")
}
```

---

## Summary

The Dependency Injection architecture provides:

- ✅ Centralized dependency management
- ✅ Loose coupling through interfaces
- ✅ Easy testing with mock dependencies
- ✅ Lazy initialization with factories
- ✅ Singleton pattern for shared resources
- ✅ Automatic dependency resolution
- ✅ Circular dependency detection
- ✅ Thread-safe operations
- ✅ Module-based organization

By following this architecture, you ensure maintainable, testable, and flexible code that scales with your application's complexity.
