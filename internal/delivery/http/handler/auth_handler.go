package handler

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"hr-management-system/internal/config"
	"hr-management-system/internal/delivery/http/dto"
	"hr-management-system/internal/delivery/http/middleware"
	"hr-management-system/internal/delivery/http/response"
	"hr-management-system/internal/domain/entity"
	"hr-management-system/internal/infrastructure/cache"
	"hr-management-system/internal/infrastructure/database"
	"hr-management-system/internal/infrastructure/email"
	"hr-management-system/internal/infrastructure/logger"
	"hr-management-system/internal/infrastructure/queue"
	"hr-management-system/internal/security"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	db     *database.Database
	cache  *cache.RedisCache
	queue  *queue.Queue
	email  *email.EmailService
	log    *logger.Logger
	cfg    *config.Config
}

func NewAuthHandler(
	db *database.Database,
	cache *cache.RedisCache,
	queue *queue.Queue,
	emailSvc *email.EmailService,
	log *logger.Logger,
	cfg *config.Config,
) *AuthHandler {
	return &AuthHandler{
		db:    db,
		cache: cache,
		queue: queue,
		email: emailSvc,
		log:   log,
		cfg:   cfg,
	}
}

// Login handles user authentication
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "common.validation_error", nil)
		return
	}

	ctx := c.Request.Context()
	clientIP := c.ClientIP()
	
	// Check login attempts
	loginManager := security.NewLoginAttemptManager(h.cache)
	isLocked, lockDuration, err := loginManager.IsLocked(ctx, req.Email)
	if err == nil && isLocked {
		response.Error(c, 429, "ACCOUNT_LOCKED", "auth.account_locked", map[string]string{
			"duration": lockDuration.String(),
		})
		return
	}

	// Find user
	var user entity.User
	err = h.db.QueryRowContext(ctx, `
		SELECT id, email, phone, password, status, email_verified_at, 
		       two_factor_enabled, preferred_language, failed_login_attempts, locked_until
		FROM users WHERE email = $1 AND deleted_at IS NULL
	`, req.Email).Scan(
		&user.ID, &user.Email, &user.Phone, &user.Password, &user.Status,
		&user.EmailVerifiedAt, &user.TwoFactorEnabled, &user.PreferredLanguage,
		&user.FailedLoginAttempts, &user.LockedUntil,
	)

	if err == sql.ErrNoRows {
		loginManager.RecordFailedAttempt(ctx, req.Email)
		h.log.LogAuthAttempt(req.Email, clientIP, false, "user not found")
		response.Unauthorized(c, "auth.login_failed")
		return
	}

	if err != nil {
		h.log.WithError(err).Error("Failed to query user")
		response.InternalError(c, err)
		return
	}

	// Check password
	if !security.CheckPassword(req.Password, user.Password) {
		loginManager.RecordFailedAttempt(ctx, req.Email)
		h.log.LogAuthAttempt(req.Email, clientIP, false, "invalid password")
		response.Unauthorized(c, "auth.login_failed")
		return
	}

	// Check status
	if user.Status != entity.UserStatusActive {
		h.log.LogAuthAttempt(req.Email, clientIP, false, "account inactive")
		response.Unauthorized(c, "auth.account_inactive")
		return
	}

	// Check 2FA
	if user.TwoFactorEnabled {
		// Generate and send OTP
		otp, _ := security.GenerateOTP(6)
		h.cache.SetOTP(ctx, user.Email, security.HashOTP(otp), h.cfg.Security.OTPExpiry)
		
		h.queue.SendOTP(ctx, queue.OTPPayload{
			Email:    user.Email,
			OTP:      otp,
			Type:     "two_factor",
			Language: user.PreferredLanguage,
		})

		response.OK(c, "auth.two_factor_required", gin.H{
			"requires_2fa": true,
			"email":        user.Email,
		})
		return
	}

	// Clear login attempts
	loginManager.ClearAttempts(ctx, req.Email)

	// Get roles and permissions
	roles, permissions := h.getUserRolesAndPermissions(ctx, user.ID)

	// Generate tokens
	tokenPair, err := security.GenerateTokenPair(
		user.ID.String(),
		user.Email,
		roles,
		permissions,
		&h.cfg.JWT,
	)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	// Update last login
	h.db.ExecContext(ctx, `
		UPDATE users SET last_login_at = NOW(), last_login_ip = $1, 
		       failed_login_attempts = 0 WHERE id = $2
	`, clientIP, user.ID)

	// Store session
	h.storeSession(ctx, user.ID, tokenPair, c)

	h.log.LogAuthAttempt(req.Email, clientIP, true, "")

	response.OK(c, "auth.login_success", dto.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		TokenType:    tokenPair.TokenType,
		User: dto.UserResponse{
			ID:                user.ID,
			Email:             user.Email,
			Phone:             user.Phone,
			Status:            string(user.Status),
			TwoFactorEnabled:  user.TwoFactorEnabled,
			PreferredLanguage: user.PreferredLanguage,
			Roles:             roles,
			Permissions:       permissions,
		},
	})
}

