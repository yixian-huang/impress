package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadWithBootstrap_GeneratesEphemeralJWT(t *testing.T) {
	cleanupBootstrapEnv()
	os.Setenv("SETUP_BOOTSTRAP", "true")
	defer cleanupBootstrapEnv()

	result, err := LoadWithBootstrap()
	require.NoError(t, err)
	assert.True(t, result.BootstrapMode)
	assert.NotEmpty(t, result.Config.JWTSecret)
	assert.NotEmpty(t, result.Config.JWTRefreshSecret)
}

func TestWriteEnvFile_CreatesEnvAndDirs(t *testing.T) {
	dir := t.TempDir()
	path, err := WriteEnvFile(dir, EnvFileParams{
		Port:             9090,
		DBDSN:            "file:" + filepath.Join(dir, "data", "test.db") + "?cache=shared&mode=rwc",
		JWTSecret:        "secret",
		JWTRefreshSecret: "refresh",
	})
	require.NoError(t, err)
	assert.FileExists(t, path)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "PORT=9090")
	assert.Contains(t, content, "JWT_SECRET=secret")
}

func cleanupBootstrapEnv() {
	os.Unsetenv("SETUP_BOOTSTRAP")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("JWT_REFRESH_SECRET")
	os.Unsetenv("PORT")
	os.Unsetenv("DB_DSN")
	os.Unsetenv("ENV")
}
