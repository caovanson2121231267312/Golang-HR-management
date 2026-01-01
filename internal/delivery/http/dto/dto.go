package dto

import (
	"time"

	"github.com/google/uuid"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Remember bool   `json:"remember"`
}

type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    time.Time    `json:"expires_at"`
	TokenType    string       `json:"token_type"`
	User         UserResponse `json:"user"`
}

type RegisterRequest struct {
	Email           string `json:"email" binding:"required,email"`
	Phone           string `json:"phone" binding:"required"`
	Password        string `json:"password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
	FirstName       string `json:"first_name" binding:"required"`
	LastName        string `json:"last_name" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token           string `json:"token" binding:"required"`
	Password        string `json:"password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
}

type SendOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	Type  string `json:"type" binding:"required,oneof=email_verification phone_verification password_reset login two_factor"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=6"`
	Type  string `json:"type" binding:"required"`
}

type TwoFactorRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

// ==================== USER ====================

type UserResponse struct {
	ID                uuid.UUID  `json:"id"`
	Email             string     `json:"email"`
	Phone             string     `json:"phone"`
	Status            string     `json:"status"`
	EmailVerifiedAt   *time.Time `json:"email_verified_at,omitempty"`
	LastLoginAt       *time.Time `json:"last_login_at,omitempty"`
	TwoFactorEnabled  bool       `json:"two_factor_enabled"`
	PreferredLanguage string     `json:"preferred_language"`
	Roles             []string   `json:"roles"`
	Permissions       []string   `json:"permissions"`
	CreatedAt         time.Time  `json:"created_at"`
}

type CreateUserRequest struct {
	Email    string   `json:"email" binding:"required,email"`
	Phone    string   `json:"phone" binding:"required"`
	Password string   `json:"password" binding:"required,min=8"`
	RoleIDs  []string `json:"role_ids"`
}

type UpdateUserRequest struct {
	Phone             *string  `json:"phone"`
	Status            *string  `json:"status"`
	PreferredLanguage *string  `json:"preferred_language"`
	TwoFactorEnabled  *bool    `json:"two_factor_enabled"`
	RoleIDs           []string `json:"role_ids"`
}

// ==================== EMPLOYEE ====================

type EmployeeResponse struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"user_id"`
	EmployeeCode     string     `json:"employee_code"`
	FirstName        string     `json:"first_name"`
	LastName         string     `json:"last_name"`
	FullName         string     `json:"full_name"`
	Gender           string     `json:"gender"`
	DateOfBirth      time.Time  `json:"date_of_birth"`
	Email            string     `json:"email"`
	Phone            string     `json:"phone"`
	IDNumber         string     `json:"id_number"`
	DepartmentID     uuid.UUID  `json:"department_id"`
	DepartmentName   string     `json:"department_name,omitempty"`
	PositionID       uuid.UUID  `json:"position_id"`
	PositionName     string     `json:"position_name,omitempty"`
	ManagerID        *uuid.UUID `json:"manager_id,omitempty"`
	ManagerName      string     `json:"manager_name,omitempty"`
	EmploymentType   string     `json:"employment_type"`
	EmploymentStatus string     `json:"employment_status"`
	JoinDate         time.Time  `json:"join_date"`
	BaseSalary       float64    `json:"base_salary"`
	Avatar           string     `json:"avatar,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type CreateEmployeeRequest struct {
	Email            string    `json:"email" binding:"required,email"`
	Phone            string    `json:"phone" binding:"required"`
	FirstName        string    `json:"first_name" binding:"required"`
	LastName         string    `json:"last_name" binding:"required"`
	Gender           string    `json:"gender" binding:"required,oneof=male female other"`
	DateOfBirth      string    `json:"date_of_birth" binding:"required"`
	PlaceOfBirth     string    `json:"place_of_birth"`
	Nationality      string    `json:"nationality"`
	MaritalStatus    string    `json:"marital_status" binding:"oneof=single married divorced widowed"`
	IDNumber         string    `json:"id_number" binding:"required"`
	IDIssuedDate     string    `json:"id_issued_date"`
	IDIssuedPlace    string    `json:"id_issued_place"`
	TaxCode          string    `json:"tax_code"`
	SocialInsuranceNo string   `json:"social_insurance_no"`
	BankAccountNo    string    `json:"bank_account_no"`
	BankName         string    `json:"bank_name"`
	BankBranch       string    `json:"bank_branch"`
	PermanentAddress string    `json:"permanent_address"`
	PermanentWardID  int       `json:"permanent_ward_id"`
	CurrentAddress   string    `json:"current_address"`
	CurrentWardID    int       `json:"current_ward_id"`
	PersonalEmail    string    `json:"personal_email"`
	PersonalPhone    string    `json:"personal_phone"`
	EmergencyContact string    `json:"emergency_contact"`
	EmergencyPhone   string    `json:"emergency_phone"`
	DepartmentID     string    `json:"department_id" binding:"required,uuid"`
	PositionID       string    `json:"position_id" binding:"required,uuid"`
	ManagerID        string    `json:"manager_id"`
	EmploymentType   string    `json:"employment_type" binding:"required,oneof=full_time part_time contract intern probation"`
	JoinDate         string    `json:"join_date" binding:"required"`
	ProbationEndDate string    `json:"probation_end_date"`
	BaseSalary       float64   `json:"base_salary" binding:"required,min=0"`
	SalaryGrade      string    `json:"salary_grade"`
	RoleIDs          []string  `json:"role_ids"`
}

