# Authentication Module Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build complete authentication system with email OTP verification, JWT tokens, and bcrypt password hashing for Lael Hospital management system.

**Architecture:** Stateless JWT-based auth with email OTP verification. Services layer for email/OTP/password/JWT operations, handlers for API endpoints, middleware for protected routes. Admin registers with OTP verification, staff registers with admin approval workflow. Login supports OTP (admin/staff) and password (staff only).

**Tech Stack:** Go 1.23, Gin, JWT (golang-jwt/jwt/v5), bcrypt, SMTP email, SQLC, MySQL

---

## Task 1: Update Configuration for Auth Services

**Files:**
- Modify: `server/config/config.go`
- Modify: `server/config/.env`

**Step 1: Add email configuration to config struct**

In `server/config/config.go`, add after `LoggingConfig`:

```go
// EmailConfig contains SMTP email configuration
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
}
```

**Step 2: Update AppConfig to include EmailConfig**

In `server/config/config.go`, add to `AppConfig` struct:

```go
type AppConfig struct {
	Environment string
	Server      ServerConfig
	Database    DBConfig
	Security    SecurityConfig
	Logging     LoggingConfig
	Email       EmailConfig  // Add this
}
```

**Step 3: Add email configuration loading**

In `LoadConfig()` function, after `Logging` configuration block:

```go
		Email: EmailConfig{
			SMTPHost:     viper.GetString("SMTP_HOST"),
			SMTPPort:     viper.GetInt("SMTP_PORT"),
			SMTPUsername: viper.GetString("SMTP_USERNAME"),
			SMTPPassword: viper.GetString("SMTP_PASSWORD"),
			FromEmail:    viper.GetString("SMTP_FROM_EMAIL"),
			FromName:     viper.GetString("SMTP_FROM_NAME"),
		},
```

**Step 4: Add email defaults**

In `setDefaults()` function, after logging defaults:

```go
	// Email defaults
	viper.SetDefault("SMTP_HOST", "smtp.gmail.com")
	viper.SetDefault("SMTP_PORT", 587)
	viper.SetDefault("SMTP_USERNAME", "")
	viper.SetDefault("SMTP_PASSWORD", "")
	viper.SetDefault("SMTP_FROM_EMAIL", "noreply@laelhospital.com")
	viper.SetDefault("SMTP_FROM_NAME", "Lael Hospital")
```

**Step 5: Add email config to .env**

In `server/config/.env`, add after logging configuration:

```env
# Email Configuration (Gmail SMTP)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM_EMAIL=noreply@laelhospital.com
SMTP_FROM_NAME=Lael Hospital
```

**Step 6: Commit configuration changes**

```bash
git add server/config/config.go server/config/.env
git commit -m "feat(config): add email SMTP configuration"
```

---

## Task 2: Create Email Service

**Files:**
- Create: `server/services/email/email.go`
- Create: `server/services/email/templates.go`

**Step 1: Create email service interface and implementation**

Create `server/services/email/email.go`:

```go
package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"github.com/leal-hospital/server/config"
	"github.com/leal-hospital/server/logger"
)

// EmailService handles email sending operations
type EmailService interface {
	SendOTP(toEmail, toName, otp string) error
	SendWelcomeEmail(toEmail, toName string) error
	SendPasswordResetConfirmation(toEmail, toName string) error
}

type emailService struct {
	config *config.EmailConfig
}

// NewEmailService creates a new email service
func NewEmailService(cfg *config.EmailConfig) EmailService {
	return &emailService{
		config: cfg,
	}
}

// SendOTP sends OTP verification email
func (s *emailService) SendOTP(toEmail, toName, otp string) error {
	subject := "Your OTP for Lael Hospital"
	body := GenerateOTPEmailBody(toName, otp)
	return s.sendEmail(toEmail, subject, body)
}

// SendWelcomeEmail sends welcome email after successful registration
func (s *emailService) SendWelcomeEmail(toEmail, toName string) error {
	subject := "Welcome to Lael Hospital"
	body := GenerateWelcomeEmailBody(toName)
	return s.sendEmail(toEmail, subject, body)
}

// SendPasswordResetConfirmation sends password reset confirmation
func (s *emailService) SendPasswordResetConfirmation(toEmail, toName string) error {
	subject := "Password Reset Successful"
	body := GeneratePasswordResetConfirmationBody(toName)
	return s.sendEmail(toEmail, subject, body)
}

// sendEmail sends an email using SMTP
func (s *emailService) sendEmail(to, subject, body string) error {
	from := s.config.FromEmail

	// Set up authentication
	auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)

	// Compose message
	message := []byte(fmt.Sprintf(
		"From: %s <%s>\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s\r\n",
		s.config.FromName, from, to, subject, body,
	))

	// Connect to SMTP server with TLS
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	// Send email
	err := smtp.SendMail(addr, auth, from, []string{to}, message)
	if err != nil {
		logger.Error("Failed to send email", "to", to, "error", err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	logger.Info("Email sent successfully", "to", to, "subject", subject)
	return nil
}

// SendEmailTLS sends email using TLS (alternative method if needed)
func (s *emailService) sendEmailTLS(to, subject, body string) error {
	from := s.config.FromEmail
	auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         s.config.SMTPHost,
	}

	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	// Connect
	conn, err := tls.Dial("tcp", addr, tlsconfig)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.config.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	// Auth
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	// Set sender and recipient
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send message
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	message := fmt.Sprintf(
		"From: %s <%s>\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s\r\n",
		s.config.FromName, from, to, subject, body,
	)

	_, err = w.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	client.Quit()

	logger.Info("Email sent successfully via TLS", "to", to, "subject", subject)
	return nil
}
```

**Step 2: Create email templates**

Create `server/services/email/templates.go`:

```go
package email

import "fmt"

// GenerateOTPEmailBody generates HTML email body for OTP
func GenerateOTPEmailBody(name, otp string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #5B2C91 0%%, #7B3FA6 100%%); color: white; padding: 20px; text-align: center; }
        .content { background: #f9f9f9; padding: 30px; margin-top: 20px; }
        .otp-box { background: white; border: 2px solid #5B2C91; padding: 20px; text-align: center; font-size: 32px; font-weight: bold; letter-spacing: 5px; color: #5B2C91; margin: 20px 0; }
        .footer { text-align: center; margin-top: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Lael Hospital</h1>
        </div>
        <div class="content">
            <h2>Hello %s,</h2>
            <p>Your One-Time Password (OTP) for verification is:</p>
            <div class="otp-box">%s</div>
            <p><strong>This OTP is valid for 5 minutes only.</strong></p>
            <p>If you did not request this OTP, please ignore this email.</p>
        </div>
        <div class="footer">
            <p>© 2026 Lael Hospital. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, name, otp)
}

// GenerateWelcomeEmailBody generates HTML email body for welcome message
func GenerateWelcomeEmailBody(name string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #5B2C91 0%%, #7B3FA6 100%%); color: white; padding: 20px; text-align: center; }
        .content { background: #f9f9f9; padding: 30px; margin-top: 20px; }
        .footer { text-align: center; margin-top: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to Lael Hospital</h1>
        </div>
        <div class="content">
            <h2>Hello %s,</h2>
            <p>Welcome to Lael Hospital Management System! Your account has been successfully created.</p>
            <p>You can now log in and start using our services.</p>
            <p>If you have any questions, please don't hesitate to contact us.</p>
        </div>
        <div class="footer">
            <p>© 2026 Lael Hospital. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, name)
}

// GeneratePasswordResetConfirmationBody generates HTML email body for password reset confirmation
func GeneratePasswordResetConfirmationBody(name string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #5B2C91 0%%, #7B3FA6 100%%); color: white; padding: 20px; text-align: center; }
        .content { background: #f9f9f9; padding: 30px; margin-top: 20px; }
        .footer { text-align: center; margin-top: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Lael Hospital</h1>
        </div>
        <div class="content">
            <h2>Hello %s,</h2>
            <p>Your password has been successfully reset.</p>
            <p>If you did not make this change, please contact us immediately.</p>
        </div>
        <div class="footer">
            <p>© 2026 Lael Hospital. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, name)
}
```

