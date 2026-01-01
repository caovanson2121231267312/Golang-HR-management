package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App        AppConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	JWT        JWTConfig
	Email      EmailConfig
	Elastic    ElasticConfig
	RateLimit  RateLimitConfig
	Security   SecurityConfig
	Logger     LoggerConfig
	Worker     WorkerConfig
}

type AppConfig struct {
	Name        string
	Environment string
	Port        string
	Debug       bool
	BaseURL     string
	FrontendURL string
	Timezone    string
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Host         string
	Port         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	CacheTTL     time.Duration
}

type JWTConfig struct {
	AccessSecret        string
	RefreshSecret       string
	AccessTokenExpiry   time.Duration
	RefreshTokenExpiry  time.Duration
	Issuer              string
	Audience            string
}

type EmailConfig struct {
	Host       string
	Port       int
	Username   string
	Password   string
	From       string
	FromName   string
	EnableTLS  bool
}

type ElasticConfig struct {
	URLs     []string
	Username string
	Password string
	Index    string
}

type RateLimitConfig struct {
	RequestsPerSecond int
	RequestsPerMinute int
	RequestsPerHour   int
	BurstSize         int
	BlockDuration     time.Duration
}

type SecurityConfig struct {
	BCryptCost           int
	OTPLength            int
	OTPExpiry            time.Duration
	MaxLoginAttempts     int
	LockoutDuration      time.Duration
	PasswordMinLength    int
	SessionTimeout       time.Duration
	CSRFTokenExpiry      time.Duration
	AllowedOrigins       []string
	TrustedProxies       []string
	EnableIPWhitelist    bool
	IPWhitelist          []string
}

type LoggerConfig struct {
	Level      string
	Format     string
	Output     string
	FilePath   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

type WorkerConfig struct {
	Concurrency  int
	RedisAddr    string
	RetryMax     int
	RetryDelay   time.Duration
	Queues       map[string]int
}

var AppConfig_ *Config

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		// .env file is optional
	}

	config := &Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "HR Management System"),
			Environment: getEnv("APP_ENV", "development"),
			Port:        getEnv("APP_PORT", "8080"),
			Debug:       getEnvBool("APP_DEBUG", true),
			BaseURL:     getEnv("APP_BASE_URL", "http://localhost:8080"),
			FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),
			Timezone:    getEnv("APP_TIMEZONE", "Asia/Ho_Chi_Minh"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			DBName:          getEnv("DB_NAME", "hr_management"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 100),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", "1h"),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnv("REDIS_PORT", "6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getEnvInt("REDIS_DB", 0),
			PoolSize:     getEnvInt("REDIS_POOL_SIZE", 100),
			MinIdleConns: getEnvInt("REDIS_MIN_IDLE_CONNS", 10),
			MaxRetries:   getEnvInt("REDIS_MAX_RETRIES", 3),
			CacheTTL:     getEnvDuration("REDIS_CACHE_TTL", "15m"),
		},
		JWT: JWTConfig{
			AccessSecret:       getEnv("JWT_ACCESS_SECRET", "your-super-secret-access-key-change-in-production"),
			RefreshSecret:      getEnv("JWT_REFRESH_SECRET", "your-super-secret-refresh-key-change-in-production"),
			AccessTokenExpiry:  getEnvDuration("JWT_ACCESS_EXPIRY", "15m"),
			RefreshTokenExpiry: getEnvDuration("JWT_REFRESH_EXPIRY", "168h"),
			Issuer:             getEnv("JWT_ISSUER", "hr-management-system"),
			Audience:           getEnv("JWT_AUDIENCE", "hr-management-users"),
		},
		Email: EmailConfig{
			Host:      getEnv("EMAIL_HOST", "smtp.gmail.com"),
			Port:      getEnvInt("EMAIL_PORT", 587),
			Username:  getEnv("EMAIL_USERNAME", ""),
			Password:  getEnv("EMAIL_PASSWORD", ""),
			From:      getEnv("EMAIL_FROM", "noreply@hrms.com"),
			FromName:  getEnv("EMAIL_FROM_NAME", "HR Management System"),
			EnableTLS: getEnvBool("EMAIL_ENABLE_TLS", true),
		},
		Elastic: ElasticConfig{
			URLs:     []string{getEnv("ELASTIC_URL", "http://localhost:9200")},
			Username: getEnv("ELASTIC_USERNAME", "elastic"),
			Password: getEnv("ELASTIC_PASSWORD", "changeme"),
			Index:    getEnv("ELASTIC_INDEX", "hr_management"),
		},
		RateLimit: RateLimitConfig{
			RequestsPerSecond: getEnvInt("RATE_LIMIT_PER_SECOND", 10),
			RequestsPerMinute: getEnvInt("RATE_LIMIT_PER_MINUTE", 100),
			RequestsPerHour:   getEnvInt("RATE_LIMIT_PER_HOUR", 1000),
			BurstSize:         getEnvInt("RATE_LIMIT_BURST", 20),
			BlockDuration:     getEnvDuration("RATE_LIMIT_BLOCK_DURATION", "1h"),
		},
		Security: SecurityConfig{
			BCryptCost:        getEnvInt("BCRYPT_COST", 12),
			OTPLength:         getEnvInt("OTP_LENGTH", 6),
			OTPExpiry:         getEnvDuration("OTP_EXPIRY", "5m"),
			MaxLoginAttempts:  getEnvInt("MAX_LOGIN_ATTEMPTS", 5),
			LockoutDuration:   getEnvDuration("LOCKOUT_DURATION", "30m"),
			PasswordMinLength: getEnvInt("PASSWORD_MIN_LENGTH", 8),
			SessionTimeout:    getEnvDuration("SESSION_TIMEOUT", "24h"),
			CSRFTokenExpiry:   getEnvDuration("CSRF_TOKEN_EXPIRY", "1h"),
			AllowedOrigins:    []string{getEnv("ALLOWED_ORIGINS", "*")},
			TrustedProxies:    []string{getEnv("TRUSTED_PROXIES", "127.0.0.1")},
			EnableIPWhitelist: getEnvBool("ENABLE_IP_WHITELIST", false),
			IPWhitelist:       []string{},
		},
		Logger: LoggerConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "json"),
			Output:     getEnv("LOG_OUTPUT", "file"),
			FilePath:   getEnv("LOG_FILE_PATH", "./logs/app.log"),
			MaxSize:    getEnvInt("LOG_MAX_SIZE", 100),
			MaxBackups: getEnvInt("LOG_MAX_BACKUPS", 5),
			MaxAge:     getEnvInt("LOG_MAX_AGE", 5),
			Compress:   getEnvBool("LOG_COMPRESS", true),
		},
		Worker: WorkerConfig{
			Concurrency: getEnvInt("WORKER_CONCURRENCY", 10),
			RedisAddr:   fmt.Sprintf("%s:%s", getEnv("REDIS_HOST", "localhost"), getEnv("REDIS_PORT", "6379")),
			RetryMax:    getEnvInt("WORKER_RETRY_MAX", 3),
			RetryDelay:  getEnvDuration("WORKER_RETRY_DELAY", "10s"),
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	}

	AppConfig_ = config
	return config, nil
}

func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue string) time.Duration {
	value := getEnv(key, defaultValue)
	duration, err := time.ParseDuration(value)
	if err != nil {
		duration, _ = time.ParseDuration(defaultValue)
	}
	return duration
}
