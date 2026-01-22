# Error Handling Architecture

## Overview

This document describes the centralized error handling system that provides consistent error management across the service. The architecture consists of three main components:

1. **Error Configuration** (`config/errors.yaml`) - Centralized error definitions
2. **Error Registry** (`payerror/loaderrorcode.go`) - Error loading and caching mechanism
3. **Error Handler** (`payerror/error.go`) - Error construction and response formatting

## Architecture Components

### 1. Error Configuration File (`errors.yaml`)

The `errors.yaml` file contains all error definitions organized by source/category. Each error has a unique code and includes metadata for both internal logging and user-facing messages.

**Structure:**
```yaml
errors:
  <source_name>:
    "<error_code>":
      type: "<error_type>"
      message: "<internal_message>"
      display_message: "<user_facing_message>"
      decline_type: "<optional_decline_category>"
      source: "<source_name>"
```

**Field Descriptions:**
- `source_name`: Category or external system (e.g., INTERNAL, GATEWAY, DATABASE)
- `error_code`: Unique identifier for the error within the source
- `type`: Error classification (BAD_REQUEST, SYSTEM_ERROR, GATEWAY_ERROR, BUSINESS_ERROR, DATABASE_ERROR)
- `message`: Technical message for logging and debugging
- `display_message`: User-friendly message shown to clients
- `decline_type`: Optional categorization (e.g., TD=Technical Decline, BD=Business Decline)
- `source`: Reference back to the source category

### 2. Error Registry

The error registry loads all errors from `errors.yaml` at application startup and caches them in memory for fast lookup during request processing.

**Key Features:**
- Thread-safe error lookups using `sync.RWMutex`
- Singleton pattern via Gin middleware
- Hierarchical key structure: `source:code` (e.g., `INTERNAL:404`, `GATEWAY:TIMEOUT`)
- Fallback handling for unmapped errors

**Registry Methods:**
```go
// Get error by code and source
GetError(code, source string) *AppError

// Get gateway-specific error
GetErrorForGateway(code, source string) *AppError

// Extract and map gateway response to error
GetGatewayError(response interface{}) *AppError
```

### 3. Error Handler

Provides structured error objects with stack traces, display messages, and response formatting.

**Error Types:**
```go
const (
    ValidationError ErrorType = "BAD_REQUEST"     // 400 - Client input errors
    SystemError     ErrorType = "SYSTEM_ERROR"    // 500 - Internal failures
    GatewayErr      ErrorType = "GATEWAY_ERROR"   // 453 - External service errors
    BusinessError   ErrorType = "BUSINESS_ERROR"  // 600 - Business logic errors
    DatabaseError   ErrorType = "DATABASE_ERROR"  // 503 - Data persistence errors
    Failure         ErrorType = "FAILED"          // General failure
)
```

**Constructor Functions:**
```go
// Create new error with formatted message
Newf(errType ErrorType, code ErrorCode, msg string, args ...interface{}) *AppError

// Wrap existing error
Wrapf(err error, errType ErrorType, code ErrorCode, msg string, args ...interface{}) *AppError
```

**Builder Methods:**
```go
WithDisplayMessage(msg string) *AppError    // Set user-facing message
WithDetails(details []ErrorDetail) *AppError // Add validation details
WithGatewayError(response interface{}) *AppError // Map gateway response
```

## Complete Example: Order Management System

Let's walk through a complete example of adding error handling for an order management system.

### Step 1: Define Errors in `errors.yaml`

Add error definitions for the order service:

