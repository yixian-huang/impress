package service

import (
	"testing"
	"time"

	"github.com/yixian-huang/inkless/backend/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestRenderTemplate_Zh(t *testing.T) {
	svc := &EmailService{}
	sub := &model.FormSubmission{
		Name:    "张三",
		Email:   "zhang@example.com",
		Phone:   "13800138000",
		Company: "测试公司",
		Message: "你好，我想咨询一下。",
		Locale:  "zh",
	}
	sub.CreatedAt = time.Date(2025, 6, 15, 14, 30, 0, 0, time.UTC)

	tmpl := "姓名: {{name}}, 邮箱: {{email}}, 电话: {{phone}}, 公司: {{company}}, 消息: {{message}}, 日期: {{date}}"
	result := svc.renderTemplate(tmpl, sub)

	assert.Contains(t, result, "张三")
	assert.Contains(t, result, "zhang@example.com")
	assert.Contains(t, result, "13800138000")
	assert.Contains(t, result, "测试公司")
	assert.Contains(t, result, "你好，我想咨询一下。")
	assert.Contains(t, result, "2025-06-15 14:30")
}

func TestRenderTemplate_En(t *testing.T) {
	svc := &EmailService{}
	sub := &model.FormSubmission{
		Name:    "John Doe",
		Email:   "john@example.com",
		Phone:   "+1-555-0100",
		Company: "Test Corp",
		Message: "Hello, I want to inquire.",
		Locale:  "en",
	}
	sub.CreatedAt = time.Date(2025, 6, 15, 14, 30, 0, 0, time.UTC)

	tmpl := "Name: {{name}}, Email: {{email}}, Phone: {{phone}}, Company: {{company}}, Message: {{message}}, Date: {{date}}"
	result := svc.renderTemplate(tmpl, sub)

	assert.Contains(t, result, "John Doe")
	assert.Contains(t, result, "john@example.com")
	assert.Contains(t, result, "+1-555-0100")
	assert.Contains(t, result, "Test Corp")
	assert.Contains(t, result, "Hello, I want to inquire.")
	assert.Contains(t, result, "Jun 15, 2025 2:30 PM")
}

func TestRenderTemplate_MissingVars(t *testing.T) {
	svc := &EmailService{}
	sub := &model.FormSubmission{
		Name:   "Alice",
		Email:  "alice@example.com",
		Locale: "zh",
	}
	sub.CreatedAt = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	tmpl := "Name: {{name}}, Phone: {{phone}}, Company: {{company}}"
	result := svc.renderTemplate(tmpl, sub)

	assert.Contains(t, result, "Alice")
	// Missing fields should be replaced with empty strings
	assert.Contains(t, result, "Phone: ,")
	assert.Contains(t, result, "Company: ")
}

func TestGetTemplate_ExactMatch(t *testing.T) {
	svc := &EmailService{}
	templates := map[string]EmailTemplate{
		"zh": {Subject: "中文主题", Body: "中文正文"},
		"en": {Subject: "English Subject", Body: "English Body"},
	}

	result := svc.getTemplate(templates, "en")
	assert.Equal(t, "English Subject", result.Subject)
	assert.Equal(t, "English Body", result.Body)
}

func TestGetTemplate_Fallback(t *testing.T) {
	svc := &EmailService{}
	templates := map[string]EmailTemplate{
		"zh": {Subject: "中文主题", Body: "中文正文"},
	}

	result := svc.getTemplate(templates, "en")
	assert.Equal(t, "中文主题", result.Subject)
	assert.Equal(t, "中文正文", result.Body)
}

func TestGetTemplate_Empty(t *testing.T) {
	svc := &EmailService{}
	templates := map[string]EmailTemplate{}

	result := svc.getTemplate(templates, "en")
	assert.Equal(t, "", result.Subject)
	assert.Equal(t, "", result.Body)
}

func TestSMTPConfig_IsConfigured(t *testing.T) {
	tests := []struct {
		name     string
		cfg      SMTPConfig
		expected bool
	}{
		{
			name: "fully configured",
			cfg: SMTPConfig{
				Host:     "smtp.example.com",
				Port:     465,
				Username: "user",
				Password: "pass",
				From:     "noreply@example.com",
			},
			expected: true,
		},
		{
			name:     "empty config",
			cfg:      SMTPConfig{},
			expected: false,
		},
		{
			name: "missing host",
			cfg: SMTPConfig{
				Port:     465,
				Username: "user",
				Password: "pass",
				From:     "noreply@example.com",
			},
			expected: false,
		},
		{
			name: "missing port",
			cfg: SMTPConfig{
				Host:     "smtp.example.com",
				Username: "user",
				Password: "pass",
				From:     "noreply@example.com",
			},
			expected: false,
		},
		{
			name: "missing from",
			cfg: SMTPConfig{
				Host: "smtp.example.com",
				Port: 465,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.cfg.IsConfigured())
		})
	}
}

func TestBuildHTMLMessage_WithReplyToAndFromName(t *testing.T) {
	msg := buildHTMLMessage(
		"noreply@example.com",
		"Example Sender",
		"user@example.com",
		"reply@example.com",
		"Test Subject",
		"<html><body>Hello</body></html>",
	)

	msgStr := string(msg)
	assert.Contains(t, msgStr, "From: Example Sender <noreply@example.com>\r\n")
	assert.Contains(t, msgStr, "To: user@example.com\r\n")
	assert.Contains(t, msgStr, "Reply-To: reply@example.com\r\n")
	assert.Contains(t, msgStr, "Subject: Test Subject\r\n")
	assert.Contains(t, msgStr, "Content-Type: text/html; charset=UTF-8\r\n")
	assert.Contains(t, msgStr, "<html><body>Hello</body></html>")
}

func TestBuildHTMLMessage_WithoutReplyToOrFromName(t *testing.T) {
	msg := buildHTMLMessage(
		"noreply@example.com",
		"",
		"user@example.com",
		"",
		"Test Subject",
		"<html><body>Hello</body></html>",
	)

	msgStr := string(msg)
	assert.Contains(t, msgStr, "From: noreply@example.com\r\n")
	assert.NotContains(t, msgStr, "Reply-To:")
	assert.Contains(t, msgStr, "Content-Type: text/html; charset=UTF-8\r\n")
}
