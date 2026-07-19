package db_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/yixian-huang/inkless/backend/internal/db/migrations"
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

	// Apply only the baseline migration (later migrations depend on tables created by AutoMigrate)
	if err := goose.UpTo(sqlDB, ".", 1); err != nil {
		t.Fatalf("goose up to 1: %v", err)
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