```yaml
errors:
  # ... existing errors ...

  ORDER_SERVICE:
    "ORD001":
      type: "BUSINESS_ERROR"
      message: "Order not found in the system"
      display_message: "The requested order could not be found. Please check the order ID."
      decline_type: "BD"
      source: "ORDER_SERVICE"

    "ORD002":
      type: "BUSINESS_ERROR"
      message: "Order already cancelled"
      display_message: "This order has already been cancelled and cannot be modified."
      decline_type: "BD"
      source: "ORDER_SERVICE"

    "ORD003":
      type: "BUSINESS_ERROR"
      message: "Insufficient inventory for order"
      display_message: "Some items in your order are out of stock. Please review your cart."
      decline_type: "BD"
      source: "ORDER_SERVICE"

    "ORD004":
      type: "VALIDATION_ERROR"
      message: "Invalid order quantity"
      display_message: "Please enter a valid quantity between 1 and 100."
      decline_type: "BD"
      source: "ORDER_SERVICE"

  PAYMENT_GATEWAY:
    "PAY001":
      type: "GATEWAY_ERROR"
      message: "Payment gateway timeout"
      display_message: "Payment processing is taking longer than expected. Please try again."
      decline_type: "TD"
      source: "PAYMENT_GATEWAY"

    "PAY002":
      type: "GATEWAY_ERROR"
      message: "Payment declined by gateway"
      display_message: "Your payment was declined. Please check your payment details and try again."
      decline_type: "BD"
      source: "PAYMENT_GATEWAY"

  INTERNAL:
    "500":
      type: "SYSTEM_ERROR"
      message: "Internal server error"
      display_message: "Something went wrong on our end. Please try again later."
      decline_type: "TD"
      source: "INTERNAL"

    "404":
      type: "BAD_REQUEST"
      message: "Resource not found"
      display_message: "The requested resource was not found."
      decline_type: "BD"
      source: "INTERNAL"
```

### Step 2: Initialize Error Registry in Application

In your `main.go` or application initialization:

```go
package main

import (
    "github.com/gin-gonic/gin"
    "your-service/payerror"
)

func main() {
    router := gin.Default()

    // Initialize error registry - loads all errors from errors.yaml
    // This middleware makes the registry available in all request contexts
    if err := payerror.InitErrorRegistry(router); err != nil {
        panic("Failed to initialize error registry: " + err.Error())
    }

    // Register routes
    registerRoutes(router)

    router.Run(":8080")
}
```

### Step 3: Use Errors in Your Service Layer

**Example Service: `services/order/order_service.go`**

```go
package order

import (
    "context"
    "github.com/gin-gonic/gin"
    "your-service/payerror"
    "your-service/domain"
)

type OrderService struct {
    orderRepo OrderRepository
}

// CreateOrder validates and creates a new order
func (s *OrderService) CreateOrder(c *gin.Context, request *CreateOrderRequest) (*domain.Order, *payerror.AppError) {
    // Get error registry from context
    registry := payerror.GetErrorRegistry(c)

    // Validate quantity
    if request.Quantity < 1 || request.Quantity > 100 {
        // Return mapped error from errors.yaml
        return nil, registry.GetError("ORD004", "ORDER_SERVICE")
    }

    // Check inventory
    inventory, err := s.orderRepo.CheckInventory(request.ProductID, request.Quantity)
    if err != nil {
        // Wrap database error
        return nil, payerror.Wrapf(
            err,
            payerror.DatabaseError,
            payerror.ErrDatabase,
            "Failed to check inventory for product %s",
            request.ProductID,
        )
    }

    if inventory < request.Quantity {
        // Return business error for insufficient inventory
        return nil, registry.GetError("ORD003", "ORDER_SERVICE")
    }

    // Create order
    order, err := s.orderRepo.CreateOrder(request)
    if err != nil {
        return nil, payerror.Wrapf(
            err,
            payerror.DatabaseError,
            payerror.ErrDatabase,
            "Failed to create order",
        )
    }

    return order, nil
}

// CancelOrder cancels an existing order
func (s *OrderService) CancelOrder(c *gin.Context, orderID string) *payerror.AppError {
    registry := payerror.GetErrorRegistry(c)

    // Fetch order
    order, err := s.orderRepo.GetOrderByID(orderID)
    if err != nil {
        return registry.GetError("ORD001", "ORDER_SERVICE")
    }

    // Check if already cancelled
    if order.Status == "CANCELLED" {
        return registry.GetError("ORD002", "ORDER_SERVICE")
    }

    // Cancel the order
    if err := s.orderRepo.UpdateOrderStatus(orderID, "CANCELLED"); err != nil {
        return payerror.Wrapf(
            err,
            payerror.DatabaseError,
            payerror.ErrDatabase,
            "Failed to cancel order %s",
            orderID,
        )
    }

    return nil
}
```

