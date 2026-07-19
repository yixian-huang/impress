package global_config

import (
	"strings"
	"testing"

	"github.com/yixian-huang/inkless/backend/internal/model"
)

func validBase() model.JSONMap {
	return model.JSONMap{
		"identity": map[string]any{
			"name":          map[string]any{"zh": "My Site"},
			"localeMode":    "mono-zh",
			"defaultLocale": "zh",
		},
		"brand":  map[string]any{},
		"author": map[string]any{"socials": []any{}},
		"footer": map[string]any{},
		"seo":    map[string]any{},
	}
}

func TestValidateGlobalConfig_AcceptsMinimalValid(t *testing.T) {
	cfg, err := validateGlobalConfig(validBase())
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if cfg.Identity.LocaleMode != LocaleModeMonoZh {
		t.Errorf("got localeMode=%q want mono-zh", cfg.Identity.LocaleMode)
	}
}

func TestValidateGlobalConfig_RejectsEmptyName(t *testing.T) {
	raw := validBase()
	raw["identity"].(map[string]any)["name"] = map[string]any{}
	_, err := validateGlobalConfig(raw)
	if err == nil || !strings.Contains(err.Error(), "identity.name") {
		t.Fatalf("expected identity.name error, got: %v", err)
	}
}

func TestValidateGlobalConfig_RejectsInvalidLocaleMode(t *testing.T) {
	raw := validBase()
	raw["identity"].(map[string]any)["localeMode"] = "klingon"
	_, err := validateGlobalConfig(raw)
	if err == nil || !strings.Contains(err.Error(), "localeMode") {
		t.Fatalf("expected localeMode error, got: %v", err)
	}
}

func TestValidateGlobalConfig_RejectsDefaultLocaleMismatch(t *testing.T) {
	raw := validBase()
	raw["identity"].(map[string]any)["localeMode"] = "mono-en"
	raw["identity"].(map[string]any)["defaultLocale"] = "zh"
	_, err := validateGlobalConfig(raw)
	if err == nil || !strings.Contains(err.Error(), "defaultLocale") {
		t.Fatalf("expected defaultLocale mismatch error, got: %v", err)
	}
}

func TestValidateGlobalConfig_RejectsLongICP(t *testing.T) {
	raw := validBase()
	raw["footer"] = map[string]any{"icp": strings.Repeat("a", 101)}
	_, err := validateGlobalConfig(raw)
	if err == nil || !strings.Contains(err.Error(), "footer.icp") {
		t.Fatalf("expected footer.icp error, got: %v", err)
	}
}

func TestValidateGlobalConfig_RejectsSocialWithoutURL(t *testing.T) {
	raw := validBase()
	raw["author"] = map[string]any{
		"socials": []any{map[string]any{"kind": "github"}},
	}
	_, err := validateGlobalConfig(raw)
	if err == nil || !strings.Contains(err.Error(), "socials[0].url") {
		t.Fatalf("expected socials[0].url error, got: %v", err)
	}
}
