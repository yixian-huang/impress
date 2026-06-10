package repository

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// IsNotFound reports whether err indicates a missing database row.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "not found") || strings.Contains(msg, "record not found")
}