### Step 4: Handle Errors in HTTP Handlers

**Example Handler: `handler/order/order_handler.go`**

```go
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "your-service/payerror"
    "your-service/services/order"
)

type OrderHandler struct {
    orderService *order.OrderService
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
    var request CreateOrderRequest

    // Parse request
    if err := c.ShouldBindJSON(&request); err != nil {
        registry := payerror.GetErrorRegistry(c)
        appErr := registry.GetError("ORD004", "ORDER_SERVICE")
        c.JSON(http.StatusBadRequest, appErr.ToResponse())
        return
    }

    // Call service
    order, appErr := h.orderService.CreateOrder(c, &request)
    if appErr != nil {
        // Convert error to HTTP response
        c.JSON(http.StatusOK, appErr.ToResponse())
        return
    }

    // Success response
    c.JSON(http.StatusOK, gin.H{
        "code": "200",
        "msg": "SUCCESS",
        "model": order,
    })
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
    orderID := c.Param("orderId")

    // Call service
    if appErr := h.orderService.CancelOrder(c, orderID); appErr != nil {
        c.JSON(http.StatusOK, appErr.ToResponse())
        return
    }

    // Success response
    c.JSON(http.StatusOK, gin.H{
        "code": "200",
        "msg": "SUCCESS",
        "model": gin.H{
            "message": "Order cancelled successfully",
        },
    })
}
```

### Step 5: Handle Gateway Errors

**Example: Processing Payment Gateway Response**

```go
package payment

import (
    "github.com/gin-gonic/gin"
    "your-service/payerror"
)

type PaymentService struct {
    gatewayClient GatewayClient
}

func (s *PaymentService) ProcessPayment(c *gin.Context, request *PaymentRequest) *payerror.AppError {
    registry := payerror.GetErrorRegistry(c)

    // Call external payment gateway
    response, err := s.gatewayClient.ProcessPayment(request)
    if err != nil {
        // Wrap network/connection error
        return payerror.Wrapf(
            err,
            payerror.GatewayErr,
            payerror.ErrGateway,
            "Failed to connect to payment gateway",
        ).WithDisplayMessage("Unable to process payment. Please try again.")
    }

    // Check gateway response status
    if response.Status != "SUCCESS" {
        // Map gateway error code to our error system
        // This will lookup "PAYMENT_GATEWAY:PAY001" or "PAYMENT_GATEWAY:PAY002"
        return registry.GetGatewayError(response)
    }

    return nil
}
```

## Error Response Format

All errors are converted to a consistent response format:

```json
{
  "code": "499",
  "msg": "FAILED",
  "model": {
    "errorCode": "ORD003",
    "message": "Insufficient inventory for order",
    "displayMessage": "Some items in your order are out of stock. Please review your cart.",
    "details": [
      {
        "fieldViolations": [
          {
            "field": "quantity",
            "description": "Requested quantity exceeds available inventory"
          }
        ]
      }
    ]
  }
}
```

## Best Practices

### 1. Error Code Naming Convention
- Use descriptive prefixes: `ORD` (Order), `PAY` (Payment), `INV` (Inventory)
- Use sequential numbering: `ORD001`, `ORD002`, etc.
- Keep codes short but meaningful

### 2. Error Messages
- **message**: Technical, detailed, for logging and debugging
- **display_message**: User-friendly, actionable, avoid technical jargon

### 3. Error Type Selection
- `BAD_REQUEST` (400): Client provided invalid input
- `BUSINESS_ERROR` (600): Business rule violation
- `GATEWAY_ERROR` (453): External service failure
- `DATABASE_ERROR` (503): Data persistence failure
- `SYSTEM_ERROR` (500): Unexpected internal error

### 4. Error Source Organization
Group related errors under the same source:
```yaml
ORDER_SERVICE:     # All order-related errors
PAYMENT_GATEWAY:   # Payment processing errors
INVENTORY_SERVICE: # Inventory management errors
INTERNAL:          # System-level errors
```

### 5. Using Error Registry
Always retrieve errors through the registry:
```go
// GOOD: Uses registry for consistent error handling
registry := payerror.GetErrorRegistry(c)
return registry.GetError("ORD001", "ORDER_SERVICE")

// AVOID: Creating errors manually bypasses configuration
return payerror.Newf(payerror.BusinessError, "ORD001", "Order not found")
```

