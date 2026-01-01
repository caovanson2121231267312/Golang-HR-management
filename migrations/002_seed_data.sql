-- HR Management System Seed Data
-- Run after migrations

-- ==================== ROLES ====================
INSERT INTO roles (id, name, slug, description, level, is_system) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'Super Admin', 'super_admin', 'Quyền cao nhất trong hệ thống', 1, TRUE),
('550e8400-e29b-41d4-a716-446655440002', 'HR Manager', 'hr_manager', 'Quản lý nhân sự', 2, TRUE),
('550e8400-e29b-41d4-a716-446655440003', 'Payroll Manager', 'payroll_manager', 'Quản lý lương', 2, TRUE),
('550e8400-e29b-41d4-a716-446655440004', 'Department Manager', 'department_manager', 'Quản lý phòng ban', 3, FALSE),
('550e8400-e29b-41d4-a716-446655440005', 'Team Leader', 'team_leader', 'Trưởng nhóm', 4, FALSE),
('550e8400-e29b-41d4-a716-446655440006', 'Employee', 'employee', 'Nhân viên', 5, TRUE);

-- ==================== PERMISSIONS ====================
INSERT INTO permissions (id, name, slug, module, description) VALUES
-- Users
('660e8400-e29b-41d4-a716-446655440001', 'View Users', 'users.view', 'users', 'Xem danh sách người dùng'),
('660e8400-e29b-41d4-a716-446655440002', 'Create Users', 'users.create', 'users', 'Tạo người dùng'),
('660e8400-e29b-41d4-a716-446655440003', 'Update Users', 'users.update', 'users', 'Cập nhật người dùng'),
('660e8400-e29b-41d4-a716-446655440004', 'Delete Users', 'users.delete', 'users', 'Xóa người dùng'),
-- Employees
('660e8400-e29b-41d4-a716-446655440011', 'View Employees', 'employees.view', 'employees', 'Xem danh sách nhân viên'),
('660e8400-e29b-41d4-a716-446655440012', 'Create Employees', 'employees.create', 'employees', 'Tạo nhân viên'),
('660e8400-e29b-41d4-a716-446655440013', 'Update Employees', 'employees.update', 'employees', 'Cập nhật nhân viên'),
('660e8400-e29b-41d4-a716-446655440014', 'Delete Employees', 'employees.delete', 'employees', 'Xóa nhân viên'),
-- Departments
('660e8400-e29b-41d4-a716-446655440021', 'View Departments', 'departments.view', 'departments', 'Xem phòng ban'),
('660e8400-e29b-41d4-a716-446655440022', 'Create Departments', 'departments.create', 'departments', 'Tạo phòng ban'),
('660e8400-e29b-41d4-a716-446655440023', 'Update Departments', 'departments.update', 'departments', 'Cập nhật phòng ban'),
('660e8400-e29b-41d4-a716-446655440024', 'Delete Departments', 'departments.delete', 'departments', 'Xóa phòng ban'),
-- Positions
('660e8400-e29b-41d4-a716-446655440031', 'View Positions', 'positions.view', 'positions', 'Xem vị trí'),
('660e8400-e29b-41d4-a716-446655440032', 'Create Positions', 'positions.create', 'positions', 'Tạo vị trí'),
('660e8400-e29b-41d4-a716-446655440033', 'Update Positions', 'positions.update', 'positions', 'Cập nhật vị trí'),
('660e8400-e29b-41d4-a716-446655440034', 'Delete Positions', 'positions.delete', 'positions', 'Xóa vị trí'),
-- Attendance
('660e8400-e29b-41d4-a716-446655440041', 'View Attendance', 'attendance.view', 'attendance', 'Xem chấm công'),
('660e8400-e29b-41d4-a716-446655440042', 'Manage Attendance', 'attendance.manage', 'attendance', 'Quản lý chấm công'),
('660e8400-e29b-41d4-a716-446655440043', 'Approve Attendance', 'attendance.approve', 'attendance', 'Phê duyệt chấm công'),
-- Leave
('660e8400-e29b-41d4-a716-446655440051', 'View Leave', 'leave.view', 'leave', 'Xem nghỉ phép'),
('660e8400-e29b-41d4-a716-446655440052', 'Manage Leave', 'leave.manage', 'leave', 'Quản lý nghỉ phép'),
('660e8400-e29b-41d4-a716-446655440053', 'Approve Leave', 'leave.approve', 'leave', 'Phê duyệt nghỉ phép'),
-- Overtime
('660e8400-e29b-41d4-a716-446655440061', 'View Overtime', 'overtime.view', 'overtime', 'Xem tăng ca'),
('660e8400-e29b-41d4-a716-446655440062', 'Manage Overtime', 'overtime.manage', 'overtime', 'Quản lý tăng ca'),
('660e8400-e29b-41d4-a716-446655440063', 'Approve Overtime', 'overtime.approve', 'overtime', 'Phê duyệt tăng ca'),
-- Payroll
('660e8400-e29b-41d4-a716-446655440071', 'View Payroll', 'payroll.view', 'payroll', 'Xem lương'),
('660e8400-e29b-41d4-a716-446655440072', 'Create Payroll', 'payroll.create', 'payroll', 'Tạo bảng lương'),
('660e8400-e29b-41d4-a716-446655440073', 'Calculate Payroll', 'payroll.calculate', 'payroll', 'Tính lương'),
('660e8400-e29b-41d4-a716-446655440074', 'Approve Payroll', 'payroll.approve', 'payroll', 'Phê duyệt lương'),
('660e8400-e29b-41d4-a716-446655440075', 'Pay Payroll', 'payroll.pay', 'payroll', 'Thanh toán lương'),
('660e8400-e29b-41d4-a716-446655440076', 'Manage Payroll', 'payroll.manage', 'payroll', 'Quản lý lương'),
-- Roles
('660e8400-e29b-41d4-a716-446655440081', 'View Roles', 'roles.view', 'roles', 'Xem vai trò'),
('660e8400-e29b-41d4-a716-446655440082', 'Create Roles', 'roles.create', 'roles', 'Tạo vai trò'),
('660e8400-e29b-41d4-a716-446655440083', 'Update Roles', 'roles.update', 'roles', 'Cập nhật vai trò'),
('660e8400-e29b-41d4-a716-446655440084', 'Delete Roles', 'roles.delete', 'roles', 'Xóa vai trò'),
-- Permissions
('660e8400-e29b-41d4-a716-446655440091', 'View Permissions', 'permissions.view', 'permissions', 'Xem quyền'),
-- Reports
('660e8400-e29b-41d4-a716-446655440101', 'View Reports', 'reports.view', 'reports', 'Xem báo cáo'),
('660e8400-e29b-41d4-a716-446655440102', 'Export Reports', 'reports.export', 'reports', 'Xuất báo cáo');

