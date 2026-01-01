package email

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"path/filepath"
	"sync"
	"io"

	"hr-management-system/internal/config"

	"gopkg.in/gomail.v2"
)

type EmailService struct {
	cfg       *config.EmailConfig
	dialer    *gomail.Dialer
	templates map[string]*template.Template
	mu        sync.RWMutex
}

var emailService *EmailService

func NewEmailService(cfg *config.EmailConfig) (*EmailService, error) {
	dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)

	emailService = &EmailService{
		cfg:       cfg,
		dialer:    dialer,
		templates: make(map[string]*template.Template),
	}

	// Load templates
	if err := emailService.loadTemplates(); err != nil {
		return nil, err
	}

	return emailService, nil
}

func GetEmailService() *EmailService {
	return emailService
}

func (e *EmailService) loadTemplates() error {
	templates := map[string]string{
		"otp":            otpTemplate,
		"password_reset": passwordResetTemplate,
		"welcome":        welcomeTemplate,
		"payslip":        payslipTemplate,
		"leave_request":  leaveRequestTemplate,
		"leave_approved": leaveApprovedTemplate,
		"overtime":       overtimeTemplate,
	}

	for name, content := range templates {
		tmpl, err := template.New(name).Parse(content)
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", name, err)
		}
		e.templates[name] = tmpl
	}

	return nil
}

func (e *EmailService) getTemplate(name string) (*template.Template, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	tmpl, ok := e.templates[name]
	if !ok {
		return nil, fmt.Errorf("template %s not found", name)
	}
	return tmpl, nil
}

func (e *EmailService) renderTemplate(name string, data interface{}) (string, error) {
	tmpl, err := e.getTemplate(name)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

type Email struct {
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	Body        string
	IsHTML      bool
	Attachments []Attachment
}

type Attachment struct {
	Filename string
	Content  []byte
	MimeType string
}

func (e *EmailService) Send(ctx context.Context, email Email) error {
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", e.cfg.FromName, e.cfg.From))
	m.SetHeader("To", email.To...)

	if len(email.Cc) > 0 {
		m.SetHeader("Cc", email.Cc...)
	}

	if len(email.Bcc) > 0 {
		m.SetHeader("Bcc", email.Bcc...)
	}

	m.SetHeader("Subject", email.Subject)

	if email.IsHTML {
		m.SetBody("text/html", email.Body)
	} else {
		m.SetBody("text/plain", email.Body)
	}

	for _, att := range email.Attachments {
		m.Attach(att.Filename, gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(att.Content)
			return err
		}))

	}

	return e.dialer.DialAndSend(m)
}

// OTP Email
type OTPEmailData struct {
	Name     string
	OTP      string
	Type     string
	Expiry   string
	AppName  string
	Language string
}

func (e *EmailService) SendOTP(ctx context.Context, to, name, otp, otpType, language string) error {
	data := OTPEmailData{
		Name:     name,
		OTP:      otp,
		Type:     otpType,
		Expiry:   "5 phút",
		AppName:  "HR Management System",
		Language: language,
	}

	body, err := e.renderTemplate("otp", data)
	if err != nil {
		return err
	}

	subject := "Mã xác thực OTP"
	if language == "en" {
		subject = "Your OTP Code"
	}

	return e.Send(ctx, Email{
		To:      []string{to},
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	})
}

// Password Reset Email
type PasswordResetData struct {
	Name      string
	ResetLink string
	Expiry    string
	AppName   string
}

func (e *EmailService) SendPasswordReset(ctx context.Context, to, name, resetLink string) error {
	data := PasswordResetData{
		Name:      name,
		ResetLink: resetLink,
		Expiry:    "1 giờ",
		AppName:   "HR Management System",
	}

	body, err := e.renderTemplate("password_reset", data)
	if err != nil {
		return err
	}

	return e.Send(ctx, Email{
		To:      []string{to},
		Subject: "Đặt lại mật khẩu",
		Body:    body,
		IsHTML:  true,
	})
}

// Welcome Email
type WelcomeData struct {
	Name         string
	Email        string
	TempPassword string
	LoginURL     string
	AppName      string
}

