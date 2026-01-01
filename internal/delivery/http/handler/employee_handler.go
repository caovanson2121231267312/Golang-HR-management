package handler

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	"context"
	"hr-management-system/internal/config"
	"hr-management-system/internal/delivery/http/dto"
	"hr-management-system/internal/delivery/http/middleware"
	"hr-management-system/internal/delivery/http/response"
	"hr-management-system/internal/infrastructure/cache"
	"hr-management-system/internal/infrastructure/database"
	"hr-management-system/internal/infrastructure/logger"
	"hr-management-system/internal/infrastructure/queue"
	"hr-management-system/internal/infrastructure/search"
	"hr-management-system/internal/security"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EmployeeHandler struct {
	db    *database.Database
	cache *cache.RedisCache
	queue *queue.Queue
	es    *search.ElasticSearch
	log   *logger.Logger
	cfg   *config.Config
}

func NewEmployeeHandler(db *database.Database, cache *cache.RedisCache, queue *queue.Queue, es *search.ElasticSearch, log *logger.Logger, cfg *config.Config) *EmployeeHandler {
	return &EmployeeHandler{db: db, cache: cache, queue: queue, es: es, log: log, cfg: cfg}
}

func (h *EmployeeHandler) List(c *gin.Context) {
	var filter dto.EmployeeFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.BadRequest(c, "common.validation_error", nil)
		return
	}

	ctx := c.Request.Context()
	pagination := database.NewPagination(filter.Page, filter.PageSize)

	baseQuery := `
		SELECT e.id, e.user_id, e.employee_code, e.first_name, e.last_name, e.full_name,
		       e.gender, e.date_of_birth, e.id_number, e.department_id, d.name,
		       e.position_id, p.name, e.manager_id, COALESCE(m.full_name, ''),
		       e.employment_type, e.employment_status, e.join_date, e.base_salary,
		       e.avatar, e.created_at, e.updated_at, u.email, u.phone
		FROM employees e
		INNER JOIN users u ON u.id = e.user_id
		INNER JOIN departments d ON d.id = e.department_id
		INNER JOIN positions p ON p.id = e.position_id
		LEFT JOIN employees m ON m.id = e.manager_id
		WHERE e.deleted_at IS NULL`

	countQuery := `SELECT COUNT(*) FROM employees e WHERE e.deleted_at IS NULL`
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(e.full_name ILIKE $%d OR e.employee_code ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}
	if filter.DepartmentID != "" {
		conditions = append(conditions, fmt.Sprintf("e.department_id = $%d", argIdx))
		args = append(args, filter.DepartmentID)
		argIdx++
	}
	if filter.EmploymentStatus != "" {
		conditions = append(conditions, fmt.Sprintf("e.employment_status = $%d", argIdx))
		args = append(args, filter.EmploymentStatus)
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

	baseQuery += fmt.Sprintf(" ORDER BY e.created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, pagination.GetLimit(), pagination.GetOffset())

	rows, err := h.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	defer rows.Close()

	var employees []dto.EmployeeResponse
	for rows.Next() {
		var emp dto.EmployeeResponse
		var managerID, avatar sql.NullString
		rows.Scan(&emp.ID, &emp.UserID, &emp.EmployeeCode, &emp.FirstName, &emp.LastName,
			&emp.FullName, &emp.Gender, &emp.DateOfBirth, &emp.IDNumber,
			&emp.DepartmentID, &emp.DepartmentName, &emp.PositionID, &emp.PositionName,
			&managerID, &emp.ManagerName, &emp.EmploymentType, &emp.EmploymentStatus,
			&emp.JoinDate, &emp.BaseSalary, &avatar, &emp.CreatedAt, &emp.UpdatedAt,
			&emp.Email, &emp.Phone)
		if managerID.Valid {
			id, _ := uuid.Parse(managerID.String)
			emp.ManagerID = &id
		}
		if avatar.Valid {
			emp.Avatar = avatar.String
		}
		employees = append(employees, emp)
	}

	response.OKWithMeta(c, "common.list", employees, pagination)
}

func (h *EmployeeHandler) Get(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	cacheKey := "employee:" + id
	var emp dto.EmployeeResponse
	if err := h.cache.Get(ctx, cacheKey, &emp); err == nil {
		response.OK(c, "common.success", emp)
		return
	}

	query := `
		SELECT e.id, e.user_id, e.employee_code, e.first_name, e.last_name, e.full_name,
		       e.gender, e.date_of_birth, e.id_number, e.department_id, d.name,
		       e.position_id, p.name, e.manager_id, COALESCE(m.full_name, ''),
		       e.employment_type, e.employment_status, e.join_date, e.base_salary,
		       e.avatar, e.created_at, e.updated_at, u.email, u.phone
		FROM employees e
		INNER JOIN users u ON u.id = e.user_id
		INNER JOIN departments d ON d.id = e.department_id
		INNER JOIN positions p ON p.id = e.position_id
		LEFT JOIN employees m ON m.id = e.manager_id
		WHERE e.id = $1 AND e.deleted_at IS NULL`

	var managerID, avatar sql.NullString
	err := h.db.QueryRowContext(ctx, query, id).Scan(
		&emp.ID, &emp.UserID, &emp.EmployeeCode, &emp.FirstName, &emp.LastName,
		&emp.FullName, &emp.Gender, &emp.DateOfBirth, &emp.IDNumber,
		&emp.DepartmentID, &emp.DepartmentName, &emp.PositionID, &emp.PositionName,
		&managerID, &emp.ManagerName, &emp.EmploymentType, &emp.EmploymentStatus,
		&emp.JoinDate, &emp.BaseSalary, &avatar, &emp.CreatedAt, &emp.UpdatedAt,
		&emp.Email, &emp.Phone)

	if err == sql.ErrNoRows {
		response.NotFound(c, "employee.not_found")
		return
	}
	if err != nil {
		response.InternalError(c, err)
		return
	}

	if managerID.Valid {
		mid, _ := uuid.Parse(managerID.String)
		emp.ManagerID = &mid
	}
	if avatar.Valid {
		emp.Avatar = avatar.String
	}

	h.cache.Set(ctx, cacheKey, emp, 15*time.Minute)
	response.OK(c, "common.success", emp)
}

func (h *EmployeeHandler) Create(c *gin.Context) {
	var req dto.CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "common.validation_error", nil)
		return
	}

	ctx := c.Request.Context()
	currentUserID := middleware.GetUserID(c)

	var exists bool
	h.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, req.Email).Scan(&exists)
	if exists {
		response.Conflict(c, "user.email_exists")
		return
	}

	employeeCode, err := h.generateEmployeeCode(ctx)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	dateOfBirth, _ := time.Parse("2006-01-02", req.DateOfBirth)
	joinDate, _ := time.Parse("2006-01-02", req.JoinDate)
	idIssuedDate, _ := time.Parse("2006-01-02", req.IDIssuedDate)

	tempPassword, _ := security.GenerateSecureToken(8)
	hashedPassword, _ := security.HashPassword(tempPassword)

	tx, err := h.db.Begin()
	if err != nil {
		response.InternalError(c, err)
		return
	}
	defer tx.Rollback()

	userID := uuid.New()
	tx.ExecContext(ctx, `
		INSERT INTO users (id, email, phone, password, status, preferred_language, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'active', 'vi', NOW(), NOW())`,
		userID, req.Email, req.Phone, hashedPassword)

	employeeID := uuid.New()
	fullName := req.FirstName + " " + req.LastName
	deptID, _ := uuid.Parse(req.DepartmentID)
	posID, _ := uuid.Parse(req.PositionID)

	var managerID *uuid.UUID
	if req.ManagerID != "" {
		mid, _ := uuid.Parse(req.ManagerID)
		managerID = &mid
	}

	tx.ExecContext(ctx, `
		INSERT INTO employees (id, user_id, employee_code, first_name, last_name, full_name, gender,
			date_of_birth, place_of_birth, nationality, marital_status, id_number, id_issued_date,
			id_issued_place, department_id, position_id, manager_id, employment_type, employment_status,
			join_date, base_salary, salary_grade, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,'active',$19,$20,$21,NOW(),NOW())`,
		employeeID, userID, employeeCode, req.FirstName, req.LastName, fullName, req.Gender,
		dateOfBirth, req.PlaceOfBirth, req.Nationality, req.MaritalStatus, req.IDNumber, idIssuedDate,
		req.IDIssuedPlace, deptID, posID, managerID, req.EmploymentType, joinDate, req.BaseSalary, req.SalaryGrade)

	for _, roleID := range req.RoleIDs {
		rid, _ := uuid.Parse(roleID)
		tx.ExecContext(ctx, `INSERT INTO user_roles (user_id, role_id, created_at, created_by) VALUES ($1, $2, NOW(), $3)`,
			userID, rid, currentUserID)
	}

	if err := tx.Commit(); err != nil {
		response.InternalError(c, err)
		return
	}

	h.queue.IndexDocument(ctx, queue.ElasticPayload{
		Index: "employees", DocumentID: employeeID.String(),
		Document: map[string]interface{}{"id": employeeID.String(), "employee_code": employeeCode, "full_name": fullName,
			"email": req.Email, "department_id": req.DepartmentID, "employment_status": "active", "created_at": time.Now()},
		Action: "index",
	})

	h.queue.LogAudit(ctx, queue.AuditLogPayload{
		UserID: currentUserID, Action: "create", TableName: "employees", RecordID: employeeID.String(),
		NewValues: req, IPAddress: c.ClientIP(), UserAgent: c.Request.UserAgent(),
	})

	response.Created(c, "employee.created", gin.H{"id": employeeID, "employee_code": employeeCode, "temp_password": tempPassword})
}