-- ==================== ROLE PERMISSIONS ====================
-- Super Admin gets all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT '550e8400-e29b-41d4-a716-446655440001', id FROM permissions;

-- HR Manager
INSERT INTO role_permissions (role_id, permission_id) VALUES
('550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440011'),
('550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440012'),
('550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440013'),
('550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440014'),
('550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440021'),
('550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440022'),
('550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440023'),
('550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440031'),
('550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440032'),
('550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440033'),
('550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440041'),
('550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440042'),
('550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440043'),
('550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440051'),
('550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440052'),
('550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440053'),
('550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440061'),
('550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440062'),
('550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440063'),
('550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440101'),
('550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440102');

-- Payroll Manager
INSERT INTO role_permissions (role_id, permission_id) VALUES
('550e8400-e29b-41d4-a716-446655440003', '660e8400-e29b-41d4-a716-446655440011'),
('550e8400-e29b-41d4-a716-446655440003', '660e8400-e29b-41d4-a716-446655440041'),
('550e8400-e29b-41d4-a716-446655440003', '660e8400-e29b-41d4-a716-446655440071'),
('550e8400-e29b-41d4-a716-446655440003', '660e8400-e29b-41d4-a716-446655440072'),
('550e8400-e29b-41d4-a716-446655440003', '660e8400-e29b-41d4-a716-446655440073'),
('550e8400-e29b-41d4-a716-446655440003', '660e8400-e29b-41d4-a716-446655440074'),
('550e8400-e29b-41d4-a716-446655440003', '660e8400-e29b-41d4-a716-446655440075'),
('550e8400-e29b-41d4-a716-446655440003', '660e8400-e29b-41d4-a716-446655440076'),
('550e8400-e29b-41d4-a716-446655440003', '660e8400-e29b-41d4-a716-446655440101'),
('550e8400-e29b-41d4-a716-446655440003', '660e8400-e29b-41d4-a716-446655440102');