func (e *EmailService) SendWelcome(ctx context.Context, to, name, tempPassword, loginURL string) error {
	data := WelcomeData{
		Name:         name,
		Email:        to,
		TempPassword: tempPassword,
		LoginURL:     loginURL,
		AppName:      "HR Management System",
	}

	body, err := e.renderTemplate("welcome", data)
	if err != nil {
		return err
	}

	return e.Send(ctx, Email{
		To:      []string{to},
		Subject: "Chào mừng đến với HR Management System",
		Body:    body,
		IsHTML:  true,
	})
}

// Payslip Email
type PayslipData struct {
	Name       string
	Period     string
	NetSalary  string
	AppName    string
}

func (e *EmailService) SendPayslip(ctx context.Context, to, name, period, netSalary string, pdfContent []byte) error {
	data := PayslipData{
		Name:      name,
		Period:    period,
		NetSalary: netSalary,
		AppName:   "HR Management System",
	}

	body, err := e.renderTemplate("payslip", data)
	if err != nil {
		return err
	}

	return e.Send(ctx, Email{
		To:      []string{to},
		Subject: fmt.Sprintf("Phiếu lương tháng %s", period),
		Body:    body,
		IsHTML:  true,
		Attachments: []Attachment{
			{
				Filename: filepath.Base(fmt.Sprintf("payslip_%s.pdf", period)),
				Content:  pdfContent,
				MimeType: "application/pdf",
			},
		},
	})
}

// Leave Request Email
type LeaveRequestData struct {
	EmployeeName string
	LeaveType    string
	StartDate    string
	EndDate      string
	TotalDays    float64
	Reason       string
	ApproveURL   string
	AppName      string
}

func (e *EmailService) SendLeaveRequest(ctx context.Context, to string, data LeaveRequestData) error {
	body, err := e.renderTemplate("leave_request", data)
	if err != nil {
		return err
	}

	return e.Send(ctx, Email{
		To:      []string{to},
		Subject: fmt.Sprintf("Yêu cầu nghỉ phép từ %s", data.EmployeeName),
		Body:    body,
		IsHTML:  true,
	})
}

// Leave Approved Email
type LeaveApprovedData struct {
	Name      string
	LeaveType string
	StartDate string
	EndDate   string
	Status    string
	Notes     string
	AppName   string
}

func (e *EmailService) SendLeaveApproved(ctx context.Context, to string, data LeaveApprovedData) error {
	body, err := e.renderTemplate("leave_approved", data)
	if err != nil {
		return err
	}

	subject := "Đơn nghỉ phép đã được phê duyệt"
	if data.Status == "rejected" {
		subject = "Đơn nghỉ phép bị từ chối"
	}

	return e.Send(ctx, Email{
		To:      []string{to},
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	})
}

