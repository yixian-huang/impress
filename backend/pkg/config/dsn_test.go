package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildDSN_SQLiteDefault(t *testing.T) {
	dsn, err := BuildDSN(DatabaseInput{Type: "sqlite"})
	require.NoError(t, err)
	assert.Contains(t, dsn, "file:./data/inkless.db")
}

func TestBuildDSN_Postgres(t *testing.T) {
	dsn, err := BuildDSN(DatabaseInput{
		Type: "postgres",
		Postgres: &PostgresInput{
			Host:     "localhost",
			Port:     5432,
			User:     "inkless",
			Password: "secret",
			DBName:   "inkless",
		},
	})
	require.NoError(t, err)
	assert.Contains(t, dsn, "postgres://")
	assert.Contains(t, dsn, "inkless")
}

func TestBuildDSN_LegacySQLitePathRemainsExplicitlyUsable(t *testing.T) {
	dsn, err := BuildDSN(DatabaseInput{Type: "sqlite", SQLitePath: "./data/impress.db"})
	require.NoError(t, err)
	assert.Equal(t, "file:./data/impress.db?cache=shared&mode=rwc", dsn)
}

func TestBuildDSN_PostgresMissingFields(t *testing.T) {
	_, err := BuildDSN(DatabaseInput{
		Type:     "postgres",
		Postgres: &PostgresInput{Host: "localhost"},
	})
	assert.Error(t, err)
}
