-- HR Management System Database Schema
-- PostgreSQL Migration

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ==================== USERS & AUTH ====================

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    phone VARCHAR(20) NOT NULL,
    password VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('active', 'inactive', 'suspended', 'pending')),
    email_verified_at TIMESTAMP,
    phone_verified_at TIMESTAMP,
    last_login_at TIMESTAMP,
    last_login_ip VARCHAR(45),
    failed_login_attempts INT DEFAULT 0,
    locked_until TIMESTAMP,
    password_changed_at TIMESTAMP,
    two_factor_enabled BOOLEAN DEFAULT FALSE,
    two_factor_secret VARCHAR(255),
    preferred_language VARCHAR(10) DEFAULT 'vi',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_status ON users(status) WHERE deleted_at IS NULL;

CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL,
    refresh_token VARCHAR(255) NOT NULL,
    user_agent TEXT,
    ip_address VARCHAR(45),
    expires_at TIMESTAMP NOT NULL,
    last_activity TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_revoked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_expires ON user_sessions(expires_at) WHERE is_revoked = FALSE;

CREATE TABLE otp_codes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code VARCHAR(10) NOT NULL,
    type VARCHAR(30) NOT NULL CHECK (type IN ('email_verification', 'phone_verification', 'password_reset', 'login', 'two_factor')),
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    attempts INT DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_otp_codes_user_type ON otp_codes(user_id, type) WHERE used_at IS NULL;

CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- ==================== RBAC ====================

CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    level INT DEFAULT 1,
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    module VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id),
    PRIMARY KEY (user_id, role_id)
);

-- ==================== ADDRESS ====================

