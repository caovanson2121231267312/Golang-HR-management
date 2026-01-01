package handler

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"hr-management-system/internal/config"
	"hr-management-system/internal/delivery/http/dto"
	"hr-management-system/internal/delivery/http/middleware"
	"hr-management-system/internal/delivery/http/response"
	"hr-management-system/internal/infrastructure/cache"
	"hr-management-system/internal/infrastructure/database"
	"hr-management-system/internal/infrastructure/logger"
	"hr-management-system/internal/infrastructure/queue"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AttendanceHandler struct {
	db    *database.Database
	cache *cache.RedisCache
	queue *queue.Queue
	log   *logger.Logger
	cfg   *config.Config
}

func NewAttendanceHandler(db *database.Database, cache *cache.RedisCache, queue *queue.Queue, log *logger.Logger, cfg *config.Config) *AttendanceHandler {
	return &AttendanceHandler{db: db, cache: cache, queue: queue, log: log, cfg: cfg}
}

// CheckIn records employee check-in
func (h *AttendanceHandler) CheckIn(c *gin.Context) {
	var req dto.CheckInRequest
	c.ShouldBindJSON(&req)

	userID := middleware.GetUserID(c)
	ctx := c.Request.Context()
	clientIP := c.ClientIP()
	today := time.Now().Format("2006-01-02")

	// Get employee ID
	var employeeID uuid.UUID
	err := h.db.QueryRowContext(ctx, `SELECT id FROM employees WHERE user_id = $1 AND deleted_at IS NULL`, userID).Scan(&employeeID)
	if err != nil {
		response.NotFound(c, "employee.not_found")
		return
	}

	// Check if already checked in today
	var existingID uuid.UUID
	err = h.db.QueryRowContext(ctx, `
		SELECT id FROM attendances WHERE employee_id = $1 AND date = $2
	`, employeeID, today).Scan(&existingID)

	if err == nil {
		response.Conflict(c, "attendance.already_checked_in")
		return
	}

	// Create attendance record
	attendanceID := uuid.New()
	now := time.Now()

	_, err = h.db.ExecContext(ctx, `
		INSERT INTO attendances (id, employee_id, date, check_in, check_in_ip, check_in_location, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'present', NOW(), NOW())
	`, attendanceID, employeeID, today, now, clientIP, req.Location)

	if err != nil {
		response.InternalError(c, err)
		return
	}

	// Log attendance
	h.db.ExecContext(ctx, `
		INSERT INTO attendance_logs (id, attendance_id, action, timestamp, ip_address, user_agent, location)
		VALUES ($1, $2, 'check_in', $3, $4, $5, $6)
	`, uuid.New(), attendanceID, now, clientIP, c.Request.UserAgent(), req.Location)

	// h.log.WithModule("attendance").WithUserID(userID).Info("Employee checked in")

	response.OK(c, "attendance.check_in", gin.H{
		"attendance_id": attendanceID,
		"check_in":      now,
	})
}

// CheckOut records employee check-out
func (h *AttendanceHandler) CheckOut(c *gin.Context) {
	var req dto.CheckOutRequest
	c.ShouldBindJSON(&req)

	userID := middleware.GetUserID(c)
	ctx := c.Request.Context()
	clientIP := c.ClientIP()
	today := time.Now().Format("2006-01-02")

	var employeeID uuid.UUID
	h.db.QueryRowContext(ctx, `SELECT id FROM employees WHERE user_id = $1`, userID).Scan(&employeeID)

	// Get today's attendance
	var attendanceID uuid.UUID
	var checkIn time.Time
	var checkOut sql.NullTime

	err := h.db.QueryRowContext(ctx, `
		SELECT id, check_in, check_out FROM attendances 
		WHERE employee_id = $1 AND date = $2
	`, employeeID, today).Scan(&attendanceID, &checkIn, &checkOut)

	if err == sql.ErrNoRows {
		response.BadRequest(c, "attendance.not_checked_in", nil)
		return
	}

	if checkOut.Valid {
		response.Conflict(c, "attendance.already_checked_out")
		return
	}

	now := time.Now()
	workingHours := now.Sub(checkIn).Hours()

	// Update attendance
	_, err = h.db.ExecContext(ctx, `
		UPDATE attendances 
		SET check_out = $1, check_out_ip = $2, check_out_location = $3, 
		    working_hours = $4, updated_at = NOW()
		WHERE id = $5
	`, now, clientIP, req.Location, workingHours, attendanceID)

	if err != nil {
		response.InternalError(c, err)
		return
	}

	// Log
	h.db.ExecContext(ctx, `
		INSERT INTO attendance_logs (id, attendance_id, action, timestamp, ip_address, user_agent, location)
		VALUES ($1, $2, 'check_out', $3, $4, $5, $6)
	`, uuid.New(), attendanceID, now, clientIP, c.Request.UserAgent(), req.Location)

	response.OK(c, "attendance.check_out", gin.H{
		"attendance_id": attendanceID,
		"check_out":     now,
		"working_hours": workingHours,
	})
}