**Step 3: Commit email service**

```bash
git add server/services/email/
git commit -m "feat(email): add email service with SMTP and templates"
```

---

## Task 3: Create OTP Service

**Files:**
- Create: `server/services/otp/otp.go`

**Step 1: Create OTP service**

Create `server/services/otp/otp.go`:

```go
package otp

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/leal-hospital/server/logger"
	"github.com/leal-hospital/server/medierror"
	"github.com/leal-hospital/server/models/db"
	"github.com/leal-hospital/server/services/email"
	"github.com/leal-hospital/server/utils/dbutils"
)

// OTPService handles OTP generation and validation
type OTPService interface {
	GenerateAndSendOTP(mobile, email, name, otpType string) error
	ValidateOTP(mobile, email, otp, otpType string) error
}

type otpService struct {
	queries      *db.Queries
	emailService email.EmailService
	otpLength    int
	otpExpiry    time.Duration
	maxRetries   int
}

// NewOTPService creates a new OTP service
func NewOTPService(queries *db.Queries, emailService email.EmailService, otpLength int, otpExpiry time.Duration, maxRetries int) OTPService {
	return &otpService{
		queries:      queries,
		emailService: emailService,
		otpLength:    otpLength,
		otpExpiry:    otpExpiry,
		maxRetries:   maxRetries,
	}
}

// GenerateAndSendOTP generates a new OTP and sends it via email
func (s *otpService) GenerateAndSendOTP(mobile, email, name, otpType string) error {
	// Generate OTP
	otp, err := s.generateOTP()
	if err != nil {
		logger.Error("Failed to generate OTP", "error", err)
		return medierror.NewFromRegistry("1004", nil) // Internal server error
	}

	// Calculate expiry
	expiry := time.Now().Add(s.otpExpiry)

	// Store OTP in database
	_, err = s.queries.CreateOTP(dbutils.GetDBContext(), db.CreateOTPParams{
		Mobile:  mobile,
		Email:   email,
		Otp:     otp,
		Expiry:  expiry,
		OtpType: db.LaelOtpOtpType(otpType),
	})
	if err != nil {
		logger.Error("Failed to create OTP in database", "error", err, "mobile", mobile, "email", email)
		return medierror.NewFromRegistry("1004", nil)
	}

	// Send OTP via email
	err = s.emailService.SendOTP(email, name, otp)
	if err != nil {
		logger.Error("Failed to send OTP email", "error", err, "email", email)
		return medierror.NewFromRegistry("1004", nil)
	}

	logger.Info("OTP generated and sent", "mobile", mobile, "email", email, "type", otpType)
	return nil
}

// ValidateOTP validates the provided OTP
func (s *otpService) ValidateOTP(mobile, email, otp, otpType string) error {
	// Fetch latest OTP for this email
	otpRecord, err := s.queries.GetLatestOTPByEmail(dbutils.GetDBContext(), db.GetLatestOTPByEmailParams{
		Email:   email,
		OtpType: db.LaelOtpOtpType(otpType),
	})
	if err != nil {
		logger.Error("OTP not found", "email", email, "type", otpType)
		return medierror.NewFromRegistry("2001", nil) // Invalid OTP
	}

	// Check if OTP is expired
	if time.Now().After(otpRecord.Expiry) {
		logger.Warn("OTP expired", "email", email, "expiry", otpRecord.Expiry)
		return medierror.NewFromRegistry("2002", nil) // OTP expired
	}

	// Check if OTP is already validated
	if otpRecord.IsValidated == 1 {
		logger.Warn("OTP already used", "email", email)
		return medierror.NewFromRegistry("2003", nil) // OTP already used
	}

	// Check retry count
	if otpRecord.RetryCount >= int32(s.maxRetries) {
		logger.Warn("Maximum OTP retry attempts exceeded", "email", email, "retries", otpRecord.RetryCount)
		return medierror.NewFromRegistry("2004", nil) // Max retries exceeded
	}

	// Validate OTP
	if otpRecord.Otp != otp {
		// Increment retry count
		err = s.queries.IncrementRetryCount(dbutils.GetDBContext(), otpRecord.ID)
		if err != nil {
			logger.Error("Failed to increment retry count", "error", err)
		}
		logger.Warn("Invalid OTP provided", "email", email, "retries", otpRecord.RetryCount+1)
		return medierror.NewFromRegistry("2001", nil) // Invalid OTP
	}

	// Mark OTP as validated
	err = s.queries.ValidateOTP(dbutils.GetDBContext(), otpRecord.ID)
	if err != nil {
		logger.Error("Failed to mark OTP as validated", "error", err)
		return medierror.NewFromRegistry("1004", nil)
	}

	logger.Info("OTP validated successfully", "email", email, "type", otpType)
	return nil
}

// generateOTP generates a random OTP of specified length
func (s *otpService) generateOTP() (string, error) {
	const digits = "0123456789"
	otp := make([]byte, s.otpLength)

	for i := range otp {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		otp[i] = digits[num.Int64()]
	}

	return string(otp), nil
}
```

**Step 2: Add OTP error codes to errors.yaml**

In `server/config/errors.yaml`, add after existing errors:

```yaml
"2001":
  message: Invalid OTP
  display_message: The OTP you entered is incorrect. Please try again.
  decline_type: "IA"
  source: AUTH

"2002":
  message: OTP expired
  display_message: Your OTP has expired. Please request a new one.
  decline_type: "EX"
  source: AUTH

"2003":
  message: OTP already used
  display_message: This OTP has already been used. Please request a new one.
  decline_type: "AU"
  source: AUTH

"2004":
  message: Maximum OTP retry attempts exceeded
  display_message: You have exceeded the maximum number of OTP attempts. Please request a new OTP.
  decline_type: "MR"
  source: AUTH
```

**Step 3: Create dbutils helper**

Create `server/utils/dbutils/context.go`:

```go
package dbutils

import "context"

// GetDBContext returns a context for database operations
func GetDBContext() context.Context {
	return context.Background()
}
```

**Step 4: Commit OTP service**

```bash
git add server/services/otp/ server/config/errors.yaml server/utils/dbutils/
git commit -m "feat(otp): add OTP service with generation and validation"
```

---

## Task 4: Create Password Service

**Files:**
- Create: `server/services/password/password.go`

**Step 1: Install bcrypt dependency**

```bash
cd server
go get golang.org/x/crypto/bcrypt
```

**Step 2: Create password service**

Create `server/services/password/password.go`:

