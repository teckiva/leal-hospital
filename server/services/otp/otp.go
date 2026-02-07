package otp

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"github.com/leal-hospital/server/logger"
	"github.com/leal-hospital/server/models/db"
)

// GenerateAndSendOTP generates a new OTP and sends it via email
func (o *OTPSvc) GenerateAndSendOTP(email, name, otpType string) error {
	functionName := "OTPSvc.GenerateAndSendOTP"

	logger.Info(functionName, "Generating OTP for email:", email, "type:", otpType)

	// Generate random OTP
	otp, err := o.generateOTP()
	if err != nil {
		logger.Error(functionName, "Failed to generate OTP:", err)
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Calculate expiry time
	expiry := time.Now().Add(o.SecurityConfig.OTPExpiration)

	// Convert otpType string to db.LaelOtpOtpType
	var dbOtpType db.LaelOtpOtpType
	switch otpType {
	case "registration":
		dbOtpType = db.LaelOtpOtpTypeRegistration
	case "login":
		dbOtpType = db.LaelOtpOtpTypeLogin
	case "forgot_password":
		dbOtpType = db.LaelOtpOtpTypeForgotPassword
	default:
		logger.Error(functionName, "Invalid OTP type:", otpType)
		return fmt.Errorf("invalid OTP type: %s", otpType)
	}

	// Store OTP in database
	ctx := context.Background()
	_, err = o.Queries.CreateOTP(ctx, db.CreateOTPParams{
		Mobile:  "", // Empty mobile for email-based OTP
		Email:   email,
		Otp:     otp,
		Expiry:  expiry,
		OtpType: dbOtpType,
	})

	if err != nil {
		logger.Error(functionName, "Failed to store OTP in database:", err)
		return fmt.Errorf("failed to store OTP: %w", err)
	}

	// Send OTP via email
	if err := o.EmailSvc.SendOTP(email, name, otp); err != nil {
		logger.Error(functionName, "Failed to send OTP email:", err)
		return fmt.Errorf("failed to send OTP email: %w", err)
	}

	logger.Info(functionName, "OTP generated and sent successfully for email:", email)
	return nil
}

// VerifyOTP verifies the provided OTP for the given email
func (o *OTPSvc) VerifyOTP(email, otp, otpType string) (bool, error) {
	functionName := "OTPSvc.VerifyOTP"

	logger.Info(functionName, "Verifying OTP for email:", email, "type:", otpType)

	// Convert otpType string to db.LaelOtpOtpType
	var dbOtpType db.LaelOtpOtpType
	switch otpType {
	case "registration":
		dbOtpType = db.LaelOtpOtpTypeRegistration
	case "login":
		dbOtpType = db.LaelOtpOtpTypeLogin
	case "forgot_password":
		dbOtpType = db.LaelOtpOtpTypeForgotPassword
	default:
		logger.Error(functionName, "Invalid OTP type:", otpType)
		return false, fmt.Errorf("invalid OTP type: %s", otpType)
	}

	// Get latest OTP from database
	ctx := context.Background()
	otpRecord, err := o.Queries.GetLatestOTPByEmail(ctx, db.GetLatestOTPByEmailParams{
		Email:   email,
		OtpType: dbOtpType,
	})

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Info(functionName, "No OTP found for email:", email)
			return false, nil
		}
		logger.Error(functionName, "Failed to fetch OTP from database:", err)
		return false, fmt.Errorf("failed to fetch OTP: %w", err)
	}

	// Check if OTP has expired
	if time.Now().After(otpRecord.Expiry) {
		logger.Info(functionName, "OTP expired for email:", email)
		return false, nil
	}

	// Check if OTP matches
	if otpRecord.Otp != otp {
		// Increment retry count
		if err := o.Queries.IncrementRetryCount(ctx, otpRecord.ID); err != nil {
			logger.Error(functionName, "Failed to increment retry count:", err)
		}
		logger.Info(functionName, "OTP mismatch for email:", email)
		return false, nil
	}

	// Mark OTP as validated
	if err := o.Queries.ValidateOTP(ctx, otpRecord.ID); err != nil {
		logger.Error(functionName, "Failed to mark OTP as validated:", err)
		return false, fmt.Errorf("failed to validate OTP: %w", err)
	}

	logger.Info(functionName, "OTP verified successfully for email:", email)
	return true, nil
}

// CleanupExpiredOTPs removes expired OTPs from the database
func (o *OTPSvc) CleanupExpiredOTPs() error {
	functionName := "OTPSvc.CleanupExpiredOTPs"

	logger.Info(functionName, "Cleaning up expired OTPs")

	ctx := context.Background()
	if err := o.Queries.DeleteExpiredOTP(ctx); err != nil {
		logger.Error(functionName, "Failed to delete expired OTPs:", err)
		return fmt.Errorf("failed to cleanup expired OTPs: %w", err)
	}

	logger.Info(functionName, "Expired OTPs cleaned up successfully")
	return nil
}

// generateOTP generates a random OTP of configured length
func (o *OTPSvc) generateOTP() (string, error) {
	functionName := "OTPSvc.generateOTP"

	otpLength := o.SecurityConfig.OTPLength
	if otpLength <= 0 {
		otpLength = 6 // Default to 6 if not configured
	}

	// Calculate max value (10^otpLength - 1)
	max := new(big.Int)
	max.Exp(big.NewInt(10), big.NewInt(int64(otpLength)), nil)

	// Generate random number
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		logger.Error(functionName, "Failed to generate random OTP:", err)
		return "", err
	}

	// Format with leading zeros
	format := fmt.Sprintf("%%0%dd", otpLength)
	otp := fmt.Sprintf(format, n)

	return otp, nil
}
