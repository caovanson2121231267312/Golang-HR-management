package entity

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Base model với soft delete
type BaseModel struct {
	ID        uuid.UUID    `json:"id" db:"id"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt sql.NullTime `json:"-" db:"deleted_at"`
}

// ==================== USER & AUTH ====================

type User struct {
	BaseModel
	Email              string         `json:"email" db:"email"`
	Phone              string         `json:"phone" db:"phone"`
	Password           string         `json:"-" db:"password"`
	Status             UserStatus     `json:"status" db:"status"`
	EmailVerifiedAt    sql.NullTime   `json:"email_verified_at" db:"email_verified_at"`
	PhoneVerifiedAt    sql.NullTime   `json:"phone_verified_at" db:"phone_verified_at"`
	LastLoginAt        sql.NullTime   `json:"last_login_at" db:"last_login_at"`
	LastLoginIP        sql.NullString `json:"last_login_ip" db:"last_login_ip"`
	FailedLoginAttempts int           `json:"-" db:"failed_login_attempts"`
	LockedUntil        sql.NullTime   `json:"-" db:"locked_until"`
	PasswordChangedAt  sql.NullTime   `json:"password_changed_at" db:"password_changed_at"`
	TwoFactorEnabled   bool           `json:"two_factor_enabled" db:"two_factor_enabled"`
	TwoFactorSecret    sql.NullString `json:"-" db:"two_factor_secret"`
	PreferredLanguage  string         `json:"preferred_language" db:"preferred_language"`
	
	// Relations
	Employee  *Employee `json:"employee,omitempty"`
	Roles     []Role    `json:"roles,omitempty"`
}

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusPending   UserStatus = "pending"
)

type UserSession struct {
	BaseModel
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	Token        string     `json:"-" db:"token"`
	RefreshToken string     `json:"-" db:"refresh_token"`
	UserAgent    string     `json:"user_agent" db:"user_agent"`
	IPAddress    string     `json:"ip_address" db:"ip_address"`
	ExpiresAt    time.Time  `json:"expires_at" db:"expires_at"`
	LastActivity time.Time  `json:"last_activity" db:"last_activity"`
	IsRevoked    bool       `json:"is_revoked" db:"is_revoked"`
}

type OTPCode struct {
	BaseModel
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Code      string    `json:"-" db:"code"`
	Type      OTPType   `json:"type" db:"type"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	UsedAt    sql.NullTime `json:"used_at" db:"used_at"`
	Attempts  int       `json:"attempts" db:"attempts"`
}

type OTPType string

const (
	OTPTypeEmailVerification OTPType = "email_verification"
	OTPTypePhoneVerification OTPType = "phone_verification"
	OTPTypePasswordReset     OTPType = "password_reset"
	OTPTypeLogin             OTPType = "login"
	OTPTypeTwoFactor         OTPType = "two_factor"
)

type PasswordResetToken struct {
	BaseModel
	UserID    uuid.UUID    `json:"user_id" db:"user_id"`
	Token     string       `json:"-" db:"token"`
	ExpiresAt time.Time    `json:"expires_at" db:"expires_at"`
	UsedAt    sql.NullTime `json:"used_at" db:"used_at"`
}

// ==================== RBAC - PERMISSION ====================

type Role struct {
	BaseModel
	Name        string `json:"name" db:"name"`
	Slug        string `json:"slug" db:"slug"`
	Description string `json:"description" db:"description"`
	Level       int    `json:"level" db:"level"`
	IsSystem    bool   `json:"is_system" db:"is_system"`
	
	Permissions []Permission `json:"permissions,omitempty"`
}

type Permission struct {
	BaseModel
	Name        string `json:"name" db:"name"`
	Slug        string `json:"slug" db:"slug"`
	Module      string `json:"module" db:"module"`
	Description string `json:"description" db:"description"`
}

type RolePermission struct {
	RoleID       uuid.UUID `json:"role_id" db:"role_id"`
	PermissionID uuid.UUID `json:"permission_id" db:"permission_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type UserRole struct {
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	RoleID    uuid.UUID `json:"role_id" db:"role_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	CreatedBy uuid.UUID `json:"created_by" db:"created_by"`
}

