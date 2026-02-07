package dto

import "github.com/leal-hospital/server/domain"

// RegisterRequest represents user registration request
// POST /api/auth/register
type RegisterRequest struct {
	Name        string `json:"name" binding:"required"`
	Mobile      string `json:"mobile" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Designation string `json:"designation" binding:"required,oneof=doctor nurse staff"`
	Password    string `json:"password" binding:"required,min=8"`
}

// ConvertRegisterRequestToDomain converts DTO request to domain model
func ConvertRegisterRequestToDomain(req *RegisterRequest) *domain.RegisterRequest {
	return &domain.RegisterRequest{
		Name:        req.Name,
		Mobile:      req.Mobile,
		Email:       req.Email,
		Designation: req.Designation,
		Password:    req.Password,
	}
}

// VerifyOTPRequest represents OTP verification request
// POST /api/auth/verify-otp
type VerifyOTPRequest struct {
	Email   string `json:"email" binding:"required,email"`
	OTP     string `json:"otp" binding:"required,len=6"`
	OTPType string `json:"otp_type" binding:"required,oneof=registration login forgot_password"`
}

// ConvertVerifyOTPRequestToDomain converts DTO request to domain model
func ConvertVerifyOTPRequestToDomain(req *VerifyOTPRequest) *domain.VerifyOTPRequest {
	return &domain.VerifyOTPRequest{
		Email:   req.Email,
		OTP:     req.OTP,
		OTPType: req.OTPType,
	}
}

// LoginRequest represents user login request
// POST /api/auth/login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// ConvertLoginRequestToDomain converts DTO request to domain model
func ConvertLoginRequestToDomain(req *LoginRequest) *domain.LoginRequest {
	return &domain.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}
}

// ForgotPasswordRequest represents forgot password request
// POST /api/auth/forgot-password
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ConvertForgotPasswordRequestToDomain converts DTO request to domain model
func ConvertForgotPasswordRequestToDomain(req *ForgotPasswordRequest) *domain.ForgotPasswordRequest {
	return &domain.ForgotPasswordRequest{
		Email: req.Email,
	}
}

// ResetPasswordRequest represents reset password request
// POST /api/auth/reset-password
type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	OTP         string `json:"otp" binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ConvertResetPasswordRequestToDomain converts DTO request to domain model
func ConvertResetPasswordRequestToDomain(req *ResetPasswordRequest) *domain.ResetPasswordRequest {
	return &domain.ResetPasswordRequest{
		Email:       req.Email,
		OTP:         req.OTP,
		NewPassword: req.NewPassword,
	}
}

// RefreshTokenRequest represents token refresh request
// POST /api/auth/refresh-token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ConvertRefreshTokenRequestToDomain converts DTO request to domain model
func ConvertRefreshTokenRequestToDomain(req *RefreshTokenRequest) *domain.RefreshTokenRequest {
	return &domain.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	}
}
