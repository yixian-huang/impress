package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ManifestFileName is the expected name of the plugin manifest file.
const ManifestFileName = "plugin.yaml"

// LoadManifest reads and parses a plugin.yaml file from the given directory.
// It returns the parsed PluginMeta or an error if the file cannot be read or parsed.
func LoadManifest(dir string) (*PluginMeta, error) {
	path := filepath.Join(dir, ManifestFileName)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest %s: %w", path, err)
	}

	return ParseManifest(data)
}

// ParseManifest parses raw YAML bytes into a PluginMeta struct.
func ParseManifest(data []byte) (*PluginMeta, error) {
	var meta PluginMeta
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse manifest YAML: %w", err)
	}
	return &meta, nil
}

// LoadAndValidateManifest reads, parses, and validates a plugin.yaml from the given directory.
func LoadAndValidateManifest(dir string) (*PluginMeta, error) {
	meta, err := LoadManifest(dir)
	if err != nil {
		return nil, err
	}

	if err := meta.Validate(); err != nil {
		return nil, err
	}

	return meta, nil
}

// ValidateManifest validates a PluginMeta struct.
// This is a convenience wrapper around PluginMeta.Validate().
func ValidateManifest(meta *PluginMeta) error {
	return meta.Validate()
}

// ValidateSupportedSettings rejects secret-bearing schemas until encrypted
// persistence and response masking are available for external plugins.
func ValidateSupportedSettings(meta *PluginMeta) error {
	if schemaContainsSecret(meta.SettingsSchema) {
		return fmt.Errorf(
			"plugin %q declares secret settings, which are not supported by the external plugin beta",
			meta.ID,
		)
	}
	return nil
}

// ValidateExternalRuntimeContract limits the beta to capabilities that are
// fully wired through the external process lifecycle.
func ValidateExternalRuntimeContract(meta *PluginMeta) error {
	if len(meta.Dependencies) > 0 {
		return fmt.Errorf("plugin %q declares dependencies, which are not supported by the external plugin beta", meta.ID)
	}
	if len(meta.Routes) > 0 {
		return fmt.Errorf("plugin %q declares routes, which are not supported by the external plugin beta", meta.ID)
	}
	if meta.FrontendEntry != "" {
		return fmt.Errorf("plugin %q declares a frontend entry, which is not supported by the external plugin beta", meta.ID)
	}
	if len(meta.Providers) == 0 {
		return fmt.Errorf("plugin %q must declare at least one supported provider", meta.ID)
	}

	seen := make(map[string]struct{}, len(meta.Providers))
	for _, declaration := range meta.Providers {
		if declaration.Type == "storage" {
			return fmt.Errorf(
				"plugin %q declares storage provider %q; external storage providers are reserved until runtime ownership is coordinated",
				meta.ID,
				declaration.Name,
			)
		}
		if declaration.Name != declaration.Type {
			return fmt.Errorf(
				"plugin %q provider %q must use the canonical registration name %q",
				meta.ID,
				declaration.Name,
				declaration.Type,
			)
		}
		if _, duplicate := seen[declaration.Name]; duplicate {
			return fmt.Errorf("plugin %q declares duplicate provider %q", meta.ID, declaration.Name)
		}
		seen[declaration.Name] = struct{}{}
	}
	return nil
}

func schemaContainsSecret(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			normalized := strings.ToLower(strings.TrimSpace(key))
			if normalized == "secret" {
				if enabled, ok := child.(bool); ok && enabled {
					return true
				}
			}
			if normalized == "format" || normalized == "type" {
				if text, ok := child.(string); ok {
					switch strings.ToLower(strings.TrimSpace(text)) {
					case "password", "secret":
						return true
					}
				}
			}
			if schemaContainsSecret(child) {
				return true
			}
		}
	case []any:
		for _, child := range typed {
			if schemaContainsSecret(child) {
				return true
			}
		}
	}
	return false
}