// GetMyAttendance returns current user's attendance
func (h *AttendanceHandler) GetMyAttendance(c *gin.Context) {
	userID := middleware.GetUserID(c)
	ctx := c.Request.Context()

	startDate := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	rows, err := h.db.QueryContext(ctx, `
		SELECT a.id, a.date, a.check_in, a.check_out, a.working_hours, a.overtime_hours, a.status, a.notes
		FROM attendances a
		INNER JOIN employees e ON e.id = a.employee_id
		WHERE e.user_id = $1 AND a.date BETWEEN $2 AND $3
		ORDER BY a.date DESC
	`, userID, startDate, endDate)

	if err != nil {
		response.InternalError(c, err)
		return
	}
	defer rows.Close()

	var attendances []map[string]interface{}
	for rows.Next() {
		var id uuid.UUID
		var date time.Time
		var checkIn, checkOut sql.NullTime
		var workingHours, overtimeHours float64
		var status string
		var notes sql.NullString

		rows.Scan(&id, &date, &checkIn, &checkOut, &workingHours, &overtimeHours, &status, &notes)

		att := map[string]interface{}{
			"id":             id,
			"date":           date.Format("2006-01-02"),
			"working_hours":  workingHours,
			"overtime_hours": overtimeHours,
			"status":         status,
		}
		if checkIn.Valid {
			att["check_in"] = checkIn.Time
		}
		if checkOut.Valid {
			att["check_out"] = checkOut.Time
		}
		if notes.Valid {
			att["notes"] = notes.String
		}
		attendances = append(attendances, att)
	}

	response.OK(c, "common.list", attendances)
}

