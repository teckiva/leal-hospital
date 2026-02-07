package medierror

import (
	"fmt"
	"runtime"
)

// ErrorType represents the type of error
type ErrorType string

const (
	ValidationError ErrorType = "BAD_REQUEST"
	SystemError     ErrorType = "SYSTEM_ERROR"
	GatewayErr      ErrorType = "GATEWAY_ERROR"
	BusinessError   ErrorType = "BUSINESS_ERROR"
	DatabaseError   ErrorType = "DATABASE_ERROR"
	Failure         ErrorType = "FAILED"
)

// ErrorCode represents an error code
type ErrorCode string

// Common HTTP error codes
const (
	ErrBadRequest   ErrorCode = "400"
	ErrUnauthorized ErrorCode = "401"
	ErrForbidden    ErrorCode = "403"
	ErrNotFound     ErrorCode = "404"
	ErrInternal     ErrorCode = "500"
	ErrDatabase     ErrorCode = "503"
)

// AppError represents an application error with context
type AppError struct {
	errType        ErrorType
	code           ErrorCode
	message        string
	displayMessage string
	declineType    string
	source         string
	stack          string
	cause          error
	details        []ErrorDetail
}

// ErrorDetail represents field-level validation error details
type ErrorDetail struct {
	FieldViolations []FieldViolation `json:"fieldViolations,omitempty"`
}

// FieldViolation represents a single field validation error
type FieldViolation struct {
	Field       string `json:"field"`
	Description string `json:"description"`
}

// ErrorResponse represents the API error response structure
type ErrorResponse struct {
	Code  string      `json:"code"`
	Msg   string      `json:"msg"`
	Model ErrorModel  `json:"model"`
}