```go
package password

import (
	"fmt"
	"regexp"

	"golang.org/x/crypto/bcrypt"

	"github.com/leal-hospital/server/logger"
	"github.com/leal-hospital/server/medierror"
)

const (
	// bcrypt cost factor (10-12 is recommended)
	bcryptCost = 12
	// Minimum password length
	minPasswordLength = 8
)

// PasswordService handles password hashing and validation
type PasswordService interface {
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, password string) error
	ValidatePasswordStrength(password string) error
}

type passwordService struct{}

// NewPasswordService creates a new password service
func NewPasswordService() PasswordService {
	return &passwordService{}
}

// HashPassword hashes a plain text password using bcrypt
func (s *passwordService) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		logger.Error("Failed to hash password", "error", err)
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// VerifyPassword verifies a plain text password against a hashed password
func (s *passwordService) VerifyPassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			logger.Warn("Password verification failed - incorrect password")
			return medierror.NewFromRegistry("2005", nil) // Invalid password
		}
		logger.Error("Password verification error", "error", err)
		return medierror.NewFromRegistry("1004", nil)
	}
	return nil
}

// ValidatePasswordStrength validates password meets security requirements
// Requirements: Minimum 8 characters, at least one letter, one number, one special character
func (s *passwordService) ValidatePasswordStrength(password string) error {
	if len(password) < minPasswordLength {
		return medierror.NewFromRegistry("2006", nil) // Password too short
	}

	// Check for at least one letter
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	if !hasLetter {
		return medierror.NewFromRegistry("2007", nil) // Password must contain letter
	}

	// Check for at least one digit
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasDigit {
		return medierror.NewFromRegistry("2008", nil) // Password must contain digit
	}

	// Check for at least one special character
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)
	if !hasSpecial {
		return medierror.NewFromRegistry("2009", nil) // Password must contain special char
	}

	return nil
}
```

**Step 3: Add password error codes to errors.yaml**

In `server/config/errors.yaml`, add:

```yaml
"2005":
  message: Invalid password
  display_message: The password you entered is incorrect.
  decline_type: "IP"
  source: AUTH

"2006":
  message: Password too short
  display_message: Password must be at least 8 characters long.
  decline_type: "PS"
  source: AUTH

"2007":
  message: Password must contain letters
  display_message: Password must contain at least one letter.
  decline_type: "PL"
  source: AUTH

"2008":
  message: Password must contain digits
  display_message: Password must contain at least one number.
  decline_type: "PD"
  source: AUTH

"2009":
  message: Password must contain special characters
  display_message: Password must contain at least one special character (!@#$%^&*).
  decline_type: "PC"
  source: AUTH
```

**Step 4: Commit password service**

```bash
git add server/services/password/ server/config/errors.yaml go.mod go.sum
git commit -m "feat(password): add password service with bcrypt hashing"
```

---

## Task 5: Create JWT Service

**Files:**
- Create: `server/services/jwt/jwt.go`

**Step 1: Install JWT dependency**

```bash
cd server
go get github.com/golang-jwt/jwt/v5
```

**Step 2: Create JWT service**

Create `server/services/jwt/jwt.go`:

```go
package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/leal-hospital/server/logger"
	"github.com/leal-hospital/server/medierror"
)

// Claims represents JWT claims
type Claims struct {
	UserID      int64  `json:"user_id"`
	Email       string `json:"email"`
	Mobile      string `json:"mobile"`
	Name        string `json:"name"`
	IsAdmin     bool   `json:"is_admin"`
	Designation string `json:"designation"`
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// JWTService handles JWT token operations
type JWTService interface {
	GenerateTokenPair(userID int64, email, mobile, name, designation string, isAdmin bool) (*TokenPair, error)
	ValidateAccessToken(tokenString string) (*Claims, error)
	ValidateRefreshToken(tokenString string) (*Claims, error)
	RefreshAccessToken(refreshTokenString string) (string, error)
}

type jwtService struct {
	secretKey              []byte
	accessTokenExpiration  time.Duration
	refreshTokenExpiration time.Duration
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string, accessExpiration, refreshExpiration time.Duration) JWTService {
	return &jwtService{
		secretKey:              []byte(secretKey),
		accessTokenExpiration:  accessExpiration,
		refreshTokenExpiration: refreshExpiration,
	}
}

// GenerateTokenPair generates access and refresh tokens
func (s *jwtService) GenerateTokenPair(userID int64, email, mobile, name, designation string, isAdmin bool) (*TokenPair, error) {
	now := time.Now()
	accessExpiry := now.Add(s.accessTokenExpiration)
	refreshExpiry := now.Add(s.refreshTokenExpiration)

	// Generate access token
	accessClaims := Claims{
		UserID:      userID,
		Email:       email,
		Mobile:      mobile,
		Name:        name,
		IsAdmin:     isAdmin,
		Designation: designation,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "lael-hospital",
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(s.secretKey)
	if err != nil {
		logger.Error("Failed to sign access token", "error", err)
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	refreshClaims := Claims{
		UserID:      userID,
		Email:       email,
		Mobile:      mobile,
		Name:        name,
		IsAdmin:     isAdmin,
		Designation: designation,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "lael-hospital",
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(s.secretKey)
	if err != nil {
		logger.Error("Failed to sign refresh token", "error", err)
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	logger.Info("Token pair generated", "user_id", userID, "email", email)

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    accessExpiry,
	}, nil
}

// ValidateAccessToken validates an access token and returns claims
func (s *jwtService) ValidateAccessToken(tokenString string) (*Claims, error) {
	return s.validateToken(tokenString)
}

// ValidateRefreshToken validates a refresh token and returns claims
func (s *jwtService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return s.validateToken(tokenString)
}

// RefreshAccessToken generates a new access token from a valid refresh token
func (s *jwtService) RefreshAccessToken(refreshTokenString string) (string, error) {
	// Validate refresh token
	claims, err := s.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", err
	}

	// Generate new access token
	now := time.Now()
	accessExpiry := now.Add(s.accessTokenExpiration)

	newClaims := Claims{
		UserID:      claims.UserID,
		Email:       claims.Email,
		Mobile:      claims.Mobile,
		Name:        claims.Name,
		IsAdmin:     claims.IsAdmin,
		Designation: claims.Designation,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "lael-hospital",
			Subject:   fmt.Sprintf("%d", claims.UserID),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	accessTokenString, err := accessToken.SignedString(s.secretKey)
	if err != nil {
		logger.Error("Failed to sign new access token", "error", err)
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	logger.Info("Access token refreshed", "user_id", claims.UserID, "email", claims.Email)
	return accessTokenString, nil
}

// validateToken validates a JWT token and returns claims
func (s *jwtService) validateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		logger.Warn("Token validation failed", "error", err)
		return nil, medierror.NewFromRegistry("2010", nil) // Invalid token
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		logger.Warn("Invalid token claims")
		return nil, medierror.NewFromRegistry("2010", nil)
	}

	// Check expiration
	if claims.ExpiresAt.Before(time.Now()) {
		logger.Warn("Token expired", "user_id", claims.UserID)
		return nil, medierror.NewFromRegistry("2011", nil) // Token expired
	}

	return claims, nil
}
```

**Step 3: Add JWT error codes to errors.yaml**

In `server/config/errors.yaml`, add:

```yaml
"2010":
  message: Invalid token
  display_message: Your session token is invalid. Please log in again.
  decline_type: "IT"
  source: AUTH

"2011":
  message: Token expired
  display_message: Your session has expired. Please log in again.
  decline_type: "TE"
  source: AUTH
```

**Step 4: Commit JWT service**

```bash
git add server/services/jwt/ server/config/errors.yaml go.mod go.sum
git commit -m "feat(jwt): add JWT service with token generation and validation"
```

---

## Task 6: Create Auth Middleware