CREATE TABLE provinces (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    name_en VARCHAR(100),
    code VARCHAR(20) NOT NULL UNIQUE,
    type VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE districts (
    id SERIAL PRIMARY KEY,
    province_id INT NOT NULL REFERENCES provinces(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    name_en VARCHAR(100),
    code VARCHAR(20) NOT NULL UNIQUE,
    type VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_districts_province ON districts(province_id);

CREATE TABLE wards (
    id SERIAL PRIMARY KEY,
    district_id INT NOT NULL REFERENCES districts(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    name_en VARCHAR(100),
    code VARCHAR(20) NOT NULL UNIQUE,
    type VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_wards_district ON wards(district_id);

-- ==================== DEPARTMENTS & POSITIONS ====================

CREATE TABLE departments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    parent_id UUID REFERENCES departments(id),
    manager_id UUID,
    level INT DEFAULT 1,
    path TEXT,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    budget_code VARCHAR(50),
    cost_center VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_departments_parent ON departments(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_departments_code ON departments(code) WHERE deleted_at IS NULL;

CREATE TABLE positions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    level INT DEFAULT 1,
    min_salary DECIMAL(15,2) DEFAULT 0,
    max_salary DECIMAL(15,2) DEFAULT 0,
    requirements TEXT,
    responsibilities TEXT,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- ==================== EMPLOYEES ====================

CREATE TABLE employees (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    employee_code VARCHAR(20) NOT NULL UNIQUE,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    gender VARCHAR(10) CHECK (gender IN ('male', 'female', 'other')),
    date_of_birth DATE NOT NULL,
    place_of_birth VARCHAR(255),
    nationality VARCHAR(100) DEFAULT 'Việt Nam',
    religion VARCHAR(100),
    marital_status VARCHAR(20) CHECK (marital_status IN ('single', 'married', 'divorced', 'widowed')),
    id_number VARCHAR(20) NOT NULL,
    id_issued_date DATE,
    id_issued_place VARCHAR(255),
    tax_code VARCHAR(20),
    social_insurance_no VARCHAR(20),
    health_insurance_no VARCHAR(20),
    bank_account_no VARCHAR(30),
    bank_name VARCHAR(100),
    bank_branch VARCHAR(100),
    avatar VARCHAR(500),
    permanent_address TEXT,
    permanent_ward_id INT REFERENCES wards(id),
    current_address TEXT,
    current_ward_id INT REFERENCES wards(id),
    personal_email VARCHAR(255),
    personal_phone VARCHAR(20),
    emergency_contact VARCHAR(100),
    emergency_phone VARCHAR(20),
    department_id UUID NOT NULL REFERENCES departments(id),
    position_id UUID NOT NULL REFERENCES positions(id),
    manager_id UUID REFERENCES employees(id),
    employment_type VARCHAR(20) NOT NULL CHECK (employment_type IN ('full_time', 'part_time', 'contract', 'intern', 'probation')),
    employment_status VARCHAR(20) DEFAULT 'active' CHECK (employment_status IN ('active', 'inactive', 'on_leave', 'resigned', 'terminated')),
    join_date DATE NOT NULL,
    probation_end_date DATE,
    contract_start_date DATE,
    contract_end_date DATE,
    resignation_date DATE,
    base_salary DECIMAL(15,2) NOT NULL DEFAULT 0,
    salary_grade VARCHAR(20),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_employees_code ON employees(employee_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_employees_department ON employees(department_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_employees_position ON employees(position_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_employees_manager ON employees(manager_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_employees_status ON employees(employment_status) WHERE deleted_at IS NULL;
CREATE INDEX idx_employees_full_name ON employees(full_name) WHERE deleted_at IS NULL;

-- Add foreign key for department manager
ALTER TABLE departments ADD CONSTRAINT fk_dept_manager FOREIGN KEY (manager_id) REFERENCES employees(id);

-- ==================== WORK SHIFTS ====================

CREATE TABLE work_shifts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(20) NOT NULL UNIQUE,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    break_start TIME,
    break_end TIME,
    working_hours DECIMAL(4,2) NOT NULL,
    is_night_shift BOOLEAN DEFAULT FALSE,
    description TEXT,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE employee_shifts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    shift_id UUID NOT NULL REFERENCES work_shifts(id),
    date DATE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(employee_id, date)
);

-- ==================== ATTENDANCE ====================

CREATE TABLE attendances (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id),
    date DATE NOT NULL,
    check_in TIMESTAMP,
    check_out TIMESTAMP,
    check_in_ip VARCHAR(45),
    check_out_ip VARCHAR(45),
    check_in_location TEXT,
    check_out_location TEXT,
    working_hours DECIMAL(4,2) DEFAULT 0,
    overtime_hours DECIMAL(4,2) DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'present' CHECK (status IN ('present', 'absent', 'late', 'early_leave', 'half_day', 'on_leave', 'holiday', 'weekend')),
    notes TEXT,
    approved_by UUID REFERENCES employees(id),
    approved_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(employee_id, date)
);

CREATE INDEX idx_attendances_employee ON attendances(employee_id, date) WHERE deleted_at IS NULL;
CREATE INDEX idx_attendances_date ON attendances(date) WHERE deleted_at IS NULL;
CREATE INDEX idx_attendances_status ON attendances(status) WHERE deleted_at IS NULL;

CREATE TABLE attendance_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    attendance_id UUID NOT NULL REFERENCES attendances(id) ON DELETE CASCADE,
    action VARCHAR(20) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    location TEXT,
    device_info TEXT
);

-- ==================== LEAVE MANAGEMENT ====================

CREATE TABLE leave_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(20) NOT NULL UNIQUE,
    description TEXT,
    default_days INT DEFAULT 0,
    max_carry_over INT DEFAULT 0,
    is_paid BOOLEAN DEFAULT TRUE,
    requires_approval BOOLEAN DEFAULT TRUE,
    color VARCHAR(7) DEFAULT '#3B82F6',
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE leave_balances (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id),
    leave_type_id UUID NOT NULL REFERENCES leave_types(id),
    year INT NOT NULL,
    total_days DECIMAL(4,1) DEFAULT 0,
    used_days DECIMAL(4,1) DEFAULT 0,
    pending_days DECIMAL(4,1) DEFAULT 0,
    carried_over DECIMAL(4,1) DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(employee_id, leave_type_id, year)
);

CREATE TABLE leave_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id),
    leave_type_id UUID NOT NULL REFERENCES leave_types(id),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    total_days DECIMAL(4,1) NOT NULL,
    reason TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled')),
    approved_by UUID REFERENCES employees(id),
    approved_at TIMESTAMP,
    approver_notes TEXT,
    attachments TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_leave_requests_employee ON leave_requests(employee_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_leave_requests_status ON leave_requests(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_leave_requests_dates ON leave_requests(start_date, end_date) WHERE deleted_at IS NULL;

-- ==================== OVERTIME ====================

CREATE TABLE overtime_policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    weekday_multiplier DECIMAL(3,2) DEFAULT 1.5,
    weekend_multiplier DECIMAL(3,2) DEFAULT 2.0,
    holiday_multiplier DECIMAL(3,2) DEFAULT 3.0,
    night_multiplier DECIMAL(3,2) DEFAULT 1.3,
    max_hours_per_day DECIMAL(4,2) DEFAULT 4,
    max_hours_per_month DECIMAL(5,2) DEFAULT 40,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE overtime_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id),
    date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    hours DECIMAL(4,2) NOT NULL,
    reason TEXT NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('weekday', 'weekend', 'holiday', 'night')),
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled', 'completed')),
    approved_by UUID REFERENCES employees(id),
    approved_at TIMESTAMP,
    approver_notes TEXT,
    multiplier DECIMAL(3,2) DEFAULT 1.5,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_overtime_requests_employee ON overtime_requests(employee_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_overtime_requests_status ON overtime_requests(status) WHERE deleted_at IS NULL;

-- ==================== PAYROLL ====================

CREATE TABLE payroll_periods (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    year INT NOT NULL,
    month INT NOT NULL CHECK (month BETWEEN 1 AND 12),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    pay_date DATE NOT NULL,
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'processing', 'pending', 'approved', 'paid', 'cancelled')),
    processed_by UUID REFERENCES users(id),
    processed_at TIMESTAMP,
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMP,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(year, month)
);

CREATE TABLE payslips (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id),
    payroll_period_id UUID NOT NULL REFERENCES payroll_periods(id),
    employee_code VARCHAR(20) NOT NULL,
    employee_name VARCHAR(255) NOT NULL,
    department_name VARCHAR(255),
    position_name VARCHAR(255),
    working_days DECIMAL(4,1) DEFAULT 0,
    actual_working_days DECIMAL(4,1) DEFAULT 0,
    leave_days DECIMAL(4,1) DEFAULT 0,
    absent_days DECIMAL(4,1) DEFAULT 0,
    overtime_hours DECIMAL(5,2) DEFAULT 0,
    base_salary DECIMAL(15,2) DEFAULT 0,
    overtime_pay DECIMAL(15,2) DEFAULT 0,
    allowances DECIMAL(15,2) DEFAULT 0,
    bonuses DECIMAL(15,2) DEFAULT 0,
    other_earnings DECIMAL(15,2) DEFAULT 0,
    gross_earnings DECIMAL(15,2) DEFAULT 0,
    social_insurance DECIMAL(15,2) DEFAULT 0,
    health_insurance DECIMAL(15,2) DEFAULT 0,
    unemployment_insurance DECIMAL(15,2) DEFAULT 0,
    personal_income_tax DECIMAL(15,2) DEFAULT 0,
    other_deductions DECIMAL(15,2) DEFAULT 0,
    total_deductions DECIMAL(15,2) DEFAULT 0,
    net_salary DECIMAL(15,2) DEFAULT 0,
    earnings_details JSONB,
    deductions_details JSONB,
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'confirmed', 'paid')),
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(employee_id, payroll_period_id)
);

CREATE INDEX idx_payslips_employee ON payslips(employee_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_payslips_period ON payslips(payroll_period_id) WHERE deleted_at IS NULL;

CREATE TABLE allowances (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(20) NOT NULL UNIQUE,
    type VARCHAR(50),
    amount DECIMAL(15,2) DEFAULT 0,
    is_fixed BOOLEAN DEFAULT TRUE,
    is_taxable BOOLEAN DEFAULT TRUE,
    description TEXT,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE employee_allowances (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id),
    allowance_id UUID NOT NULL REFERENCES allowances(id),
    amount DECIMAL(15,2) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE deductions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(20) NOT NULL UNIQUE,
    type VARCHAR(50),
    percentage DECIMAL(5,2) DEFAULT 0,
    fixed_amount DECIMAL(15,2) DEFAULT 0,
    is_required BOOLEAN DEFAULT FALSE,
    description TEXT,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE tax_brackets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    min_income DECIMAL(15,2) NOT NULL,
    max_income DECIMAL(15,2) NOT NULL,
    tax_rate DECIMAL(5,2) NOT NULL,
    deduction DECIMAL(15,2) DEFAULT 0,
    year INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ==================== HOLIDAYS ====================

CREATE TABLE holidays (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    date DATE NOT NULL,
    type VARCHAR(50),
    description TEXT,
    is_recurring BOOLEAN DEFAULT FALSE,
    year INT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_holidays_date ON holidays(date) WHERE deleted_at IS NULL;

-- ==================== NOTIFICATIONS ====================

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    type VARCHAR(50) NOT NULL,
    data JSONB,
    read_at TIMESTAMP,
    action_url VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_notifications_user ON notifications(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_notifications_read ON notifications(user_id, read_at) WHERE deleted_at IS NULL;

-- ==================== AUDIT LOGS ====================

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(50) NOT NULL,
    table_name VARCHAR(100) NOT NULL,
    record_id UUID NOT NULL,
    old_values JSONB,
    new_values JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_table ON audit_logs(table_name, record_id);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at);

-- ==================== SYSTEM SETTINGS ====================

CREATE TABLE system_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(100) NOT NULL UNIQUE,
    value TEXT NOT NULL,
    type VARCHAR(20) DEFAULT 'string',
    "group" VARCHAR(50),
    label VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ==================== INDEXES FOR PERFORMANCE ====================

CREATE INDEX idx_audit_logs_created_desc ON audit_logs(created_at DESC);
CREATE INDEX idx_attendances_date_desc ON attendances(date DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_payslips_period_employee ON payslips(payroll_period_id, employee_id) WHERE deleted_at IS NULL;
