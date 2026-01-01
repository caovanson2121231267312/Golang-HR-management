package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hr-management-system/internal/config"
	"hr-management-system/internal/infrastructure/cache"
	"hr-management-system/internal/infrastructure/database"
	"hr-management-system/internal/infrastructure/logger"
	"hr-management-system/internal/infrastructure/queue"

	"github.com/robfig/cron/v3"
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

	log.Info("Starting HR Management Scheduler...")

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

	jobQueue, err := queue.NewQueue(&cfg.Worker)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize job queue")
	}
	defer jobQueue.Close()

	scheduler := NewScheduler(db, redisCache, jobQueue, log, cfg)

	loc, _ := time.LoadLocation(cfg.App.Timezone)
	c := cron.New(cron.WithLocation(loc), cron.WithSeconds())

	// Daily attendance reminder at 8:00 AM (Mon-Fri)
	c.AddFunc("0 0 8 * * 1-5", func() {
		log.Info("Running: Daily attendance reminder")
		scheduler.SendAttendanceReminder()
	})

	// Daily attendance report at 6:00 PM (Mon-Fri)
	c.AddFunc("0 0 18 * * 1-5", func() {
		log.Info("Running: Daily attendance report")
		scheduler.GenerateDailyAttendanceReport()
	})

	// Weekly attendance summary on Monday at 9:00 AM
	c.AddFunc("0 0 9 * * 1", func() {
		log.Info("Running: Weekly attendance summary")
		scheduler.GenerateWeeklyAttendanceSummary()
	})

	// Monthly payroll reminder on 25th at 9:00 AM
	c.AddFunc("0 0 9 25 * *", func() {
		log.Info("Running: Monthly payroll reminder")
		scheduler.SendPayrollReminder()
	})

	// Check leave balance expiry on 1st of each month
	c.AddFunc("0 0 0 1 * *", func() {
		log.Info("Running: Leave balance check")
		scheduler.CheckLeaveBalanceExpiry()
	})

	// Birthday notifications at 8:00 AM daily
	c.AddFunc("0 0 8 * * *", func() {
		log.Info("Running: Birthday notifications")
		scheduler.SendBirthdayNotifications()
	})

	// Contract expiry check daily at 9:00 AM
	c.AddFunc("0 0 9 * * *", func() {
		log.Info("Running: Contract expiry check")
		scheduler.CheckContractExpiry()
	})

	// Clean expired sessions every hour
	c.AddFunc("0 0 * * * *", func() {
		log.Info("Running: Session cleanup")
		scheduler.CleanExpiredSessions()
	})

	// Clean old audit logs weekly (keep 90 days)
	c.AddFunc("0 0 2 * * 0", func() {
		log.Info("Running: Audit log cleanup")
		scheduler.CleanOldAuditLogs()
	})

	// Elasticsearch sync daily at 3:00 AM
	c.AddFunc("0 0 3 * * *", func() {
		log.Info("Running: Elasticsearch sync")
		scheduler.SyncElasticsearch()
	})

	c.Start()
	log.Info("Scheduler started successfully")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down scheduler...")
	c.Stop()
	log.Info("Scheduler stopped")
}

type Scheduler struct {
	db    *database.Database
	cache *cache.RedisCache
	queue *queue.Queue
	log   *logger.Logger
	cfg   *config.Config
}

func NewScheduler(db *database.Database, cache *cache.RedisCache, q *queue.Queue, log *logger.Logger, cfg *config.Config) *Scheduler {
	return &Scheduler{db: db, cache: cache, queue: q, log: log, cfg: cfg}
}

func (s *Scheduler) SendAttendanceReminder() {
	ctx := context.Background()
	today := time.Now().Format("2006-01-02")

	rows, err := s.db.QueryContext(ctx, `
		SELECT u.id, u.email, e.full_name FROM users u
		INNER JOIN employees e ON e.user_id = u.id
		WHERE e.employment_status = 'active'
		AND e.id NOT IN (SELECT employee_id FROM attendances WHERE date = $1)
	`, today)
	if err != nil {
		s.log.WithError(err).Error("Failed to get employees without attendance")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var userID, email, name string
		rows.Scan(&userID, &email, &name)
		s.queue.SendNotification(ctx, queue.NotificationPayload{
			UserID:  userID,
			Title:   "Nhắc nhở chấm công",
			Message: "Bạn chưa chấm công hôm nay. Vui lòng chấm công ngay.",
			Type:    "attendance_reminder",
		})
	}
}

func (s *Scheduler) GenerateDailyAttendanceReport() {
	ctx := context.Background()
	s.queue.GenerateReport(ctx, queue.ReportPayload{
		ReportType:  "daily_attendance",
		Format:      "excel",
		Filters:     map[string]interface{}{"date": time.Now().Format("2006-01-02")},
		RequestedBy: "system",
	})
}

func (s *Scheduler) GenerateWeeklyAttendanceSummary() {
	ctx := context.Background()
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -7)
	s.queue.GenerateReport(ctx, queue.ReportPayload{
		ReportType:  "weekly_attendance_summary",
		Format:      "excel",
		Filters:     map[string]interface{}{"start_date": startDate.Format("2006-01-02"), "end_date": endDate.Format("2006-01-02")},
		RequestedBy: "system",
	})
}