**Files:**
- Create: `server/middleware/auth.go`

**Step 1: Create auth middleware**

Create `server/middleware/auth.go`:

```go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/leal-hospital/server/logger"
	"github.com/leal-hospital/server/medierror"
	"github.com/leal-hospital/server/services/jwt"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(jwtService jwt.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("Missing authorization header")
			errResp := medierror.NewFromRegistry("2012", nil) // Unauthorized
			c.JSON(http.StatusUnauthorized, errResp)
			c.Abort()
			return
		}

		// Check Bearer format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Warn("Invalid authorization header format")
			errResp := medierror.NewFromRegistry("2012", nil)
			c.JSON(http.StatusUnauthorized, errResp)
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			if appErr, ok := err.(*medierror.AppError); ok {
				c.JSON(http.StatusUnauthorized, appErr)
			} else {
				errResp := medierror.NewFromRegistry("2010", nil)
				c.JSON(http.StatusUnauthorized, errResp)
			}
			c.Abort()
			return
		}

		// Store claims in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("mobile", claims.Mobile)
		c.Set("name", claims.Name)
		c.Set("is_admin", claims.IsAdmin)
		c.Set("designation", claims.Designation)

		logger.Info("User authenticated", "user_id", claims.UserID, "email", claims.Email)

		c.Next()
	}
}

// AdminOnlyMiddleware ensures only admins can access the route
func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists || !isAdmin.(bool) {
			logger.Warn("Non-admin attempted to access admin route", "user_id", c.GetInt64("user_id"))
			errResp := medierror.NewFromRegistry("2013", nil) // Forbidden
			c.JSON(http.StatusForbidden, errResp)
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetUserIDFromContext retrieves user ID from context
func GetUserIDFromContext(c *gin.Context) int64 {
	userID, _ := c.Get("user_id")
	return userID.(int64)
}

// GetUserEmailFromContext retrieves user email from context
func GetUserEmailFromContext(c *gin.Context) string {
	email, _ := c.Get("email")
	return email.(string)
}

// IsAdminFromContext checks if user is admin from context
func IsAdminFromContext(c *gin.Context) bool {
	isAdmin, _ := c.Get("is_admin")
	return isAdmin.(bool)
}
```

**Step 2: Add auth error codes to errors.yaml**

In `server/config/errors.yaml`, add:

```yaml
"2012":
  message: Unauthorized
  display_message: You must be logged in to access this resource.
  decline_type: "UA"
  source: AUTH

"2013":
  message: Forbidden
  display_message: You do not have permission to access this resource.
  decline_type: "FB"
  source: AUTH
```

**Step 3: Commit auth middleware**

```bash
git add server/middleware/auth.go server/config/errors.yaml
git commit -m "feat(middleware): add JWT auth middleware with admin guard"
```

---

## Task 7: Create Auth Module - Registration Handler

**Files:**
- Create: `server/modules/auth/handler.go`
- Create: `server/modules/auth/dto.go`
- Create: `server/modules/auth/module.go`

**Step 1: Create DTOs**

Create `server/modules/auth/dto.go`:

```go
package auth

// RegisterAdminRequest represents admin registration request
type RegisterAdminRequest struct {
	Name     string `json:"name" binding:"required"`
	Mobile   string `json:"mobile" binding:"required,len=10"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// RegisterStaffRequest represents staff registration request
type RegisterStaffRequest struct {
	Name        string `json:"name" binding:"required"`
	Mobile      string `json:"mobile" binding:"required,len=10"`
	Email       string `json:"email" binding:"required,email"`
	Designation string `json:"designation" binding:"required,oneof=doctor nurse staff"`
	Password    string `json:"password" binding:"required,min=8"`
}

// SendOTPRequest represents OTP sending request
type SendOTPRequest struct {
	Mobile string `json:"mobile" binding:"required,len=10"`
	Email  string `json:"email" binding:"required,email"`
}

// VerifyOTPRequest represents OTP verification request
type VerifyOTPRequest struct {
	Mobile  string `json:"mobile" binding:"required,len=10"`
	Email   string `json:"email" binding:"required,email"`
	OTP     string `json:"otp" binding:"required,len=6"`
	OTPType string `json:"otp_type" binding:"required,oneof=registration login forgot_password"`
}

// LoginWithOTPRequest represents login with OTP request
type LoginWithOTPRequest struct {
	Mobile string `json:"mobile" binding:"required,len=10"`
	Email  string `json:"email" binding:"required,email"`
	OTP    string `json:"otp" binding:"required,len=6"`
}

// LoginWithPasswordRequest represents login with password request
type LoginWithPasswordRequest struct {
	Mobile   string `json:"mobile" binding:"required,len=10"`
	Password string `json:"password" binding:"required"`
}

// ForgotPasswordRequest represents forgot password request
type ForgotPasswordRequest struct {
	Mobile string `json:"mobile" binding:"required,len=10"`
	Email  string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents password reset request
type ResetPasswordRequest struct {
	Mobile      string `json:"mobile" binding:"required,len=10"`
	Email       string `json:"email" binding:"required,email"`
	OTP         string `json:"otp" binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// RefreshTokenRequest represents token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	User struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		Mobile      string `json:"mobile"`
		Email       string `json:"email"`
		Designation string `json:"designation"`
		IsAdmin     bool   `json:"is_admin"`
	} `json:"user"`
	Tokens struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresAt    string `json:"expires_at"`
	} `json:"tokens"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
}
```

**Step 2: Commit DTOs**

```bash
git add server/modules/auth/dto.go
git commit -m "feat(auth): add auth DTOs for request/response"
```

---

## Task 8: Create Auth Handler - Registration

**Step 1: Create handler struct and constructor**

Create `server/modules/auth/handler.go`:

```go
package auth

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/leal-hospital/server/logger"
	"github.com/leal-hospital/server/medierror"
	"github.com/leal-hospital/server/models/db"
	jwtService "github.com/leal-hospital/server/services/jwt"
	"github.com/leal-hospital/server/services/otp"
	"github.com/leal-hospital/server/services/password"
	"github.com/leal-hospital/server/utils/dbutils"
)

// Handler handles auth-related HTTP requests
type Handler struct {
	queries         *db.Queries
	otpService      otp.OTPService
	passwordService password.PasswordService
	jwtService      jwtService.JWTService
}

// NewHandler creates a new auth handler
func NewHandler(
	queries *db.Queries,
	otpService otp.OTPService,
	passwordService password.PasswordService,
	jwtService jwtService.JWTService,
) *Handler {
	return &Handler{
		queries:         queries,
		otpService:      otpService,
		passwordService: passwordService,
		jwtService:      jwtService,
	}
}

