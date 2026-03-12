package sdk

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// FieldType represents a JSON-Schema-compatible field type.
type FieldType string

const (
	FieldTypeString  FieldType = "string"
	FieldTypeNumber  FieldType = "number"
	FieldTypeBoolean FieldType = "boolean"
	FieldTypeInteger FieldType = "integer"
)

// SettingsField describes a single settings field derived from the plugin's JSON Schema.
type SettingsField struct {
	Key         string    // JSON property name
	Type        FieldType // "string" | "number" | "integer" | "boolean"
	Title       string    // human-readable label (from JSON Schema "title")
	Description string    // tooltip / help text
	Required    bool      // whether the field is mandatory
	Default     string    // default value serialised as string
	Enum        []string  // allowed values (for "string" fields)
	MinLength   int       // minimum string length (0 = no constraint)
	MaxLength   int       // maximum string length (0 = no constraint)
	Minimum     *float64  // minimum numeric value
	Maximum     *float64  // maximum numeric value
}

// SettingsManager parses and validates plugin settings against a JSON-Schema-derived
// field definition list. It carries the current live values and exposes typed accessors.
type SettingsManager struct {
	fields  []SettingsField
	current map[string]string // current values (string form)
}

// newSettingsManager builds a SettingsManager from a raw SettingsSchema map
// (as declared in plugin.yaml) and the current config values.
func newSettingsManager(schema map[string]any, config map[string]string) *SettingsManager {
	sm := &SettingsManager{
		current: make(map[string]string, len(config)),
	}
	for k, v := range config {
		sm.current[k] = v
	}
	sm.fields = parseSchema(schema)
	return sm
}

// Fields returns the parsed list of settings fields.
func (sm *SettingsManager) Fields() []SettingsField {
	return sm.fields
}

// Get returns the current string value for key, or the field's default if not set.
func (sm *SettingsManager) Get(key string) string {
	if v, ok := sm.current[key]; ok {
		return v
	}
	for _, f := range sm.fields {
		if f.Key == key {
			return f.Default
		}
	}
	return ""
}

// GetBool returns the boolean value for key. Returns false and an error if the
// value cannot be parsed as a boolean.
func (sm *SettingsManager) GetBool(key string) (bool, error) {
	raw := sm.Get(key)
	if raw == "" {
		return false, nil
	}
	return strconv.ParseBool(raw)
}

// GetInt returns the integer value for key. Returns 0 and an error if the
// value cannot be parsed as an integer.
func (sm *SettingsManager) GetInt(key string) (int, error) {
	raw := sm.Get(key)
	if raw == "" {
		return 0, nil
	}
	return strconv.Atoi(raw)
}

// GetFloat returns the float64 value for key. Returns 0 and an error if the
// value cannot be parsed as a float.
func (sm *SettingsManager) GetFloat(key string) (float64, error) {
	raw := sm.Get(key)
	if raw == "" {
		return 0, nil
	}
	return strconv.ParseFloat(raw, 64)
}

// Set updates the in-memory value for key without persistence.
// Use the Store's UpdateSettings to persist changes.
func (sm *SettingsManager) Set(key, value string) {
	sm.current[key] = value
}

// Validate checks that all required fields are present and all values satisfy
// the declared type constraints. Returns a combined error on failure.
func (sm *SettingsManager) Validate() error {
	var errs []string

	for _, f := range sm.fields {
		raw, set := sm.current[f.Key]
		if !set || raw == "" {
			if f.Required {
				errs = append(errs, fmt.Sprintf("field %q is required", f.Key))
			}
			continue
		}

		switch f.Type {
		case FieldTypeBoolean:
			if _, err := strconv.ParseBool(raw); err != nil {
				errs = append(errs, fmt.Sprintf("field %q must be a boolean, got %q", f.Key, raw))
			}

		case FieldTypeInteger:
			n, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				errs = append(errs, fmt.Sprintf("field %q must be an integer, got %q", f.Key, raw))
				break
			}
			if f.Minimum != nil && float64(n) < *f.Minimum {
				errs = append(errs, fmt.Sprintf("field %q must be >= %.0f", f.Key, *f.Minimum))
			}
			if f.Maximum != nil && float64(n) > *f.Maximum {
				errs = append(errs, fmt.Sprintf("field %q must be <= %.0f", f.Key, *f.Maximum))
			}

		case FieldTypeNumber:
			v, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				errs = append(errs, fmt.Sprintf("field %q must be a number, got %q", f.Key, raw))
				break
			}
			if f.Minimum != nil && v < *f.Minimum {
				errs = append(errs, fmt.Sprintf("field %q must be >= %v", f.Key, *f.Minimum))
			}
			if f.Maximum != nil && v > *f.Maximum {
				errs = append(errs, fmt.Sprintf("field %q must be <= %v", f.Key, *f.Maximum))
			}

		case FieldTypeString:
			if f.MinLength > 0 && len(raw) < f.MinLength {
				errs = append(errs, fmt.Sprintf("field %q must be at least %d characters", f.Key, f.MinLength))
			}
			if f.MaxLength > 0 && len(raw) > f.MaxLength {
				errs = append(errs, fmt.Sprintf("field %q must be at most %d characters", f.Key, f.MaxLength))
			}
			if len(f.Enum) > 0 && !containsString(f.Enum, raw) {
				errs = append(errs, fmt.Sprintf("field %q must be one of %v, got %q", f.Key, f.Enum, raw))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("settings validation failed: %s", joinStrings(errs, "; "))
	}
	return nil
}

