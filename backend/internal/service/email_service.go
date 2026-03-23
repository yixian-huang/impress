package service

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/smtp"
	"strings"
	"time"

	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/repository"
)

// ---------- Config types ----------

// EmailConfig is the top-level email configuration stored in SiteConfig key "email".
type EmailConfig struct {
	SMTP      SMTPConfig      `json:"smtp"`
	Receiver  ReceiverConfig  `json:"receiver"`
	AutoReply AutoReplyConfig `json:"autoReply"`
	Templates TemplatesConfig `json:"templates"`
}

// SMTPConfig holds SMTP server settings.
type SMTPConfig struct {
	Host               string `json:"host"`
	Port               int    `json:"port"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	From               string `json:"from"`
	FromName           string `json:"fromName"`
	UseTLS             bool   `json:"useTLS"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify"`
}

// IsConfigured returns true when the minimum SMTP fields are present.
func (s *SMTPConfig) IsConfigured() bool {
	return s.Host != "" && s.Port > 0 && s.From != ""
}

// ReceiverConfig controls forwarding to admin inbox(es).
type ReceiverConfig struct {
	Enabled bool   `json:"enabled"`
	Emails  string `json:"emails"`
}

// AutoReplyConfig controls the auto-reply feature.
type AutoReplyConfig struct {
	Enabled bool `json:"enabled"`
}

// TemplatesConfig groups templates by type, each with locale variants.
type TemplatesConfig struct {
	AutoReply map[string]EmailTemplate `json:"autoReply"`
	Forward   map[string]EmailTemplate `json:"forward"`
}

// EmailTemplate holds the subject and HTML body template.
type EmailTemplate struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// ---------- Service ----------

// EmailService handles loading email config and rendering/sending emails.
type EmailService struct {
	siteConfigRepo repository.SiteConfigRepository
}

// NewEmailService creates a new EmailService.
func NewEmailService(repo repository.SiteConfigRepository) *EmailService {
	return &EmailService{siteConfigRepo: repo}
}

// LoadConfig reads the "email" SiteConfig and returns the parsed EmailConfig, or nil.
func (s *EmailService) LoadConfig(ctx context.Context) *EmailConfig {
	doc, err := s.siteConfigRepo.FindByKey(ctx, model.SiteConfigKeyEmail)
	if err != nil {
		slog.Warn("email config not found", "error", err)
		return nil
	}
	if len(doc.PublishedConfig) == 0 {
		return nil
	}
	raw, err := json.Marshal(doc.PublishedConfig)
	if err != nil {
		slog.Error("failed to marshal email config", "error", err)
		return nil
	}
	var cfg EmailConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		slog.Error("failed to unmarshal email config", "error", err)
		return nil
	}
	return &cfg
}

// renderTemplate replaces placeholders in tmpl with values from submission.
func (s *EmailService) renderTemplate(tmpl string, submission *model.FormSubmission) string {
	locale := submission.Locale
	if locale == "" {
		locale = "zh"
	}

	var dateStr string
	if locale == "en" {
		dateStr = submission.CreatedAt.Format("Jan 02, 2006 3:04 PM")
	} else {
		dateStr = submission.CreatedAt.Format("2006-01-02 15:04")
	}

	r := strings.NewReplacer(
		"{{name}}", submission.Name,
		"{{email}}", submission.Email,
		"{{phone}}", submission.Phone,
		"{{company}}", submission.Company,
		"{{message}}", submission.Message,
		"{{date}}", dateStr,
	)
	return r.Replace(tmpl)
}

// getTemplate returns the template for the given locale, falling back to "zh".
func (s *EmailService) getTemplate(templates map[string]EmailTemplate, locale string) EmailTemplate {
	if t, ok := templates[locale]; ok {
		return t
	}
	if t, ok := templates["zh"]; ok {
		return t
	}
	return EmailTemplate{}
}

// ---------- SMTP Transport ----------

// buildHTMLMessage constructs an RFC 5322 compliant email message with HTML content.
func buildHTMLMessage(from, fromName, to, replyTo, subject, body string) []byte {
	var b strings.Builder
	if fromName != "" {
		fmt.Fprintf(&b, "From: %s <%s>\r\n", fromName, from)
	} else {
		fmt.Fprintf(&b, "From: %s\r\n", from)
	}
	fmt.Fprintf(&b, "To: %s\r\n", to)
	if replyTo != "" {
		fmt.Fprintf(&b, "Reply-To: %s\r\n", replyTo)
	}
	fmt.Fprintf(&b, "Subject: %s\r\n", subject)
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	return []byte(b.String())
}

// sendMail dispatches the message via TLS or STARTTLS based on configuration.
func (s *EmailService) sendMail(cfg *SMTPConfig, to, replyTo, subject, htmlBody string) error {
	msg := buildHTMLMessage(cfg.From, cfg.FromName, to, replyTo, subject, htmlBody)
	if cfg.UseTLS {
		return s.sendTLS(cfg, to, msg)
	}
	return s.sendSTARTTLS(cfg, to, msg)
}