// RegisterAdmin handles admin registration
func (h *Handler) RegisterAdmin(c *gin.Context) {
	var req RegisterAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid registration request", "error", err)
		errResp := medierror.NewFromRegistry("1000", nil)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	// Validate password strength
	if err := h.passwordService.ValidatePasswordStrength(req.Password); err != nil {
		if appErr, ok := err.(*medierror.AppError); ok {
			c.JSON(http.StatusBadRequest, appErr)
		} else {
			errResp := medierror.NewFromRegistry("1004", nil)
			c.JSON(http.StatusInternalServerError, errResp)
		}
		return
	}

	// Check if mobile already exists
	_, err := h.queries.GetUserByMobile(dbutils.GetDBContext(), req.Mobile)
	if err == nil {
		logger.Warn("Mobile already registered", "mobile", req.Mobile)
		errResp := medierror.NewFromRegistry("2014", nil)
		c.JSON(http.StatusConflict, errResp)
		return
	}

	// Check if email already exists
	_, err = h.queries.GetUserByEmail(dbutils.GetDBContext(), req.Email)
	if err == nil {
		logger.Warn("Email already registered", "email", req.Email)
		errResp := medierror.NewFromRegistry("2015", nil)
		c.JSON(http.StatusConflict, errResp)
		return
	}

	// Hash password
	hashedPassword, err := h.passwordService.HashPassword(req.Password)
	if err != nil {
		errResp := medierror.NewFromRegistry("1004", nil)
		c.JSON(http.StatusInternalServerError, errResp)
		return
	}

	// Create user
	result, err := h.queries.CreateUser(dbutils.GetDBContext(), db.CreateUserParams{
		Name:         req.Name,
		Mobile:       req.Mobile,
		Email:        req.Email,
		Designation:  "doctor", // Admin is always a doctor
		IsAdmin:      1,
		IsApproved:   1, // Admin is auto-approved
		PasswordHash: sql.NullString{String: hashedPassword, Valid: true},
	})
	if err != nil {
		logger.Error("Failed to create admin user", "error", err)
		errResp := medierror.NewFromRegistry("1004", nil)
		c.JSON(http.StatusInternalServerError, errResp)
		return
	}

	userID, _ := result.LastInsertId()

	// Send OTP for verification
	err = h.otpService.GenerateAndSendOTP(req.Mobile, req.Email, req.Name, "registration")
	if err != nil {
		logger.Error("Failed to send OTP", "error", err)
		// User created but OTP failed - still return success
	}

	logger.Info("Admin registered successfully", "user_id", userID, "email", req.Email)

	c.JSON(http.StatusCreated, MessageResponse{
		Message: "Admin registration successful. Please verify your email with the OTP sent.",
	})
}

// RegisterStaff handles staff registration
func (h *Handler) RegisterStaff(c *gin.Context) {
	var req RegisterStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid registration request", "error", err)
		errResp := medierror.NewFromRegistry("1000", nil)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	// Validate password strength
	if err := h.passwordService.ValidatePasswordStrength(req.Password); err != nil {
		if appErr, ok := err.(*medierror.AppError); ok {
			c.JSON(http.StatusBadRequest, appErr)
		} else {
			errResp := medierror.NewFromRegistry("1004", nil)
			c.JSON(http.StatusInternalServerError, errResp)
		}
		return
	}

	// Check if mobile already exists
	_, err := h.queries.GetUserByMobile(dbutils.GetDBContext(), req.Mobile)
	if err == nil {
		logger.Warn("Mobile already registered", "mobile", req.Mobile)
		errResp := medierror.NewFromRegistry("2014", nil)
		c.JSON(http.StatusConflict, errResp)
		return
	}

	// Check if email already exists
	_, err = h.queries.GetUserByEmail(dbutils.GetDBContext(), req.Email)
	if err == nil {
		logger.Warn("Email already registered", "email", req.Email)
		errResp := medierror.NewFromRegistry("2015", nil)
		c.JSON(http.StatusConflict, errResp)
		return
	}

	// Hash password
	hashedPassword, err := h.passwordService.HashPassword(req.Password)
	if err != nil {
		errResp := medierror.NewFromRegistry("1004", nil)
		c.JSON(http.StatusInternalServerError, errResp)
		return
	}

	// Create user (not approved yet - needs admin approval)
	result, err := h.queries.CreateUser(dbutils.GetDBContext(), db.CreateUserParams{
		Name:         req.Name,
		Mobile:       req.Mobile,
		Email:        req.Email,
		Designation:  db.LaelUsersDesignation(req.Designation),
		IsAdmin:      0,
		IsApproved:   0, // Staff needs admin approval
		PasswordHash: sql.NullString{String: hashedPassword, Valid: true},
	})
	if err != nil {
		logger.Error("Failed to create staff user", "error", err)
		errResp := medierror.NewFromRegistry("1004", nil)
		c.JSON(http.StatusInternalServerError, errResp)
		return
	}

	userID, _ := result.LastInsertId()

	logger.Info("Staff registration submitted", "user_id", userID, "email", req.Email)

	c.JSON(http.StatusCreated, MessageResponse{
		Message: "Registration submitted successfully. Please wait for admin approval.",
	})
}
```

**Step 2: Add registration error codes to errors.yaml**

In `server/config/errors.yaml`, add:

```yaml
"2014":
  message: Mobile number already registered
  display_message: This mobile number is already registered. Please use a different number or log in.
  decline_type: "ME"
  source: AUTH

"2015":
  message: Email already registered
  display_message: This email address is already registered. Please use a different email or log in.
  decline_type: "EE"
  source: AUTH
```

**Step 3: Commit registration handler**

```bash
git add server/modules/auth/handler.go server/config/errors.yaml
git commit -m "feat(auth): add admin and staff registration handlers"
```

---

## Task 9: Create Auth Handler - Login

**Step 1: Add login methods to handler**

In `server/modules/auth/handler.go`, add after `RegisterStaff`:

```go
// LoginWithOTP handles login with OTP (admin and staff)
func (h *Handler) LoginWithOTP(c *gin.Context) {
	var req LoginWithOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid login request", "error", err)
		errResp := medierror.NewFromRegistry("1000", nil)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	// Validate OTP
	err := h.otpService.ValidateOTP(req.Mobile, req.Email, req.OTP, "login")
	if err != nil {
		if appErr, ok := err.(*medierror.AppError); ok {
			c.JSON(http.StatusUnauthorized, appErr)
		} else {
			errResp := medierror.NewFromRegistry("1004", nil)
			c.JSON(http.StatusInternalServerError, errResp)
		}
		return
	}

	// Get user by mobile
	user, err := h.queries.GetUserByMobile(dbutils.GetDBContext(), req.Mobile)
	if err != nil {
		logger.Warn("User not found", "mobile", req.Mobile)
		errResp := medierror.NewFromRegistry("2016", nil)
		c.JSON(http.StatusNotFound, errResp)
		return
	}

	// Check if staff is approved
	if user.IsAdmin == 0 && user.IsApproved == 0 {
		logger.Warn("Staff not approved yet", "user_id", user.ID)
		errResp := medierror.NewFromRegistry("2017", nil)
		c.JSON(http.StatusForbidden, errResp)
		return
	}

	// Check if user is active
	if user.Status != "active" {
		logger.Warn("User account not active", "user_id", user.ID, "status", user.Status)
		errResp := medierror.NewFromRegistry("2018", nil)
		c.JSON(http.StatusForbidden, errResp)
		return
	}

	// Generate JWT tokens
	tokens, err := h.jwtService.GenerateTokenPair(
		user.ID,
		user.Email,
		user.Mobile,
		user.Name,
		string(user.Designation),
		user.IsAdmin == 1,
	)
	if err != nil {
		logger.Error("Failed to generate tokens", "error", err)
		errResp := medierror.NewFromRegistry("1004", nil)
		c.JSON(http.StatusInternalServerError, errResp)
		return
	}

	// Update last login
	err = h.queries.UpdateLastLogin(dbutils.GetDBContext(), user.ID)
	if err != nil {
		logger.Error("Failed to update last login", "error", err)
	}

	// Prepare response
	var response AuthResponse
	response.User.ID = user.ID
	response.User.Name = user.Name
	response.User.Mobile = user.Mobile
	response.User.Email = user.Email
	response.User.Designation = string(user.Designation)
	response.User.IsAdmin = user.IsAdmin == 1
	response.Tokens.AccessToken = tokens.AccessToken
	response.Tokens.RefreshToken = tokens.RefreshToken
	response.Tokens.ExpiresAt = tokens.ExpiresAt.Format(time.RFC3339)

	logger.Info("User logged in successfully", "user_id", user.ID, "email", user.Email)

	c.JSON(http.StatusOK, response)
}