// JSONSchema returns a JSON Schema document (as a map) derived from the
// parsed field definitions. This is used by the admin UI to render the
// settings panel.
func (sm *SettingsManager) JSONSchema() map[string]any {
	props := make(map[string]any, len(sm.fields))
	required := make([]string, 0)

	for _, f := range sm.fields {
		prop := map[string]any{
			"type": string(f.Type),
		}
		if f.Title != "" {
			prop["title"] = f.Title
		}
		if f.Description != "" {
			prop["description"] = f.Description
		}
		if f.Default != "" {
			prop["default"] = f.Default
		}
		if len(f.Enum) > 0 {
			prop["enum"] = f.Enum
		}
		if f.MinLength > 0 {
			prop["minLength"] = f.MinLength
		}
		if f.MaxLength > 0 {
			prop["maxLength"] = f.MaxLength
		}
		if f.Minimum != nil {
			prop["minimum"] = *f.Minimum
		}
		if f.Maximum != nil {
			prop["maximum"] = *f.Maximum
		}
		props[f.Key] = prop
		if f.Required {
			required = append(required, f.Key)
		}
	}

	schema := map[string]any{
		"type":       "object",
		"properties": props,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

// MarshalJSON returns the JSON encoding of the current settings values.
func (sm *SettingsManager) MarshalJSON() ([]byte, error) {
	return json.Marshal(sm.current)
}

// --- internal helpers ---

// parseSchema converts a raw SettingsSchema map (from plugin.yaml) into a
// typed []SettingsField. Unknown or malformed entries are silently skipped.
func parseSchema(schema map[string]any) []SettingsField {
	if schema == nil {
		return nil
	}

	// Accept both flat map (key -> field def) and JSON-Schema style
	// { "properties": { key: fieldDef }, "required": [...] }.
	requiredSet := map[string]bool{}

	props, ok := schema["properties"]
	if !ok {
		// treat schema itself as the properties map
		props = schema
	}

	propsMap, ok := props.(map[string]any)
	if !ok {
		return nil
	}

	// Parse required array if present
	if req, ok := schema["required"]; ok {
		switch v := req.(type) {
		case []string:
			for _, k := range v {
				requiredSet[k] = true
			}
		case []any:
			for _, item := range v {
				if s, ok := item.(string); ok {
					requiredSet[s] = true
				}
			}
		}
	}

	fields := make([]SettingsField, 0, len(propsMap))
	for key, raw := range propsMap {
		def, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		f := SettingsField{
			Key:      key,
			Required: requiredSet[key],
		}

		if t, ok := def["type"].(string); ok {
			f.Type = FieldType(t)
		}
		if s, ok := def["title"].(string); ok {
			f.Title = s
		}
		if s, ok := def["description"].(string); ok {
			f.Description = s
		}
		if s, ok := def["default"].(string); ok {
			f.Default = s
		}
		if n, ok := def["minLength"].(int); ok {
			f.MinLength = n
		}
		if n, ok := def["maxLength"].(int); ok {
			f.MaxLength = n
		}
		if v, ok := parseNumber(def["minimum"]); ok {
			f.Minimum = &v
		}
		if v, ok := parseNumber(def["maximum"]); ok {
			f.Maximum = &v
		}
		if e, ok := def["enum"].([]any); ok {
			for _, item := range e {
				if s, ok := item.(string); ok {
					f.Enum = append(f.Enum, s)
				}
			}
		}

		fields = append(fields, f)
	}
	return fields
}

// parseNumber attempts to extract a float64 from various numeric types.
func parseNumber(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	}
	return 0, false
}

// containsString reports whether s is in the slice.
func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// joinStrings joins a slice with a separator (avoids importing strings in tests).
func joinStrings(parts []string, sep string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += sep
		}
		result += p
	}
	return result
}