// loginAuth implements the SMTP LOGIN authentication mechanism.
// Many Chinese email providers (QQ, 163, etc.) only support LOGIN, not PLAIN.
type loginAuth struct {
	username, password string
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte(a.username), nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		}
	}
	return nil, nil
}

// smtpAuth returns the appropriate auth mechanism: tries LOGIN first (for QQ/163 compatibility),
// falls back to PLAIN if LOGIN is not advertised.
func smtpAuth(username, password, host string, extensions map[string]string) smtp.Auth {
	// If server advertises AUTH mechanisms, check for LOGIN support
	if mechs, ok := extensions["AUTH"]; ok {
		if strings.Contains(strings.ToUpper(mechs), "LOGIN") {
			return &loginAuth{username: username, password: password}
		}
	}
	// Default: try LOGIN anyway (many servers support it without advertising)
	return &loginAuth{username: username, password: password}
}

// sendSTARTTLS connects via plain TCP and upgrades with STARTTLS.
func (s *EmailService) sendSTARTTLS(cfg *SMTPConfig, to string, msg []byte) error {
	addr := net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", cfg.Port))
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("dial %s: %w", addr, err)
	}

	client, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("smtp new client: %w", err)
	}
	defer client.Close()

	tlsCfg := &tls.Config{
		ServerName:         cfg.Host,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
	}
	if err := client.StartTLS(tlsCfg); err != nil {
		return fmt.Errorf("starttls: %w", err)
	}

	if cfg.Username != "" {
		auth := &loginAuth{username: cfg.Username, password: cfg.Password}
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("auth: %w", err)
		}
	}

	if err := client.Mail(cfg.From); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("rcpt to: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close data: %w", err)
	}

	return client.Quit()
}

// sendTLS connects via implicit TLS (typically port 465).
func (s *EmailService) sendTLS(cfg *SMTPConfig, to string, msg []byte) error {
	addr := net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", cfg.Port))
	tlsCfg := &tls.Config{
		ServerName:         cfg.Host,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
	}
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, tlsCfg)
	if err != nil {
		return fmt.Errorf("tls dial %s: %w", addr, err)
	}

	client, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("smtp new client: %w", err)
	}
	defer client.Close()

	if cfg.Username != "" {
		auth := &loginAuth{username: cfg.Username, password: cfg.Password}
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("auth: %w", err)
		}
	}

	if err := client.Mail(cfg.From); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("rcpt to: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close data: %w", err)
	}

	return client.Quit()
}

// ---------- High-level send methods ----------

// SendAutoReply sends an auto-reply email to the form submitter.
func (s *EmailService) SendAutoReply(ctx context.Context, submission *model.FormSubmission, cfg *EmailConfig) error {
	if !cfg.AutoReply.Enabled {
		return nil
	}
	if submission.Email == "" {
		return nil
	}

	tmpl := s.getTemplate(cfg.Templates.AutoReply, submission.Locale)
	if tmpl.Subject == "" && tmpl.Body == "" {
		return nil
	}

	subject := s.renderTemplate(tmpl.Subject, submission)
	body := s.renderTemplate(tmpl.Body, submission)
	return s.sendMail(&cfg.SMTP, submission.Email, "", subject, body)
}

// SendForward sends the form submission to all configured receivers, with Reply-To set to the submitter's email.
func (s *EmailService) SendForward(ctx context.Context, submission *model.FormSubmission, cfg *EmailConfig) error {
	if !cfg.Receiver.Enabled || cfg.Receiver.Emails == "" {
		return nil
	}

	tmpl := s.getTemplate(cfg.Templates.Forward, submission.Locale)
	if tmpl.Subject == "" && tmpl.Body == "" {
		return nil
	}

	subject := s.renderTemplate(tmpl.Subject, submission)
	body := s.renderTemplate(tmpl.Body, submission)

	var lastErr error
	for _, raw := range strings.Split(cfg.Receiver.Emails, ",") {
		to := strings.TrimSpace(raw)
		if to == "" {
			continue
		}
		if err := s.sendMail(&cfg.SMTP, to, submission.Email, subject, body); err != nil {
			slog.Error("forward email failed", "to", to, "error", err)
			lastErr = err
		}
	}
	return lastErr
}

// SendTest sends a test email to verify SMTP configuration.
func (s *EmailService) SendTest(ctx context.Context, to string, cfg *EmailConfig) error {
	if !cfg.SMTP.IsConfigured() {
		return fmt.Errorf("SMTP not configured")
	}

	subject := "Impress CMS - Email Test / 邮件测试"
	body := `<html><body>
<h2>Email Configuration Test / 邮件配置测试</h2>
<p>If you receive this email, your SMTP settings are working correctly.</p>
<p>如果您收到此邮件，说明 SMTP 配置正确。</p>
</body></html>`

	return s.sendMail(&cfg.SMTP, to, "", subject, body)
}
