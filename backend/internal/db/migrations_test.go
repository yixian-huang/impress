package db_test

import (
	"database/sql"
	"testing"

	"github.com/pressly/goose/v3"
	"blotting-consultancy/internal/db/migrations"
	_ "github.com/mattn/go-sqlite3"
)

func TestGooseMigrationsEmbed(t *testing.T) {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer sqlDB.Close()

	goose.SetBaseFS(migrations.EmbedMigrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("set dialect: %v", err)
	}

	// Should run baseline without error
	if err := goose.Up(sqlDB, "."); err != nil {
		t.Fatalf("goose up: %v", err)
	}

	// Verify version is 1
	ver, err := goose.GetDBVersion(sqlDB)
	if err != nil {
		t.Fatalf("get version: %v", err)
	}
	if ver != 1 {
		t.Errorf("expected version 1, got %d", ver)
	}
}
