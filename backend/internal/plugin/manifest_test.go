package plugin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validFullManifest = `
id: s3-storage
name: S3 Storage
nameZh: S3 对象存储
version: 1.0.0
description: Store media files in Amazon S3
author: Inkless Team
license: MIT
homepage: https://example.com
minAppVersion: 1.0.0
dependencies:
  - pluginId: base-plugin
    minVersion: 0.1.0
permissions:
  - network:outbound
  - filesystem:read
providers:
  - type: storage
    name: s3-storage
routes:
  - method: GET
    path: /buckets
  - method: POST
    path: /test-connection
frontendEntry: frontend/dist/index.js
settingsSchema:
  type: object
  required:
    - endpoint
    - bucket
  properties:
    endpoint:
      type: string
      title: S3 Endpoint
    bucket:
      type: string
      title: Bucket Name
`

const validMinimalManifest = `
id: my-plugin
name: My Plugin
version: 0.1.0
`

func TestParseManifest_FullFields(t *testing.T) {
	meta, err := ParseManifest([]byte(validFullManifest))
	require.NoError(t, err)

	assert.Equal(t, "s3-storage", meta.ID)
	assert.Equal(t, "S3 Storage", meta.Name)
	assert.Equal(t, "S3 对象存储", meta.NameZh)
	assert.Equal(t, "1.0.0", meta.Version)
	assert.Equal(t, "Store media files in Amazon S3", meta.Description)
	assert.Equal(t, "Inkless Team", meta.Author)
	assert.Equal(t, "MIT", meta.License)
	assert.Equal(t, "https://example.com", meta.Homepage)
	assert.Equal(t, "1.0.0", meta.MinAppVersion)

	require.Len(t, meta.Dependencies, 1)
	assert.Equal(t, "base-plugin", meta.Dependencies[0].PluginID)
	assert.Equal(t, "0.1.0", meta.Dependencies[0].MinVersion)

	require.Len(t, meta.Permissions, 2)
	assert.Equal(t, PermNetworkOutbound, meta.Permissions[0])
	assert.Equal(t, PermFileSystemRead, meta.Permissions[1])

	require.Len(t, meta.Providers, 1)
	assert.Equal(t, "storage", meta.Providers[0].Type)
	assert.Equal(t, "s3-storage", meta.Providers[0].Name)

	require.Len(t, meta.Routes, 2)
	assert.Equal(t, "GET", meta.Routes[0].Method)
	assert.Equal(t, "/buckets", meta.Routes[0].Path)

	assert.Equal(t, "frontend/dist/index.js", meta.FrontendEntry)

	require.NotNil(t, meta.SettingsSchema)
	assert.Equal(t, "object", meta.SettingsSchema["type"])
}

func TestParseManifest_MinimalFields(t *testing.T) {
	meta, err := ParseManifest([]byte(validMinimalManifest))
	require.NoError(t, err)

	assert.Equal(t, "my-plugin", meta.ID)
	assert.Equal(t, "My Plugin", meta.Name)
	assert.Equal(t, "0.1.0", meta.Version)
	assert.Empty(t, meta.Dependencies)
	assert.Empty(t, meta.Permissions)
	assert.Empty(t, meta.Providers)
	assert.Empty(t, meta.Routes)
}

