package sdk

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestSettings(schema map[string]any, config map[string]string) *SettingsManager {
	return newSettingsManager(schema, config)
}

func TestSettingsManager_Get_Default(t *testing.T) {
	schema := map[string]any{
		"properties": map[string]any{
			"timeout": map[string]any{
				"type":    "integer",
				"default": "30",
			},
		},
	}
	sm := newTestSettings(schema, nil)
	assert.Equal(t, "30", sm.Get("timeout"))
}

func TestSettingsManager_Get_Override(t *testing.T) {
	schema := map[string]any{
		"properties": map[string]any{
			"timeout": map[string]any{
				"type":    "integer",
				"default": "30",
			},
		},
	}
	sm := newTestSettings(schema, map[string]string{"timeout": "60"})
	assert.Equal(t, "60", sm.Get("timeout"))
}

func TestSettingsManager_GetBool(t *testing.T) {
	sm := newTestSettings(nil, map[string]string{"enabled": "true", "debug": "false"})
	v, err := sm.GetBool("enabled")
	require.NoError(t, err)
	assert.True(t, v)

	v, err = sm.GetBool("debug")
	require.NoError(t, err)
	assert.False(t, v)
}

func TestSettingsManager_GetInt(t *testing.T) {
	sm := newTestSettings(nil, map[string]string{"workers": "4"})
	n, err := sm.GetInt("workers")
	require.NoError(t, err)
	assert.Equal(t, 4, n)
}

func TestSettingsManager_GetFloat(t *testing.T) {
	sm := newTestSettings(nil, map[string]string{"threshold": "0.75"})
	f, err := sm.GetFloat("threshold")
	require.NoError(t, err)
	assert.InDelta(t, 0.75, f, 1e-9)
}

func TestSettingsManager_Set(t *testing.T) {
	sm := newTestSettings(nil, nil)
	sm.Set("key", "val")
	assert.Equal(t, "val", sm.Get("key"))
}

func TestSettingsManager_Validate_Required(t *testing.T) {
	schema := map[string]any{
		"properties": map[string]any{
			"apiKey": map[string]any{"type": "string"},
		},
		"required": []any{"apiKey"},
	}
	sm := newTestSettings(schema, nil)
	err := sm.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `"apiKey" is required`)
}

func TestSettingsManager_Validate_RequiredPresent(t *testing.T) {
	schema := map[string]any{
		"properties": map[string]any{
			"apiKey": map[string]any{"type": "string"},
		},
		"required": []any{"apiKey"},
	}
	sm := newTestSettings(schema, map[string]string{"apiKey": "abc123"})
	assert.NoError(t, sm.Validate())
}

func TestSettingsManager_Validate_BadBoolean(t *testing.T) {
	schema := map[string]any{
		"properties": map[string]any{
			"flag": map[string]any{"type": "boolean"},
		},
	}
	sm := newTestSettings(schema, map[string]string{"flag": "notabool"})
	err := sm.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a boolean")
}

func TestSettingsManager_Validate_IntegerRange(t *testing.T) {
	min := 1.0
	max := 10.0
	schema := map[string]any{
		"properties": map[string]any{
			"count": map[string]any{
				"type":    "integer",
				"minimum": min,
				"maximum": max,
			},
		},
	}
	sm := newTestSettings(schema, map[string]string{"count": "0"})
	err := sm.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), ">= 1")

	sm2 := newTestSettings(schema, map[string]string{"count": "11"})
	err = sm2.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "<= 10")

	sm3 := newTestSettings(schema, map[string]string{"count": "5"})
	assert.NoError(t, sm3.Validate())
}

func TestSettingsManager_Validate_StringLength(t *testing.T) {
	schema := map[string]any{
		"properties": map[string]any{
			"name": map[string]any{
				"type":      "string",
				"minLength": 3,
				"maxLength": 10,
			},
		},
	}
	sm := newTestSettings(schema, map[string]string{"name": "ab"})
	assert.Error(t, sm.Validate())

	sm2 := newTestSettings(schema, map[string]string{"name": "toolongstring"})
	assert.Error(t, sm2.Validate())

	sm3 := newTestSettings(schema, map[string]string{"name": "alice"})
	assert.NoError(t, sm3.Validate())
}

func TestSettingsManager_Validate_Enum(t *testing.T) {
	schema := map[string]any{
		"properties": map[string]any{
			"level": map[string]any{
				"type": "string",
				"enum": []any{"debug", "info", "warn", "error"},
			},
		},
	}
	sm := newTestSettings(schema, map[string]string{"level": "verbose"})
	assert.Error(t, sm.Validate())

	sm2 := newTestSettings(schema, map[string]string{"level": "info"})
	assert.NoError(t, sm2.Validate())
}

func TestSettingsManager_JSONSchema(t *testing.T) {
	schema := map[string]any{
		"properties": map[string]any{
			"apiKey": map[string]any{
				"type":        "string",
				"title":       "API Key",
				"description": "Your API key",
			},
		},
		"required": []any{"apiKey"},
	}
	sm := newTestSettings(schema, nil)
	js := sm.JSONSchema()
	assert.Equal(t, "object", js["type"])

	props, ok := js["properties"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, props, "apiKey")

	req, ok := js["required"].([]string)
	require.True(t, ok)
	assert.Contains(t, req, "apiKey")
}

func TestSettingsManager_MarshalJSON(t *testing.T) {
	sm := newTestSettings(nil, map[string]string{"k": "v"})
	data, err := sm.MarshalJSON()
	require.NoError(t, err)

	var m map[string]string
	require.NoError(t, json.Unmarshal(data, &m))
	assert.Equal(t, "v", m["k"])
}

func TestSettingsManager_NilSchema(t *testing.T) {
	sm := newTestSettings(nil, map[string]string{"x": "1"})
	assert.Equal(t, "1", sm.Get("x"))
	assert.NoError(t, sm.Validate())
	assert.Empty(t, sm.Fields())
}
