package api

import (
	"context"
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/leal-hospital/server/domain"
	"github.com/leal-hospital/server/handler/auth/dto"
	handlerUtils "github.com/leal-hospital/server/handler/utils"
	"github.com/leal-hospital/server/logger"
	"github.com/leal-hospital/server/medierror"
	"github.com/leal-hospital/server/models/db"
	authPersistence "github.com/leal-hospital/server/persistence/auth"
	"github.com/leal-hospital/server/services/jwt"
	"github.com/leal-hospital/server/services/otp"
	"github.com/leal-hospital/server/services/password"
	"github.com/leal-hospital/server/utils"
)

// AuthHandler defines the interface for authentication operations
type AuthHandler interface {
	Register() gin.HandlerFunc
	VerifyOTP() gin.HandlerFunc
	Login() gin.HandlerFunc
	ForgotPassword() gin.HandlerFunc
	ResetPassword() gin.HandlerFunc
	RefreshToken() gin.HandlerFunc
}

// authHandler implements the AuthHandler interface
type authHandler struct {
	authPers    authPersistence.AuthPersistence
	otpSvc      otp.OTPSvcDriver
	passwordSvc password.PasswordSvcDriver
	jwtSvc      jwt.JWTSvcDriver
}

// NewAuthHandler creates a new instance of AuthHandler
func NewAuthHandler(
	authPers authPersistence.AuthPersistence,
	otpSvc otp.OTPSvcDriver,
	passwordSvc password.PasswordSvcDriver,
	jwtSvc jwt.JWTSvcDriver,
) AuthHandler {
	return &authHandler{
		authPers:    authPers,
		otpSvc:      otpSvc,
		passwordSvc: passwordSvc,
		jwtSvc:      jwtSvc,
	}
}

// Register handles user registration
func (h *authHandler) Register() gin.HandlerFunc {
	const functionName = "handler.auth.api.Register"

	return func(c *gin.Context) {
		// Step 1: Parse and validate request body
		var registerRequestDTO dto.RegisterRequest
		if err := c.ShouldBindJSON(&registerRequestDTO); err != nil {
			logger.Error(functionName, "invalid request body:", err)
			appErr := handlerUtils.FormatValidationError(registerRequestDTO, err)
			c.JSON(200, appErr.ToResponse())
			return
		}

		// Step 2: Convert DTO to domain
		req := dto.ConvertRegisterRequestToDomain(&registerRequestDTO)

		// Step 3: Business logic
		ctx := context.Background()

		// Check if user already exists
		existingUser, err := h.authPers.GetUserByEmail(ctx, req.Email)
		if err == nil && existingUser.ID > 0 {
			logger.Info(functionName, "User already exists:", req.Email)
			appErr := medierror.ErrUserAlreadyExists(nil)
			c.JSON(200, appErr.ToResponse())
			return
		}

		// Hash password
		passwordHash, err := h.passwordSvc.HashPassword(req.Password)
		if err != nil {
			logger.Error(functionName, "Failed to hash password:", err)
			appErr := medierror.ErrInternalServer()
			c.JSON(200, appErr.ToResponse())
			return
		}

		// Convert designation string to enum
		var designation db.LaelUsersDesignation
		switch req.Designation {
		case "doctor":
			designation = db.LaelUsersDesignationDoctor
		case "nurse":
			designation = db.LaelUsersDesignationNurse
		case "staff":
			designation = db.LaelUsersDesignationStaff
		default:
			appErr := medierror.ErrBadRequestWithMsg("Invalid designation", "Invalid designation", nil)
			c.JSON(200, appErr.ToResponse())
			return
		}

		// Create user
		userID, err := h.authPers.CreateUser(ctx, db.CreateUserParams{
			Name:         req.Name,
			Mobile:       req.Mobile,
			Email:        req.Email,
			Designation:  designation,
			IsAdmin:      false,
			IsApproved:   false,
			PasswordHash: sql.NullString{String: passwordHash, Valid: true},
		})

		if err != nil {
			logger.Error(functionName, "Failed to create user:", err)
			appErr := medierror.ErrInternalServer()
			c.JSON(200, appErr.ToResponse())
			return
		}

		// Generate and send OTP
		if err := h.otpSvc.GenerateAndSendOTP(req.Email, req.Name, "registration"); err != nil {
			logger.Error(functionName, "Failed to send OTP:", err)
			appErr := medierror.ErrInternalServer().WithDisplayMessage("User created but failed to send OTP. Please try login.")
			c.JSON(200, appErr.ToResponse())
			return
		}

		logger.Info(functionName, "User registered successfully:", userID)

		// Step 4: Build domain response and convert to DTO
		domainResp := &domain.RegisterResponse{
			Message: "Registration successful. Please verify OTP sent to your email.",
			UserID:  userID,
		}
		responseModel := dto.ConvertRegisterDomainToResponse(domainResp)

		// Step 5: Return success response
		response := utils.ResponseWithModel("200", "success", responseModel)
		c.JSON(200, response)
	}
}

