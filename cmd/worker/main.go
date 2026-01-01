package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hr-management-system/internal/config"
	"hr-management-system/internal/infrastructure/cache"
	"hr-management-system/internal/infrastructure/database"
	"hr-management-system/internal/infrastructure/email"
	"hr-management-system/internal/infrastructure/logger"
	"hr-management-system/internal/infrastructure/queue"
	"hr-management-system/internal/infrastructure/search"

	"github.com/hibiken/asynq"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	log, err := logger.NewLogger(&cfg.Logger)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	log.Info("Starting HR Management Worker...")

	// Initialize dependencies
	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	redisCache, err := cache.NewRedisCache(&cfg.Redis)
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to Redis")
	}
	defer redisCache.Close()

	es, _ := search.NewElasticSearch(&cfg.Elastic)
	emailSvc, _ := email.NewEmailService(&cfg.Email)

	// Create worker handlers
	handlers := NewHandlers(db, redisCache, es, emailSvc, log, cfg)

	// Create Asynq server
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.Worker.RedisAddr},
		asynq.Config{
			Concurrency: cfg.Worker.Concurrency,
			Queues:      cfg.Worker.Queues,
			RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
				return time.Duration(n) * cfg.Worker.RetryDelay
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.WithFields(map[string]interface{}{
					"task_type": task.Type(),
					"error":     err.Error(),
				}).Error("Task failed")
			}),
		},
	)

	// Register handlers
	mux := asynq.NewServeMux()
	mux.HandleFunc(queue.TypeEmailSend, handlers.HandleEmailSend)
	mux.HandleFunc(queue.TypeEmailOTP, handlers.HandleEmailOTP)
	mux.HandleFunc(queue.TypeEmailPasswordReset, handlers.HandleEmailPasswordReset)
	mux.HandleFunc(queue.TypeEmailPayslip, handlers.HandleEmailPayslip)
	mux.HandleFunc(queue.TypePayrollCalculate, handlers.HandlePayrollCalculate)
	mux.HandleFunc(queue.TypeReportGenerate, handlers.HandleReportGenerate)
	mux.HandleFunc(queue.TypeNotificationSend, handlers.HandleNotificationSend)
	mux.HandleFunc(queue.TypeElasticIndex, handlers.HandleElasticIndex)
	mux.HandleFunc(queue.TypeElasticDelete, handlers.HandleElasticDelete)
	mux.HandleFunc(queue.TypeAuditLog, handlers.HandleAuditLog)

	// Start server
	go func() {
		if err := srv.Run(mux); err != nil {
			log.WithError(err).Fatal("Worker server failed")
		}
	}()

	log.Info("Worker started successfully")

	// Wait for shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down worker...")
	srv.Shutdown()
	log.Info("Worker stopped")
}

type Handlers struct {
	db       *database.Database
	cache    *cache.RedisCache
	es       *search.ElasticSearch
	email    *email.EmailService
	log      *logger.Logger
	cfg      *config.Config
}

func NewHandlers(db *database.Database, cache *cache.RedisCache, es *search.ElasticSearch, emailSvc *email.EmailService, log *logger.Logger, cfg *config.Config) *Handlers {
	return &Handlers{db: db, cache: cache, es: es, email: emailSvc, log: log, cfg: cfg}
}

func (h *Handlers) HandleEmailSend(ctx context.Context, t *asynq.Task) error {
	var payload queue.EmailPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	start := time.Now()
	err := h.email.Send(ctx, email.Email{
		To:      payload.To,
		Cc:      payload.Cc,
		Bcc:     payload.Bcc,
		Subject: payload.Subject,
		Body:    payload.Body,
		IsHTML:  true,
	})

	h.log.LogJobExecution(queue.TypeEmailSend, t.ResultWriter().TaskID(), time.Since(start), err)
	return err
}

func (h *Handlers) HandleEmailOTP(ctx context.Context, t *asynq.Task) error {
	var payload queue.OTPPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	start := time.Now()
	err := h.email.SendOTP(ctx, payload.Email, "", payload.OTP, payload.Type, payload.Language)
	h.log.LogJobExecution(queue.TypeEmailOTP, t.ResultWriter().TaskID(), time.Since(start), err)
	return err
}

func (h *Handlers) HandleEmailPasswordReset(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		Email     string `json:"email"`
		Name      string `json:"name"`
		ResetLink string `json:"reset_link"`
	}
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	return h.email.SendPasswordReset(ctx, payload.Email, payload.Name, payload.ResetLink)
}

func (h *Handlers) HandleEmailPayslip(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		Email      string `json:"email"`
		Name       string `json:"name"`
		Period     string `json:"period"`
		NetSalary  string `json:"net_salary"`
		PDFContent []byte `json:"pdf_content"`
	}
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	return h.email.SendPayslip(ctx, payload.Email, payload.Name, payload.Period, payload.NetSalary, payload.PDFContent)
}

func (h *Handlers) HandlePayrollCalculate(ctx context.Context, t *asynq.Task) error {
	var payload queue.PayrollPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	h.log.WithField("period_id", payload.PeriodID).Info("Calculating payroll")
	// Payroll calculation logic here
	return nil
}

func (h *Handlers) HandleReportGenerate(ctx context.Context, t *asynq.Task) error {
	var payload queue.ReportPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	h.log.WithFields(map[string]interface{}{
		"report_type": payload.ReportType,
		"format":      payload.Format,
	}).Info("Generating report")
	// Report generation logic here
	return nil
}

func (h *Handlers) HandleNotificationSend(ctx context.Context, t *asynq.Task) error {
	var payload queue.NotificationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	_, err := h.db.ExecContext(ctx, `
		INSERT INTO notifications (id, user_id, title, message, type, data, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW(), NOW())
	`, payload.UserID, payload.Title, payload.Message, payload.Type, payload.Data)

	return err
}

func (h *Handlers) HandleElasticIndex(ctx context.Context, t *asynq.Task) error {
	if h.es == nil {
		return nil
	}

	var payload queue.ElasticPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	return h.es.Index(ctx, payload.Index, payload.DocumentID, payload.Document)
}

func (h *Handlers) HandleElasticDelete(ctx context.Context, t *asynq.Task) error {
	if h.es == nil {
		return nil
	}

	var payload queue.ElasticPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	return h.es.Delete(ctx, payload.Index, payload.DocumentID)
}

func (h *Handlers) HandleAuditLog(ctx context.Context, t *asynq.Task) error {
	var payload queue.AuditLogPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	oldJSON, _ := json.Marshal(payload.OldValues)
	newJSON, _ := json.Marshal(payload.NewValues)

	_, err := h.db.ExecContext(ctx, `
		INSERT INTO audit_logs (id, user_id, action, table_name, record_id, old_values, new_values, ip_address, user_agent, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, NOW())
	`, payload.UserID, payload.Action, payload.TableName, payload.RecordID, oldJSON, newJSON, payload.IPAddress, payload.UserAgent)

	return err
}