// LoginWithPassword handles login with password (staff only)
func (h *Handler) LoginWithPassword(c *gin.Context) {
	var req LoginWithPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid login request", "error", err)
		errResp := medierror.NewFromRegistry("1000", nil)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	// Get user by mobile
	user, err := h.queries.GetUserByMobile(dbutils.GetDBContext(), req.Mobile)
	if err != nil {
		logger.Warn("User not found", "mobile", req.Mobile)
		errResp := medierror.NewFromRegistry("2016", nil)
		c.JSON(http.StatusNotFound, errResp)
		return
	}

	// Check if password exists
	if !user.PasswordHash.Valid {
		logger.Warn("User has no password set", "user_id", user.ID)
		errResp := medierror.NewFromRegistry("2005", nil)
		c.JSON(http.StatusUnauthorized, errResp)
		return
	}

	// Verify password
	err = h.passwordService.VerifyPassword(user.PasswordHash.String, req.Password)
	if err != nil {
		if appErr, ok := err.(*medierror.AppError); ok {
			c.JSON(http.StatusUnauthorized, appErr)
		} else {
			errResp := medierror.NewFromRegistry("1004", nil)
			c.JSON(http.StatusInternalServerError, errResp)
		}
		return
	}

	// Check if staff is approved
	if user.IsAdmin == 0 && user.IsApproved == 0 {
		logger.Warn("Staff not approved yet", "user_id", user.ID)
		errResp := medierror.NewFromRegistry("2017", nil)
		c.JSON(http.StatusForbidden, errResp)
		return
	}

	// Check if user is active
	if user.Status != "active" {
		logger.Warn("User account not active", "user_id", user.ID, "status", user.Status)
		errResp := medierror.NewFromRegistry("2018", nil)
		c.JSON(http.StatusForbidden, errResp)
		return
	}

	// Generate JWT tokens
	tokens, err := h.jwtService.GenerateTokenPair(
		user.ID,
		user.Email,
		user.Mobile,
		user.Name,
		string(user.Designation),
		user.IsAdmin == 1,
	)
	if err != nil {
		logger.Error("Failed to generate tokens", "error", err)
		errResp := medierror.NewFromRegistry("1004", nil)
		c.JSON(http.StatusInternalServerError, errResp)
		return
	}

	// Update last login
	err = h.queries.UpdateLastLogin(dbutils.GetDBContext(), user.ID)
	if err != nil {
		logger.Error("Failed to update last login", "error", err)
	}

	// Prepare response
	var response AuthResponse
	response.User.ID = user.ID
	response.User.Name = user.Name
	response.User.Mobile = user.Mobile
	response.User.Email = user.Email
	response.User.Designation = string(user.Designation)
	response.User.IsAdmin = user.IsAdmin == 1
	response.Tokens.AccessToken = tokens.AccessToken
	response.Tokens.RefreshToken = tokens.RefreshToken
	response.Tokens.ExpiresAt = tokens.ExpiresAt.Format(time.RFC3339)

	logger.Info("User logged in successfully", "user_id", user.ID, "email", user.Email)

	c.JSON(http.StatusOK, response)
}

// SendLoginOTP sends OTP for login
func (h *Handler) SendLoginOTP(c *gin.Context) {
	var req SendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid OTP request", "error", err)
		errResp := medierror.NewFromRegistry("1000", nil)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	// Check if user exists
	user, err := h.queries.GetUserByMobile(dbutils.GetDBContext(), req.Mobile)
	if err != nil {
		logger.Warn("User not found", "mobile", req.Mobile)
		errResp := medierror.NewFromRegistry("2016", nil)
		c.JSON(http.StatusNotFound, errResp)
		return
	}

	// Send OTP
	err = h.otpService.GenerateAndSendOTP(req.Mobile, user.Email, user.Name, "login")
	if err != nil {
		if appErr, ok := err.(*medierror.AppError); ok {
			c.JSON(http.StatusInternalServerError, appErr)
		} else {
			errResp := medierror.NewFromRegistry("1004", nil)
			c.JSON(http.StatusInternalServerError, errResp)
		}
		return
	}

	logger.Info("Login OTP sent", "mobile", req.Mobile, "email", user.Email)

	c.JSON(http.StatusOK, MessageResponse{
		Message: "OTP sent successfully to your registered email.",
	})
}
```

**Step 2: Add login error codes to errors.yaml**

In `server/config/errors.yaml`, add:

```yaml
"2016":
  message: User not found
  display_message: No account found with this mobile number.
  decline_type: "NF"
  source: AUTH

"2017":
  message: Account pending approval
  display_message: Your account is pending admin approval. Please wait.
  decline_type: "PA"
  source: AUTH

"2018":
  message: Account not active
  display_message: Your account is not active. Please contact admin.
  decline_type: "NA"
  source: AUTH
```

**Step 3: Commit login handler**

```bash
git add server/modules/auth/handler.go server/config/errors.yaml
git commit -m "feat(auth): add login handlers with OTP and password"
```

---

## Task 10: Create Auth Handler - Password Reset & Token Refresh

**Step 1: Add password reset and token refresh methods**

In `server/modules/auth/handler.go`, add after login methods:

```go
// ForgotPassword initiates password reset process
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid forgot password request", "error", err)
		errResp := medierror.NewFromRegistry("1000", nil)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	// Check if user exists
	user, err := h.queries.GetUserByMobile(dbutils.GetDBContext(), req.Mobile)
	if err != nil {
		logger.Warn("User not found", "mobile", req.Mobile)
		errResp := medierror.NewFromRegistry("2016", nil)
		c.JSON(http.StatusNotFound, errResp)
		return
	}

	// Verify email matches
	if user.Email != req.Email {
		logger.Warn("Email does not match", "mobile", req.Mobile, "provided_email", req.Email)
		errResp := medierror.NewFromRegistry("2016", nil)
		c.JSON(http.StatusNotFound, errResp)
		return
	}

	// Send OTP
	err = h.otpService.GenerateAndSendOTP(req.Mobile, req.Email, user.Name, "forgot_password")
	if err != nil {
		if appErr, ok := err.(*medierror.AppError); ok {
			c.JSON(http.StatusInternalServerError, appErr)
		} else {
			errResp := medierror.NewFromRegistry("1004", nil)
			c.JSON(http.StatusInternalServerError, errResp)
		}
		return
	}

	logger.Info("Password reset OTP sent", "mobile", req.Mobile, "email", req.Email)

	c.JSON(http.StatusOK, MessageResponse{
		Message: "Password reset OTP sent to your email.",
	})
}