type UpdateEmployeeRequest struct {
	FirstName        *string  `json:"first_name"`
	LastName         *string  `json:"last_name"`
	Phone            *string  `json:"phone"`
	Gender           *string  `json:"gender"`
	MaritalStatus    *string  `json:"marital_status"`
	DepartmentID     *string  `json:"department_id"`
	PositionID       *string  `json:"position_id"`
	ManagerID        *string  `json:"manager_id"`
	EmploymentType   *string  `json:"employment_type"`
	EmploymentStatus *string  `json:"employment_status"`
	BaseSalary       *float64 `json:"base_salary"`
	SalaryGrade      *string  `json:"salary_grade"`
	BankAccountNo    *string  `json:"bank_account_no"`
	BankName         *string  `json:"bank_name"`
	BankBranch       *string  `json:"bank_branch"`
	CurrentAddress   *string  `json:"current_address"`
	CurrentWardID    *int     `json:"current_ward_id"`
}

type EmployeeFilter struct {
	Search           string `form:"search"`
	DepartmentID     string `form:"department_id"`
	PositionID       string `form:"position_id"`
	EmploymentType   string `form:"employment_type"`
	EmploymentStatus string `form:"employment_status"`
	Page             int    `form:"page,default=1"`
	PageSize         int    `form:"page_size,default=20"`
	SortBy           string `form:"sort_by,default=created_at"`
	SortOrder        string `form:"sort_order,default=desc"`
}

// ==================== DEPARTMENT ====================

type DepartmentResponse struct {
	ID          uuid.UUID             `json:"id"`
	Name        string                `json:"name"`
	Code        string                `json:"code"`
	Description string                `json:"description,omitempty"`
	ParentID    *uuid.UUID            `json:"parent_id,omitempty"`
	ParentName  string                `json:"parent_name,omitempty"`
	ManagerID   *uuid.UUID            `json:"manager_id,omitempty"`
	ManagerName string                `json:"manager_name,omitempty"`
	Level       int                   `json:"level"`
	Status      string                `json:"status"`
	EmployeeCount int                 `json:"employee_count"`
	Children    []DepartmentResponse  `json:"children,omitempty"`
	CreatedAt   time.Time             `json:"created_at"`
}

type CreateDepartmentRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	ParentID    string `json:"parent_id"`
	ManagerID   string `json:"manager_id"`
	BudgetCode  string `json:"budget_code"`
	CostCenter  string `json:"cost_center"`
}

type UpdateDepartmentRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	ParentID    *string `json:"parent_id"`
	ManagerID   *string `json:"manager_id"`
	Status      *string `json:"status"`
	BudgetCode  *string `json:"budget_code"`
	CostCenter  *string `json:"cost_center"`
}

// ==================== POSITION ====================

type PositionResponse struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Code           string    `json:"code"`
	Description    string    `json:"description,omitempty"`
	Level          int       `json:"level"`
	MinSalary      float64   `json:"min_salary"`
	MaxSalary      float64   `json:"max_salary"`
	Status         string    `json:"status"`
	EmployeeCount  int       `json:"employee_count"`
	CreatedAt      time.Time `json:"created_at"`
}

type CreatePositionRequest struct {
	Name           string  `json:"name" binding:"required"`
	Code           string  `json:"code" binding:"required"`
	Description    string  `json:"description"`
	Level          int     `json:"level" binding:"min=1"`
	MinSalary      float64 `json:"min_salary" binding:"min=0"`
	MaxSalary      float64 `json:"max_salary" binding:"gtefield=MinSalary"`
	Requirements   string  `json:"requirements"`
	Responsibilities string `json:"responsibilities"`
}

// ==================== ATTENDANCE ====================

