package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/text/language"
)

//go:embed locales/*.json
var localesFS embed.FS

type I18n struct {
	translations map[string]map[string]string
	defaultLang  string
	mu           sync.RWMutex
}

var instance *I18n

func New(defaultLang string) (*I18n, error) {
	i := &I18n{
		translations: make(map[string]map[string]string),
		defaultLang:  defaultLang,
	}

	// Load embedded translations
	if err := i.loadEmbeddedTranslations(); err != nil {
		return nil, err
	}

	instance = i
	return i, nil
}

func Get() *I18n {
	return instance
}

func (i *I18n) loadEmbeddedTranslations() error {
	entries, err := localesFS.ReadDir("locales")
	if err != nil {
		// If no embedded files, load defaults
		return i.loadDefaults()
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}

		lang := strings.TrimSuffix(name, ".json")
		content, err := localesFS.ReadFile("locales/" + name)
		if err != nil {
			continue
		}

		var translations map[string]string
		if err := json.Unmarshal(content, &translations); err != nil {
			continue
		}

		i.translations[lang] = translations
	}

	if len(i.translations) == 0 {
		return i.loadDefaults()
	}

	return nil
}

func (i *I18n) loadDefaults() error {
	i.translations["vi"] = viTranslations
	i.translations["en"] = enTranslations
	return nil
}

func (i *I18n) T(lang, key string, args ...interface{}) string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	// Try requested language
	if trans, ok := i.translations[lang]; ok {
		if msg, ok := trans[key]; ok {
			if len(args) > 0 {
				return fmt.Sprintf(msg, args...)
			}
			return msg
		}
	}

	// Fallback to default language
	if trans, ok := i.translations[i.defaultLang]; ok {
		if msg, ok := trans[key]; ok {
			if len(args) > 0 {
				return fmt.Sprintf(msg, args...)
			}
			return msg
		}
	}

	// Return key if not found
	return key
}

func (i *I18n) Translate(lang, key string, args ...interface{}) string {
	return i.T(lang, key, args...)
}

func (i *I18n) GetSupportedLanguages() []string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	langs := make([]string, 0, len(i.translations))
	for lang := range i.translations {
		langs = append(langs, lang)
	}
	return langs
}

func (i *I18n) HasLanguage(lang string) bool {
	i.mu.RLock()
	defer i.mu.RUnlock()

	_, ok := i.translations[lang]
	return ok
}

func (i *I18n) SetDefaultLanguage(lang string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.defaultLang = lang
}

func (i *I18n) AddTranslation(lang, key, value string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if _, ok := i.translations[lang]; !ok {
		i.translations[lang] = make(map[string]string)
	}
	i.translations[lang][key] = value
}

func (i *I18n) LoadTranslations(lang string, translations map[string]string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if _, ok := i.translations[lang]; !ok {
		i.translations[lang] = make(map[string]string)
	}

	for k, v := range translations {
		i.translations[lang][k] = v
	}
}

// Helper function to detect language from Accept-Language header
func DetectLanguage(acceptLanguage string, supported []string) string {
	if acceptLanguage == "" {
		return "vi" // Default to Vietnamese
	}

	tags := make([]language.Tag, len(supported))
	for i, lang := range supported {
		tags[i] = language.Make(lang)
	}

	matcher := language.NewMatcher(tags)
	tag, _ := language.MatchStrings(matcher, acceptLanguage)
	base, _ := tag.Base()
	return base.String()
}

// Convenience function
func T(lang, key string, args ...interface{}) string {
	if instance == nil {
		return key
	}
	return instance.T(lang, key, args...)
}

