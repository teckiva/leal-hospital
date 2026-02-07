package email

import "github.com/leal-hospital/server/config"

// EmailSvcDriver defines the interface for email operations
type EmailSvcDriver interface {
	SendOTP(toEmail, toName, otp string) error
	SendWelcomeEmail(toEmail, toName string) error
	SendPasswordResetConfirmation(toEmail, toName string) error
}

// EmailSvc implements EmailSvcDriver interface
type EmailSvc struct {
	Config *config.EmailConfig
}

// NewEmailSvc creates a new email service instance
func NewEmailSvc(cfg *config.EmailConfig) EmailSvcDriver {
	return &EmailSvc{
		Config: cfg,
	}
}