// Email templates
var otpTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #2563eb; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background: #f9fafb; }
        .otp-box { background: #2563eb; color: white; font-size: 32px; letter-spacing: 8px; padding: 20px; text-align: center; margin: 20px 0; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.AppName}}</h1>
        </div>
        <div class="content">
            <p>Xin chào {{.Name}},</p>
            <p>Mã xác thực OTP của bạn là:</p>
            <div class="otp-box">{{.OTP}}</div>
            <p>Mã này sẽ hết hạn sau {{.Expiry}}.</p>
            <p>Nếu bạn không yêu cầu mã này, vui lòng bỏ qua email này.</p>
        </div>
        <div class="footer">
            <p>&copy; {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

var passwordResetTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #2563eb; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background: #f9fafb; }
        .button { background: #2563eb; color: white; padding: 12px 24px; text-decoration: none; display: inline-block; margin: 20px 0; border-radius: 4px; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.AppName}}</h1>
        </div>
        <div class="content">
            <p>Xin chào {{.Name}},</p>
            <p>Bạn đã yêu cầu đặt lại mật khẩu. Nhấn vào nút bên dưới để tiếp tục:</p>
            <p style="text-align: center;">
                <a href="{{.ResetLink}}" class="button">Đặt lại mật khẩu</a>
            </p>
            <p>Link này sẽ hết hạn sau {{.Expiry}}.</p>
            <p>Nếu bạn không yêu cầu đặt lại mật khẩu, vui lòng bỏ qua email này.</p>
        </div>
        <div class="footer">
            <p>&copy; {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

var welcomeTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #2563eb; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background: #f9fafb; }
        .credentials { background: #e5e7eb; padding: 15px; margin: 20px 0; border-radius: 4px; }
        .button { background: #2563eb; color: white; padding: 12px 24px; text-decoration: none; display: inline-block; margin: 20px 0; border-radius: 4px; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Chào mừng đến với {{.AppName}}</h1>
        </div>
        <div class="content">
            <p>Xin chào {{.Name}},</p>
            <p>Tài khoản của bạn đã được tạo thành công. Dưới đây là thông tin đăng nhập:</p>
            <div class="credentials">
                <p><strong>Email:</strong> {{.Email}}</p>
                <p><strong>Mật khẩu tạm thời:</strong> {{.TempPassword}}</p>
            </div>
            <p>Vui lòng đổi mật khẩu sau khi đăng nhập lần đầu.</p>
            <p style="text-align: center;">
                <a href="{{.LoginURL}}" class="button">Đăng nhập ngay</a>
            </p>
        </div>
        <div class="footer">
            <p>&copy; {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

var payslipTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #2563eb; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background: #f9fafb; }
        .salary-box { background: #10b981; color: white; font-size: 24px; padding: 20px; text-align: center; margin: 20px 0; border-radius: 4px; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Phiếu lương tháng {{.Period}}</h1>
        </div>
        <div class="content">
            <p>Xin chào {{.Name}},</p>
            <p>Phiếu lương tháng {{.Period}} của bạn đã sẵn sàng.</p>
            <div class="salary-box">
                <p>Lương thực nhận</p>
                <p style="font-size: 32px; margin: 0;">{{.NetSalary}} VNĐ</p>
            </div>
            <p>Chi tiết phiếu lương được đính kèm trong file PDF.</p>
        </div>
        <div class="footer">
            <p>&copy; {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

var leaveRequestTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #f59e0b; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background: #f9fafb; }
        .info-box { background: #e5e7eb; padding: 15px; margin: 20px 0; border-radius: 4px; }
        .button { padding: 12px 24px; text-decoration: none; display: inline-block; margin: 10px 5px; border-radius: 4px; }
        .approve { background: #10b981; color: white; }
        .reject { background: #ef4444; color: white; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Yêu cầu nghỉ phép mới</h1>
        </div>
        <div class="content">
            <p>Nhân viên <strong>{{.EmployeeName}}</strong> đã gửi yêu cầu nghỉ phép:</p>
            <div class="info-box">
                <p><strong>Loại nghỉ phép:</strong> {{.LeaveType}}</p>
                <p><strong>Từ ngày:</strong> {{.StartDate}}</p>
                <p><strong>Đến ngày:</strong> {{.EndDate}}</p>
                <p><strong>Tổng số ngày:</strong> {{.TotalDays}}</p>
                <p><strong>Lý do:</strong> {{.Reason}}</p>
            </div>
            <p style="text-align: center;">
                <a href="{{.ApproveURL}}" class="button approve">Xem chi tiết</a>
            </p>
        </div>
        <div class="footer">
            <p>&copy; {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

var leaveApprovedTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #10b981; color: white; padding: 20px; text-align: center; }
        .header.rejected { background: #ef4444; }
        .content { padding: 20px; background: #f9fafb; }
        .info-box { background: #e5e7eb; padding: 15px; margin: 20px 0; border-radius: 4px; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header {{if eq .Status "rejected"}}rejected{{end}}">
            <h1>{{if eq .Status "approved"}}Đơn nghỉ phép đã được duyệt{{else}}Đơn nghỉ phép bị từ chối{{end}}</h1>
        </div>
        <div class="content">
            <p>Xin chào {{.Name}},</p>
            <p>Đơn xin nghỉ phép của bạn đã được {{if eq .Status "approved"}}phê duyệt{{else}}từ chối{{end}}.</p>
            <div class="info-box">
                <p><strong>Loại nghỉ phép:</strong> {{.LeaveType}}</p>
                <p><strong>Từ ngày:</strong> {{.StartDate}}</p>
                <p><strong>Đến ngày:</strong> {{.EndDate}}</p>
                {{if .Notes}}<p><strong>Ghi chú:</strong> {{.Notes}}</p>{{end}}
            </div>
        </div>
        <div class="footer">
            <p>&copy; {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

var overtimeTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #8b5cf6; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background: #f9fafb; }
        .info-box { background: #e5e7eb; padding: 15px; margin: 20px 0; border-radius: 4px; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Thông báo tăng ca</h1>
        </div>
        <div class="content">
            <p>Thông báo liên quan đến yêu cầu tăng ca.</p>
        </div>
        <div class="footer">
            <p>&copy; HR Management System. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`
