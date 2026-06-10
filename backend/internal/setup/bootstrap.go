package setup

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"blotting-consultancy/internal/db"
	"blotting-consultancy/pkg/config"
	"gorm.io/gorm/logger"
)

var (
	ErrEnvAlreadyConfigured = errors.New("environment file already configured")
	ErrEnvConfigNotAllowed  = errors.New("environment configuration not allowed in current mode")
)

// BootstrapInput is the payload for POST /setup/save-env.
type BootstrapInput struct {
	Port     int                 `json:"port"`
	Database config.DatabaseInput `json:"database"`
}

// BootstrapResult is returned after persisting .env.
type BootstrapResult struct {
	Success         bool   `json:"success"`
	RestartRequired bool   `json:"restartRequired"`
	EnvPath         string `json:"envPath"`
}

// TestDatabase verifies that the provided database settings can connect.
func TestDatabase(ctx context.Context, in config.DatabaseInput) error {
	dsn, err := config.BuildDSN(in)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	database, err := db.Init(db.InitOptions{
		DSN:      dsn,
		LogLevel: logger.Silent,
	})
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	sqlDB, err := database.DB.DB()
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}
	defer sqlDB.Close()

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := database.HealthCheck(pingCtx); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	return nil
}

// SaveEnv writes .env for bootstrap installs. Only allowed before a persisted env exists.
func SaveEnv(bootstrapMode bool, outputDir string, in BootstrapInput) (*BootstrapResult, error) {
	if !bootstrapMode {
		return nil, ErrEnvConfigNotAllowed
	}
	if config.EnvFileConfigured() {
		return nil, ErrEnvAlreadyConfigured
	}

	port, err := config.ParsePort(in.Port)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	dsn, err := config.BuildDSN(in.Database)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	if err := TestDatabase(context.Background(), in.Database); err != nil {
		return nil, err
	}

	if outputDir == "" {
		outputDir = "."
	}

	secret, refresh, err := config.GenerateJWTSecrets()
	if err != nil {
		return nil, err
	}

	envPath, err := config.WriteEnvFile(outputDir, config.EnvFileParams{
		Port:             port,
		DBDSN:            dsn,
		JWTSecret:        secret,
		JWTRefreshSecret: refresh,
		Env:              "development",
		UploadDir:        "./uploads",
	})
	if err != nil {
		return nil, err
	}

	return &BootstrapResult{
		Success:         true,
		RestartRequired: true,
		EnvPath:         envPath,
	}, nil
}

// WorkingDirectory returns the directory used for bootstrap file writes.
func WorkingDirectory() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}
