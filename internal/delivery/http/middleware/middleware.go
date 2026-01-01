package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"hr-management-system/internal/config"
	"hr-management-system/internal/delivery/http/response"
	"hr-management-system/internal/i18n"
	"hr-management-system/internal/infrastructure/cache"
	"hr-management-system/internal/infrastructure/logger"
	"hr-management-system/internal/security"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ==================== REQUEST ID ====================

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// ==================== LOGGER ====================

func Logger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		userAgent := c.Request.UserAgent()

		if raw != "" {
			path = path + "?" + raw
		}

		log.LogHTTPRequest(method, path, statusCode, latency, clientIP, userAgent)
	}
}

// ==================== CORS ====================

func CORS(cfg *config.SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		
		allowed := false
		for _, o := range cfg.AllowedOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID, X-CSRF-Token, Accept-Language")
			c.Header("Access-Control-Expose-Headers", "Content-Length, X-Request-ID")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// ==================== SECURITY HEADERS ====================

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.Next()
	}
}

// ==================== JWT AUTH ====================

func JWTAuth(jwtCfg *config.JWTConfig, redisCache *cache.RedisCache) gin.HandlerFunc {
	blacklist := security.NewSessionBlacklist(redisCache)

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "auth.token_invalid")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(c, "auth.token_invalid")
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := security.ValidateAccessToken(tokenString, jwtCfg)
		if err != nil {
			response.Unauthorized(c, "auth.token_expired")
			c.Abort()
			return
		}

		// Check if session is blacklisted
		isBlacklisted, err := blacklist.IsBlacklisted(c.Request.Context(), claims.SessionID)
		if err == nil && isBlacklisted {
			response.Unauthorized(c, "auth.token_invalid")
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)
		c.Set("permissions", claims.Permissions)
		c.Set("session_id", claims.SessionID)
		c.Set("claims", claims)

		c.Next()
	}
}

// ==================== OPTIONAL AUTH ====================

func OptionalAuth(jwtCfg *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]
		claims, err := security.ValidateAccessToken(tokenString, jwtCfg)
		if err != nil {
			c.Next()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)
		c.Set("permissions", claims.Permissions)
		c.Set("session_id", claims.SessionID)
		c.Set("claims", claims)

		c.Next()
	}
}

// ==================== PERMISSION CHECK ====================

func RequirePermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userPerms, exists := c.Get("permissions")
		if !exists {
			response.Forbidden(c, "permission.denied")
			c.Abort()
			return
		}

		perms, ok := userPerms.([]string)
		if !ok {
			response.Forbidden(c, "permission.denied")
			c.Abort()
			return
		}

		if !security.HasAnyPermission(perms, permissions) {
			response.Forbidden(c, "permission.denied")
			c.Abort()
			return
		}

		c.Next()
	}
}

func RequireAllPermissions(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userPerms, exists := c.Get("permissions")
		if !exists {
			response.Forbidden(c, "permission.denied")
			c.Abort()
			return
		}

		perms, ok := userPerms.([]string)
		if !ok {
			response.Forbidden(c, "permission.denied")
			c.Abort()
			return
		}

		if !security.HasAllPermissions(perms, permissions) {
			response.Forbidden(c, "permission.denied")
			c.Abort()
			return
		}

		c.Next()
	}
}

// ==================== ROLE CHECK ====================

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Get("roles")
		if !exists {
			response.Forbidden(c, "permission.denied")
			c.Abort()
			return
		}

		uRoles, ok := userRoles.([]string)
		if !ok {
			response.Forbidden(c, "permission.denied")
			c.Abort()
			return
		}

		hasRole := false
		for _, r := range roles {
			for _, ur := range uRoles {
				if r == ur {
					hasRole = true
					break
				}
			}
		}

		if !hasRole {
			response.Forbidden(c, "permission.denied")
			c.Abort()
			return
		}

		c.Next()
	}
}

// ==================== RATE LIMITER ====================

func RateLimiter(redisCache *cache.RedisCache, cfg *config.RateLimitConfig) gin.HandlerFunc {
	limiter := security.NewRateLimiter(redisCache)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		// Check rate limit
		result, err := limiter.CheckIP(
			c.Request.Context(),
			clientIP,
			int64(cfg.RequestsPerMinute),
			time.Minute,
		)

		if err != nil {
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.RequestsPerMinute))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
		c.Header("X-RateLimit-Reset", result.ResetAt.Format(time.RFC3339))

		if !result.Allowed {
			c.Header("Retry-After", fmt.Sprintf("%d", int(time.Until(result.ResetAt).Seconds())))
			response.TooManyRequests(c, "rate_limit.exceeded")
			c.Abort()
			return
		}

		c.Next()
	}
}

