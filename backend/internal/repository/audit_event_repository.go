package repository

import (
	"context"
	"time"

	"github.com/yixian-huang/inkless/backend/internal/model"
)

// AuditEventRepository defines the interface for audit event data access
type AuditEventRepository interface {
	// Create creates a new audit event record
	Create(ctx context.Context, event *model.AuditEvent) error

	// List returns a filtered, paginated list of audit events
	List(ctx context.Context, offset, limit int, action, actor string, from, to *time.Time) ([]model.AuditEvent, int64, error)
}
