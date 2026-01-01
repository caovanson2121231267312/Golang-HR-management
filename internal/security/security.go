package security

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net"
	"regexp"
	"strings"
	"time"
	"unicode"

	"hr-management-system/internal/config"
	"hr-management-system/internal/infrastructure/cache"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var cfg *config.SecurityConfig

func Init(c *config.SecurityConfig) {
	cfg = c
}

// ==================== PASSWORD ====================

type PasswordValidator struct {
	MinLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireNumber    bool
	RequireSpecial   bool
}

func DefaultPasswordValidator() *PasswordValidator {
	return &PasswordValidator{
		MinLength:        8,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireNumber:    true,
		RequireSpecial:   true,
	}
}

func (v *PasswordValidator) Validate(password string) error {
	if len(password) < v.MinLength {
		return fmt.Errorf("password must be at least %d characters", v.MinLength)
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsNumber(c):
			hasNumber = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSpecial = true
		}
	}

	if v.RequireUppercase && !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if v.RequireLowercase && !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if v.RequireNumber && !hasNumber {
		return errors.New("password must contain at least one number")
	}
	if v.RequireSpecial && !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	return nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cfg.BCryptCost)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ==================== JWT ====================

type TokenClaims struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions,omitempty"`
	SessionID   string   `json:"session_id"`
	TokenType   string   `json:"token_type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

func GenerateTokenPair(userID, email string, roles, permissions []string, jwtCfg *config.JWTConfig) (*TokenPair, error) {
	sessionID := uuid.New().String()
	now := time.Now()

	// Access Token
	accessClaims := TokenClaims{
		UserID:      userID,
		Email:       email,
		Roles:       roles,
		Permissions: permissions,
		SessionID:   sessionID,
		TokenType:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jwtCfg.Issuer,
			Subject:   userID,
			Audience:  jwt.ClaimStrings{jwtCfg.Audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(jwtCfg.AccessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(jwtCfg.AccessSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Refresh Token
	refreshClaims := TokenClaims{
		UserID:    userID,
		SessionID: sessionID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jwtCfg.Issuer,
			Subject:   userID,
			Audience:  jwt.ClaimStrings{jwtCfg.Audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(jwtCfg.RefreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(jwtCfg.RefreshSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    now.Add(jwtCfg.AccessTokenExpiry),
		TokenType:    "Bearer",
	}, nil
}

func ValidateAccessToken(tokenString string, jwtCfg *config.JWTConfig) (*TokenClaims, error) {
	return validateToken(tokenString, jwtCfg.AccessSecret, "access")
}

func ValidateRefreshToken(tokenString string, jwtCfg *config.JWTConfig) (*TokenClaims, error) {
	return validateToken(tokenString, jwtCfg.RefreshSecret, "refresh")
}

func validateToken(tokenString, secret, expectedType string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("invalid token type: expected %s, got %s", expectedType, claims.TokenType)
	}

	return claims, nil
}

// ==================== OTP ====================

func GenerateOTP(length int) (string, error) {
	if length <= 0 {
		length = 6
	}

	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(length)), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%0*d", length, n), nil
}

func HashOTP(otp string) string {
	hash := sha256.Sum256([]byte(otp))
	return hex.EncodeToString(hash[:])
}

func VerifyOTP(otp, hash string) bool {
	return HashOTP(otp) == hash
}

// ==================== TOKEN GENERATION ====================

func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func GenerateCSRFToken() (string, error) {
	return GenerateSecureToken(32)
}

func GenerateSessionID() string {
	return uuid.New().String()
}

// ==================== RATE LIMITER ====================

type RateLimiter struct {
	cache     *cache.RedisCache
	keyPrefix string
}

func NewRateLimiter(c *cache.RedisCache) *RateLimiter {
	return &RateLimiter{
		cache:     c,
		keyPrefix: "rl:",
	}
}

type RateLimitResult struct {
	Allowed   bool
	Remaining int64
	ResetAt   time.Time
}

func (r *RateLimiter) Check(ctx context.Context, key string, limit int64, window time.Duration) (*RateLimitResult, error) {
	allowed, remaining, err := r.cache.RateLimit(ctx, r.keyPrefix+key, limit, window)
	if err != nil {
		return nil, err
	}

	return &RateLimitResult{
		Allowed:   allowed,
		Remaining: remaining,
		ResetAt:   time.Now().Add(window),
	}, nil
}

func (r *RateLimiter) CheckIP(ctx context.Context, ip string, limit int64, window time.Duration) (*RateLimitResult, error) {
	return r.Check(ctx, "ip:"+ip, limit, window)
}

func (r *RateLimiter) CheckUser(ctx context.Context, userID string, limit int64, window time.Duration) (*RateLimitResult, error) {
	return r.Check(ctx, "user:"+userID, limit, window)
}

func (r *RateLimiter) CheckEndpoint(ctx context.Context, ip, endpoint string, limit int64, window time.Duration) (*RateLimitResult, error) {
	key := fmt.Sprintf("endpoint:%s:%s", ip, endpoint)
	return r.Check(ctx, key, limit, window)
}

// ==================== LOGIN ATTEMPTS ====================

type LoginAttemptManager struct {
	cache *cache.RedisCache
}

func NewLoginAttemptManager(c *cache.RedisCache) *LoginAttemptManager {
	return &LoginAttemptManager{cache: c}
}

func (m *LoginAttemptManager) RecordFailedAttempt(ctx context.Context, identifier string) (int, error) {
	key := "login_attempts:" + identifier
	count, err := m.cache.Incr(ctx, key)
	if err != nil {
		return 0, err
	}

	if count == 1 {
		m.cache.Expire(ctx, key, cfg.LockoutDuration)
	}

	return int(count), nil
}

func (m *LoginAttemptManager) GetAttempts(ctx context.Context, identifier string) (int, error) {
	key := "login_attempts:" + identifier
	var count int
	err := m.cache.Get(ctx, key, &count)
	if err == cache.ErrCacheMiss {
		return 0, nil
	}
	return count, err
}

func (m *LoginAttemptManager) IsLocked(ctx context.Context, identifier string) (bool, time.Duration, error) {
	attempts, err := m.GetAttempts(ctx, identifier)
	if err != nil {
		return false, 0, err
	}

	if attempts >= cfg.MaxLoginAttempts {
		ttl, err := m.cache.TTL(ctx, "login_attempts:"+identifier)
		if err != nil {
			return true, cfg.LockoutDuration, nil
		}
		return true, ttl, nil
	}

	return false, 0, nil
}

func (m *LoginAttemptManager) ClearAttempts(ctx context.Context, identifier string) error {
	return m.cache.Delete(ctx, "login_attempts:"+identifier)
}

// ==================== SESSION BLACKLIST ====================

type SessionBlacklist struct {
	cache *cache.RedisCache
}

func NewSessionBlacklist(c *cache.RedisCache) *SessionBlacklist {
	return &SessionBlacklist{cache: c}
}

func (s *SessionBlacklist) Add(ctx context.Context, sessionID string, ttl time.Duration) error {
	return s.cache.Set(ctx, "blacklist:"+sessionID, true, ttl)
}

func (s *SessionBlacklist) IsBlacklisted(ctx context.Context, sessionID string) (bool, error) {
	exists, err := s.cache.Exists(ctx, "blacklist:"+sessionID)
	return exists, err
}

func (s *SessionBlacklist) Remove(ctx context.Context, sessionID string) error {
	return s.cache.Delete(ctx, "blacklist:"+sessionID)
}

// ==================== IP VALIDATION ====================

func IsPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	privateBlocks := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
	}

	for _, block := range privateBlocks {
		_, cidr, _ := net.ParseCIDR(block)
		if cidr.Contains(parsedIP) {
			return true
		}
	}

	return false
}

func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

func GetClientIP(remoteAddr string, xForwardedFor string, xRealIP string, trustedProxies []string) string {
	// Check X-Real-IP first
	if xRealIP != "" && isFromTrustedProxy(remoteAddr, trustedProxies) {
		return strings.TrimSpace(xRealIP)
	}

	// Check X-Forwarded-For
	if xForwardedFor != "" && isFromTrustedProxy(remoteAddr, trustedProxies) {
		ips := strings.Split(xForwardedFor, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if IsValidIP(ip) && !IsPrivateIP(ip) {
				return ip
			}
		}
	}

	// Fallback to remote address
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

func isFromTrustedProxy(remoteAddr string, trustedProxies []string) bool {
	host, _, _ := net.SplitHostPort(remoteAddr)
	for _, proxy := range trustedProxies {
		if host == proxy {
			return true
		}
	}
	return false
}

// ==================== INPUT SANITIZATION ====================

func SanitizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func IsValidPhone(phone string) bool {
	phoneRegex := regexp.MustCompile(`^(\+84|84|0)[0-9]{9,10}$`)
	return phoneRegex.MatchString(strings.ReplaceAll(phone, " ", ""))
}

func SanitizePhone(phone string) string {
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	if strings.HasPrefix(phone, "+84") {
		phone = "0" + phone[3:]
	} else if strings.HasPrefix(phone, "84") {
		phone = "0" + phone[2:]
	}
	return phone
}

// ==================== PERMISSION CHECK ====================

func HasPermission(userPermissions []string, requiredPermission string) bool {
	for _, p := range userPermissions {
		if p == requiredPermission || p == "*" {
			return true
		}
		// Check wildcard permissions (e.g., "users.*" matches "users.create")
		if strings.HasSuffix(p, ".*") {
			prefix := strings.TrimSuffix(p, ".*")
			if strings.HasPrefix(requiredPermission, prefix) {
				return true
			}
		}
	}
	return false
}

func HasAnyPermission(userPermissions []string, requiredPermissions []string) bool {
	for _, rp := range requiredPermissions {
		if HasPermission(userPermissions, rp) {
			return true
		}
	}
	return false
}

func HasAllPermissions(userPermissions []string, requiredPermissions []string) bool {
	for _, rp := range requiredPermissions {
		if !HasPermission(userPermissions, rp) {
			return false
		}
	}
	return true
}

// ==================== FINGERPRINT ====================

func GenerateDeviceFingerprint(userAgent, ip, acceptLanguage string) string {
	data := fmt.Sprintf("%s|%s|%s", userAgent, ip, acceptLanguage)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