// VerifyOTP handles OTP verification
func (h *authHandler) VerifyOTP() gin.HandlerFunc {
	const functionName = "handler.auth.api.VerifyOTP"

	return func(c *gin.Context) {
		// Step 1: Bind JSON to DTO
		var verifyRequestDTO dto.VerifyOTPRequest
		if err := c.ShouldBindJSON(&verifyRequestDTO); err != nil {
			logger.Error(functionName, "invalid request body:", err)
			appErr := handlerUtils.FormatValidationError(verifyRequestDTO, err)
			c.JSON(200, appErr.ToResponse())
			return
		}

		// Step 2: Convert DTO to domain
		req := dto.ConvertVerifyOTPRequestToDomain(&verifyRequestDTO)

		// Step 3: Business logic
		// Verify OTP
		isValid, err := h.otpSvc.VerifyOTP(req.Email, req.OTP, req.OTPType)
		if err != nil {
			logger.Error(functionName, "OTP verification failed:", err)
			appErr := medierror.ErrInternalServer()
			c.JSON(200, appErr.ToResponse())
			return
		}

		if !isValid {
			logger.Info(functionName, "Invalid OTP for email:", req.Email)
			appErr := medierror.ErrInvalidOTP(nil)
			c.JSON(200, appErr.ToResponse())
			return
		}

		// Get user details
		ctx := context.Background()
		user, err := h.authPers.GetUserByEmail(ctx, req.Email)
		if err != nil {
			logger.Error(functionName, "Failed to fetch user:", err)
			appErr := medierror.ErrUserNotFound(err)
			c.JSON(200, appErr.ToResponse())
			return
		}

		// Generate tokens
		accessToken, err := h.jwtSvc.GenerateAccessToken(user.ID, user.Email, user.IsAdmin)
		if err != nil {
			logger.Error(functionName, "Failed to generate access token:", err)
			appErr := medierror.ErrInternalServer()
			c.JSON(200, appErr.ToResponse())
			return
		}

		refreshToken, err := h.jwtSvc.GenerateRefreshToken(user.ID)
		if err != nil {
			logger.Error(functionName, "Failed to generate refresh token:", err)
			appErr := medierror.ErrInternalServer()
			c.JSON(200, appErr.ToResponse())
			return
		}

		logger.Info(functionName, "OTP verified successfully for email:", req.Email)

		// Step 4: Build domain response and convert to DTO
		domainResp := &domain.VerifyOTPResponse{
			Message:      "OTP verified successfully",
			IsVerified:   true,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}
		responseModel := dto.ConvertVerifyOTPDomainToResponse(domainResp)

		// Step 5: Return response
		response := utils.ResponseWithModel("200", "success", responseModel)
		c.JSON(200, response)
	}
}

// Login handles user login
func (h *authHandler) Login() gin.HandlerFunc {
	const functionName = "handler.auth.api.Login"

	return func(c *gin.Context) {
		// Step 1: Bind JSON to DTO
		var loginRequestDTO dto.LoginRequest
		if err := c.ShouldBindJSON(&loginRequestDTO); err != nil {
			logger.Error(functionName, "invalid request body:", err)
			appErr := handlerUtils.FormatValidationError(loginRequestDTO, err)
			c.JSON(200, appErr.ToResponse())
			return
		}

		// Step 2: Convert DTO to domain
		req := dto.ConvertLoginRequestToDomain(&loginRequestDTO)

		logger.Info(functionName, "Login request for email:", req.Email)

		ctx := context.Background()

		// Step 3: Business logic
		// Get user by email
		user, err := h.authPers.GetUserByEmail(ctx, req.Email)
		if err != nil {
			if err == sql.ErrNoRows {
				logger.Info(functionName, "User not found:", req.Email)
				appErr := medierror.ErrInvalidCredentials(nil)
				c.JSON(200, appErr.ToResponse())
				return
			}
			logger.Error(functionName, "Failed to fetch user:", err)
			appErr := medierror.ErrInternalServer()
			c.JSON(200, appErr.ToResponse())
			return
		}

		// Check if user is approved
		if !user.IsApproved {
			logger.Info(functionName, "User not approved:", req.Email)
			appErr := medierror.ErrStaffNotApproved(nil)
			c.JSON(200, appErr.ToResponse())
			return
		}

		// Verify password
		if !user.PasswordHash.Valid {
			logger.Error(functionName, "User has no password set:", req.Email)
			appErr := medierror.ErrInvalidCredentials(nil)
			c.JSON(200, appErr.ToResponse())
			return
		}

		if err := h.passwordSvc.VerifyPassword(user.PasswordHash.String, req.Password); err != nil {
			logger.Info(functionName, "Password verification failed for:", req.Email)
			appErr := medierror.ErrInvalidCredentials(nil)
			c.JSON(200, appErr.ToResponse())
			return
		}

		// Generate and send OTP for 2FA
		if err := h.otpSvc.GenerateAndSendOTP(req.Email, user.Name, "login"); err != nil {
			logger.Error(functionName, "Failed to send OTP:", err)
			appErr := medierror.ErrInternalServer()
			c.JSON(200, appErr.ToResponse())
			return
		}

		logger.Info(functionName, "OTP sent successfully for login:", req.Email)

		// Step 4: Build domain response and convert to DTO
		domainResp := &domain.LoginResponse{
			Message: "OTP sent to your email. Please verify to complete login.",
		}
		responseModel := dto.ConvertLoginDomainToResponse(domainResp)

		// Step 5: Return response
		response := utils.ResponseWithModel("200", "success", responseModel)
		c.JSON(200, response)
	}
}