type AttendanceResponse struct {
	ID              uuid.UUID  `json:"id"`
	EmployeeID      uuid.UUID  `json:"employee_id"`
	EmployeeName    string     `json:"employee_name"`
	EmployeeCode    string     `json:"employee_code"`
	Date            time.Time  `json:"date"`
	CheckIn         *time.Time `json:"check_in"`
	CheckOut        *time.Time `json:"check_out"`
	WorkingHours    float64    `json:"working_hours"`
	OvertimeHours   float64    `json:"overtime_hours"`
	Status          string     `json:"status"`
	Notes           string     `json:"notes,omitempty"`
	ApprovedBy      *uuid.UUID `json:"approved_by,omitempty"`
	ApproverName    string     `json:"approver_name,omitempty"`
}

type CheckInRequest struct {
	Location  string `json:"location"`
	Notes     string `json:"notes"`
}

type CheckOutRequest struct {
	Location string `json:"location"`
	Notes    string `json:"notes"`
}

type AttendanceFilter struct {
	EmployeeID   string `form:"employee_id"`
	DepartmentID string `form:"department_id"`
	Status       string `form:"status"`
	StartDate    string `form:"start_date"`
	EndDate      string `form:"end_date"`
	Page         int    `form:"page,default=1"`
	PageSize     int    `form:"page_size,default=20"`
}

// ==================== LEAVE ====================