### 6. Adding Field Validation Errors
For input validation failures, include field-level details:
```go
details := []payerror.ErrorDetail{
    {
        FieldViolations: []payerror.FieldViolation{
            {
                Field:       "email",
                Description: "Invalid email format",
            },
            {
                Field:       "quantity",
                Description: "Must be between 1 and 100",
            },
        },
    },
}

return registry.GetError("ORD004", "ORDER_SERVICE").WithDetails(details)
```

## Common Patterns

### Pattern 1: Resource Not Found
```go
user, err := repo.GetUserByID(userID)
if err != nil {
    return registry.GetError("USR001", "USER_SERVICE")
}
```

### Pattern 2: Validation Failure
```go
if !isValidEmail(request.Email) {
    return registry.GetError("USR002", "USER_SERVICE")
}
```

### Pattern 3: Database Error Wrapping
```go
if err := repo.SaveUser(user); err != nil {
    return payerror.Wrapf(
        err,
        payerror.DatabaseError,
        payerror.ErrDatabase,
        "Failed to save user %s", user.ID,
    )
}
```

### Pattern 4: Gateway Error Handling
```go
response, err := gatewayClient.Call(request)
if err != nil {
    return payerror.Wrapf(err, payerror.GatewayErr, payerror.ErrGateway, "Gateway call failed")
}

if response.Status != "SUCCESS" {
    return registry.GetGatewayError(response)
}
```

## Testing Error Handling

### Unit Test Example
```go
func TestCreateOrder_InsufficientInventory(t *testing.T) {
    // Setup
    mockRepo := &MockOrderRepository{
        InventoryQty: 5,
    }
    service := &OrderService{orderRepo: mockRepo}

    // Create test context with error registry
    c, _ := gin.CreateTestContext(nil)
    registry, _ := payerror.InitializeErrorRegistry()
    c.Set(payerror.ErrorRegistryKey, registry)

    request := &CreateOrderRequest{
        ProductID: "PROD123",
        Quantity:  10, // More than available
    }

    // Execute
    order, appErr := service.CreateOrder(c, request)

    // Assert
    assert.Nil(t, order)
    assert.NotNil(t, appErr)
    assert.Equal(t, "ORD003", string(appErr.code))
    assert.Contains(t, appErr.displayMessage, "out of stock")
}
```

## Troubleshooting

### Error Registry Not Found
**Symptom:** `GetErrorRegistry` returns `nil`

**Solution:** Ensure `InitErrorRegistry` is called before registering routes:
```go
if err := payerror.InitErrorRegistry(router); err != nil {
    panic(err)
}
```

### Error Code Not Mapped
**Symptom:** Returns "Error code mapping not found"

**Solution:**
1. Check `errors.yaml` has the error defined
2. Verify source name and code match exactly (case-sensitive)
3. Restart application to reload error registry

### Custom Error Messages Not Showing
**Symptom:** Users see generic error messages

**Solution:** Use `WithDisplayMessage()` to override:
```go
return registry.GetError("ORD001", "ORDER_SERVICE").
    WithDisplayMessage("Custom message for this specific case")
```

## Migration Guide

### Adding New Errors
1. Add error definition to `errors.yaml` under appropriate source
2. No code changes needed - registry auto-loads on restart
3. Use `registry.GetError(code, source)` in your code

### Modifying Existing Errors
1. Update error message or display_message in `errors.yaml`
2. Restart application to reload
3. No code changes needed if error code remains the same

### Deprecating Errors
1. Keep error definition in `errors.yaml` for backward compatibility
2. Add comment in YAML indicating deprecation
3. Update code to use new error code
4. Remove old error after grace period

---

## Summary

The error handling architecture provides:
- ✅ Centralized error configuration
- ✅ Consistent error responses across the service
- ✅ User-friendly messages separate from technical details
- ✅ Fast in-memory error lookups
- ✅ Easy error definition and management
- ✅ Gateway error mapping support
- ✅ Stack traces for debugging

By following this architecture, you ensure consistent, maintainable, and user-friendly error handling throughout the service.