// ForgotPassword handles forgot password request
func (h *authHandler) ForgotPassword() gin.HandlerFunc {
	const functionName = "handler.auth.api.ForgotPassword"

	return func(c *gin.Context) {
		// Step 1: Bind JSON to DTO
		var forgotRequestDTO dto.ForgotPasswordRequest
		if err := c.ShouldBindJSON(&forgotRequestDTO); err != nil {
			logger.Error(functionName, "invalid request body:", err)
			appErr := handlerUtils.FormatValidationError(forgotRequestDTO, err)
			c.JSON(200, appErr.ToResponse())
			return
		}

		// Step 2: Convert DTO to domain
		req := dto.ConvertForgotPasswordRequestToDomain(&forgotRequestDTO)

		logger.Info(functionName, "Forgot password request for email:", req.Email)

		ctx := context.Background()

		// Step 3: Business logic
		// Check if user exists
		user, err := h.authPers.GetUserByEmail(ctx, req.Email)
		if err != nil {
			if err == sql.ErrNoRows {
				// Don't reveal if email exists or not for security
				logger.Info(functionName, "User not found:", req.Email)
				domainResp := &domain.ForgotPasswordResponse{
					Message: "If the email exists, an OTP has been sent.",
				}
				responseModel := dto.ConvertForgotPasswordDomainToResponse(domainResp)
				response := utils.ResponseWithModel("200", "success", responseModel)
				c.JSON(200, response)
				return
			}
			logger.Error(functionName, "Failed to fetch user:", err)
			appErr := medierror.ErrInternalServer()
			c.JSON(200, appErr.ToResponse())
			return
		}

		// Generate and send OTP
		if err := h.otpSvc.GenerateAndSendOTP(req.Email, user.Name, "forgot_password"); err != nil {
			logger.Error(functionName, "Failed to send OTP:", err)
			appErr := medierror.ErrInternalServer()
			c.JSON(200, appErr.ToResponse())
			return
		}

		logger.Info(functionName, "Password reset OTP sent for:", req.Email)

		// Step 4: Build domain response and convert to DTO
		domainResp := &domain.ForgotPasswordResponse{
			Message: "If the email exists, an OTP has been sent.",
		}
		responseModel := dto.ConvertForgotPasswordDomainToResponse(domainResp)

		// Step 5: Return response
		response := utils.ResponseWithModel("200", "success", responseModel)
		c.JSON(200, response)
	}
}

