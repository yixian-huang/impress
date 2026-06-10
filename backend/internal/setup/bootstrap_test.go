package setup

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"blotting-consultancy/pkg/config"

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
		Database: config.DatabaseInput{
			Type:       "sqlite",
			SQLitePath: filepath.Join(dir, "data", "impress.db"),
		},
	})
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.True(t, result.RestartRequired)
	assert.FileExists(t, result.EnvPath)
}

func TestSaveEnv_RejectsWhenNotBootstrap(t *testing.T) {
	_, err := SaveEnv(false, ".", BootstrapInput{
		Database: config.DatabaseInput{Type: "sqlite", SQLitePath: ":memory:"},
	})
	assert.ErrorIs(t, err, ErrEnvConfigNotAllowed)
}
