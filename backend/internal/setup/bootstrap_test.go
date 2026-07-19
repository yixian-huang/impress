package setup

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/yixian-huang/inkless/backend/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestDatabase_SQLiteMemory(t *testing.T) {
	err := TestDatabase(context.Background(), config.DatabaseInput{
		Type:       "sqlite",
		SQLitePath: ":memory:",
	})
	require.NoError(t, err)
}

func TestSaveEnv_BootstrapWritesFile(t *testing.T) {
	dir := t.TempDir()
	oldWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldWD)

	result, err := SaveEnv(true, dir, BootstrapInput{
		Port: 8088,
		Env:  "production",
		Database: config.DatabaseInput{
			Type:       "sqlite",
			SQLitePath: filepath.Join(dir, "data", "inkless.db"),
		},
	})
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.True(t, result.RestartRequired)
	assert.FileExists(t, result.EnvPath)

	data, err := os.ReadFile(result.EnvPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "ENV=production")
}

func TestSaveEnv_RejectsWhenNotAllowed(t *testing.T) {
	_, err := SaveEnv(false, ".", BootstrapInput{
		Database: config.DatabaseInput{Type: "sqlite", SQLitePath: ":memory:"},
	})
	assert.ErrorIs(t, err, ErrEnvConfigNotAllowed)
}

func TestValidatePostgresHost_RejectsPublicIP(t *testing.T) {
	err := validatePostgresHost("8.8.8.8")
	assert.Error(t, err)
}
