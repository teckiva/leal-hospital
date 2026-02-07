package otp

import (
	"github.com/leal-hospital/server/config"
	"github.com/leal-hospital/server/models/db"
	"github.com/leal-hospital/server/services/email"
)

// OTPSvcDriver defines the interface for OTP operations
type OTPSvcDriver interface {
	GenerateAndSendOTP(email, name, otpType string) error
	VerifyOTP(email, otp, otpType string) (bool, error)
	CleanupExpiredOTPs() error
}

// OTPSvc implements OTPSvcDriver interface
type OTPSvc struct {
	Queries        *db.Queries
	SecurityConfig *config.SecurityConfig
	EmailSvc       email.EmailSvcDriver
}

// NewOTPSvc creates a new OTP service instance
func NewOTPSvc(queries *db.Queries, securityConfig *config.SecurityConfig, emailSvc email.EmailSvcDriver) OTPSvcDriver {
	return &OTPSvc{
		Queries:        queries,
		SecurityConfig: securityConfig,
		EmailSvc:       emailSvc,
	}
}