-- ==================== PROVINCES (Sample) ====================
INSERT INTO provinces (id, name, name_en, code, type) VALUES
(1, 'Hà Nội', 'Hanoi', '01', 'Thành phố Trung ương'),
(2, 'Hồ Chí Minh', 'Ho Chi Minh', '79', 'Thành phố Trung ương'),
(3, 'Đà Nẵng', 'Da Nang', '48', 'Thành phố Trung ương'),
(4, 'Hải Phòng', 'Hai Phong', '31', 'Thành phố Trung ương'),
(5, 'Cần Thơ', 'Can Tho', '92', 'Thành phố Trung ương');

-- ==================== DISTRICTS (Sample) ====================
INSERT INTO districts (id, province_id, name, name_en, code, type) VALUES
(1, 1, 'Ba Đình', 'Ba Dinh', '001', 'Quận'),
(2, 1, 'Hoàn Kiếm', 'Hoan Kiem', '002', 'Quận'),
(3, 1, 'Đống Đa', 'Dong Da', '006', 'Quận'),
(4, 1, 'Cầu Giấy', 'Cau Giay', '005', 'Quận'),
(5, 2, 'Quận 1', 'District 1', '760', 'Quận'),
(6, 2, 'Quận 3', 'District 3', '770', 'Quận'),
(7, 2, 'Quận 7', 'District 7', '778', 'Quận');

-- ==================== WARDS (Sample) ====================
INSERT INTO wards (id, district_id, name, name_en, code, type) VALUES
(1, 1, 'Phường Phúc Xá', 'Phuc Xa Ward', '00001', 'Phường'),
(2, 1, 'Phường Trúc Bạch', 'Truc Bach Ward', '00004', 'Phường'),
(3, 2, 'Phường Hàng Bạc', 'Hang Bac Ward', '00034', 'Phường'),
(4, 5, 'Phường Bến Nghé', 'Ben Nghe Ward', '26734', 'Phường'),
(5, 5, 'Phường Bến Thành', 'Ben Thanh Ward', '26737', 'Phường');

-- ==================== DEPARTMENTS ====================
INSERT INTO departments (id, name, code, description, level, path, status) VALUES
('770e8400-e29b-41d4-a716-446655440001', 'Ban Giám đốc', 'BGD', 'Ban lãnh đạo công ty', 1, '/BGD', 'active'),
('770e8400-e29b-41d4-a716-446655440002', 'Phòng Nhân sự', 'HR', 'Quản lý nhân sự', 2, '/BGD/HR', 'active'),
('770e8400-e29b-41d4-a716-446655440003', 'Phòng Kế toán', 'ACC', 'Quản lý tài chính kế toán', 2, '/BGD/ACC', 'active'),
('770e8400-e29b-41d4-a716-446655440004', 'Phòng IT', 'IT', 'Công nghệ thông tin', 2, '/BGD/IT', 'active'),
('770e8400-e29b-41d4-a716-446655440005', 'Phòng Kinh doanh', 'SALES', 'Kinh doanh và bán hàng', 2, '/BGD/SALES', 'active'),
('770e8400-e29b-41d4-a716-446655440006', 'Phòng Marketing', 'MKT', 'Marketing và truyền thông', 2, '/BGD/MKT', 'active');

