# Email Settings for Contact Form Notifications

## Overview

Add email configuration to system settings, enabling automatic email notifications when users submit the contact form. Two independent email channels: auto-reply to the user and forwarding to a configured admin inbox.

## Requirements

1. **SMTP configuration** — admin configures SMTP server credentials in the backend
2. **Auto-reply** — send a confirmation email to the user after form submission (toggleable)
3. **Forward notification** — forward form content to a configured admin inbox (toggleable)
4. **Reply-To** — forwarded emails set `Reply-To` to the user's email so admin can reply directly
5. **HTML templates** — 4 templates (auto-reply zh/en + forward zh/en), editable via CodeMirror 6 HTML editor
6. **Template variables** — `{{name}}`, `{{email}}`, `{{phone}}`, `{{company}}`, `{{message}}`, `{{date}}` (date format: `2026-03-17 14:30` in zh, `Mar 17, 2026 2:30 PM` in en)
7. **Test email** — admin can send a test email to verify SMTP configuration
8. **Non-blocking** — email sending must not block or fail the form submission

## Data Model

Email configuration is stored in `SiteConfig` with key `"email"`, reusing the existing draft/publish versioning mechanism.

### Config Structure

```json
{
  "smtp": {
    "host": "smtp.gmail.com",
    "port": 587,
    "username": "noreply@example.com",
    "password": "app-password-here",
    "from": "noreply@example.com",
    "fromName": "印迹咨询",
    "useTLS": false,
    "insecureSkipVerify": false
  },
  "receiver": {
    "enabled": true,
    "email": "admin@example.com"
  },
  "autoReply": {
    "enabled": true
  },
  "templates": {
    "autoReply": {
      "zh": { "subject": "感谢您的联系", "body": "<html>..." },
      "en": { "subject": "Thank you for contacting us", "body": "<html>..." }
    },
    "forward": {
      "zh": { "subject": "新的联系表单：{{name}}", "body": "<html>..." },
      "en": { "subject": "New contact form: {{name}}", "body": "<html>..." }
    }
  }
}
```

### Why SiteConfig

- Email config is site-wide, fits naturally alongside `"global"` and `"theme"` keys
- Only change needed: add `"email"` to the allowed key validation in `site_config.go`

### Save Semantics

Email config uses **immediate publish** (same as the theme handler): `PUT` writes to both `draftConfig` and `publishedConfig` simultaneously. There is no separate draft/publish workflow. The email sending flow always reads `publishedConfig`, which is always in sync with the latest save.

### Password Handling

- Stored as plaintext in the database (same trust model as other SiteConfig data)
- API response masks the password field as `"****"`
- PUT request: if password field is `"****"`, the backend preserves the existing password

## Backend Architecture

### New Files

**`internal/service/email_service.go`** — Core email service

- `EmailService` struct with dependency on `SiteConfigRepository`
- Methods:
  - `SendAutoReply(ctx, submission, config)` — render auto-reply template by submission locale, send to user
  - `SendForward(ctx, submission, config)` — render forward template, send to receiver email with `Reply-To: <user email>`
  - `SendTest(ctx, to, config)` — send test email to verify SMTP works
  - `LoadConfig(ctx)` — read and parse email config from SiteConfig published config
  - `renderTemplate(templateStr, submission)` — replace `{{var}}` placeholders using `strings.NewReplacer`
- SMTP sending: implements its own SMTP transport (TLS/STARTTLS), following the same patterns as `internal/plugins/email/plugin.go` but as independent code. The existing plugin's methods are unexported and not directly importable. `EmailService` builds its own RFC 5322 message with `Content-Type: text/html; charset=UTF-8` and `Reply-To` header support. The existing plugin's `buildMessage()` is NOT reused because it hardcodes `text/plain` and has no `Reply-To` support.
- Does **not** implement `NotifierProvider`. The form submission handler calls `EmailService` methods directly. The provider registry retains `LogNotifier` as a generic fallback; `EmailService` is a domain-specific service, not a generic notifier.

**`internal/handler/email_settings/handler.go`** — Admin API handler

Dedicated handler (not reusing the theme handler) because email config requires password masking on read and test-email functionality — logic that doesn't exist in the generic SiteConfig flow.

- `GET /admin/email-settings` — returns published email config with password masked
- `PUT /admin/email-settings` — writes to both draftConfig and publishedConfig simultaneously (immediate publish, same as theme handler). Preserves password if masked value `"****"` is sent.
- `POST /admin/email-settings/test` — accepts `{ "to": "test@example.com" }`, sends test email using current config, returns success/error message

### Modified Files

**`internal/model/site_config.go`**
- Add `"email"` to the set of valid SiteConfig keys

**`internal/handler/form_submission/handler.go`**
- Add `EmailService` as a dependency to the `Handler` struct (constructor signature changes: `NewHandler(repo, emailSvc)`)
- In `HandlePublicSubmit`: after saving submission to DB and returning 201, launch a goroutine with a detached context (request context is cancelled after response):
  ```go
  go func() {
      bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
      defer cancel()

      config := emailSvc.LoadConfig(bgCtx)
      if config == nil || !config.SMTP.IsConfigured() {
          return
      }
      if config.Receiver.Enabled {
          if err := emailSvc.SendForward(bgCtx, submission, config); err != nil {
              logger.Error("forward email failed", "error", err)
          }
      }
      if config.AutoReply.Enabled {
          if err := emailSvc.SendAutoReply(bgCtx, submission, config); err != nil {
              logger.Error("auto-reply email failed", "error", err)
          }
      }
  }()
  ```

