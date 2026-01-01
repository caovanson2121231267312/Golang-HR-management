package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"hr-management-system/internal/config"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
	cfg     *config.LoggerConfig
	file    *os.File
	mu      sync.Mutex
	logDir  string
}

var logger *Logger

func NewLogger(cfg *config.LoggerConfig) (*Logger, error) {
	log := logrus.New()

	// Set level
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	// Set formatter
	if cfg.Format == "json" {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				filename := filepath.Base(f.File)
				return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)
			},
		})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
		})
	}

	log.SetReportCaller(true)

	logger = &Logger{
		Logger: log,
		cfg:    cfg,
		logDir: filepath.Dir(cfg.FilePath),
	}

	// Setup output
	if cfg.Output == "file" || cfg.Output == "both" {
		if err := logger.setupFileOutput(); err != nil {
			return nil, err
		}
	}

	if cfg.Output == "stdout" || cfg.Output == "both" {
		if cfg.Output == "both" && logger.file != nil {
			log.SetOutput(io.MultiWriter(os.Stdout, logger.file))
		} else {
			log.SetOutput(os.Stdout)
		}
	}

	// Start log rotation goroutine
	go logger.rotateLogsDaily()

	// Cleanup old logs (keep only 5 days)
	go logger.cleanupOldLogs()

	return logger, nil
}

func GetLogger() *Logger {
	return logger
}

func (l *Logger) setupFileOutput() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Create log directory
	if err := os.MkdirAll(l.logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file with date suffix
	logFile := l.getLogFileName()
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	l.file = file
	l.Logger.SetOutput(file)

	return nil
}

func (l *Logger) getLogFileName() string {
	base := filepath.Base(l.cfg.FilePath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	date := time.Now().Format("2006-01-02")
	return filepath.Join(l.logDir, fmt.Sprintf("%s-%s%s", name, date, ext))
}

func (l *Logger) rotateLogsDaily() {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 1, 0, now.Location())
		time.Sleep(time.Until(next))

		l.mu.Lock()
		if l.file != nil {
			l.file.Close()
		}
		l.setupFileOutput()
		l.mu.Unlock()

		l.Info("Log file rotated")
	}
}

func (l *Logger) cleanupOldLogs() {
	// Cleanup logs older than 5 days
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Run immediately on start
	l.doCleanup()

	for range ticker.C {
		l.doCleanup()
	}
}

func (l *Logger) doCleanup() {
	cutoff := time.Now().AddDate(0, 0, -l.cfg.MaxAge)

	err := filepath.Walk(l.logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if info.ModTime().Before(cutoff) {
			l.WithField("file", path).Info("Removing old log file")
			os.Remove(path)
		}

		return nil
	})

	if err != nil {
		l.WithError(err).Error("Failed to cleanup old logs")
	}
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// Convenience methods with context
func (l *Logger) WithUserID(userID string) *logrus.Entry {
	return l.WithField("user_id", userID)
}

func (l *Logger) WithRequestID(requestID string) *logrus.Entry {
	return l.WithField("request_id", requestID)
}

func (l *Logger) WithModule(module string) *logrus.Entry {
	return l.WithField("module", module)
}

func (l *Logger) WithAction(action string) *logrus.Entry {
	return l.WithField("action", action)
}

func (l *Logger) WithIP(ip string) *logrus.Entry {
	return l.WithField("ip", ip)
}

// HTTP request logging
func (l *Logger) LogHTTPRequest(method, path string, statusCode int, latency time.Duration, ip, userAgent string) {
	l.WithFields(logrus.Fields{
		"method":      method,
		"path":        path,
		"status_code": statusCode,
		"latency_ms":  latency.Milliseconds(),
		"ip":          ip,
		"user_agent":  userAgent,
	}).Info("HTTP Request")
}

// Database query logging
func (l *Logger) LogDBQuery(query string, duration time.Duration, err error) {
	entry := l.WithFields(logrus.Fields{
		"query":       query,
		"duration_ms": duration.Milliseconds(),
	})

	if err != nil {
		entry.WithError(err).Error("Database query failed")
	} else {
		entry.Debug("Database query executed")
	}
}

// Audit logging
func (l *Logger) LogAudit(userID, action, resource, resourceID string, oldValue, newValue interface{}) {
	l.WithFields(logrus.Fields{
		"user_id":     userID,
		"action":      action,
		"resource":    resource,
		"resource_id": resourceID,
		"old_value":   oldValue,
		"new_value":   newValue,
	}).Info("Audit log")
}

// Security logging
func (l *Logger) LogSecurityEvent(eventType, userID, ip, description string) {
	l.WithFields(logrus.Fields{
		"event_type":  eventType,
		"user_id":     userID,
		"ip":          ip,
		"description": description,
	}).Warn("Security event")
}

// Auth logging
func (l *Logger) LogAuthAttempt(email, ip string, success bool, reason string) {
	entry := l.WithFields(logrus.Fields{
		"email":   email,
		"ip":      ip,
		"success": success,
		"reason":  reason,
	})

	if success {
		entry.Info("Authentication successful")
	} else {
		entry.Warn("Authentication failed")
	}
}

// Job/Queue logging
func (l *Logger) LogJobExecution(jobType, jobID string, duration time.Duration, err error) {
	entry := l.WithFields(logrus.Fields{
		"job_type":    jobType,
		"job_id":      jobID,
		"duration_ms": duration.Milliseconds(),
	})

	if err != nil {
		entry.WithError(err).Error("Job execution failed")
	} else {
		entry.Info("Job executed successfully")
	}
}

// Cache logging
func (l *Logger) LogCacheOperation(operation, key string, hit bool) {
	l.WithFields(logrus.Fields{
		"operation": operation,
		"key":       key,
		"cache_hit": hit,
	}).Debug("Cache operation")
}

// Email logging
func (l *Logger) LogEmailSent(to, subject string, success bool, err error) {
	entry := l.WithFields(logrus.Fields{
		"to":      to,
		"subject": subject,
		"success": success,
	})

	if err != nil {
		entry.WithError(err).Error("Email sending failed")
	} else {
		entry.Info("Email sent successfully")
	}
}