// ResetPassword handles password reset with OTP
func (h *authHandler) ResetPassword() gin.HandlerFunc {
	const functionName = "handler.auth.api.ResetPassword"

	return func(c *gin.Context) {
		// Step 1: Bind JSON to DTO
		var resetRequestDTO dto.ResetPasswordRequest
		if err := c.ShouldBindJSON(&resetRequestDTO); err != nil {
			logger.Error(functionName, "invalid request body:", err)
			appErr := handlerUtils.FormatValidationError(resetRequestDTO, err)
			c.JSON(200, appErr.ToResponse())
			return
		}

		// Step 2: Convert DTO to domain
		req := dto.ConvertResetPasswordRequestToDomain(&resetRequestDTO)

		logger.Info(functionName, "Password reset request for email:", req.Email)

		// Step 3: Business logic
		// Verify OTP
		isValid, err := h.otpSvc.VerifyOTP(req.Email, req.OTP, "forgot_password")
		if err != nil {
			logger.Error(functionName, "OTP verification failed:", err)
			appErr := medierror.ErrInternalServer()
			c.JSON(200, appErr.ToResponse())
			return
		}

		if !isValid {
			logger.Info(functionName, "Invalid OTP for email:", req.Email)
			appErr := medierror.ErrInvalidOTP(nil)
			c.JSON(200, appErr.ToResponse())
			return
		}

		// Get user
		ctx := context.Background()
		user, err := h.authPers.GetUserByEmail(ctx, req.Email)
		if err != nil {
			logger.Error(functionName, "Failed to fetch user:", err)
			appErr := medierror.ErrUserNotFound(err)
			c.JSON(200, appErr.ToResponse())
			return
		}

		// Hash new password
		passwordHash, err := h.passwordSvc.HashPassword(req.NewPassword)
		if err != nil {
			logger.Error(functionName, "Failed to hash password:", err)
			appErr := medierror.ErrInternalServer()
			c.JSON(200, appErr.ToResponse())
			return
		}

		// Update password
		if err := h.authPers.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
			PasswordHash: sql.NullString{String: passwordHash, Valid: true},
			ID:           user.ID,
		}); err != nil {
			logger.Error(functionName, "Failed to update password:", err)
			appErr := medierror.ErrInternalServer()
			c.JSON(200, appErr.ToResponse())
			return
		}

		logger.Info(functionName, "Password reset successfully for:", req.Email)

		// Step 4: Build domain response and convert to DTO
		domainResp := &domain.ResetPasswordResponse{
			Message: "Password reset successfully",
		}
		responseModel := dto.ConvertResetPasswordDomainToResponse(domainResp)

		// Step 5: Return response
		response := utils.ResponseWithModel("200", "success", responseModel)
		c.JSON(200, response)
	}
}

// RefreshToken handles token refresh
func (h *authHandler) RefreshToken() gin.HandlerFunc {
	const functionName = "handler.auth.api.RefreshToken"

	return func(c *gin.Context) {
		// Step 1: Bind JSON to DTO
		var refreshRequestDTO dto.RefreshTokenRequest
		if err := c.ShouldBindJSON(&refreshRequestDTO); err != nil {
			logger.Error(functionName, "invalid request body:", err)
			appErr := handlerUtils.FormatValidationError(refreshRequestDTO, err)
			c.JSON(200, appErr.ToResponse())
			return
		}

		// Step 2: Convert DTO to domain
		req := dto.ConvertRefreshTokenRequestToDomain(&refreshRequestDTO)

		logger.Info(functionName, "Token refresh request")

		// Step 3: Business logic
		// Validate refresh token
		claims, err := h.jwtSvc.ValidateToken(req.RefreshToken)
		if err != nil {
			logger.Error(functionName, "Token validation failed:", err)
			appErr := medierror.ErrSessionExpired(err)
			c.JSON(200, appErr.ToResponse())
			return
		}

		// Check if token is refresh token
		if claims.Type != "refresh" {
			logger.Info(functionName, "Invalid token type:", claims.Type)
			appErr := medierror.ErrUnauthorizedWithMsg("Invalid token type", "Refresh token required", nil)
			c.JSON(200, appErr.ToResponse())
			return
		}

		// Get user details
		ctx := context.Background()
		user, err := h.authPers.GetUserByID(ctx, claims.UserID)
		if err != nil {
			logger.Error(functionName, "Failed to fetch user:", err)
			appErr := medierror.ErrUserNotFound(err)
			c.JSON(200, appErr.ToResponse())
			return
		}

		// Generate new tokens
		accessToken, err := h.jwtSvc.GenerateAccessToken(user.ID, user.Email, user.IsAdmin)
		if err != nil {
			logger.Error(functionName, "Failed to generate access token:", err)
			appErr := medierror.ErrInternalServer()
			c.JSON(200, appErr.ToResponse())
			return
		}

		refreshToken, err := h.jwtSvc.GenerateRefreshToken(user.ID)
		if err != nil {
			logger.Error(functionName, "Failed to generate refresh token:", err)
			appErr := medierror.ErrInternalServer()
			c.JSON(200, appErr.ToResponse())
			return
		}

		logger.Info(functionName, "Tokens refreshed successfully for user:", user.ID)

		// Step 4: Build domain response and convert to DTO
		domainResp := &domain.RefreshTokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}
		responseModel := dto.ConvertRefreshTokenDomainToResponse(domainResp)

		// Step 5: Return response
		response := utils.ResponseWithModel("200", "success", responseModel)
		c.JSON(200, response)
	}
}