-- Set parent_id
UPDATE departments SET parent_id = '770e8400-e29b-41d4-a716-446655440001' WHERE code != 'BGD';

-- ==================== POSITIONS ====================
INSERT INTO positions (id, name, code, description, level, min_salary, max_salary, status) VALUES
('880e8400-e29b-41d4-a716-446655440001', 'Giám đốc', 'GD', 'Giám đốc điều hành', 1, 50000000, 100000000, 'active'),
('880e8400-e29b-41d4-a716-446655440002', 'Phó Giám đốc', 'PGD', 'Phó Giám đốc', 2, 40000000, 70000000, 'active'),
('880e8400-e29b-41d4-a716-446655440003', 'Trưởng phòng', 'TP', 'Trưởng phòng ban', 3, 25000000, 45000000, 'active'),
('880e8400-e29b-41d4-a716-446655440004', 'Phó phòng', 'PP', 'Phó phòng ban', 4, 20000000, 35000000, 'active'),
('880e8400-e29b-41d4-a716-446655440005', 'Trưởng nhóm', 'TN', 'Trưởng nhóm', 5, 15000000, 25000000, 'active'),
('880e8400-e29b-41d4-a716-446655440006', 'Chuyên viên cao cấp', 'CVCC', 'Chuyên viên cao cấp', 6, 12000000, 20000000, 'active'),
('880e8400-e29b-41d4-a716-446655440007', 'Chuyên viên', 'CV', 'Chuyên viên', 7, 8000000, 15000000, 'active'),
('880e8400-e29b-41d4-a716-446655440008', 'Nhân viên', 'NV', 'Nhân viên', 8, 6000000, 12000000, 'active'),
('880e8400-e29b-41d4-a716-446655440009', 'Thực tập sinh', 'TTS', 'Thực tập sinh', 9, 3000000, 6000000, 'active');

-- ==================== LEAVE TYPES ====================
INSERT INTO leave_types (id, name, code, description, default_days, max_carry_over, is_paid, color) VALUES
('990e8400-e29b-41d4-a716-446655440001', 'Nghỉ phép năm', 'ANNUAL', 'Nghỉ phép hàng năm theo luật', 12, 5, TRUE, '#3B82F6'),
('990e8400-e29b-41d4-a716-446655440002', 'Nghỉ ốm', 'SICK', 'Nghỉ ốm đau', 30, 0, TRUE, '#EF4444'),
('990e8400-e29b-41d4-a716-446655440003', 'Nghỉ việc riêng', 'PERSONAL', 'Nghỉ việc riêng không lương', 5, 0, FALSE, '#F59E0B'),
('990e8400-e29b-41d4-a716-446655440004', 'Nghỉ cưới', 'WEDDING', 'Nghỉ đám cưới bản thân', 3, 0, TRUE, '#EC4899'),
('990e8400-e29b-41d4-a716-446655440005', 'Nghỉ tang', 'FUNERAL', 'Nghỉ tang gia đình', 3, 0, TRUE, '#6B7280'),
('990e8400-e29b-41d4-a716-446655440006', 'Nghỉ thai sản', 'MATERNITY', 'Nghỉ thai sản', 180, 0, TRUE, '#8B5CF6');