func (h *EmployeeHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "common.validation_error", nil)
		return
	}

	ctx := c.Request.Context()
	currentUserID := middleware.GetUserID(c)

	var exists bool
	h.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM employees WHERE id = $1 AND deleted_at IS NULL)`, id).Scan(&exists)
	if !exists {
		response.NotFound(c, "employee.not_found")
		return
	}

	updates := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIdx := 1

	if req.FirstName != nil {
		updates = append(updates, fmt.Sprintf("first_name = $%d", argIdx))
		args = append(args, *req.FirstName)
		argIdx++
	}
	if req.LastName != nil {
		updates = append(updates, fmt.Sprintf("last_name = $%d", argIdx))
		args = append(args, *req.LastName)
		argIdx++
	}
	if req.DepartmentID != nil {
		updates = append(updates, fmt.Sprintf("department_id = $%d", argIdx))
		args = append(args, *req.DepartmentID)
		argIdx++
	}
	if req.PositionID != nil {
		updates = append(updates, fmt.Sprintf("position_id = $%d", argIdx))
		args = append(args, *req.PositionID)
		argIdx++
	}
	if req.BaseSalary != nil {
		updates = append(updates, fmt.Sprintf("base_salary = $%d", argIdx))
		args = append(args, *req.BaseSalary)
		argIdx++
	}
	if req.EmploymentStatus != nil {
		updates = append(updates, fmt.Sprintf("employment_status = $%d", argIdx))
		args = append(args, *req.EmploymentStatus)
		argIdx++
	}

	args = append(args, id)
	query := fmt.Sprintf(`UPDATE employees SET %s WHERE id = $%d`, strings.Join(updates, ", "), argIdx)
	h.db.ExecContext(ctx, query, args...)

	h.cache.Delete(ctx, "employee:"+id)

	h.queue.LogAudit(ctx, queue.AuditLogPayload{
		UserID: currentUserID, Action: "update", TableName: "employees", RecordID: id,
		NewValues: req, IPAddress: c.ClientIP(), UserAgent: c.Request.UserAgent(),
	})

	response.OK(c, "employee.updated", nil)
}

func (h *EmployeeHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()
	currentUserID := middleware.GetUserID(c)

	result, err := h.db.ExecContext(ctx, `UPDATE employees SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		response.NotFound(c, "employee.not_found")
		return
	}

	h.db.ExecContext(ctx, `UPDATE users SET status = 'inactive' WHERE id = (SELECT user_id FROM employees WHERE id = $1)`, id)
	h.cache.Delete(ctx, "employee:"+id)

	h.queue.IndexDocument(ctx, queue.ElasticPayload{Index: "employees", DocumentID: id, Action: "delete"})
	h.queue.LogAudit(ctx, queue.AuditLogPayload{
		UserID: currentUserID, Action: "delete", TableName: "employees", RecordID: id,
		IPAddress: c.ClientIP(), UserAgent: c.Request.UserAgent(),
	})

	response.OK(c, "employee.deleted", nil)
}

func (h *EmployeeHandler) Search(c *gin.Context) {
	query := c.Query("q")
	page, size := 1, 20
	fmt.Sscanf(c.DefaultQuery("page", "1"), "%d", &page)
	fmt.Sscanf(c.DefaultQuery("size", "20"), "%d", &size)

	filters := make(map[string]interface{})
	if dept := c.Query("department_id"); dept != "" {
		filters["department_id"] = dept
	}

	ctx := c.Request.Context()
	result, err := h.es.SearchEmployees(ctx, query, filters, page, size)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, "common.success", gin.H{"total": result.Total, "hits": result.Hits, "page": page, "size": size})
}

func (h *EmployeeHandler) generateEmployeeCode(ctx context.Context) (string, error) {
	var lastCode string

	err := h.db.QueryRowContext(ctx, `
		SELECT employee_code
		FROM employees
		WHERE employee_code LIKE 'NV%'
		ORDER BY employee_code DESC
		LIMIT 1
	`).Scan(&lastCode)

	num := 1

	if err != nil {
		if err != sql.ErrNoRows {
			return "", err
		}
	} else {
		fmt.Sscanf(lastCode, "NV%d", &num)
		num++
	}

	return fmt.Sprintf("NV%06d", num), nil
}