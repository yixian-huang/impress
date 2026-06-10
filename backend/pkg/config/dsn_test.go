package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildDSN_SQLiteDefault(t *testing.T) {
	dsn, err := BuildDSN(DatabaseInput{Type: "sqlite"})
	require.NoError(t, err)
	assert.Contains(t, dsn, "file:./data/impress.db")
}

func TestBuildDSN_Postgres(t *testing.T) {
	dsn, err := BuildDSN(DatabaseInput{
		Type: "postgres",
		Postgres: &PostgresInput{
			Host:   "localhost",
			Port:   5432,
			User:   "impress",
			Password: "secret",
			DBName: "impress",
		},
	})
	require.NoError(t, err)
	assert.Contains(t, dsn, "postgres://")
	assert.Contains(t, dsn, "impress")
}

func TestBuildDSN_PostgresMissingFields(t *testing.T) {
	_, err := BuildDSN(DatabaseInput{
		Type:     "postgres",
		Postgres: &PostgresInput{Host: "localhost"},
	})
	assert.Error(t, err)
}
