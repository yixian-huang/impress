package db

import (
	"fmt"

	"gorm.io/gorm"
)

// Migrator handles database schema migrations
type Migrator struct {
	db *DB
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *DB) *Migrator {
	return &Migrator{db: db}
}

// AutoMigrate runs GORM's auto-migration for the provided models
// This creates tables and updates schema based on struct definitions
func (m *Migrator) AutoMigrate(models ...interface{}) error {
	if err := m.db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("auto migration failed: %w", err)
	}
	return nil
}

// DropTable drops the tables for the provided models
// Use with caution - this is destructive
func (m *Migrator) DropTable(models ...interface{}) error {
	if err := m.db.Migrator().DropTable(models...); err != nil {
		return fmt.Errorf("drop table failed: %w", err)
	}
	return nil
}

// HasTable checks if a table exists for the given model
func (m *Migrator) HasTable(model interface{}) bool {
	return m.db.Migrator().HasTable(model)
}

// Migration represents a single migration operation
type Migration struct {
	ID   string
	Up   func(*gorm.DB) error
	Down func(*gorm.DB) error
}

// MigrationHistory tracks applied migrations
type MigrationHistory struct {
	ID        uint   `gorm:"primaryKey"`
	Migration string `gorm:"uniqueIndex;not null"`
	AppliedAt int64  `gorm:"autoCreateTime"`
}

// RunMigrations applies a list of manual migrations in order
// Tracks which migrations have been applied in migration_history table
func (m *Migrator) RunMigrations(migrations []Migration) error {
	// Ensure migration_history table exists
	if err := m.db.AutoMigrate(&MigrationHistory{}); err != nil {
		return fmt.Errorf("failed to create migration_history table: %w", err)
	}

	for _, migration := range migrations {
		// Check if migration already applied
		var history MigrationHistory
		result := m.db.Where("migration = ?", migration.ID).First(&history)

		if result.Error == nil {
			// Migration already applied, skip
			continue
		}

		if result.Error != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to check migration history for %s: %w", migration.ID, result.Error)
		}

		// Apply migration
		if err := migration.Up(m.db.DB); err != nil {
			return fmt.Errorf("migration %s failed: %w", migration.ID, err)
		}

		// Record migration
		history = MigrationHistory{Migration: migration.ID}
		if err := m.db.Create(&history).Error; err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migration.ID, err)
		}
	}

	return nil
}

// RollbackMigration rolls back a single migration by ID
func (m *Migrator) RollbackMigration(migrations []Migration, migrationID string) error {
	// Find the migration
	var target *Migration
	for i := range migrations {
		if migrations[i].ID == migrationID {
			target = &migrations[i]
			break
		}
	}

	if target == nil {
		return fmt.Errorf("migration %s not found", migrationID)
	}

	// Check if migration was applied
	var history MigrationHistory
	result := m.db.Where("migration = ?", migrationID).First(&history)
	if result.Error == gorm.ErrRecordNotFound {
		return fmt.Errorf("migration %s was not applied", migrationID)
	}
	if result.Error != nil {
		return fmt.Errorf("failed to check migration history: %w", result.Error)
	}

	// Run rollback
	if err := target.Down(m.db.DB); err != nil {
		return fmt.Errorf("rollback of migration %s failed: %w", migrationID, err)
	}

	// Remove migration record
	if err := m.db.Delete(&history).Error; err != nil {
		return fmt.Errorf("failed to remove migration record for %s: %w", migrationID, err)
	}

	return nil
}