// ErrorModel represents the error details in the response
type ErrorModel struct {
	ErrorCode      string        `json:"errorCode"`
	Message        string        `json:"message"`
	DisplayMessage string        `json:"displayMessage"`
	DeclineType    string        `json:"declineType,omitempty"`
	Source         string        `json:"source,omitempty"`
	Details        []ErrorDetail `json:"details,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("[%s:%s] %s: %v", e.source, e.code, e.message, e.cause)
	}
	return fmt.Sprintf("[%s:%s] %s", e.source, e.code, e.message)
}

// Code returns the error code
func (e *AppError) Code() ErrorCode {
	return e.code
}

// Message returns the internal error message
func (e *AppError) Message() string {
	return e.message
}

// DisplayMessage returns the user-friendly error message
func (e *AppError) DisplayMessage() string {
	return e.displayMessage
}

// DeclineType returns the decline type
func (e *AppError) DeclineType() string {
	return e.declineType
}

// Source returns the error source
func (e *AppError) Source() string {
	return e.source
}

// Stack returns the stack trace
func (e *AppError) Stack() string {
	return e.stack
}

// Cause returns the underlying cause error
func (e *AppError) Cause() error {
	return e.cause
}

// Type returns the error type
func (e *AppError) Type() ErrorType {
	return e.errType
}

// Unwrap returns the underlying cause error (for Go 1.13+ error unwrapping)
func (e *AppError) Unwrap() error {
	return e.cause
}

// WithDisplayMessage sets a custom display message
func (e *AppError) WithDisplayMessage(msg string) *AppError {
	e.displayMessage = msg
	return e
}

// WithDetails adds field validation details
func (e *AppError) WithDetails(details []ErrorDetail) *AppError {
	e.details = details
	return e
}

// ToResponse converts the AppError to an API response format
func (e *AppError) ToResponse() ErrorResponse {
	return ErrorResponse{
		Code: "499",
		Msg:  string(Failure),
		Model: ErrorModel{
			ErrorCode:      string(e.code),
			Message:        e.message,
			DisplayMessage: e.displayMessage,
			DeclineType:    e.declineType,
			Source:         e.source,
			Details:        e.details,
		},
	}
}

// Newf creates a new AppError with formatted message
func Newf(errType ErrorType, code ErrorCode, msg string, args ...any) *AppError {
	message := fmt.Sprintf(msg, args...)
	return &AppError{
		errType:        errType,
		code:           code,
		message:        message,
		displayMessage: message,
		source:         "INTERNAL",
		stack:          captureStackTrace(),
	}
}

// Wrapf wraps an existing error with additional context
func Wrapf(err error, errType ErrorType, code ErrorCode, msg string, args ...any) *AppError {
	message := fmt.Sprintf(msg, args...)
	return &AppError{
		errType:        errType,
		code:           code,
		message:        message,
		displayMessage: message,
		source:         "INTERNAL",
		stack:          captureStackTrace(),
		cause:          err,
	}
}

// NewFromRegistry creates an AppError from registry lookup
func NewFromRegistry(code ErrorCode, cause error) *AppError {
	registry := GetGlobalRegistry()
	if registry == nil {
		return &AppError{
			errType:        SystemError,
			code:           code,
			message:        "Error registry not initialized",
			displayMessage: "An error occurred",
			declineType:    "TD",
			source:         "INTERNAL",
			stack:          captureStackTrace(),
			cause:          cause,
		}
	}

	config, err := registry.GetError(code)
	if err != nil {
		return &AppError{
			errType:        SystemError,
			code:           code,
			message:        fmt.Sprintf("Unknown error: %s", code),
			displayMessage: "An error occurred",
			declineType:    "TD",
			source:         "INTERNAL",
			stack:          captureStackTrace(),
			cause:          cause,
		}
	}

	errType := getErrorTypeFromSource(config.Source)
	return &AppError{
		errType:        errType,
		code:           code,
		message:        config.Message,
		displayMessage: config.DisplayMessage,
		declineType:    config.DeclineType,
		source:         config.Source,
		stack:          captureStackTrace(),
		cause:          cause,
	}
}

// getErrorTypeFromSource maps error source to error type
func getErrorTypeFromSource(source string) ErrorType {
	switch source {
	case "INTERNAL":
		return SystemError
	case "AUTH_SERVICE", "USER_SERVICE", "PATIENT_SERVICE", "OPD_SERVICE":
		return BusinessError
	default:
		return SystemError
	}
}

// captureStackTrace captures the current stack trace
func captureStackTrace() string {
	const maxStackDepth = 32
	pc := make([]uintptr, maxStackDepth)
	n := runtime.Callers(3, pc)
	if n == 0 {
		return ""
	}

	frames := runtime.CallersFrames(pc[:n])
	stackTrace := ""

	for {
		frame, more := frames.Next()
		stackTrace += fmt.Sprintf("\n\t%s:%d %s", frame.File, frame.Line, frame.Function)
		if !more {
			break
		}
	}

	return stackTrace
}

// Common error constructors

// ErrBadRequestWithMsg creates a bad request error with custom message
func ErrBadRequestWithMsg(msg string, displayMsg string, cause error) *AppError {
	return &AppError{
		errType:        ValidationError,
		code:           ErrBadRequest,
		message:        msg,
		displayMessage: displayMsg,
		source:         "INTERNAL",
		stack:          captureStackTrace(),
		cause:          cause,
	}
}

// ErrUnauthorizedWithMsg creates an unauthorized error with custom message
func ErrUnauthorizedWithMsg(msg string, displayMsg string, cause error) *AppError {
	return &AppError{
		errType:        ValidationError,
		code:           ErrUnauthorized,
		message:        msg,
		displayMessage: displayMsg,
		source:         "INTERNAL",
		stack:          captureStackTrace(),
		cause:          cause,
	}
}

// ErrForbiddenWithMsg creates a forbidden error with custom message
func ErrForbiddenWithMsg(msg string, displayMsg string, cause error) *AppError {
	return &AppError{
		errType:        ValidationError,
		code:           ErrForbidden,
		message:        msg,
		displayMessage: displayMsg,
		source:         "INTERNAL",
		stack:          captureStackTrace(),
		cause:          cause,
	}
}

// ErrNotFoundWithMsg creates a not found error with custom message
func ErrNotFoundWithMsg(msg string, displayMsg string, cause error) *AppError {
	return &AppError{
		errType:        ValidationError,
		code:           ErrNotFound,
		message:        msg,
		displayMessage: displayMsg,
		source:         "INTERNAL",
		stack:          captureStackTrace(),
		cause:          cause,
	}
}

// ErrInternalWithMsg creates an internal server error with custom message
func ErrInternalWithMsg(msg string, displayMsg string, cause error) *AppError {
	return &AppError{
		errType:        SystemError,
		code:           ErrInternal,
		message:        msg,
		displayMessage: displayMsg,
		source:         "INTERNAL",
		stack:          captureStackTrace(),
		cause:          cause,
	}
}

// ErrDatabaseWithMsg creates a database error with custom message
func ErrDatabaseWithMsg(msg string, displayMsg string, cause error) *AppError {
	return &AppError{
		errType:        DatabaseError,
		code:           ErrDatabase,
		message:        msg,
		displayMessage: displayMsg,
		source:         "INTERNAL",
		stack:          captureStackTrace(),
		cause:          cause,
	}
}

// ErrInternalServer creates an internal server error (1004)
func ErrInternalServer() *AppError {
	return NewFromRegistry("1004", nil)
}

// Business error constructors using registry

// ErrInvalidCredentials creates an invalid credentials error (2000)
func ErrInvalidCredentials(cause error) *AppError {
	return NewFromRegistry("2000", cause)
}

// ErrOTPExpired creates an OTP expired error (2001)
func ErrOTPExpired(cause error) *AppError {
	return NewFromRegistry("2001", cause)
}

// ErrInvalidOTP creates an invalid OTP error (2002)
func ErrInvalidOTP(cause error) *AppError {
	return NewFromRegistry("2002", cause)
}

// ErrSessionExpired creates a session expired error (2003)
func ErrSessionExpired(cause error) *AppError {
	return NewFromRegistry("2003", cause)
}

// ErrMaxOTPAttempts creates a max OTP attempts error (2004)
func ErrMaxOTPAttempts(cause error) *AppError {
	return NewFromRegistry("2004", cause)
}

// ErrStaffNotApproved creates a staff not approved error (2005)
func ErrStaffNotApproved(cause error) *AppError {
	return NewFromRegistry("2005", cause)
}

// ErrUserNotFound creates a user not found error (3000)
func ErrUserNotFound(cause error) *AppError {
	return NewFromRegistry("3000", cause)
}

// ErrUserAlreadyExists creates a user already exists error (3001)
func ErrUserAlreadyExists(cause error) *AppError {
	return NewFromRegistry("3001", cause)
}

// ErrInvalidMobile creates an invalid mobile number error (3002)
func ErrInvalidMobile(cause error) *AppError {
	return NewFromRegistry("3002", cause)
}

// ErrInvalidPassword creates an invalid password format error (3003)
func ErrInvalidPassword(cause error) *AppError {
	return NewFromRegistry("3003", cause)
}

// ErrPatientNotFound creates a patient not found error (4000)
func ErrPatientNotFound(cause error) *AppError {
	return NewFromRegistry("4000", cause)
}

// ErrInvalidPatientData creates an invalid patient data error (4001)
func ErrInvalidPatientData(cause error) *AppError {
	return NewFromRegistry("4001", cause)
}

// ErrPatientAlreadyRegistered creates a patient already registered error (4002)
func ErrPatientAlreadyRegistered(cause error) *AppError {
	return NewFromRegistry("4002", cause)
}

// ErrOPDNotFound creates an OPD not found error (5000)
func ErrOPDNotFound(cause error) *AppError {
	return NewFromRegistry("5000", cause)
}

// ErrInvalidOPDData creates an invalid OPD data error (5001)
func ErrInvalidOPDData(cause error) *AppError {
	return NewFromRegistry("5001", cause)
}

// ErrOPDAlreadyExists creates an OPD already exists error (5002)
func ErrOPDAlreadyExists(cause error) *AppError {
	return NewFromRegistry("5002", cause)
}

// Helper functions

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError converts an error to AppError if possible
func GetAppError(err error) (*AppError, bool) {
	if err == nil {
		return nil, false
	}
	appErr, ok := err.(*AppError)
	return appErr, ok
}
