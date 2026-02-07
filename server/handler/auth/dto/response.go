package dto

import "github.com/leal-hospital/server/domain"

// RegisterResponse represents user registration response
type RegisterResponse struct {
	Message string `json:"message"`
	UserID  int64  `json:"user_id,omitempty"`
}

// ConvertRegisterDomainToResponse converts domain response to DTO response
func ConvertRegisterDomainToResponse(dom *domain.RegisterResponse) *RegisterResponse {
	return &RegisterResponse{
		Message: dom.Message,
		UserID:  dom.UserID,
	}
}

// VerifyOTPResponse represents OTP verification response
type VerifyOTPResponse struct {
	Message      string `json:"message"`
	IsVerified   bool   `json:"is_verified"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// ConvertVerifyOTPDomainToResponse converts domain response to DTO response
func ConvertVerifyOTPDomainToResponse(dom *domain.VerifyOTPResponse) *VerifyOTPResponse {
	return &VerifyOTPResponse{
		Message:      dom.Message,
		IsVerified:   dom.IsVerified,
		AccessToken:  dom.AccessToken,
		RefreshToken: dom.RefreshToken,
	}
}

// LoginResponse represents user login response (OTP sent)
type LoginResponse struct {
	Message string `json:"message"`
}

// ConvertLoginDomainToResponse converts domain response to DTO response
func ConvertLoginDomainToResponse(dom *domain.LoginResponse) *LoginResponse {
	return &LoginResponse{
		Message: dom.Message,
	}
}

// ForgotPasswordResponse represents forgot password response
type ForgotPasswordResponse struct {
	Message string `json:"message"`
}

// ConvertForgotPasswordDomainToResponse converts domain response to DTO response
func ConvertForgotPasswordDomainToResponse(dom *domain.ForgotPasswordResponse) *ForgotPasswordResponse {
	return &ForgotPasswordResponse{
		Message: dom.Message,
	}
}

// ResetPasswordResponse represents reset password response
type ResetPasswordResponse struct {
	Message string `json:"message"`
}

// ConvertResetPasswordDomainToResponse converts domain response to DTO response
func ConvertResetPasswordDomainToResponse(dom *domain.ResetPasswordResponse) *ResetPasswordResponse {
	return &ResetPasswordResponse{
		Message: dom.Message,
	}
}

// RefreshTokenResponse represents token refresh response
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// ConvertRefreshTokenDomainToResponse converts domain response to DTO response
func ConvertRefreshTokenDomainToResponse(dom *domain.RefreshTokenResponse) *RefreshTokenResponse {
	return &RefreshTokenResponse{
		AccessToken:  dom.AccessToken,
		RefreshToken: dom.RefreshToken,
	}
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error string `json:"error"`
}