// ResetPassword resets user password
func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid reset password request", "error", err)
		errResp := medierror.NewFromRegistry("1000", nil)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	// Validate OTP
	err := h.otpService.ValidateOTP(req.Mobile, req.Email, req.OTP, "forgot_password")
	if err != nil {
		if appErr, ok := err.(*medierror.AppError); ok {
			c.JSON(http.StatusUnauthorized, appErr)
		} else {
			errResp := medierror.NewFromRegistry("1004", nil)
			c.JSON(http.StatusInternalServerError, errResp)
		}
		return
	}

	// Validate password strength
	if err := h.passwordService.ValidatePasswordStrength(req.NewPassword); err != nil {
		if appErr, ok := err.(*medierror.AppError); ok {
			c.JSON(http.StatusBadRequest, appErr)
		} else {
			errResp := medierror.NewFromRegistry("1004", nil)
			c.JSON(http.StatusInternalServerError, errResp)
		}
		return
	}

	// Get user
	user, err := h.queries.GetUserByMobile(dbutils.GetDBContext(), req.Mobile)
	if err != nil {
		logger.Warn("User not found", "mobile", req.Mobile)
		errResp := medierror.NewFromRegistry("2016", nil)
		c.JSON(http.StatusNotFound, errResp)
		return
	}

	// Hash new password
	hashedPassword, err := h.passwordService.HashPassword(req.NewPassword)
	if err != nil {
		errResp := medierror.NewFromRegistry("1004", nil)
		c.JSON(http.StatusInternalServerError, errResp)
		return
	}

	// Update password
	err = h.queries.UpdateUserPassword(dbutils.GetDBContext(), db.UpdateUserPasswordParams{
		PasswordHash: sql.NullString{String: hashedPassword, Valid: true},
		ID:           user.ID,
	})
	if err != nil {
		logger.Error("Failed to update password", "error", err)
		errResp := medierror.NewFromRegistry("1004", nil)
		c.JSON(http.StatusInternalServerError, errResp)
		return
	}

	logger.Info("Password reset successfully", "user_id", user.ID)

	c.JSON(http.StatusOK, MessageResponse{
		Message: "Password reset successfully. You can now log in with your new password.",
	})
}

// RefreshToken refreshes access token using refresh token
func (h *Handler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid refresh token request", "error", err)
		errResp := medierror.NewFromRegistry("1000", nil)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	// Refresh access token
	newAccessToken, err := h.jwtService.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		if appErr, ok := err.(*medierror.AppError); ok {
			c.JSON(http.StatusUnauthorized, appErr)
		} else {
			errResp := medierror.NewFromRegistry("2010", nil)
			c.JSON(http.StatusUnauthorized, errResp)
		}
		return
	}

	// Validate refresh token to get user info
	claims, err := h.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		errResp := medierror.NewFromRegistry("2010", nil)
		c.JSON(http.StatusUnauthorized, errResp)
		return
	}

	logger.Info("Token refreshed", "user_id", claims.UserID)

	c.JSON(http.StatusOK, gin.H{
		"access_token": newAccessToken,
		"expires_at":   time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	})
}

// GetCurrentUser returns current logged-in user info
func (h *Handler) GetCurrentUser(c *gin.Context) {
	userID := c.GetInt64("user_id")

	user, err := h.queries.GetUserByID(dbutils.GetDBContext(), userID)
	if err != nil {
		logger.Error("Failed to get user", "error", err, "user_id", userID)
		errResp := medierror.NewFromRegistry("2016", nil)
		c.JSON(http.StatusNotFound, errResp)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          user.ID,
		"name":        user.Name,
		"mobile":      user.Mobile,
		"email":       user.Email,
		"designation": user.Designation,
		"is_admin":    user.IsAdmin == 1,
		"status":      user.Status,
		"created_on":  user.CreatedOn,
	})
}
```

**Step 2: Commit password reset and token refresh**

```bash
git add server/modules/auth/handler.go
git commit -m "feat(auth): add forgot password, reset password, and token refresh"
```

---

## Task 11: Create Auth Module Registration

**Step 1: Create module file**

Create `server/modules/auth/module.go`:

```go
package auth

import (
	"github.com/gin-gonic/gin"

	"github.com/leal-hospital/server/di"
	"github.com/leal-hospital/server/middleware"
	"github.com/leal-hospital/server/models/db"
	jwtService "github.com/leal-hospital/server/services/jwt"
)

// Module represents the auth module
type Module struct{}

// Name returns the module name
func (m *Module) Name() string {
	return "auth"
}

// Configure registers dependencies in DI container
func (m *Module) Configure(container *di.Container) {
	// Auth handler is registered as a factory
	container.RegisterFactory((*Handler)(nil), func(c *di.Container) any {
		queries := c.Get((*db.Queries)(nil)).(*db.Queries)
		otpSvc := c.Get((*interface{ otp.OTPService })(nil))
		passwordSvc := c.Get((*interface{ password.PasswordService })(nil))
		jwtSvc := c.Get((*interface{ jwtService.JWTService })(nil))

		return NewHandler(queries, otpSvc, passwordSvc, jwtSvc)
	})
}

// RegisterRoutes registers auth routes
func (m *Module) RegisterRoutes(router *gin.Engine, container *di.Container) {
	handler := container.Get((*Handler)(nil)).(*Handler)
	jwtSvc := container.Get((*interface{ jwtService.JWTService })(nil)).(jwtService.JWTService)

	authGroup := router.Group("/api/auth")
	{
		// Public routes
		authGroup.POST("/register/admin", handler.RegisterAdmin)
		authGroup.POST("/register/staff", handler.RegisterStaff)
		authGroup.POST("/login/otp", handler.LoginWithOTP)
		authGroup.POST("/login/password", handler.LoginWithPassword)
		authGroup.POST("/otp/send-login", handler.SendLoginOTP)
		authGroup.POST("/forgot-password", handler.ForgotPassword)
		authGroup.POST("/reset-password", handler.ResetPassword)
		authGroup.POST("/refresh-token", handler.RefreshToken)

		// Protected routes
		protected := authGroup.Group("")
		protected.Use(middleware.AuthMiddleware(jwtSvc))
		{
			protected.GET("/me", handler.GetCurrentUser)
		}
	}
}
```

**Step 2: Commit auth module**

```bash
git add server/modules/auth/module.go
git commit -m "feat(auth): add auth module with DI and route registration"
```

---

## Task 12: Update App Bootstrap to Include Auth Module

**Step 1: Update app.go to register services and auth module**

In `server/app/app.go`, update the `Bootstrap` method:

```go
// After database initialization and before middleware setup, add:

	// Initialize services
	emailService := email.NewEmailService(&a.config.Email)

	otpService := otp.NewOTPService(
		queries,
		emailService,
		a.config.Security.OTPLength,
		a.config.Security.OTPExpiration,
		5, // max retries
	)

	passwordService := password.NewPasswordService()

	jwtService := jwt.NewJWTService(
		a.config.Security.JWTSecret,
		a.config.Security.JWTExpiration,
		a.config.Security.RefreshExpiration,
	)

	// Register services in DI container
	a.container.Register(emailService)
	a.container.Register(otpService)
	a.container.Register(passwordService)
	a.container.Register(jwtService)
	a.container.Register(queries)
```

**Step 2: Update main.go to include auth module**

In `server/cmd/main.go`, update the modules slice:

```go
	// Register modules
	modules := []app.Module{
		&auth.Module{},
	}
```

Add import:
```go
	"github.com/leal-hospital/server/modules/auth"