// Vietnamese translations
var viTranslations = map[string]string{
	// Common
	"common.success":              "Thành công",
	"common.error":                "Đã xảy ra lỗi",
	"common.not_found":            "Không tìm thấy",
	"common.unauthorized":         "Không có quyền truy cập",
	"common.forbidden":            "Truy cập bị từ chối",
	"common.bad_request":          "Yêu cầu không hợp lệ",
	"common.internal_error":       "Lỗi hệ thống",
	"common.validation_error":     "Dữ liệu không hợp lệ",
	"common.created":              "Tạo thành công",
	"common.updated":              "Cập nhật thành công",
	"common.deleted":              "Xóa thành công",
	"common.list":                 "Lấy danh sách thành công",
	
	// Auth
	"auth.login_success":          "Đăng nhập thành công",
	"auth.login_failed":           "Email hoặc mật khẩu không đúng",
	"auth.logout_success":         "Đăng xuất thành công",
	"auth.token_expired":          "Phiên đăng nhập đã hết hạn",
	"auth.token_invalid":          "Token không hợp lệ",
	"auth.account_locked":         "Tài khoản đã bị khóa. Vui lòng thử lại sau %s",
	"auth.account_inactive":       "Tài khoản chưa được kích hoạt",
	"auth.email_not_verified":     "Email chưa được xác thực",
	"auth.password_reset_sent":    "Email đặt lại mật khẩu đã được gửi",
	"auth.password_reset_success": "Đặt lại mật khẩu thành công",
	"auth.password_changed":       "Đổi mật khẩu thành công",
	"auth.old_password_incorrect": "Mật khẩu cũ không đúng",
	"auth.refresh_success":        "Làm mới token thành công",
	"auth.two_factor_required":    "Yêu cầu xác thực 2 bước",
	"auth.two_factor_invalid":     "Mã xác thực không đúng",
	
	// OTP
	"otp.sent":                    "Mã OTP đã được gửi",
	"otp.verified":                "Xác thực OTP thành công",
	"otp.invalid":                 "Mã OTP không đúng",
	"otp.expired":                 "Mã OTP đã hết hạn",
	"otp.too_many_attempts":       "Vượt quá số lần thử. Vui lòng yêu cầu mã mới",
	
	// User
	"user.created":                "Tạo người dùng thành công",
	"user.updated":                "Cập nhật người dùng thành công",
	"user.deleted":                "Xóa người dùng thành công",
	"user.not_found":              "Không tìm thấy người dùng",
	"user.email_exists":           "Email đã được sử dụng",
	"user.phone_exists":           "Số điện thoại đã được sử dụng",
	
	// Employee
	"employee.created":            "Tạo nhân viên thành công",
	"employee.updated":            "Cập nhật nhân viên thành công",
	"employee.deleted":            "Xóa nhân viên thành công",
	"employee.not_found":          "Không tìm thấy nhân viên",
	"employee.code_exists":        "Mã nhân viên đã tồn tại",
	
	// Department
	"department.created":          "Tạo phòng ban thành công",
	"department.updated":          "Cập nhật phòng ban thành công",
	"department.deleted":          "Xóa phòng ban thành công",
	"department.not_found":        "Không tìm thấy phòng ban",
	"department.has_employees":    "Không thể xóa phòng ban còn nhân viên",
	
	// Attendance
	"attendance.check_in":         "Chấm công vào thành công",
	"attendance.check_out":        "Chấm công ra thành công",
	"attendance.already_checked_in": "Đã chấm công vào hôm nay",
	"attendance.not_checked_in":   "Chưa chấm công vào",
	"attendance.already_checked_out": "Đã chấm công ra hôm nay",
	
	// Leave
	"leave.created":               "Tạo đơn nghỉ phép thành công",
	"leave.approved":              "Phê duyệt đơn nghỉ phép thành công",
	"leave.rejected":              "Từ chối đơn nghỉ phép",
	"leave.cancelled":             "Hủy đơn nghỉ phép thành công",
	"leave.not_found":             "Không tìm thấy đơn nghỉ phép",
	"leave.insufficient_balance":  "Số ngày phép không đủ",
	"leave.overlap":               "Ngày nghỉ trùng với đơn khác",
	
	// Overtime
	"overtime.created":            "Tạo đề xuất tăng ca thành công",
	"overtime.approved":           "Phê duyệt tăng ca thành công",
	"overtime.rejected":           "Từ chối tăng ca",
	"overtime.not_found":          "Không tìm thấy đề xuất tăng ca",
	"overtime.max_hours_exceeded": "Vượt quá số giờ tăng ca tối đa",
	
	// Payroll
	"payroll.generated":           "Tạo bảng lương thành công",
	"payroll.approved":            "Phê duyệt bảng lương thành công",
	"payroll.paid":                "Thanh toán lương thành công",
	"payroll.not_found":           "Không tìm thấy bảng lương",
	"payslip.sent":                "Gửi phiếu lương thành công",
	
	// Validation
	"validation.required":         "Trường %s là bắt buộc",
	"validation.email":            "Email không hợp lệ",
	"validation.phone":            "Số điện thoại không hợp lệ",
	"validation.min_length":       "Trường %s phải có ít nhất %d ký tự",
	"validation.max_length":       "Trường %s không được vượt quá %d ký tự",
	"validation.min_value":        "Giá trị tối thiểu là %d",
	"validation.max_value":        "Giá trị tối đa là %d",
	"validation.date_format":      "Định dạng ngày không hợp lệ",
	"validation.future_date":      "Ngày phải là ngày trong tương lai",
	"validation.past_date":        "Ngày phải là ngày trong quá khứ",
	
	// Rate Limit
	"rate_limit.exceeded":         "Quá nhiều yêu cầu. Vui lòng thử lại sau",
	
	// Permission
	"permission.denied":           "Bạn không có quyền thực hiện hành động này",
}

