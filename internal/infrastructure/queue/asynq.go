package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"hr-management-system/internal/config"

	"github.com/hibiken/asynq"
)

// Task types
const (
	TypeEmailSend           = "email:send"
	TypeEmailOTP            = "email:otp"
	TypeEmailPasswordReset  = "email:password_reset"
	TypeEmailPayslip        = "email:payslip"
	TypePayrollCalculate    = "payroll:calculate"
	TypePayrollGenerate     = "payroll:generate"
	TypeReportGenerate      = "report:generate"
	TypeReportExport        = "report:export"
	TypeNotificationSend    = "notification:send"
	TypeAttendanceSync      = "attendance:sync"
	TypeDataSync            = "data:sync"
	TypeCacheInvalidate     = "cache:invalidate"
	TypeElasticIndex        = "elastic:index"
	TypeElasticDelete       = "elastic:delete"
	TypeAuditLog            = "audit:log"
)

// Queue priorities
const (
	QueueCritical = "critical"
	QueueDefault  = "default"
	QueueLow      = "low"
)

type Queue struct {
	client    *asynq.Client
	inspector *asynq.Inspector
}

var queue *Queue

func NewQueue(cfg *config.WorkerConfig) (*Queue, error) {
	redisOpt := asynq.RedisClientOpt{Addr: cfg.RedisAddr}
	
	client := asynq.NewClient(redisOpt)
	inspector := asynq.NewInspector(redisOpt)
	
	queue = &Queue{
		client:    client,
		inspector: inspector,
	}
	
	return queue, nil
}

func GetQueue() *Queue {
	return queue
}

func (q *Queue) Close() error {
	return q.client.Close()
}

// Enqueue task
func (q *Queue) Enqueue(ctx context.Context, taskType string, payload interface{}, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	task := asynq.NewTask(taskType, data)
	return q.client.EnqueueContext(ctx, task, opts...)
}

// Enqueue with default options
func (q *Queue) EnqueueDefault(ctx context.Context, taskType string, payload interface{}) (*asynq.TaskInfo, error) {
	return q.Enqueue(ctx, taskType, payload,
		asynq.Queue(QueueDefault),
		asynq.MaxRetry(3),
		asynq.Timeout(5*time.Minute),
	)
}

// Enqueue critical task
func (q *Queue) EnqueueCritical(ctx context.Context, taskType string, payload interface{}) (*asynq.TaskInfo, error) {
	return q.Enqueue(ctx, taskType, payload,
		asynq.Queue(QueueCritical),
		asynq.MaxRetry(5),
		asynq.Timeout(10*time.Minute),
	)
}

// Enqueue low priority task
func (q *Queue) EnqueueLow(ctx context.Context, taskType string, payload interface{}) (*asynq.TaskInfo, error) {
	return q.Enqueue(ctx, taskType, payload,
		asynq.Queue(QueueLow),
		asynq.MaxRetry(2),
		asynq.Timeout(30*time.Minute),
	)
}

// Schedule task
func (q *Queue) Schedule(ctx context.Context, taskType string, payload interface{}, processAt time.Time, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	task := asynq.NewTask(taskType, data)
	opts = append(opts, asynq.ProcessAt(processAt))
	return q.client.EnqueueContext(ctx, task, opts...)
}

// Schedule in duration
func (q *Queue) ScheduleIn(ctx context.Context, taskType string, payload interface{}, delay time.Duration) (*asynq.TaskInfo, error) {
	return q.Schedule(ctx, taskType, payload, time.Now().Add(delay))
}

// Task payloads
type EmailPayload struct {
	To          []string          `json:"to"`
	Cc          []string          `json:"cc,omitempty"`
	Bcc         []string          `json:"bcc,omitempty"`
	Subject     string            `json:"subject"`
	Body        string            `json:"body"`
	Template    string            `json:"template,omitempty"`
	TemplateData map[string]interface{} `json:"template_data,omitempty"`
	Attachments []AttachmentPayload `json:"attachments,omitempty"`
}

type AttachmentPayload struct {
	Filename string `json:"filename"`
	Content  []byte `json:"content"`
	MimeType string `json:"mime_type"`
}

type OTPPayload struct {
	Email    string `json:"email"`
	Phone    string `json:"phone,omitempty"`
	OTP      string `json:"otp"`
	Type     string `json:"type"`
	Language string `json:"language"`
}

type PayrollPayload struct {
	PeriodID   string `json:"period_id"`
	EmployeeID string `json:"employee_id,omitempty"`
	Action     string `json:"action"`
}

