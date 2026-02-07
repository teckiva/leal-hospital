package domain

// User represents user domain model
type User struct {
	ID           int64
	Name         string
	Mobile       string
	Email        string
	Designation  string
	Status       string
	IsAdmin      bool
	IsApproved   bool
	PasswordHash string
}

// TokenClaims represents JWT token claims (domain model)
type TokenClaims struct {
	UserID  int64
	Email   string
	IsAdmin bool
	Type    string // "access" or "refresh"
}

// ============ Request Domain Models ============

// RegisterRequest represents registration request in domain layer
type RegisterRequest struct {
	Name        string
	Mobile      string
	Email       string
	Designation string
	Password    string
}

// VerifyOTPRequest represents OTP verification request in domain layer
type VerifyOTPRequest struct {
	Email   string
	OTP     string
	OTPType string
}

// LoginRequest represents login request in domain layer
type LoginRequest struct {
	Email    string
	Password string
}

// ForgotPasswordRequest represents forgot password request in domain layer
type ForgotPasswordRequest struct {
	Email string
}

// ResetPasswordRequest represents reset password request in domain layer
type ResetPasswordRequest struct {
	Email       string
	OTP         string
	NewPassword string
}

// RefreshTokenRequest represents refresh token request in domain layer
type RefreshTokenRequest struct {
	RefreshToken string
}

// ============ Response Domain Models ============

// RegisterResponse represents registration response in domain layer
type RegisterResponse struct {
	Message string
	UserID  int64
}

// VerifyOTPResponse represents OTP verification response in domain layer
type VerifyOTPResponse struct {
	Message      string
	IsVerified   bool
	AccessToken  string
	RefreshToken string
}

// LoginResponse represents login response in domain layer
type LoginResponse struct {
	Message string
}

// ForgotPasswordResponse represents forgot password response in domain layer
type ForgotPasswordResponse struct {
	Message string
}

// ResetPasswordResponse represents reset password response in domain layer
type ResetPasswordResponse struct {
	Message string
}

// RefreshTokenResponse represents refresh token response in domain layer
type RefreshTokenResponse struct {
	AccessToken  string
	RefreshToken string
}