// Verify2FA handles two-factor authentication
func (h *AuthHandler) Verify2FA(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Code  string `json:"code" binding:"required,len=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "common.validation_error", nil)
		return
	}

	ctx := c.Request.Context()

	// Get stored OTP
	storedHash, err := h.cache.GetOTP(ctx, req.Email)
	if err != nil {
		response.BadRequest(c, "otp.expired", nil)
		return
	}

	// Verify OTP
	if !security.VerifyOTP(req.Code, storedHash) {
		response.BadRequest(c, "otp.invalid", nil)
		return
	}

	// Delete OTP
	h.cache.DeleteOTP(ctx, req.Email)

	// Get user
	var user entity.User
	err = h.db.QueryRowContext(ctx, `
		SELECT id, email, phone, status, preferred_language
		FROM users WHERE email = $1 AND deleted_at IS NULL
	`, req.Email).Scan(&user.ID, &user.Email, &user.Phone, &user.Status, &user.PreferredLanguage)

	if err != nil {
		response.InternalError(c, err)
		return
	}

	// Get roles and permissions
	roles, permissions := h.getUserRolesAndPermissions(ctx, user.ID)

	// Generate tokens
	tokenPair, err := security.GenerateTokenPair(
		user.ID.String(),
		user.Email,
		roles,
		permissions,
		&h.cfg.JWT,
	)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	// Store session
	h.storeSession(ctx, user.ID, tokenPair, c)

	response.OK(c, "auth.login_success", dto.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		TokenType:    tokenPair.TokenType,
		User: dto.UserResponse{
			ID:                user.ID,
			Email:             user.Email,
			Phone:             user.Phone,
			Status:            string(user.Status),
			PreferredLanguage: user.PreferredLanguage,
			Roles:             roles,
			Permissions:       permissions,
		},
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "common.validation_error", nil)
		return
	}

	claims, err := security.ValidateRefreshToken(req.RefreshToken, &h.cfg.JWT)
	if err != nil {
		response.Unauthorized(c, "auth.token_expired")
		return
	}

	ctx := c.Request.Context()

	// Get user
	var user entity.User
	err = h.db.QueryRowContext(ctx, `
		SELECT id, email, phone, status, preferred_language
		FROM users WHERE id = $1 AND deleted_at IS NULL
	`, claims.UserID).Scan(&user.ID, &user.Email, &user.Phone, &user.Status, &user.PreferredLanguage)

	if err != nil {
		response.Unauthorized(c, "auth.token_invalid")
		return
	}

	// Check status
	if user.Status != entity.UserStatusActive {
		response.Unauthorized(c, "auth.account_inactive")
		return
	}

	// Get roles and permissions
	roles, permissions := h.getUserRolesAndPermissions(ctx, user.ID)

	// Generate new tokens
	tokenPair, err := security.GenerateTokenPair(
		user.ID.String(),
		user.Email,
		roles,
		permissions,
		&h.cfg.JWT,
	)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	// Blacklist old session
	blacklist := security.NewSessionBlacklist(h.cache)
	blacklist.Add(ctx, claims.SessionID, h.cfg.JWT.RefreshTokenExpiry)

	// Store new session
	h.storeSession(ctx, user.ID, tokenPair, c)

	response.OK(c, "auth.refresh_success", gin.H{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_at":    tokenPair.ExpiresAt,
		"token_type":    tokenPair.TokenType,
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID := middleware.GetUserID(c)
	if sessionID == "" {
		response.OK(c, "auth.logout_success", nil)
		return
	}

	ctx := c.Request.Context()

	// Blacklist session
	blacklist := security.NewSessionBlacklist(h.cache)
	blacklist.Add(ctx, sessionID, h.cfg.JWT.RefreshTokenExpiry)

	// Invalidate user cache
	userID := middleware.GetUserID(c)
	h.cache.InvalidateUserCache(ctx, userID)
	h.cache.InvalidatePermissions(ctx, userID)

	response.OK(c, "auth.logout_success", nil)
}

// SendOTP sends OTP to user
func (h *AuthHandler) SendOTP(c *gin.Context) {
	var req dto.SendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "common.validation_error", nil)
		return
	}

	ctx := c.Request.Context()
	lang := middleware.GetLanguage(c)

	// Check rate limit for OTP
	limiter := security.NewRateLimiter(h.cache)
	result, _ := limiter.Check(ctx, "otp:"+req.Email, 3, 5*time.Minute)
	if result != nil && !result.Allowed {
		response.TooManyRequests(c, "otp.too_many_attempts")
		return
	}

	// Get user name
	var userName string
	h.db.QueryRowContext(ctx, `
		SELECT COALESCE(e.full_name, u.email) 
		FROM users u 
		LEFT JOIN employees e ON e.user_id = u.id 
		WHERE u.email = $1
	`, req.Email).Scan(&userName)

	if userName == "" {
		userName = req.Email
	}

	// Generate OTP
	otp, err := security.GenerateOTP(6)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	// Store OTP hash
	h.cache.SetOTP(ctx, req.Email+":"+req.Type, security.HashOTP(otp), h.cfg.Security.OTPExpiry)

	// Send OTP via queue
	h.queue.SendOTP(ctx, queue.OTPPayload{
		Email:    req.Email,
		OTP:      otp,
		Type:     req.Type,
		Language: lang,
	})

	response.OK(c, "otp.sent", nil)
}

// VerifyOTP verifies OTP code
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req dto.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "common.validation_error", nil)
		return
	}

	ctx := c.Request.Context()

	// Get stored OTP hash
	storedHash, err := h.cache.GetOTP(ctx, req.Email+":"+req.Type)
	if err != nil {
		response.BadRequest(c, "otp.expired", nil)
		return
	}

	// Verify OTP
	if !security.VerifyOTP(req.OTP, storedHash) {
		response.BadRequest(c, "otp.invalid", nil)
		return
	}

	// Delete OTP
	h.cache.DeleteOTP(ctx, req.Email+":"+req.Type)

	// Handle based on type
	switch req.Type {
	case "email_verification":
		h.db.ExecContext(ctx, `
			UPDATE users SET email_verified_at = NOW() WHERE email = $1
		`, req.Email)
	}

	response.OK(c, "otp.verified", nil)
}

// ForgotPassword initiates password reset
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "common.validation_error", nil)
		return
	}

	ctx := c.Request.Context()

	// Check if user exists
	var userID uuid.UUID
	var userName string
	err := h.db.QueryRowContext(ctx, `
		SELECT u.id, COALESCE(e.full_name, u.email)
		FROM users u
		LEFT JOIN employees e ON e.user_id = u.id
		WHERE u.email = $1 AND u.deleted_at IS NULL
	`, req.Email).Scan(&userID, &userName)

	if err == sql.ErrNoRows {
		// Don't reveal if user exists
		response.OK(c, "auth.password_reset_sent", nil)
		return
	}

	if err != nil {
		response.InternalError(c, err)
		return
	}

	// Generate reset token
	token, err := security.GenerateSecureToken(32)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	// Store token
	_, err = h.db.ExecContext(ctx, `
		INSERT INTO password_reset_tokens (id, user_id, token, expires_at)
		VALUES ($1, $2, $3, $4)
	`, uuid.New(), userID, token, time.Now().Add(time.Hour))

	if err != nil {
		response.InternalError(c, err)
		return
	}

	// Send reset email
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", h.cfg.App.FrontendURL, token)
	h.email.SendPasswordReset(ctx, req.Email, userName, resetLink)

	response.OK(c, "auth.password_reset_sent", nil)
}

// ResetPassword handles password reset
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "common.validation_error", nil)
		return
	}

	// Validate password
	validator := security.DefaultPasswordValidator()
	if err := validator.Validate(req.Password); err != nil {
		response.BadRequest(c, "common.validation_error", map[string]string{
			"password": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	// Find valid reset token
	var tokenRecord entity.PasswordResetToken
	err := h.db.QueryRowContext(ctx, `
		SELECT id, user_id, expires_at, used_at
		FROM password_reset_tokens
		WHERE token = $1 AND expires_at > NOW() AND used_at IS NULL
	`, req.Token).Scan(&tokenRecord.ID, &tokenRecord.UserID, &tokenRecord.ExpiresAt, &tokenRecord.UsedAt)

	if err == sql.ErrNoRows {
		response.BadRequest(c, "auth.token_invalid", nil)
		return
	}

	if err != nil {
		response.InternalError(c, err)
		return
	}

	// Hash new password
	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	// Update password
	_, err = h.db.ExecContext(ctx, `
		UPDATE users SET password = $1, password_changed_at = NOW() WHERE id = $2
	`, hashedPassword, tokenRecord.UserID)

	if err != nil {
		response.InternalError(c, err)
		return
	}

	// Mark token as used
	h.db.ExecContext(ctx, `
		UPDATE password_reset_tokens SET used_at = NOW() WHERE id = $1
	`, tokenRecord.ID)

	// Invalidate all sessions
	h.cache.DeleteByPattern(ctx, fmt.Sprintf("session:%s:*", tokenRecord.UserID))

	response.OK(c, "auth.password_reset_success", nil)
}

// ChangePassword handles password change for authenticated users
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "common.validation_error", nil)
		return
	}

	userID := middleware.GetUserID(c)
	ctx := c.Request.Context()

	// Get current password
	var currentHash string
	err := h.db.QueryRowContext(ctx, `
		SELECT password FROM users WHERE id = $1
	`, userID).Scan(&currentHash)

	if err != nil {
		response.InternalError(c, err)
		return
	}

	// Verify old password
	if !security.CheckPassword(req.OldPassword, currentHash) {
		response.BadRequest(c, "auth.old_password_incorrect", nil)
		return
	}

	// Validate new password
	validator := security.DefaultPasswordValidator()
	if err := validator.Validate(req.NewPassword); err != nil {
		response.BadRequest(c, "common.validation_error", map[string]string{
			"new_password": err.Error(),
		})
		return
	}

	// Hash new password
	hashedPassword, err := security.HashPassword(req.NewPassword)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	// Update password
	_, err = h.db.ExecContext(ctx, `
		UPDATE users SET password = $1, password_changed_at = NOW() WHERE id = $2
	`, hashedPassword, userID)

	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, "auth.password_changed", nil)
}

// GetProfile returns current user profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	ctx := c.Request.Context()

	var user entity.User
	err := h.db.QueryRowContext(ctx, `
		SELECT id, email, phone, status, email_verified_at, last_login_at,
		       two_factor_enabled, preferred_language, created_at
		FROM users WHERE id = $1
	`, userID).Scan(
		&user.ID, &user.Email, &user.Phone, &user.Status,
		&user.EmailVerifiedAt, &user.LastLoginAt,
		&user.TwoFactorEnabled, &user.PreferredLanguage, &user.CreatedAt,
	)

	if err != nil {
		response.InternalError(c, err)
		return
	}

	roles, permissions := h.getUserRolesAndPermissions(ctx, user.ID)

	var emailVerified, lastLogin *time.Time
	if user.EmailVerifiedAt.Valid {
		emailVerified = &user.EmailVerifiedAt.Time
	}
	if user.LastLoginAt.Valid {
		lastLogin = &user.LastLoginAt.Time
	}

	response.OK(c, "common.success", dto.UserResponse{
		ID:                user.ID,
		Email:             user.Email,
		Phone:             user.Phone,
		Status:            string(user.Status),
		EmailVerifiedAt:   emailVerified,
		LastLoginAt:       lastLogin,
		TwoFactorEnabled:  user.TwoFactorEnabled,
		PreferredLanguage: user.PreferredLanguage,
		Roles:             roles,
		Permissions:       permissions,
		CreatedAt:         user.CreatedAt,
	})
}

// Helper methods

func (h *AuthHandler) getUserRolesAndPermissions(ctx context.Context, userID uuid.UUID) ([]string, []string) {
	// Try cache first
	perms, err := h.cache.GetPermissions(ctx, userID.String())
	if err == nil && len(perms) > 0 {
		// Get roles from cache
		var roles []string
		h.cache.Get(ctx, "roles:"+userID.String(), &roles)
		return roles, perms
	}

	// Query from database
	rows, err := h.db.QueryContext(ctx, `
		SELECT DISTINCT r.slug
		FROM roles r
		INNER JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1
	`, userID)

	var roles []string
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var role string
			if err := rows.Scan(&role); err == nil {
				roles = append(roles, role)
			}
		}
	}

	// Query permissions
	rows, err = h.db.QueryContext(ctx, `
		SELECT DISTINCT p.slug
		FROM permissions p
		INNER JOIN role_permissions rp ON rp.permission_id = p.id
		INNER JOIN user_roles ur ON ur.role_id = rp.role_id
		WHERE ur.user_id = $1
	`, userID)

	var permissions []string
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var perm string
			if err := rows.Scan(&perm); err == nil {
				permissions = append(permissions, perm)
			}
		}
	}

	// Cache results
	h.cache.SetPermissions(ctx, userID.String(), permissions)
	h.cache.Set(ctx, "roles:"+userID.String(), roles, time.Hour)

	return roles, permissions
}

func (h *AuthHandler) storeSession(ctx context.Context, userID uuid.UUID, tokens *security.TokenPair, c *gin.Context) {
	session := entity.UserSession{
		BaseModel: entity.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
		},
		UserID:       userID,
		Token:        tokens.AccessToken[:50], // Store partial for reference
		RefreshToken: tokens.RefreshToken[:50],
		UserAgent:    c.Request.UserAgent(),
		IPAddress:    c.ClientIP(),
		ExpiresAt:    tokens.ExpiresAt,
		LastActivity: time.Now(),
	}

	h.db.ExecContext(ctx, `
		INSERT INTO user_sessions (id, user_id, token, refresh_token, user_agent, ip_address, expires_at, last_activity)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, session.ID, session.UserID, session.Token, session.RefreshToken,
		session.UserAgent, session.IPAddress, session.ExpiresAt, session.LastActivity)
}
