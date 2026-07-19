package model_test

import (
	"github.com/yixian-huang/inkless/backend/internal/model"
	"testing"
)

func TestSiteConfig_Validate_ValidKeys(t *testing.T) {
	for _, k := range []string{"global", "theme", "system"} {
		sc := &model.SiteConfig{Key: k}
		if err := sc.Validate(); err != nil {
			t.Errorf("unexpected error for key %q: %v", k, err)
		}
	}
}

func TestSiteConfig_Validate_InvalidKey(t *testing.T) {
	sc := &model.SiteConfig{Key: "invalid"}
	if err := sc.Validate(); err == nil {
		t.Error("expected error for invalid key")
	}
}

func TestSiteConfig_Validate_EmptyKey(t *testing.T) {
	sc := &model.SiteConfig{Key: ""}
	if err := sc.Validate(); err == nil {
		t.Error("expected error for empty key")
	}
}