**`cmd/server/main.go`**
- Create `EmailService` with `siteConfigRepo`
- Pass `EmailService` to form submission handler: `formHandler := form_submission.NewHandler(formRepo, emailSvc)`
- Register email settings routes under `/admin/email-settings`

## Email Sending Flow

```
User submits contact form
  → POST /public/form-submissions
  → handler saves to DB → returns 201 immediately
  → goroutine starts:
      1. Read SiteConfig key="email" (published config)
      2. If SMTP not configured → skip silently
      3. If receiver.enabled → send forward email
         - To: receiver.email
         - Reply-To: submission.email (user's email)
         - Subject/Body: forward template[submission.locale]
      4. If autoReply.enabled → send auto-reply email
         - To: submission.email (user's email)
         - Subject/Body: autoReply template[submission.locale]
      5. On any send failure → log error, continue
```

### Error Handling

- **SMTP not configured**: silent skip, no error
- **Send failure**: log error with submission ID, do not retry
- **Template variable missing**: render as empty string
- **Locale not found**: fall back to `"zh"` template
- **Config read failure**: log error, skip email sending

### Why No Retry

Form volume is low (consultancy website). Failed emails are not critical — admin can see all submissions in the admin panel. Adding retry queues would be overengineering.

## Frontend Architecture

### New Files

**`src/api/emailSettings.ts`** — API client

```typescript
export async function getEmailSettings(): Promise<EmailConfig>
export async function updateEmailSettings(config: EmailConfig): Promise<EmailConfig>
export async function sendTestEmail(to: string): Promise<{ message: string }>
```

**`src/pages/admin/email-settings/page.tsx`** — Settings page

Three-tab layout:

1. **SMTP 配置 Tab**
   - SMTP server form (host, port, username, password, from, fromName)
   - TLS checkbox, skip-verify checkbox
   - Receiver section: enable toggle + email input
   - Auto-reply section: enable toggle
   - "发送测试邮件" button + "保存配置" button

2. **自动回复模板 Tab**
   - zh/en language toggle (segmented control)
   - Subject input with variable hint text
   - HTML body editor (CodeMirror 6 with `@codemirror/lang-html`, lighter than Monaco ~5-10MB)
   - "预览邮件" button (renders HTML in sandboxed iframe) + "保存模板" button

3. **转发通知模板 Tab**
   - Same layout as auto-reply tab

### Modified Files

**`src/router/config.tsx`**
- Add lazy-loaded route: `/admin/email-settings` → `pages/admin/email-settings/page.tsx`

**Admin sidebar component** (`AdminSidebar.tsx` or equivalent)
- Add "邮箱设置" navigation entry with mail icon in the "系统" (System) group, near "存储配置"

### Editor Choice

CodeMirror 6 (`@codemirror/view`, `@codemirror/lang-html`, `@codemirror/theme-one-dark`). Chosen over Monaco Editor because Monaco requires ~5-10MB of web worker scripts and special Vite configuration. CodeMirror 6 is modular, tree-shakeable, and provides HTML syntax highlighting, auto-completion, and bracket matching with a much smaller footprint.

### Default Templates

The system ships with built-in default templates for all 4 slots. When email config is first created (or a template field is empty), these defaults are used. The admin UI pre-populates empty template editors with the defaults. Default templates include basic branded HTML with variable placeholders — simple and functional, not elaborate.

## Plugin System Impact

### What Changes

- `EmailService` implements its own SMTP transport following the same patterns as `internal/plugins/email/plugin.go` (the plugin's methods are unexported)
- This validates that plugin code can be consumed as a library (in-process) without going through the gRPC process model
- The provider registry pattern is proven: `EmailService` is a concrete service that could later be wrapped as a `NotifierProvider` if generic notification dispatch is needed

### What Doesn't Change

- Plugin Manager, gRPC host, state machine — untouched
- Plugin manifest/sandbox — untouched
- `internal/plugins/email/plugin.go` — not modified; its patterns are referenced but code is independent
- Provider registry — `LogNotifier` remains as the default generic notifier

### Future Path

Future notification channels (webhook, SMS, Slack) can either: (a) implement `NotifierProvider` for the registry, or (b) follow the `EmailService` pattern as domain-specific services. The plugin's SMTP transport is now a reusable building block.

## File Change Summary

### Backend — New
| File | Purpose |
|------|---------|
| `internal/service/email_service.go` | Email sending service (template rendering + SMTP) |
| `internal/handler/email_settings/handler.go` | Admin API for email config |

### Backend — Modified
| File | Change |
|------|--------|
| `internal/model/site_config.go` | Add `"email"` to valid keys |
| `internal/handler/form_submission/handler.go` | Add EmailService dep, inject async email sending after submit |
| `cmd/server/main.go` | Create EmailService, pass to form handler, register email settings routes |

### Frontend — New
| File | Purpose |
|------|---------|
| `src/api/emailSettings.ts` | Email settings API client |
| `src/pages/admin/email-settings/page.tsx` | Email settings admin page |

### Frontend — Modified
| File | Change |
|------|--------|
| `src/router/config.tsx` | Add email settings route |
| Admin sidebar component | Add navigation entry |

## Out of Scope

- Email delivery tracking / open tracking
- Retry queues or dead letter handling
- Multi-recipient forwarding (single receiver email only)
- Attachment support
- DKIM/SPF configuration (handled at DNS level, not in app)
- Plugin Manager changes or gRPC protocol updates