func (s *Scheduler) SendPayrollReminder() {
	ctx := context.Background()
	month := time.Now().Format("2006-01")

	rows, _ := s.db.QueryContext(ctx, `
		SELECT u.id FROM users u
		INNER JOIN user_roles ur ON ur.user_id = u.id
		INNER JOIN roles r ON r.id = ur.role_id
		WHERE r.slug IN ('hr_manager', 'payroll_manager')
	`)
	defer rows.Close()

	for rows.Next() {
		var userID string
		rows.Scan(&userID)
		s.queue.SendNotification(ctx, queue.NotificationPayload{
			UserID:  userID,
			Title:   "Nhắc nhở tính lương",
			Message: fmt.Sprintf("Đến thời điểm tính lương tháng %s.", month),
			Type:    "payroll_reminder",
		})
	}
}

func (s *Scheduler) CheckLeaveBalanceExpiry() {
	ctx := context.Background()
	lastYear := time.Now().Year() - 1
	s.db.ExecContext(ctx, `
		UPDATE leave_balances SET carried_over = LEAST(total_days - used_days, 
			(SELECT max_carry_over FROM leave_types WHERE id = leave_type_id))
		WHERE year = $1
	`, lastYear)
	s.log.Info("Leave balance expiry check completed")
}

func (s *Scheduler) SendBirthdayNotifications() {
	ctx := context.Background()
	today := time.Now().Format("01-02")

	rows, _ := s.db.QueryContext(ctx, `
		SELECT e.full_name, u.id FROM employees e
		INNER JOIN users u ON u.id = e.user_id
		WHERE TO_CHAR(e.date_of_birth, 'MM-DD') = $1 AND e.employment_status = 'active'
	`, today)
	defer rows.Close()

	for rows.Next() {
		var name, userID string
		rows.Scan(&name, &userID)
		s.queue.SendNotification(ctx, queue.NotificationPayload{
			UserID:  userID,
			Title:   "Chúc mừng sinh nhật!",
			Message: "Chúc bạn một ngày sinh nhật vui vẻ! 🎂",
			Type:    "birthday",
		})
	}
}

func (s *Scheduler) CheckContractExpiry() {
	ctx := context.Background()
	warningDate := time.Now().AddDate(0, 0, 30).Format("2006-01-02")

	rows, _ := s.db.QueryContext(ctx, `
		SELECT e.full_name, e.contract_end_date, u.id FROM employees e
		INNER JOIN users u ON u.id = e.user_id
		WHERE e.contract_end_date <= $1 AND e.employment_status = 'active'
	`, warningDate)
	defer rows.Close()

	for rows.Next() {
		var name, endDate, userID string
		rows.Scan(&name, &endDate, &userID)
		s.queue.SendNotification(ctx, queue.NotificationPayload{
			UserID:  userID,
			Title:   "Thông báo hợp đồng sắp hết hạn",
			Message: fmt.Sprintf("Hợp đồng của bạn sẽ hết hạn vào ngày %s", endDate),
			Type:    "contract_expiry",
		})
	}
}

func (s *Scheduler) CleanExpiredSessions() {
	ctx := context.Background()
	result, _ := s.db.ExecContext(ctx, `DELETE FROM user_sessions WHERE expires_at < NOW() OR is_revoked = true`)
	affected, _ := result.RowsAffected()
	s.log.WithField("count", affected).Info("Cleaned expired sessions")
}

func (s *Scheduler) CleanOldAuditLogs() {
	ctx := context.Background()
	cutoff := time.Now().AddDate(0, 0, -90).Format("2006-01-02")
	result, _ := s.db.ExecContext(ctx, `DELETE FROM audit_logs WHERE created_at < $1`, cutoff)
	affected, _ := result.RowsAffected()
	s.log.WithField("count", affected).Info("Cleaned old audit logs")
}

func (s *Scheduler) SyncElasticsearch() {
	ctx := context.Background()

	rows, _ := s.db.QueryContext(ctx, `
		SELECT e.id, e.employee_code, e.full_name, u.email, e.department_id, d.name,
		       e.position_id, p.name, e.employment_status, e.employment_type, e.join_date
		FROM employees e
		INNER JOIN users u ON u.id = e.user_id
		INNER JOIN departments d ON d.id = e.department_id
		INNER JOIN positions p ON p.id = e.position_id
		WHERE e.deleted_at IS NULL AND e.updated_at > NOW() - INTERVAL '1 day'
	`)
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id, code, name, email, deptID, deptName, posID, posName, status, empType string
		var joinDate time.Time
		rows.Scan(&id, &code, &name, &email, &deptID, &deptName, &posID, &posName, &status, &empType, &joinDate)

		s.queue.IndexDocument(ctx, queue.ElasticPayload{
			Index: "employees", DocumentID: id,
			Document: map[string]interface{}{
				"id": id, "employee_code": code, "full_name": name, "email": email,
				"department_id": deptID, "department_name": deptName, "position_id": posID,
				"position_name": posName, "employment_status": status, "employment_type": empType,
				"join_date": joinDate, "updated_at": time.Now(),
			},
			Action: "index",
		})
		count++
	}
	s.log.WithField("count", count).Info("Elasticsearch sync completed")
}