// ==================== ADDRESS (3 bảng) ====================

type Province struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	NameEn    string    `json:"name_en" db:"name_en"`
	Code      string    `json:"code" db:"code"`
	Type      string    `json:"type" db:"type"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type District struct {
	ID         int       `json:"id" db:"id"`
	ProvinceID int       `json:"province_id" db:"province_id"`
	Name       string    `json:"name" db:"name"`
	NameEn     string    `json:"name_en" db:"name_en"`
	Code       string    `json:"code" db:"code"`
	Type       string    `json:"type" db:"type"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	
	Province *Province `json:"province,omitempty"`
}

type Ward struct {
	ID         int       `json:"id" db:"id"`
	DistrictID int       `json:"district_id" db:"district_id"`
	Name       string    `json:"name" db:"name"`
	NameEn     string    `json:"name_en" db:"name_en"`
	Code       string    `json:"code" db:"code"`
	Type       string    `json:"type" db:"type"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	
	District *District `json:"district,omitempty"`
}

// ==================== EMPLOYEE ====================

type Employee struct {
	BaseModel
	UserID           uuid.UUID      `json:"user_id" db:"user_id"`
	EmployeeCode     string         `json:"employee_code" db:"employee_code"`
	FirstName        string         `json:"first_name" db:"first_name"`
	LastName         string         `json:"last_name" db:"last_name"`
	FullName         string         `json:"full_name" db:"full_name"`
	Gender           Gender         `json:"gender" db:"gender"`
	DateOfBirth      time.Time      `json:"date_of_birth" db:"date_of_birth"`
	PlaceOfBirth     string         `json:"place_of_birth" db:"place_of_birth"`
	Nationality      string         `json:"nationality" db:"nationality"`
	Religion         sql.NullString `json:"religion" db:"religion"`
	MaritalStatus    MaritalStatus  `json:"marital_status" db:"marital_status"`
	IDNumber         string         `json:"id_number" db:"id_number"`
	IDIssuedDate     time.Time      `json:"id_issued_date" db:"id_issued_date"`
	IDIssuedPlace    string         `json:"id_issued_place" db:"id_issued_place"`
	TaxCode          sql.NullString `json:"tax_code" db:"tax_code"`
	SocialInsuranceNo sql.NullString `json:"social_insurance_no" db:"social_insurance_no"`
	HealthInsuranceNo sql.NullString `json:"health_insurance_no" db:"health_insurance_no"`
	BankAccountNo    sql.NullString `json:"bank_account_no" db:"bank_account_no"`
	BankName         sql.NullString `json:"bank_name" db:"bank_name"`
	BankBranch       sql.NullString `json:"bank_branch" db:"bank_branch"`
	Avatar           sql.NullString `json:"avatar" db:"avatar"`
	
	// Address
	PermanentAddress  sql.NullString `json:"permanent_address" db:"permanent_address"`
	PermanentWardID   sql.NullInt32  `json:"permanent_ward_id" db:"permanent_ward_id"`
	CurrentAddress    sql.NullString `json:"current_address" db:"current_address"`
	CurrentWardID     sql.NullInt32  `json:"current_ward_id" db:"current_ward_id"`
	
	// Contact
	PersonalEmail    sql.NullString `json:"personal_email" db:"personal_email"`
	PersonalPhone    sql.NullString `json:"personal_phone" db:"personal_phone"`
	EmergencyContact sql.NullString `json:"emergency_contact" db:"emergency_contact"`
	EmergencyPhone   sql.NullString `json:"emergency_phone" db:"emergency_phone"`
	
	// Employment
	DepartmentID    uuid.UUID        `json:"department_id" db:"department_id"`
	PositionID      uuid.UUID        `json:"position_id" db:"position_id"`
	ManagerID       uuid.NullUUID    `json:"manager_id" db:"manager_id"`
	EmploymentType  EmploymentType   `json:"employment_type" db:"employment_type"`
	EmploymentStatus EmploymentStatus `json:"employment_status" db:"employment_status"`
	JoinDate        time.Time        `json:"join_date" db:"join_date"`
	ProbationEndDate sql.NullTime    `json:"probation_end_date" db:"probation_end_date"`
	ContractStartDate sql.NullTime   `json:"contract_start_date" db:"contract_start_date"`
	ContractEndDate  sql.NullTime    `json:"contract_end_date" db:"contract_end_date"`
	ResignationDate  sql.NullTime    `json:"resignation_date" db:"resignation_date"`
	
	// Salary
	BaseSalary      float64 `json:"base_salary" db:"base_salary"`
	SalaryGrade     string  `json:"salary_grade" db:"salary_grade"`
	
	// Relations
	User       *User       `json:"user,omitempty"`
	Department *Department `json:"department,omitempty"`
	Position   *Position   `json:"position,omitempty"`
	Manager    *Employee   `json:"manager,omitempty"`
}

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
	GenderOther  Gender = "other"
)

type MaritalStatus string

const (
	MaritalStatusSingle   MaritalStatus = "single"
	MaritalStatusMarried  MaritalStatus = "married"
	MaritalStatusDivorced MaritalStatus = "divorced"
	MaritalStatusWidowed  MaritalStatus = "widowed"
)

type EmploymentType string

const (
	EmploymentTypeFullTime   EmploymentType = "full_time"
	EmploymentTypePartTime   EmploymentType = "part_time"
	EmploymentTypeContract   EmploymentType = "contract"
	EmploymentTypeIntern     EmploymentType = "intern"
	EmploymentTypeProbation  EmploymentType = "probation"
)

type EmploymentStatus string

const (
	EmploymentStatusActive     EmploymentStatus = "active"
	EmploymentStatusInactive   EmploymentStatus = "inactive"
	EmploymentStatusOnLeave    EmploymentStatus = "on_leave"
	EmploymentStatusResigned   EmploymentStatus = "resigned"
	EmploymentStatusTerminated EmploymentStatus = "terminated"
)

// ==================== DEPARTMENT & POSITION ====================

type Department struct {
	BaseModel
	Name            string         `json:"name" db:"name"`
	Code            string         `json:"code" db:"code"`
	Description     sql.NullString `json:"description" db:"description"`
	ParentID        uuid.NullUUID  `json:"parent_id" db:"parent_id"`
	ManagerID       uuid.NullUUID  `json:"manager_id" db:"manager_id"`
	Level           int            `json:"level" db:"level"`
	Path            string         `json:"path" db:"path"`
	Status          DeptStatus     `json:"status" db:"status"`
	BudgetCode      sql.NullString `json:"budget_code" db:"budget_code"`
	CostCenter      sql.NullString `json:"cost_center" db:"cost_center"`
	
	Parent      *Department `json:"parent,omitempty"`
	Manager     *Employee   `json:"manager,omitempty"`
	Employees   []Employee  `json:"employees,omitempty"`
}

type DeptStatus string

const (
	DeptStatusActive   DeptStatus = "active"
	DeptStatusInactive DeptStatus = "inactive"
)

type Position struct {
	BaseModel
	Name           string         `json:"name" db:"name"`
	Code           string         `json:"code" db:"code"`
	Description    sql.NullString `json:"description" db:"description"`
	Level          int            `json:"level" db:"level"`
	MinSalary      float64        `json:"min_salary" db:"min_salary"`
	MaxSalary      float64        `json:"max_salary" db:"max_salary"`
	Requirements   sql.NullString `json:"requirements" db:"requirements"`
	Responsibilities sql.NullString `json:"responsibilities" db:"responsibilities"`
	Status         PositionStatus `json:"status" db:"status"`
}

type PositionStatus string

const (
	PositionStatusActive   PositionStatus = "active"
	PositionStatusInactive PositionStatus = "inactive"
)

// ==================== ATTENDANCE ====================

type Attendance struct {
	BaseModel
	EmployeeID     uuid.UUID          `json:"employee_id" db:"employee_id"`
	Date           time.Time          `json:"date" db:"date"`
	CheckIn        sql.NullTime       `json:"check_in" db:"check_in"`
	CheckOut       sql.NullTime       `json:"check_out" db:"check_out"`
	CheckInIP      sql.NullString     `json:"check_in_ip" db:"check_in_ip"`
	CheckOutIP     sql.NullString     `json:"check_out_ip" db:"check_out_ip"`
	CheckInLocation sql.NullString    `json:"check_in_location" db:"check_in_location"`
	CheckOutLocation sql.NullString   `json:"check_out_location" db:"check_out_location"`
	WorkingHours   float64            `json:"working_hours" db:"working_hours"`
	OvertimeHours  float64            `json:"overtime_hours" db:"overtime_hours"`
	Status         AttendanceStatus   `json:"status" db:"status"`
	Notes          sql.NullString     `json:"notes" db:"notes"`
	ApprovedBy     uuid.NullUUID      `json:"approved_by" db:"approved_by"`
	ApprovedAt     sql.NullTime       `json:"approved_at" db:"approved_at"`
	
	Employee   *Employee `json:"employee,omitempty"`
	Approver   *Employee `json:"approver,omitempty"`
}

type AttendanceStatus string

const (
	AttendanceStatusPresent    AttendanceStatus = "present"
	AttendanceStatusAbsent     AttendanceStatus = "absent"
	AttendanceStatusLate       AttendanceStatus = "late"
	AttendanceStatusEarlyLeave AttendanceStatus = "early_leave"
	AttendanceStatusHalfDay    AttendanceStatus = "half_day"
	AttendanceStatusOnLeave    AttendanceStatus = "on_leave"
	AttendanceStatusHoliday    AttendanceStatus = "holiday"
	AttendanceStatusWeekend    AttendanceStatus = "weekend"
)

type AttendanceLog struct {
	ID         uuid.UUID     `json:"id" db:"id"`
	AttendanceID uuid.UUID   `json:"attendance_id" db:"attendance_id"`
	Action     string        `json:"action" db:"action"`
	Timestamp  time.Time     `json:"timestamp" db:"timestamp"`
	IPAddress  string        `json:"ip_address" db:"ip_address"`
	UserAgent  string        `json:"user_agent" db:"user_agent"`
	Location   sql.NullString `json:"location" db:"location"`
	DeviceInfo sql.NullString `json:"device_info" db:"device_info"`
}

type WorkShift struct {
	BaseModel
	Name           string    `json:"name" db:"name"`
	Code           string    `json:"code" db:"code"`
	StartTime      time.Time `json:"start_time" db:"start_time"`
	EndTime        time.Time `json:"end_time" db:"end_time"`
	BreakStart     sql.NullTime `json:"break_start" db:"break_start"`
	BreakEnd       sql.NullTime `json:"break_end" db:"break_end"`
	WorkingHours   float64   `json:"working_hours" db:"working_hours"`
	IsNightShift   bool      `json:"is_night_shift" db:"is_night_shift"`
	Description    sql.NullString `json:"description" db:"description"`
	Status         string    `json:"status" db:"status"`
}

type EmployeeShift struct {
	ID          uuid.UUID `json:"id" db:"id"`
	EmployeeID  uuid.UUID `json:"employee_id" db:"employee_id"`
	ShiftID     uuid.UUID `json:"shift_id" db:"shift_id"`
	Date        time.Time `json:"date" db:"date"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// ==================== LEAVE MANAGEMENT ====================

type LeaveType struct {
	BaseModel
	Name             string  `json:"name" db:"name"`
	Code             string  `json:"code" db:"code"`
	Description      string  `json:"description" db:"description"`
	DefaultDays      int     `json:"default_days" db:"default_days"`
	MaxCarryOver     int     `json:"max_carry_over" db:"max_carry_over"`
	IsPaid           bool    `json:"is_paid" db:"is_paid"`
	RequiresApproval bool    `json:"requires_approval" db:"requires_approval"`
	Color            string  `json:"color" db:"color"`
	Status           string  `json:"status" db:"status"`
}

type LeaveBalance struct {
	BaseModel
	EmployeeID   uuid.UUID `json:"employee_id" db:"employee_id"`
	LeaveTypeID  uuid.UUID `json:"leave_type_id" db:"leave_type_id"`
	Year         int       `json:"year" db:"year"`
	TotalDays    float64   `json:"total_days" db:"total_days"`
	UsedDays     float64   `json:"used_days" db:"used_days"`
	PendingDays  float64   `json:"pending_days" db:"pending_days"`
	CarriedOver  float64   `json:"carried_over" db:"carried_over"`
	
	Employee  *Employee  `json:"employee,omitempty"`
	LeaveType *LeaveType `json:"leave_type,omitempty"`
}

type LeaveRequest struct {
	BaseModel
	EmployeeID    uuid.UUID        `json:"employee_id" db:"employee_id"`
	LeaveTypeID   uuid.UUID        `json:"leave_type_id" db:"leave_type_id"`
	StartDate     time.Time        `json:"start_date" db:"start_date"`
	EndDate       time.Time        `json:"end_date" db:"end_date"`
	TotalDays     float64          `json:"total_days" db:"total_days"`
	Reason        string           `json:"reason" db:"reason"`
	Status        LeaveStatus      `json:"status" db:"status"`
	ApprovedBy    uuid.NullUUID    `json:"approved_by" db:"approved_by"`
	ApprovedAt    sql.NullTime     `json:"approved_at" db:"approved_at"`
	ApproverNotes sql.NullString   `json:"approver_notes" db:"approver_notes"`
	Attachments   sql.NullString   `json:"attachments" db:"attachments"`
	
	Employee  *Employee  `json:"employee,omitempty"`
	LeaveType *LeaveType `json:"leave_type,omitempty"`
	Approver  *Employee  `json:"approver,omitempty"`
}

type LeaveStatus string

const (
	LeaveStatusPending   LeaveStatus = "pending"
	LeaveStatusApproved  LeaveStatus = "approved"
	LeaveStatusRejected  LeaveStatus = "rejected"
	LeaveStatusCancelled LeaveStatus = "cancelled"
)

// ==================== OVERTIME ====================

type OvertimeRequest struct {
	BaseModel
	EmployeeID     uuid.UUID        `json:"employee_id" db:"employee_id"`
	Date           time.Time        `json:"date" db:"date"`
	StartTime      time.Time        `json:"start_time" db:"start_time"`
	EndTime        time.Time        `json:"end_time" db:"end_time"`
	Hours          float64          `json:"hours" db:"hours"`
	Reason         string           `json:"reason" db:"reason"`
	Type           OvertimeType     `json:"type" db:"type"`
	Status         OvertimeStatus   `json:"status" db:"status"`
	ApprovedBy     uuid.NullUUID    `json:"approved_by" db:"approved_by"`
	ApprovedAt     sql.NullTime     `json:"approved_at" db:"approved_at"`
	ApproverNotes  sql.NullString   `json:"approver_notes" db:"approver_notes"`
	Multiplier     float64          `json:"multiplier" db:"multiplier"`
	
	Employee *Employee `json:"employee,omitempty"`
	Approver *Employee `json:"approver,omitempty"`
}

type OvertimeType string

const (
	OvertimeTypeWeekday  OvertimeType = "weekday"
	OvertimeTypeWeekend  OvertimeType = "weekend"
	OvertimeTypeHoliday  OvertimeType = "holiday"
	OvertimeTypeNight    OvertimeType = "night"
)

type OvertimeStatus string

const (
	OvertimeStatusPending   OvertimeStatus = "pending"
	OvertimeStatusApproved  OvertimeStatus = "approved"
	OvertimeStatusRejected  OvertimeStatus = "rejected"
	OvertimeStatusCancelled OvertimeStatus = "cancelled"
	OvertimeStatusCompleted OvertimeStatus = "completed"
)

type OvertimePolicy struct {
	BaseModel
	Name              string  `json:"name" db:"name"`
	WeekdayMultiplier float64 `json:"weekday_multiplier" db:"weekday_multiplier"`
	WeekendMultiplier float64 `json:"weekend_multiplier" db:"weekend_multiplier"`
	HolidayMultiplier float64 `json:"holiday_multiplier" db:"holiday_multiplier"`
	NightMultiplier   float64 `json:"night_multiplier" db:"night_multiplier"`
	MaxHoursPerDay    float64 `json:"max_hours_per_day" db:"max_hours_per_day"`
	MaxHoursPerMonth  float64 `json:"max_hours_per_month" db:"max_hours_per_month"`
	Status            string  `json:"status" db:"status"`
}

// ==================== PAYROLL ====================

type PayrollPeriod struct {
	BaseModel
	Name        string        `json:"name" db:"name"`
	Year        int           `json:"year" db:"year"`
	Month       int           `json:"month" db:"month"`
	StartDate   time.Time     `json:"start_date" db:"start_date"`
	EndDate     time.Time     `json:"end_date" db:"end_date"`
	PayDate     time.Time     `json:"pay_date" db:"pay_date"`
	Status      PayrollStatus `json:"status" db:"status"`
	ProcessedBy uuid.NullUUID `json:"processed_by" db:"processed_by"`
	ProcessedAt sql.NullTime  `json:"processed_at" db:"processed_at"`
	ApprovedBy  uuid.NullUUID `json:"approved_by" db:"approved_by"`
	ApprovedAt  sql.NullTime  `json:"approved_at" db:"approved_at"`
	Notes       sql.NullString `json:"notes" db:"notes"`
}

type PayrollStatus string

const (
	PayrollStatusDraft      PayrollStatus = "draft"
	PayrollStatusProcessing PayrollStatus = "processing"
	PayrollStatusPending    PayrollStatus = "pending"
	PayrollStatusApproved   PayrollStatus = "approved"
	PayrollStatusPaid       PayrollStatus = "paid"
	PayrollStatusCancelled  PayrollStatus = "cancelled"
)

type Payslip struct {
	BaseModel
	EmployeeID        uuid.UUID        `json:"employee_id" db:"employee_id"`
	PayrollPeriodID   uuid.UUID        `json:"payroll_period_id" db:"payroll_period_id"`
	EmployeeCode      string           `json:"employee_code" db:"employee_code"`
	EmployeeName      string           `json:"employee_name" db:"employee_name"`
	DepartmentName    string           `json:"department_name" db:"department_name"`
	PositionName      string           `json:"position_name" db:"position_name"`
	
	// Working Info
	WorkingDays       float64          `json:"working_days" db:"working_days"`
	ActualWorkingDays float64          `json:"actual_working_days" db:"actual_working_days"`
	LeaveDays         float64          `json:"leave_days" db:"leave_days"`
	AbsentDays        float64          `json:"absent_days" db:"absent_days"`
	OvertimeHours     float64          `json:"overtime_hours" db:"overtime_hours"`
	
	// Earnings
	BaseSalary        float64          `json:"base_salary" db:"base_salary"`
	OvertimePay       float64          `json:"overtime_pay" db:"overtime_pay"`
	Allowances        float64          `json:"allowances" db:"allowances"`
	Bonuses           float64          `json:"bonuses" db:"bonuses"`
	OtherEarnings     float64          `json:"other_earnings" db:"other_earnings"`
	GrossEarnings     float64          `json:"gross_earnings" db:"gross_earnings"`
	
	// Deductions
	SocialInsurance   float64          `json:"social_insurance" db:"social_insurance"`
	HealthInsurance   float64          `json:"health_insurance" db:"health_insurance"`
	UnemploymentIns   float64          `json:"unemployment_insurance" db:"unemployment_insurance"`
	PersonalIncomeTax float64          `json:"personal_income_tax" db:"personal_income_tax"`
	OtherDeductions   float64          `json:"other_deductions" db:"other_deductions"`
	TotalDeductions   float64          `json:"total_deductions" db:"total_deductions"`
	
	// Net
	NetSalary         float64          `json:"net_salary" db:"net_salary"`
	
	// Details JSON
	EarningsDetails   string           `json:"earnings_details" db:"earnings_details"`
	DeductionsDetails string           `json:"deductions_details" db:"deductions_details"`
	
	Status            PayslipStatus    `json:"status" db:"status"`
	Notes             sql.NullString   `json:"notes" db:"notes"`
	
	Employee      *Employee      `json:"employee,omitempty"`
	PayrollPeriod *PayrollPeriod `json:"payroll_period,omitempty"`
}

type PayslipStatus string

const (
	PayslipStatusDraft     PayslipStatus = "draft"
	PayslipStatusConfirmed PayslipStatus = "confirmed"
	PayslipStatusPaid      PayslipStatus = "paid"
)

type Allowance struct {
	BaseModel
	Name        string  `json:"name" db:"name"`
	Code        string  `json:"code" db:"code"`
	Type        string  `json:"type" db:"type"`
	Amount      float64 `json:"amount" db:"amount"`
	IsFixed     bool    `json:"is_fixed" db:"is_fixed"`
	IsTaxable   bool    `json:"is_taxable" db:"is_taxable"`
	Description string  `json:"description" db:"description"`
	Status      string  `json:"status" db:"status"`
}

type EmployeeAllowance struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	EmployeeID  uuid.UUID  `json:"employee_id" db:"employee_id"`
	AllowanceID uuid.UUID  `json:"allowance_id" db:"allowance_id"`
	Amount      float64    `json:"amount" db:"amount"`
	StartDate   time.Time  `json:"start_date" db:"start_date"`
	EndDate     sql.NullTime `json:"end_date" db:"end_date"`
	Status      string     `json:"status" db:"status"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type Deduction struct {
	BaseModel
	Name        string  `json:"name" db:"name"`
	Code        string  `json:"code" db:"code"`
	Type        string  `json:"type" db:"type"`
	Percentage  float64 `json:"percentage" db:"percentage"`
	FixedAmount float64 `json:"fixed_amount" db:"fixed_amount"`
	IsRequired  bool    `json:"is_required" db:"is_required"`
	Description string  `json:"description" db:"description"`
	Status      string  `json:"status" db:"status"`
}

type TaxBracket struct {
	ID         uuid.UUID `json:"id" db:"id"`
	MinIncome  float64   `json:"min_income" db:"min_income"`
	MaxIncome  float64   `json:"max_income" db:"max_income"`
	TaxRate    float64   `json:"tax_rate" db:"tax_rate"`
	Deduction  float64   `json:"deduction" db:"deduction"`
	Year       int       `json:"year" db:"year"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// ==================== HOLIDAYS ====================

type Holiday struct {
	BaseModel
	Name        string    `json:"name" db:"name"`
	Date        time.Time `json:"date" db:"date"`
	Type        string    `json:"type" db:"type"`
	Description string    `json:"description" db:"description"`
	IsRecurring bool      `json:"is_recurring" db:"is_recurring"`
	Year        int       `json:"year" db:"year"`
}

// ==================== AUDIT LOG ====================

type AuditLog struct {
	ID          uuid.UUID      `json:"id" db:"id"`
	UserID      uuid.NullUUID  `json:"user_id" db:"user_id"`
	Action      string         `json:"action" db:"action"`
	TableName   string         `json:"table_name" db:"table_name"`
	RecordID    uuid.UUID      `json:"record_id" db:"record_id"`
	OldValues   sql.NullString `json:"old_values" db:"old_values"`
	NewValues   sql.NullString `json:"new_values" db:"new_values"`
	IPAddress   string         `json:"ip_address" db:"ip_address"`
	UserAgent   string         `json:"user_agent" db:"user_agent"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
}

// ==================== NOTIFICATIONS ====================

type Notification struct {
	BaseModel
	UserID     uuid.UUID      `json:"user_id" db:"user_id"`
	Title      string         `json:"title" db:"title"`
	Message    string         `json:"message" db:"message"`
	Type       string         `json:"type" db:"type"`
	Data       sql.NullString `json:"data" db:"data"`
	ReadAt     sql.NullTime   `json:"read_at" db:"read_at"`
	ActionURL  sql.NullString `json:"action_url" db:"action_url"`
}

// ==================== SETTINGS ====================

type SystemSetting struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Key       string    `json:"key" db:"key"`
	Value     string    `json:"value" db:"value"`
	Type      string    `json:"type" db:"type"`
	Group     string    `json:"group" db:"group"`
	Label     string    `json:"label" db:"label"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
