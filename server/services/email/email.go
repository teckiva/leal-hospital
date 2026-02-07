package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"github.com/leal-hospital/server/logger"
	"github.com/leal-hospital/server/utils"
)

// SendOTP sends an OTP email to the specified recipient
func (e *EmailSvc) SendOTP(toEmail, toName, otp string) error {
	functionName := "EmailSvc.SendOTP"

	logger.Info(functionName, "Sending OTP email to:", toEmail)

	emailContent := utils.GenerateOTPEmail(toName, otp)

	if err := e.sendEmail(toEmail, emailContent.Subject, emailContent.Body); err != nil {
		logger.Error(functionName, "Failed to send OTP email:", err)
		return fmt.Errorf("failed to send OTP email: %w", err)
	}

	logger.Info(functionName, "OTP email sent successfully to:", toEmail)
	return nil
}

// SendWelcomeEmail sends a welcome email to new users
func (e *EmailSvc) SendWelcomeEmail(toEmail, toName string) error {
	functionName := "EmailSvc.SendWelcomeEmail"

	logger.Info(functionName, "Sending welcome email to:", toEmail)

	emailContent := utils.GenerateWelcomeEmail(toName)

	if err := e.sendEmail(toEmail, emailContent.Subject, emailContent.Body); err != nil {
		logger.Error(functionName, "Failed to send welcome email:", err)
		return fmt.Errorf("failed to send welcome email: %w", err)
	}

	logger.Info(functionName, "Welcome email sent successfully to:", toEmail)
	return nil
}

// SendPasswordResetConfirmation sends a password reset confirmation email
func (e *EmailSvc) SendPasswordResetConfirmation(toEmail, toName string) error {
	functionName := "EmailSvc.SendPasswordResetConfirmation"

	logger.Info(functionName, "Sending password reset confirmation to:", toEmail)

	emailContent := utils.GeneratePasswordResetConfirmation(toName)

	if err := e.sendEmail(toEmail, emailContent.Subject, emailContent.Body); err != nil {
		logger.Error(functionName, "Failed to send password reset confirmation:", err)
		return fmt.Errorf("failed to send password reset confirmation: %w", err)
	}

	logger.Info(functionName, "Password reset confirmation sent successfully to:", toEmail)
	return nil
}

// sendEmail sends an email using SMTP with STARTTLS
func (e *EmailSvc) sendEmail(to, subject, body string) error {
	// Setup authentication
	auth := smtp.PlainAuth(
		"",
		e.Config.SMTPUsername,
		e.Config.SMTPPassword,
		e.Config.SMTPHost,
	)

	// Compose message
	from := fmt.Sprintf("%s <%s>", e.Config.FromName, e.Config.FromEmail)
	message := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s",
		from, to, subject, body,
	)

	// Send email
	addr := fmt.Sprintf("%s:%d", e.Config.SMTPHost, e.Config.SMTPPort)
	err := smtp.SendMail(
		addr,
		auth,
		e.Config.FromEmail,
		[]string{to},
		[]byte(message),
	)

	if err != nil {
		// Try with explicit TLS if STARTTLS fails
		return e.sendEmailTLS(to, subject, body)
	}

	return nil
}

// sendEmailTLS sends an email using SMTP with explicit TLS
func (e *EmailSvc) sendEmailTLS(to, subject, body string) error {
	// Setup authentication
	auth := smtp.PlainAuth(
		"",
		e.Config.SMTPUsername,
		e.Config.SMTPPassword,
		e.Config.SMTPHost,
	)

	// Compose message
	from := fmt.Sprintf("%s <%s>", e.Config.FromName, e.Config.FromEmail)
	message := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s",
		from, to, subject, body,
	)

	// Setup TLS configuration
	tlsConfig := &tls.Config{
		ServerName: e.Config.SMTPHost,
	}

	// Connect to SMTP server with TLS
	addr := fmt.Sprintf("%s:%d", e.Config.SMTPHost, e.Config.SMTPPort)
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect with TLS: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, e.Config.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Authenticate
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Set sender and recipient
	if err = client.Mail(e.Config.FromEmail); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send message body
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = writer.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	// Send quit command
	return client.Quit()
}