// List all attendances (for managers/HR)
func (h *AttendanceHandler) List(c *gin.Context) {
	var filter dto.AttendanceFilter
	c.ShouldBindQuery(&filter)

	ctx := c.Request.Context()
	pagination := database.NewPagination(filter.Page, filter.PageSize)

	baseQuery := `
		SELECT a.id, a.employee_id, e.full_name, e.employee_code, a.date, 
		       a.check_in, a.check_out, a.working_hours, a.overtime_hours, a.status, a.notes
		FROM attendances a
		INNER JOIN employees e ON e.id = a.employee_id
		WHERE a.deleted_at IS NULL`

	countQuery := `SELECT COUNT(*) FROM attendances a INNER JOIN employees e ON e.id = a.employee_id WHERE a.deleted_at IS NULL`

	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.EmployeeID != "" {
		conditions = append(conditions, fmt.Sprintf("a.employee_id = $%d", argIdx))
		args = append(args, filter.EmployeeID)
		argIdx++
	}
	if filter.DepartmentID != "" {
		conditions = append(conditions, fmt.Sprintf("e.department_id = $%d", argIdx))
		args = append(args, filter.DepartmentID)
		argIdx++
	}
	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("a.status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.StartDate != "" {
		conditions = append(conditions, fmt.Sprintf("a.date >= $%d", argIdx))
		args = append(args, filter.StartDate)
		argIdx++
	}
	if filter.EndDate != "" {
		conditions = append(conditions, fmt.Sprintf("a.date <= $%d", argIdx))
		args = append(args, filter.EndDate)
		argIdx++
	}

	if len(conditions) > 0 {
		whereClause := " AND " + strings.Join(conditions, " AND ")
		baseQuery += whereClause
		countQuery += whereClause
	}

	var total int
	h.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	pagination.SetTotal(total)

	baseQuery += fmt.Sprintf(" ORDER BY a.date DESC, e.full_name LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, pagination.GetLimit(), pagination.GetOffset())

	rows, err := h.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	defer rows.Close()

	var attendances []dto.AttendanceResponse
	for rows.Next() {
		var att dto.AttendanceResponse
		var checkIn, checkOut sql.NullTime
		var notes sql.NullString

		rows.Scan(&att.ID, &att.EmployeeID, &att.EmployeeName, &att.EmployeeCode, &att.Date,
			&checkIn, &checkOut, &att.WorkingHours, &att.OvertimeHours, &att.Status, &notes)

		if checkIn.Valid {
			att.CheckIn = &checkIn.Time
		}
		if checkOut.Valid {
			att.CheckOut = &checkOut.Time
		}
		if notes.Valid {
			att.Notes = notes.String
		}
		attendances = append(attendances, att)
	}

	response.OKWithMeta(c, "common.list", attendances, pagination)
}

// GetSummary returns attendance summary
func (h *AttendanceHandler) GetSummary(c *gin.Context) {
	ctx := c.Request.Context()
	month := c.DefaultQuery("month", time.Now().Format("2006-01"))
	departmentID := c.Query("department_id")

	startDate := month + "-01"
	endDate := month + "-31"

	query := `
		SELECT 
			COUNT(DISTINCT e.id) as total_employees,
			COUNT(CASE WHEN a.status = 'present' THEN 1 END) as present_count,
			COUNT(CASE WHEN a.status = 'absent' THEN 1 END) as absent_count,
			COUNT(CASE WHEN a.status = 'late' THEN 1 END) as late_count,
			COUNT(CASE WHEN a.status = 'on_leave' THEN 1 END) as leave_count,
			COALESCE(SUM(a.working_hours), 0) as total_working_hours,
			COALESCE(SUM(a.overtime_hours), 0) as total_overtime_hours
		FROM employees e
		LEFT JOIN attendances a ON a.employee_id = e.id AND a.date BETWEEN $1 AND $2
		WHERE e.deleted_at IS NULL AND e.employment_status = 'active'`

	args := []interface{}{startDate, endDate}
	if departmentID != "" {
		query += " AND e.department_id = $3"
		args = append(args, departmentID)
	}

	var summary struct {
		TotalEmployees     int     `json:"total_employees"`
		PresentCount       int     `json:"present_count"`
		AbsentCount        int     `json:"absent_count"`
		LateCount          int     `json:"late_count"`
		LeaveCount         int     `json:"leave_count"`
		TotalWorkingHours  float64 `json:"total_working_hours"`
		TotalOvertimeHours float64 `json:"total_overtime_hours"`
	}

	h.db.QueryRowContext(ctx, query, args...).Scan(
		&summary.TotalEmployees, &summary.PresentCount, &summary.AbsentCount,
		&summary.LateCount, &summary.LeaveCount, &summary.TotalWorkingHours, &summary.TotalOvertimeHours,
	)

	response.OK(c, "common.success", summary)
}

// GetTodayStatus returns today's attendance status
func (h *AttendanceHandler) GetTodayStatus(c *gin.Context) {
	userID := middleware.GetUserID(c)
	ctx := c.Request.Context()
	today := time.Now().Format("2006-01-02")

	var att struct {
		ID           uuid.UUID  `json:"id"`
		CheckIn      *time.Time `json:"check_in"`
		CheckOut     *time.Time `json:"check_out"`
		WorkingHours float64    `json:"working_hours"`
		Status       string     `json:"status"`
	}

	var checkIn, checkOut sql.NullTime

	err := h.db.QueryRowContext(ctx, `
		SELECT a.id, a.check_in, a.check_out, a.working_hours, a.status
		FROM attendances a
		INNER JOIN employees e ON e.id = a.employee_id
		WHERE e.user_id = $1 AND a.date = $2
	`, userID, today).Scan(&att.ID, &checkIn, &checkOut, &att.WorkingHours, &att.Status)

	if err == sql.ErrNoRows {
		response.OK(c, "common.success", gin.H{
			"checked_in":  false,
			"checked_out": false,
		})
		return
	}

	if checkIn.Valid {
		att.CheckIn = &checkIn.Time
	}
	if checkOut.Valid {
		att.CheckOut = &checkOut.Time
	}

	response.OK(c, "common.success", gin.H{
		"id":            att.ID,
		"checked_in":    checkIn.Valid,
		"checked_out":   checkOut.Valid,
		"check_in":      att.CheckIn,
		"check_out":     att.CheckOut,
		"working_hours": att.WorkingHours,
		"status":        att.Status,
	})
}