// English translations
var enTranslations = map[string]string{
	// Common
	"common.success":              "Success",
	"common.error":                "An error occurred",
	"common.not_found":            "Not found",
	"common.unauthorized":         "Unauthorized",
	"common.forbidden":            "Access denied",
	"common.bad_request":          "Bad request",
	"common.internal_error":       "Internal server error",
	"common.validation_error":     "Validation error",
	"common.created":              "Created successfully",
	"common.updated":              "Updated successfully",
	"common.deleted":              "Deleted successfully",
	"common.list":                 "Retrieved successfully",
	
	// Auth
	"auth.login_success":          "Login successful",
	"auth.login_failed":           "Invalid email or password",
	"auth.logout_success":         "Logout successful",
	"auth.token_expired":          "Session expired",
	"auth.token_invalid":          "Invalid token",
	"auth.account_locked":         "Account locked. Please try again in %s",
	"auth.account_inactive":       "Account not activated",
	"auth.email_not_verified":     "Email not verified",
	"auth.password_reset_sent":    "Password reset email sent",
	"auth.password_reset_success": "Password reset successful",
	"auth.password_changed":       "Password changed successfully",
	"auth.old_password_incorrect": "Old password is incorrect",
	"auth.refresh_success":        "Token refreshed successfully",
	"auth.two_factor_required":    "Two-factor authentication required",
	"auth.two_factor_invalid":     "Invalid verification code",
	
	// OTP
	"otp.sent":                    "OTP sent successfully",
	"otp.verified":                "OTP verified successfully",
	"otp.invalid":                 "Invalid OTP",
	"otp.expired":                 "OTP expired",
	"otp.too_many_attempts":       "Too many attempts. Please request a new code",
	
	// User
	"user.created":                "User created successfully",
	"user.updated":                "User updated successfully",
	"user.deleted":                "User deleted successfully",
	"user.not_found":              "User not found",
	"user.email_exists":           "Email already in use",
	"user.phone_exists":           "Phone number already in use",
	
	// Employee
	"employee.created":            "Employee created successfully",
	"employee.updated":            "Employee updated successfully",
	"employee.deleted":            "Employee deleted successfully",
	"employee.not_found":          "Employee not found",
	"employee.code_exists":        "Employee code already exists",
	
	// Department
	"department.created":          "Department created successfully",
	"department.updated":          "Department updated successfully",
	"department.deleted":          "Department deleted successfully",
	"department.not_found":        "Department not found",
	"department.has_employees":    "Cannot delete department with employees",
	
	// Attendance
	"attendance.check_in":         "Checked in successfully",
	"attendance.check_out":        "Checked out successfully",
	"attendance.already_checked_in": "Already checked in today",
	"attendance.not_checked_in":   "Not checked in yet",
	"attendance.already_checked_out": "Already checked out today",
	
	// Leave
	"leave.created":               "Leave request created",
	"leave.approved":              "Leave request approved",
	"leave.rejected":              "Leave request rejected",
	"leave.cancelled":             "Leave request cancelled",
	"leave.not_found":             "Leave request not found",
	"leave.insufficient_balance":  "Insufficient leave balance",
	"leave.overlap":               "Leave dates overlap with another request",
	
	// Overtime
	"overtime.created":            "Overtime request created",
	"overtime.approved":           "Overtime request approved",
	"overtime.rejected":           "Overtime request rejected",
	"overtime.not_found":          "Overtime request not found",
	"overtime.max_hours_exceeded": "Maximum overtime hours exceeded",
	
	// Payroll
	"payroll.generated":           "Payroll generated successfully",
	"payroll.approved":            "Payroll approved successfully",
	"payroll.paid":                "Payroll paid successfully",
	"payroll.not_found":           "Payroll not found",
	"payslip.sent":                "Payslip sent successfully",
	
	// Validation
	"validation.required":         "%s is required",
	"validation.email":            "Invalid email address",
	"validation.phone":            "Invalid phone number",
	"validation.min_length":       "%s must be at least %d characters",
	"validation.max_length":       "%s must not exceed %d characters",
	"validation.min_value":        "Minimum value is %d",
	"validation.max_value":        "Maximum value is %d",
	"validation.date_format":      "Invalid date format",
	"validation.future_date":      "Date must be in the future",
	"validation.past_date":        "Date must be in the past",
	
	// Rate Limit
	"rate_limit.exceeded":         "Too many requests. Please try again later",
	
	// Permission
	"permission.denied":           "You don't have permission to perform this action",
}