-- ==================== WORK SHIFTS ====================
INSERT INTO work_shifts (id, name, code, start_time, end_time, break_start, break_end, working_hours, is_night_shift) VALUES
('aa0e8400-e29b-41d4-a716-446655440001', 'Ca hành chính', 'HC', '08:00', '17:00', '12:00', '13:00', 8, FALSE),
('aa0e8400-e29b-41d4-a716-446655440002', 'Ca sáng', 'S', '06:00', '14:00', '11:00', '11:30', 7.5, FALSE),
('aa0e8400-e29b-41d4-a716-446655440003', 'Ca chiều', 'C', '14:00', '22:00', '18:00', '18:30', 7.5, FALSE),
('aa0e8400-e29b-41d4-a716-446655440004', 'Ca đêm', 'D', '22:00', '06:00', '02:00', '02:30', 7.5, TRUE);

-- ==================== OVERTIME POLICY ====================
INSERT INTO overtime_policies (id, name, weekday_multiplier, weekend_multiplier, holiday_multiplier, night_multiplier, max_hours_per_day, max_hours_per_month) VALUES
('bb0e8400-e29b-41d4-a716-446655440001', 'Chính sách tăng ca chuẩn', 1.5, 2.0, 3.0, 1.3, 4, 40);

-- ==================== ALLOWANCES ====================
INSERT INTO allowances (id, name, code, type, amount, is_fixed, is_taxable) VALUES
('cc0e8400-e29b-41d4-a716-446655440001', 'Phụ cấp ăn trưa', 'LUNCH', 'meal', 730000, TRUE, FALSE),
('cc0e8400-e29b-41d4-a716-446655440002', 'Phụ cấp xăng xe', 'FUEL', 'transport', 500000, TRUE, FALSE),
('cc0e8400-e29b-41d4-a716-446655440003', 'Phụ cấp điện thoại', 'PHONE', 'communication', 300000, TRUE, TRUE),
('cc0e8400-e29b-41d4-a716-446655440004', 'Phụ cấp trách nhiệm', 'RESP', 'responsibility', 2000000, TRUE, TRUE),
('cc0e8400-e29b-41d4-a716-446655440005', 'Phụ cấp độc hại', 'HAZARD', 'hazard', 500000, TRUE, FALSE);

-- ==================== DEDUCTIONS ====================
INSERT INTO deductions (id, name, code, type, percentage, is_required) VALUES
('dd0e8400-e29b-41d4-a716-446655440001', 'BHXH', 'SI', 'insurance', 8.0, TRUE),
('dd0e8400-e29b-41d4-a716-446655440002', 'BHYT', 'HI', 'insurance', 1.5, TRUE),
('dd0e8400-e29b-41d4-a716-446655440003', 'BHTN', 'UI', 'insurance', 1.0, TRUE);

-- ==================== TAX BRACKETS (Vietnam 2024) ====================
INSERT INTO tax_brackets (id, min_income, max_income, tax_rate, deduction, year) VALUES
('ee0e8400-e29b-41d4-a716-446655440001', 0, 5000000, 5, 0, 2024),
('ee0e8400-e29b-41d4-a716-446655440002', 5000000, 10000000, 10, 250000, 2024),
('ee0e8400-e29b-41d4-a716-446655440003', 10000000, 18000000, 15, 750000, 2024),
('ee0e8400-e29b-41d4-a716-446655440004', 18000000, 32000000, 20, 1650000, 2024),
('ee0e8400-e29b-41d4-a716-446655440005', 32000000, 52000000, 25, 3250000, 2024),
('ee0e8400-e29b-41d4-a716-446655440006', 52000000, 80000000, 30, 5850000, 2024),
('ee0e8400-e29b-41d4-a716-446655440007', 80000000, 999999999999, 35, 9850000, 2024);

