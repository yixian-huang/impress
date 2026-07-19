package comment

import (
	"testing"

	"github.com/yixian-huang/inkless/backend/internal/model"
)

func TestAuthorNameFromGlobalConfig(t *testing.T) {
	cfg := model.JSONMap{
		"author": map[string]interface{}{"name": "Alice"},
	}
	if got := authorNameFromGlobalConfig(cfg); got != "Alice" {
		t.Fatalf("author name = %q, want Alice", got)
	}
}

func TestAuthorNameFromIdentityFallback(t *testing.T) {
	cfg := model.JSONMap{
		"identity": map[string]interface{}{
			"name": map[string]interface{}{"zh": "站点名"},
		},
	}
	if got := authorNameFromGlobalConfig(cfg); got != "站点名" {
		t.Fatalf("identity fallback = %q", got)
	}
}