// Endpoint specific rate limiter
func EndpointRateLimiter(redisCache *cache.RedisCache, limit int64, window time.Duration) gin.HandlerFunc {
	limiter := security.NewRateLimiter(redisCache)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		endpoint := c.FullPath()

		result, err := limiter.CheckEndpoint(c.Request.Context(), clientIP, endpoint, limit, window)
		if err != nil {
			c.Next()
			return
		}

		if !result.Allowed {
			response.TooManyRequests(c, "rate_limit.exceeded")
			c.Abort()
			return
		}

		c.Next()
	}
}

// ==================== IP WHITELIST ====================

func IPWhitelist(cfg *config.SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.EnableIPWhitelist {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		allowed := false

		for _, ip := range cfg.IPWhitelist {
			if clientIP == ip {
				allowed = true
				break
			}
		}

		if !allowed {
			response.Forbidden(c, "common.forbidden")
			c.Abort()
			return
		}

		c.Next()
	}
}

// ==================== CSRF ====================

func CSRF(redisCache *cache.RedisCache, cfg *config.SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip for GET, HEAD, OPTIONS
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		token := c.GetHeader("X-CSRF-Token")
		if token == "" {
			token = c.PostForm("_csrf")
		}

		if token == "" {
			response.Forbidden(c, "common.forbidden")
			c.Abort()
			return
		}

		// Validate token from cache
		sessionID, _ := c.Get("session_id")
		if sessionID == nil {
			response.Forbidden(c, "common.forbidden")
			c.Abort()
			return
		}

		var storedToken string
		err := redisCache.Get(c.Request.Context(), fmt.Sprintf("csrf:%s", sessionID), &storedToken)
		if err != nil || storedToken != token {
			response.Forbidden(c, "common.forbidden")
			c.Abort()
			return
		}

		c.Next()
	}
}

// ==================== LANGUAGE ====================

func Language() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetHeader("Accept-Language")
		if lang == "" {
			lang = c.Query("lang")
		}
		if lang == "" {
			lang = "vi"
		}

		supported := i18n.Get().GetSupportedLanguages()
		lang = i18n.DetectLanguage(lang, supported)

		c.Set("language", lang)
		c.Next()
	}
}

// ==================== TIMEOUT ====================

func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		done := make(chan struct{})
		go func() {
			c.Next()
			close(done)
		}()

		select {
		case <-done:
			return
		case <-ctx.Done():
			c.AbortWithStatus(http.StatusGatewayTimeout)
		}
	}
}

// ==================== RECOVERY ====================

func Recovery(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.WithFields(map[string]interface{}{
					"error":      err,
					"path":       c.Request.URL.Path,
					"method":     c.Request.Method,
					"ip":         c.ClientIP(),
					"user_agent": c.Request.UserAgent(),
				}).Error("Panic recovered")

				response.InternalError(c, fmt.Errorf("%v", err))
				c.Abort()
			}
		}()
		c.Next()
	}
}

// ==================== HELPER FUNCTIONS ====================

func GetUserID(c *gin.Context) string {
	userID, _ := c.Get("user_id")
	if userID == nil {
		return ""
	}
	return userID.(string)
}

func GetEmail(c *gin.Context) string {
	email, _ := c.Get("email")
	if email == nil {
		return ""
	}
	return email.(string)
}

func GetRoles(c *gin.Context) []string {
	roles, _ := c.Get("roles")
	if roles == nil {
		return []string{}
	}
	return roles.([]string)
}

func GetPermissions(c *gin.Context) []string {
	perms, _ := c.Get("permissions")
	if perms == nil {
		return []string{}
	}
	return perms.([]string)
}

func GetLanguage(c *gin.Context) string {
	lang, _ := c.Get("language")
	if lang == nil {
		return "vi"
	}
	return lang.(string)
}

func GetRequestID(c *gin.Context) string {
	requestID, _ := c.Get("request_id")
	if requestID == nil {
		return ""
	}
	return requestID.(string)
}