```

**Step 3: Commit app integration**

```bash
git add server/app/app.go server/cmd/main.go
git commit -m "feat(auth): integrate auth module into app bootstrap"
```

---

## Task 13: Update Database Schema and Test

**Step 1: Drop and recreate database with new schema**

```bash
mysql -u abhay.sahani -p
```

In MySQL:
```sql
DROP DATABASE IF EXISTS lael;
CREATE DATABASE lael;
USE lael;
SOURCE /Users/abhay.sahani/Documents/MyRepos/leal-hospital/server/models/dbConf/schema.sql;
SHOW TABLES;
DESCRIBE lael_users;
DESCRIBE lael_otp;
EXIT;
```

**Step 2: Test server starts without errors**

```bash
cd server
go run cmd/main.go
```

Expected output: Server starts on port 8080 without errors

**Step 3: Commit database migration**

```bash
git add server/models/dbConf/schema.sql
git commit -m "feat(db): update schema with mandatory email field"
```

---

## Task 14: Manual API Testing

**Step 1: Test admin registration**

```bash
curl -X POST http://localhost:8080/api/auth/register/admin \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Dr. Santosh Kumar",
    "mobile": "9876543210",
    "email": "santosh@laelhospital.com",
    "password": "Admin@123"
  }'
```

Expected: 201 Created with success message

**Step 2: Check email for OTP (check your SMTP logs or email)**

**Step 3: Test login with OTP**

```bash
curl -X POST http://localhost:8080/api/auth/otp/send-login \
  -H "Content-Type: application/json" \
  -d '{
    "mobile": "9876543210",
    "email": "santosh@laelhospital.com"
  }'
```

```bash
curl -X POST http://localhost:8080/api/auth/login/otp \
  -H "Content-Type: application/json" \
  -d '{
    "mobile": "9876543210",
    "email": "santosh@laelhospital.com",
    "otp": "123456"
  }'
```

Expected: 200 OK with access_token and refresh_token

**Step 4: Test protected route**

```bash
export TOKEN="<access_token_from_step_3>"

curl -X GET http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

Expected: 200 OK with user details

**Step 5: Test staff registration**

```bash
curl -X POST http://localhost:8080/api/auth/register/staff \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Nurse Priya",
    "mobile": "9876543211",
    "email": "priya@laelhospital.com",
    "designation": "nurse",
    "password": "Nurse@123"
  }'
```

Expected: 201 Created with approval pending message

**Step 6: Test login with password (should fail - not approved)**

```bash
curl -X POST http://localhost:8080/api/auth/login/password \
  -H "Content-Type: application/json" \
  -d '{
    "mobile": "9876543211",
    "password": "Nurse@123"
  }'
```

Expected: 403 Forbidden - account pending approval

---

## Task 15: Create Admin Approval Endpoints (Bonus)

**Files:**
- Create: `server/modules/admin/handler.go`
- Create: `server/modules/admin/module.go`

**Step 1: Create admin handler**

Create `server/modules/admin/handler.go`:

```go
package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/leal-hospital/server/logger"
	"github.com/leal-hospital/server/medierror"
	"github.com/leal-hospital/server/middleware"
	"github.com/leal-hospital/server/models/db"
	"github.com/leal-hospital/server/utils/dbutils"
)

// Handler handles admin operations
type Handler struct {
	queries *db.Queries
}

// NewHandler creates a new admin handler
func NewHandler(queries *db.Queries) *Handler {
	return &Handler{
		queries: queries,
	}
}

// GetPendingApprovals returns list of staff pending approval
func (h *Handler) GetPendingApprovals(c *gin.Context) {
	staffList, err := h.queries.ListPendingStaffApprovals(dbutils.GetDBContext())
	if err != nil {
		logger.Error("Failed to get pending approvals", "error", err)
		errResp := medierror.NewFromRegistry("1004", nil)
		c.JSON(http.StatusInternalServerError, errResp)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pending_staff": staffList,
	})
}

// ApproveStaff approves a staff member
func (h *Handler) ApproveStaff(c *gin.Context) {
	staffID := c.Param("id")
	adminID := middleware.GetUserIDFromContext(c)

	// Parse staff ID
	var id int64
	if _, err := fmt.Sscanf(staffID, "%d", &id); err != nil {
		errResp := medierror.NewFromRegistry("1000", nil)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	// Approve staff
	err := h.queries.ApproveStaff(dbutils.GetDBContext(), db.ApproveStaffParams{
		IsApproved: 1,
		ApprovedBy: sql.NullInt64{Int64: adminID, Valid: true},
		ID:         id,
	})
	if err != nil {
		logger.Error("Failed to approve staff", "error", err, "staff_id", id)
		errResp := medierror.NewFromRegistry("1004", nil)
		c.JSON(http.StatusInternalServerError, errResp)
		return
	}

	logger.Info("Staff approved", "staff_id", id, "approved_by", adminID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Staff approved successfully",
	})
}
```

**Step 2: Create admin module**

Create `server/modules/admin/module.go`:

```go
package admin

import (
	"github.com/gin-gonic/gin"

	"github.com/leal-hospital/server/di"
	"github.com/leal-hospital/server/middleware"
	"github.com/leal-hospital/server/models/db"
	jwtService "github.com/leal-hospital/server/services/jwt"
)

// Module represents the admin module
type Module struct{}

// Name returns the module name
func (m *Module) Name() string {
	return "admin"
}

// Configure registers dependencies
func (m *Module) Configure(container *di.Container) {
	container.RegisterFactory((*Handler)(nil), func(c *di.Container) any {
		queries := c.Get((*db.Queries)(nil)).(*db.Queries)
		return NewHandler(queries)
	})
}

// RegisterRoutes registers admin routes
func (m *Module) RegisterRoutes(router *gin.Engine, container *di.Container) {
	handler := container.Get((*Handler)(nil)).(*Handler)
	jwtSvc := container.Get((*interface{ jwtService.JWTService })(nil)).(jwtService.JWTService)

	adminGroup := router.Group("/api/admin")
	adminGroup.Use(middleware.AuthMiddleware(jwtSvc))
	adminGroup.Use(middleware.AdminOnlyMiddleware())
	{
		adminGroup.GET("/staff/pending", handler.GetPendingApprovals)
		adminGroup.POST("/staff/:id/approve", handler.ApproveStaff)
	}
}
```

**Step 3: Add admin module to main.go**

In `server/cmd/main.go`:

```go
	modules := []app.Module{
		&auth.Module{},
		&admin.Module{},
	}
```

Add import:
```go
	"github.com/leal-hospital/server/modules/admin"
```

**Step 4: Commit admin module**

```bash
git add server/modules/admin/ server/cmd/main.go
git commit -m "feat(admin): add staff approval endpoints"
```

---

## Summary

This plan implements a complete authentication system with:

**Backend:**
- Email service (SMTP) for OTP delivery
- OTP service with generation and validation
- Password service with bcrypt hashing
- JWT service with access/refresh tokens
- Auth middleware for protected routes
- Complete auth module with registration, login, password reset
- Admin module for staff approval

**Security Features:**
- Email-based OTP verification
- bcrypt password hashing (cost 12)
- JWT tokens (1hr access, 7-day refresh)
- Password strength validation
- Rate limiting on OTP retries
- Admin approval workflow for staff

**API Endpoints:**
- POST /api/auth/register/admin
- POST /api/auth/register/staff
- POST /api/auth/login/otp
- POST /api/auth/login/password
- POST /api/auth/otp/send-login
- POST /api/auth/forgot-password
- POST /api/auth/reset-password
- POST /api/auth/refresh-token
- GET /api/auth/me (protected)
- GET /api/admin/staff/pending (admin only)
- POST /api/admin/staff/:id/approve (admin only)

Next step: Frontend implementation (Login, Registration pages)