func TestParseManifest_InvalidYAML(t *testing.T) {
	_, err := ParseManifest([]byte(`{{{invalid yaml`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse manifest YAML")
}

func TestParseAndValidate_FullManifest(t *testing.T) {
	meta, err := ParseManifest([]byte(validFullManifest))
	require.NoError(t, err)
	require.NoError(t, meta.Validate())
}

func TestParseAndValidate_MinimalManifest(t *testing.T) {
	meta, err := ParseManifest([]byte(validMinimalManifest))
	require.NoError(t, err)
	require.NoError(t, meta.Validate())
}

func TestValidateManifest_InvalidID(t *testing.T) {
	meta, err := ParseManifest([]byte(`
id: AB
name: Test
version: 1.0.0
`))
	require.NoError(t, err)
	err = meta.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must match pattern")
}

func TestValidateManifest_InvalidSemver(t *testing.T) {
	meta, err := ParseManifest([]byte(`
id: my-plugin
name: Test
version: not-a-version
`))
	require.NoError(t, err)
	err = meta.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not valid semver")
}

func TestValidateManifest_UnknownPermission(t *testing.T) {
	meta, err := ParseManifest([]byte(`
id: my-plugin
name: Test
version: 1.0.0
permissions:
  - magic:wand
`))
	require.NoError(t, err)
	err = meta.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown permission")
}

func TestValidateManifest_UnknownProviderType(t *testing.T) {
	meta, err := ParseManifest([]byte(`
id: my-plugin
name: Test
version: 1.0.0
providers:
  - type: teleporter
    name: beam-me-up
`))
	require.NoError(t, err)
	err = meta.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown provider type")
}

func TestValidateManifest_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name   string
		yaml   string
		errMsg string
	}{
		{
			name:   "missing id",
			yaml:   "name: Test\nversion: 1.0.0",
			errMsg: "id is required",
		},
		{
			name:   "missing name",
			yaml:   "id: my-plugin\nversion: 1.0.0",
			errMsg: "name is required",
		},
		{
			name:   "missing version",
			yaml:   "id: my-plugin\nname: Test",
			errMsg: "version is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta, err := ParseManifest([]byte(tt.yaml))
			require.NoError(t, err)
			err = meta.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestValidateManifest_MultipleErrors(t *testing.T) {
	meta, err := ParseManifest([]byte(`
name: Test
version: not-valid
permissions:
  - bad:perm
`))
	require.NoError(t, err)
	err = meta.Validate()
	require.Error(t, err)
	// Should contain multiple error messages
	assert.Contains(t, err.Error(), "id is required")
	assert.Contains(t, err.Error(), "not valid semver")
	assert.Contains(t, err.Error(), "unknown permission")
}

func TestLoadManifest_FromDirectory(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, ManifestFileName), []byte(validMinimalManifest), 0644)
	require.NoError(t, err)

	meta, err := LoadManifest(dir)
	require.NoError(t, err)
	assert.Equal(t, "my-plugin", meta.ID)
}

func TestLoadManifest_MissingFile(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadManifest(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read manifest")
}

func TestLoadAndValidateManifest_Valid(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, ManifestFileName), []byte(validMinimalManifest), 0644)
	require.NoError(t, err)

	meta, err := LoadAndValidateManifest(dir)
	require.NoError(t, err)
	assert.Equal(t, "my-plugin", meta.ID)
}

func TestLoadAndValidateManifest_InvalidContent(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, ManifestFileName), []byte("id: AB\nname: Test\nversion: 1.0.0"), 0644)
	require.NoError(t, err)

	_, err = LoadAndValidateManifest(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must match pattern")
}

func TestValidateManifest_Wrapper(t *testing.T) {
	meta := &PluginMeta{ID: "my-plugin", Name: "Test", Version: "1.0.0"}
	assert.NoError(t, ValidateManifest(meta))

	bad := &PluginMeta{}
	assert.Error(t, ValidateManifest(bad))
}

func TestValidateSupportedSettingsRejectsSecrets(t *testing.T) {
	tests := []map[string]any{
		{
			"properties": map[string]any{
				"token": map[string]any{"type": "string", "secret": true},
			},
		},
		{
			"properties": map[string]any{
				"password": map[string]any{"type": "string", "format": "password"},
			},
		},
	}

	for _, schema := range tests {
		meta := &PluginMeta{ID: "secret-plugin", SettingsSchema: schema}
		err := ValidateSupportedSettings(meta)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "secret settings")
	}
}

func TestValidateSupportedSettingsAllowsNonSecretSchema(t *testing.T) {
	meta := &PluginMeta{
		ID: "plain-plugin",
		SettingsSchema: map[string]any{
			"outputFile": map[string]any{"type": "string"},
		},
	}
	require.NoError(t, ValidateSupportedSettings(meta))
}

func TestValidateExternalRuntimeContractRejectsUnwiredCapabilities(t *testing.T) {
	tests := []PluginMeta{
		{
			ID:           "storage-plugin",
			Providers:    []ProviderDecl{{Type: "storage", Name: "storage"}},
			Dependencies: nil,
		},
		{
			ID:        "route-plugin",
			Providers: []ProviderDecl{{Type: "notifier", Name: "notifier"}},
			Routes:    []RouteDecl{{Method: "GET", Path: "/status"}},
		},
		{
			ID:        "named-plugin",
			Providers: []ProviderDecl{{Type: "notifier", Name: "custom-notifier"}},
		},
	}

	for _, meta := range tests {
		require.Error(t, ValidateExternalRuntimeContract(&meta))
	}
}

func TestValidateExternalRuntimeContractAllowsCanonicalNotifier(t *testing.T) {
	meta := &PluginMeta{
		ID:        "notifier-plugin",
		Providers: []ProviderDecl{{Type: "notifier", Name: "notifier"}},
	}
	require.NoError(t, ValidateExternalRuntimeContract(meta))
}