type LeaveRequestResponse struct {
	ID            uuid.UUID  `json:"id"`
	EmployeeID    uuid.UUID  `json:"employee_id"`
	EmployeeName  string     `json:"employee_name"`
	LeaveTypeID   uuid.UUID  `json:"leave_type_id"`
	LeaveTypeName string     `json:"leave_type_name"`
	StartDate     time.Time  `json:"start_date"`
	EndDate       time.Time  `json:"end_date"`
	TotalDays     float64    `json:"total_days"`
	Reason        string     `json:"reason"`
	Status        string     `json:"status"`
	ApprovedBy    *uuid.UUID `json:"approved_by,omitempty"`
	ApproverName  string     `json:"approver_name,omitempty"`
	ApprovedAt    *time.Time `json:"approved_at,omitempty"`
	ApproverNotes string     `json:"approver_notes,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type CreateLeaveRequest struct {
	LeaveTypeID string `json:"leave_type_id" binding:"required,uuid"`
	StartDate   string `json:"start_date" binding:"required"`
	EndDate     string `json:"end_date" binding:"required"`
	Reason      string `json:"reason" binding:"required"`
	Attachments string `json:"attachments"`
}

type ApproveLeaveRequest struct {
	Status string `json:"status" binding:"required,oneof=approved rejected"`
	Notes  string `json:"notes"`
}

type LeaveBalanceResponse struct {
	LeaveTypeID   uuid.UUID `json:"leave_type_id"`
	LeaveTypeName string    `json:"leave_type_name"`
	Year          int       `json:"year"`
	TotalDays     float64   `json:"total_days"`
	UsedDays      float64   `json:"used_days"`
	PendingDays   float64   `json:"pending_days"`
	RemainingDays float64   `json:"remaining_days"`
}

// ==================== OVERTIME ====================

type OvertimeRequestResponse struct {
	ID            uuid.UUID  `json:"id"`
	EmployeeID    uuid.UUID  `json:"employee_id"`
	EmployeeName  string     `json:"employee_name"`
	Date          time.Time  `json:"date"`
	StartTime     time.Time  `json:"start_time"`
	EndTime       time.Time  `json:"end_time"`
	Hours         float64    `json:"hours"`
	Type          string     `json:"type"`
	Multiplier    float64    `json:"multiplier"`
	Reason        string     `json:"reason"`
	Status        string     `json:"status"`
	ApprovedBy    *uuid.UUID `json:"approved_by,omitempty"`
	ApproverName  string     `json:"approver_name,omitempty"`
	ApprovedAt    *time.Time `json:"approved_at,omitempty"`
	ApproverNotes string     `json:"approver_notes,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type CreateOvertimeRequest struct {
	Date      string `json:"date" binding:"required"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
	Type      string `json:"type" binding:"required,oneof=weekday weekend holiday night"`
	Reason    string `json:"reason" binding:"required"`
}

type ApproveOvertimeRequest struct {
	Status string `json:"status" binding:"required,oneof=approved rejected"`
	Notes  string `json:"notes"`
}

// ==================== PAYROLL ====================

type PayrollPeriodResponse struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Year        int        `json:"year"`
	Month       int        `json:"month"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     time.Time  `json:"end_date"`
	PayDate     time.Time  `json:"pay_date"`
	Status      string     `json:"status"`
	ProcessedBy *uuid.UUID `json:"processed_by,omitempty"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	ApprovedBy  *uuid.UUID `json:"approved_by,omitempty"`
	ApprovedAt  *time.Time `json:"approved_at,omitempty"`
	TotalAmount float64    `json:"total_amount"`
	EmployeeCount int      `json:"employee_count"`
	CreatedAt   time.Time  `json:"created_at"`
}

type CreatePayrollPeriodRequest struct {
	Year      int    `json:"year" binding:"required"`
	Month     int    `json:"month" binding:"required,min=1,max=12"`
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
	PayDate   string `json:"pay_date" binding:"required"`
}

type PayslipResponse struct {
	ID                uuid.UUID `json:"id"`
	EmployeeID        uuid.UUID `json:"employee_id"`
	EmployeeCode      string    `json:"employee_code"`
	EmployeeName      string    `json:"employee_name"`
	DepartmentName    string    `json:"department_name"`
	PositionName      string    `json:"position_name"`
	PeriodID          uuid.UUID `json:"period_id"`
	Year              int       `json:"year"`
	Month             int       `json:"month"`
	WorkingDays       float64   `json:"working_days"`
	ActualWorkingDays float64   `json:"actual_working_days"`
	LeaveDays         float64   `json:"leave_days"`
	AbsentDays        float64   `json:"absent_days"`
	OvertimeHours     float64   `json:"overtime_hours"`
	BaseSalary        float64   `json:"base_salary"`
	OvertimePay       float64   `json:"overtime_pay"`
	Allowances        float64   `json:"allowances"`
	Bonuses           float64   `json:"bonuses"`
	OtherEarnings     float64   `json:"other_earnings"`
	GrossEarnings     float64   `json:"gross_earnings"`
	SocialInsurance   float64   `json:"social_insurance"`
	HealthInsurance   float64   `json:"health_insurance"`
	UnemploymentIns   float64   `json:"unemployment_insurance"`
	PersonalIncomeTax float64   `json:"personal_income_tax"`
	OtherDeductions   float64   `json:"other_deductions"`
	TotalDeductions   float64   `json:"total_deductions"`
	NetSalary         float64   `json:"net_salary"`
	Status            string    `json:"status"`
}

// ==================== ROLE & PERMISSION ====================

type RoleResponse struct {
	ID          uuid.UUID            `json:"id"`
	Name        string               `json:"name"`
	Slug        string               `json:"slug"`
	Description string               `json:"description"`
	Level       int                  `json:"level"`
	IsSystem    bool                 `json:"is_system"`
	Permissions []PermissionResponse `json:"permissions,omitempty"`
	UserCount   int                  `json:"user_count"`
	CreatedAt   time.Time            `json:"created_at"`
}

type CreateRoleRequest struct {
	Name          string   `json:"name" binding:"required"`
	Slug          string   `json:"slug" binding:"required"`
	Description   string   `json:"description"`
	Level         int      `json:"level" binding:"min=1"`
	PermissionIDs []string `json:"permission_ids"`
}

type UpdateRoleRequest struct {
	Name          *string  `json:"name"`
	Description   *string  `json:"description"`
	Level         *int     `json:"level"`
	PermissionIDs []string `json:"permission_ids"`
}

type PermissionResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Module      string    `json:"module"`
	Description string    `json:"description"`
}

// ==================== ADDRESS ====================

type ProvinceResponse struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	NameEn string `json:"name_en"`
	Code   string `json:"code"`
	Type   string `json:"type"`
}

type DistrictResponse struct {
	ID         int    `json:"id"`
	ProvinceID int    `json:"province_id"`
	Name       string `json:"name"`
	NameEn     string `json:"name_en"`
	Code       string `json:"code"`
	Type       string `json:"type"`
}

type WardResponse struct {
	ID         int    `json:"id"`
	DistrictID int    `json:"district_id"`
	Name       string `json:"name"`
	NameEn     string `json:"name_en"`
	Code       string `json:"code"`
	Type       string `json:"type"`
}

// ==================== REPORT ====================

type ReportRequest struct {
	ReportType string            `json:"report_type" binding:"required"`
	Format     string            `json:"format" binding:"required,oneof=pdf excel csv"`
	StartDate  string            `json:"start_date"`
	EndDate    string            `json:"end_date"`
	Filters    map[string]string `json:"filters"`
	SendEmail  bool              `json:"send_email"`
	Email      string            `json:"email"`
}

type ReportResponse struct {
	ID         string `json:"id"`
	ReportType string `json:"report_type"`
	Status     string `json:"status"`
	FileURL    string `json:"file_url,omitempty"`
	CreatedAt  string `json:"created_at"`
}

// ==================== NOTIFICATION ====================

type NotificationResponse struct {
	ID        uuid.UUID  `json:"id"`
	Title     string     `json:"title"`
	Message   string     `json:"message"`
	Type      string     `json:"type"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
	ActionURL string     `json:"action_url,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}
