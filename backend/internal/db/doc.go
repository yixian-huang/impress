// Package db provides database connection and migration utilities for GORM.
//
// Example usage:
//
//	import (
//	  "context"
//	  "log"
//	  "time"
//	  "github.com/yixian-huang/inkless/backend/internal/db"
//	  "github.com/yixian-huang/inkless/backend/pkg/config"
//	  "gorm.io/gorm/logger"
//	)
//
//	func main() {
//	  // Load configuration
//	  cfg, err := config.Load()
//	  if err != nil {
//	    log.Fatal(err)
//	  }
//
//	  // Initialize database connection
//	  database, err := db.Init(db.InitOptions{
//	    DSN:         cfg.DBDSN,
//	    MaxOpenConn: 25,
//	    MaxIdleConn: 5,
//	    MaxLifetime: 30 * time.Minute,
//	    LogLevel:    logger.Info,
//	  })
//	  if err != nil {
//	    log.Fatalf("failed to initialize database: %v", err)
//	  }
//	  defer database.Close()
//
//	  // Health check
//	  ctx := context.Background()
//	  if err := database.HealthCheck(ctx); err != nil {
//	    log.Fatalf("database health check failed: %v", err)
//	  }
//
//	  // Run migrations
//	  migrator := db.NewMigrator(database)
//
//	  // Option 1: Auto-migration (recommended for development)
//	  if err := migrator.AutoMigrate(&User{}, &Content{}); err != nil {
//	    log.Fatalf("auto migration failed: %v", err)
//	  }
//
//	  // Option 2: Manual migrations (recommended for production)
//	  migrations := []db.Migration{
//	    {
//	      ID: "001_create_users_table",
//	      Up: func(tx *gorm.DB) error {
//	        return tx.AutoMigrate(&User{})
//	      },
//	      Down: func(tx *gorm.DB) error {
//	        return tx.Migrator().DropTable(&User{})
//	      },
//	    },
//	    {
//	      ID: "002_create_content_table",
//	      Up: func(tx *gorm.DB) error {
//	        return tx.AutoMigrate(&Content{})
//	      },
//	      Down: func(tx *gorm.DB) error {
//	        return tx.Migrator().DropTable(&Content{})
//	      },
//	    },
//	  }
//
//	  if err := migrator.RunMigrations(migrations); err != nil {
//	    log.Fatalf("migrations failed: %v", err)
//	  }
//
//	  // Rollback a specific migration if needed
//	  // if err := migrator.RollbackMigration(migrations, "002_create_content_table"); err != nil {
//	  //   log.Fatalf("rollback failed: %v", err)
//	  // }
//	}
//
// DSN Format:
//   - SQLite: "file:dev.db" or ":memory:" for in-memory database
//   - PostgreSQL: "postgres://user:pass@localhost:5432/dbname?sslmode=disable"
//
// Notes:
//   - SQLite file DSN paths are auto-created when parent directories do not exist.
//   - SQLite connections automatically enable foreign keys and busy timeout pragmas.
package db
