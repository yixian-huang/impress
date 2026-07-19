package audit

import (
	"time"

	"github.com/yixian-huang/inkless/backend/pkg/logger"
)

// Event represents a structured audit event
type Event struct {
	Timestamp time.Time              `json:"timestamp"`
	Action    string                 `json:"action"`
	Actor     string                 `json:"actor"`
	Resource  string                 `json:"resource"`
	Result    string                 `json:"result"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// Logger wraps logger for structured audit events
type Logger struct {
	log *logger.Logger
}

// NewLogger creates a new audit logger
func NewLogger(log *logger.Logger) *Logger {
	return &Logger{log: log}
}

// Log emits a structured audit event
func (a *Logger) Log(event Event) {
	event.Timestamp = time.Now()
	a.log.Info("audit_event",
		"timestamp", event.Timestamp.Format(time.RFC3339),
		"action", event.Action,
		"actor", event.Actor,
		"resource", event.Resource,
		"result", event.Result,
		"details", event.Details,
	)
}

// LogPublishSuccess records a successful publish operation
func (a *Logger) LogPublishSuccess(pageKey string, publishedVersion int, actor string, draftVersion int) {
	a.Log(Event{
		Action:   "content.publish",
		Actor:    actor,
		Resource: pageKey,
		Result:   "success",
		Details: map[string]interface{}{
			"published_version": publishedVersion,
			"draft_version":     draftVersion,
		},
	})
}

// LogPublishFailure records a failed publish operation
func (a *Logger) LogPublishFailure(pageKey string, actor string, reason string, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["reason"] = reason
	a.Log(Event{
		Action:   "content.publish",
		Actor:    actor,
		Resource: pageKey,
		Result:   "failure",
		Details:  details,
	})
}

// LogRollbackSuccess records a successful rollback operation
func (a *Logger) LogRollbackSuccess(pageKey string, publishedVersion int, sourceVersion int, actor string) {
	a.Log(Event{
		Action:   "content.rollback",
		Actor:    actor,
		Resource: pageKey,
		Result:   "success",
		Details: map[string]interface{}{
			"published_version": publishedVersion,
			"source_version":    sourceVersion,
		},
	})
}

// LogRollbackFailure records a failed rollback operation
func (a *Logger) LogRollbackFailure(pageKey string, actor string, sourceVersion int, reason string) {
	a.Log(Event{
		Action:   "content.rollback",
		Actor:    actor,
		Resource: pageKey,
		Result:   "failure",
		Details: map[string]interface{}{
			"source_version": sourceVersion,
			"reason":         reason,
		},
	})
}

// LogValidation records a validation operation
func (a *Logger) LogValidation(pageKey string, actor string, valid bool, errorCount int, translationIssueCount int) {
	result := "success"
	if !valid {
		result = "failure"
	}
	a.Log(Event{
		Action:   "content.validate",
		Actor:    actor,
		Resource: pageKey,
		Result:   result,
		Details: map[string]interface{}{
			"valid":                   valid,
			"error_count":             errorCount,
			"translation_issue_count": translationIssueCount,
		},
	})
}