-- ==================== HOLIDAYS 2024 ====================
INSERT INTO holidays (id, name, date, type, is_recurring, year) VALUES
('ff0e8400-e29b-41d4-a716-446655440001', 'Tết Dương lịch', '2024-01-01', 'national', TRUE, 2024),
('ff0e8400-e29b-41d4-a716-446655440002', 'Tết Nguyên đán', '2024-02-08', 'national', FALSE, 2024),
('ff0e8400-e29b-41d4-a716-446655440003', 'Tết Nguyên đán', '2024-02-09', 'national', FALSE, 2024),
('ff0e8400-e29b-41d4-a716-446655440004', 'Tết Nguyên đán', '2024-02-10', 'national', FALSE, 2024),
('ff0e8400-e29b-41d4-a716-446655440005', 'Tết Nguyên đán', '2024-02-11', 'national', FALSE, 2024),
('ff0e8400-e29b-41d4-a716-446655440006', 'Tết Nguyên đán', '2024-02-12', 'national', FALSE, 2024),
('ff0e8400-e29b-41d4-a716-446655440007', 'Giỗ Tổ Hùng Vương', '2024-04-18', 'national', FALSE, 2024),
('ff0e8400-e29b-41d4-a716-446655440008', 'Ngày Giải phóng', '2024-04-30', 'national', TRUE, 2024),
('ff0e8400-e29b-41d4-a716-446655440009', 'Ngày Quốc tế Lao động', '2024-05-01', 'national', TRUE, 2024),
('ff0e8400-e29b-41d4-a716-446655440010', 'Quốc khánh', '2024-09-02', 'national', TRUE, 2024);

-- ==================== SYSTEM SETTINGS ====================
INSERT INTO system_settings (id, key, value, type, "group", label) VALUES
('110e8400-e29b-41d4-a716-446655440001', 'company_name', 'HR Management System', 'string', 'company', 'Tên công ty'),
('110e8400-e29b-41d4-a716-446655440002', 'company_address', '123 Nguyễn Văn Linh, Quận 7, TP.HCM', 'string', 'company', 'Địa chỉ'),
('110e8400-e29b-41d4-a716-446655440003', 'company_phone', '028 1234 5678', 'string', 'company', 'Số điện thoại'),
('110e8400-e29b-41d4-a716-446655440004', 'company_email', 'hr@company.com', 'string', 'company', 'Email'),
('110e8400-e29b-41d4-a716-446655440005', 'company_tax_code', '0123456789', 'string', 'company', 'Mã số thuế'),
('110e8400-e29b-41d4-a716-446655440006', 'working_hours_per_day', '8', 'number', 'attendance', 'Số giờ làm việc/ngày'),
('110e8400-e29b-41d4-a716-446655440007', 'working_days_per_month', '22', 'number', 'attendance', 'Số ngày làm việc/tháng'),
('110e8400-e29b-41d4-a716-446655440008', 'late_threshold_minutes', '15', 'number', 'attendance', 'Ngưỡng đi trễ (phút)'),
('110e8400-e29b-41d4-a716-446655440009', 'personal_deduction', '11000000', 'number', 'payroll', 'Giảm trừ gia cảnh bản thân'),
('110e8400-e29b-41d4-a716-446655440010', 'dependent_deduction', '4400000', 'number', 'payroll', 'Giảm trừ gia cảnh người phụ thuộc');

-- ==================== ADMIN USER ====================
-- Password: Admin@123
INSERT INTO users (id, email, phone, password, status, email_verified_at, preferred_language) VALUES
('220e8400-e29b-41d4-a716-446655440001', 'admin@hrms.com', '0901234567', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4pGqK5eYl0AaGhMG', 'active', NOW(), 'vi');

INSERT INTO user_roles (user_id, role_id) VALUES
('220e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001');

INSERT INTO employees (id, user_id, employee_code, first_name, last_name, full_name, gender, date_of_birth, id_number, department_id, position_id, employment_type, employment_status, join_date, base_salary) VALUES
('330e8400-e29b-41d4-a716-446655440001', '220e8400-e29b-41d4-a716-446655440001', 'NV000001', 'Admin', 'System', 'System Admin', 'male', '1990-01-01', '001234567890', '770e8400-e29b-41d4-a716-446655440001', '880e8400-e29b-41d4-a716-446655440001', 'full_time', 'active', '2020-01-01', 50000000);