type ReportPayload struct {
	ReportType string                 `json:"report_type"`
	Format     string                 `json:"format"`
	Filters    map[string]interface{} `json:"filters"`
	RequestedBy string               `json:"requested_by"`
	Email      string                 `json:"email,omitempty"`
}

type NotificationPayload struct {
	UserID  string                 `json:"user_id"`
	Title   string                 `json:"title"`
	Message string                 `json:"message"`
	Type    string                 `json:"type"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

type ElasticPayload struct {
	Index      string      `json:"index"`
	DocumentID string      `json:"document_id"`
	Document   interface{} `json:"document,omitempty"`
	Action     string      `json:"action"`
}

type AuditLogPayload struct {
	UserID    string      `json:"user_id"`
	Action    string      `json:"action"`
	TableName string      `json:"table_name"`
	RecordID  string      `json:"record_id"`
	OldValues interface{} `json:"old_values,omitempty"`
	NewValues interface{} `json:"new_values,omitempty"`
	IPAddress string      `json:"ip_address"`
	UserAgent string      `json:"user_agent"`
}

// Queue inspection methods
func (q *Queue) GetQueueInfo(queueName string) (*asynq.QueueInfo, error) {
	return q.inspector.GetQueueInfo(queueName)
}

func (q *Queue) GetPendingTasks(queueName string, page, pageSize int) ([]*asynq.TaskInfo, error) {
	return q.inspector.ListPendingTasks(queueName, asynq.Page(page), asynq.PageSize(pageSize))
}

func (q *Queue) GetActiveTasks(queueName string, page, pageSize int) ([]*asynq.TaskInfo, error) {
	return q.inspector.ListActiveTasks(queueName, asynq.Page(page), asynq.PageSize(pageSize))
}

func (q *Queue) GetScheduledTasks(queueName string, page, pageSize int) ([]*asynq.TaskInfo, error) {
	return q.inspector.ListScheduledTasks(queueName, asynq.Page(page), asynq.PageSize(pageSize))
}

func (q *Queue) GetRetryTasks(queueName string, page, pageSize int) ([]*asynq.TaskInfo, error) {
	return q.inspector.ListRetryTasks(queueName, asynq.Page(page), asynq.PageSize(pageSize))
}

func (q *Queue) GetArchivedTasks(queueName string, page, pageSize int) ([]*asynq.TaskInfo, error) {
	return q.inspector.ListArchivedTasks(queueName, asynq.Page(page), asynq.PageSize(pageSize))
}

func (q *Queue) DeleteTask(queueName, taskID string) error {
	return q.inspector.DeleteTask(queueName, taskID)
}

func (q *Queue) CancelActiveTask(taskID string) error {
	return q.inspector.CancelProcessing(taskID)
}

func (q *Queue) RunTask(queueName, taskID string) error {
	return q.inspector.RunTask(queueName, taskID)
}

func (q *Queue) ArchiveTask(queueName, taskID string) error {
	return q.inspector.ArchiveTask(queueName, taskID)
}

// Helper functions for common tasks
func (q *Queue) SendEmail(ctx context.Context, payload EmailPayload) (*asynq.TaskInfo, error) {
	return q.EnqueueDefault(ctx, TypeEmailSend, payload)
}

func (q *Queue) SendOTP(ctx context.Context, payload OTPPayload) (*asynq.TaskInfo, error) {
	return q.EnqueueCritical(ctx, TypeEmailOTP, payload)
}

func (q *Queue) GenerateReport(ctx context.Context, payload ReportPayload) (*asynq.TaskInfo, error) {
	return q.EnqueueLow(ctx, TypeReportGenerate, payload)
}

func (q *Queue) CalculatePayroll(ctx context.Context, payload PayrollPayload) (*asynq.TaskInfo, error) {
	return q.EnqueueDefault(ctx, TypePayrollCalculate, payload)
}

func (q *Queue) SendNotification(ctx context.Context, payload NotificationPayload) (*asynq.TaskInfo, error) {
	return q.EnqueueDefault(ctx, TypeNotificationSend, payload)
}

func (q *Queue) IndexDocument(ctx context.Context, payload ElasticPayload) (*asynq.TaskInfo, error) {
	return q.EnqueueLow(ctx, TypeElasticIndex, payload)
}

func (q *Queue) LogAudit(ctx context.Context, payload AuditLogPayload) (*asynq.TaskInfo, error) {
	return q.EnqueueLow(ctx, TypeAuditLog, payload)
}
