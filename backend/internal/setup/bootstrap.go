package setup

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/yixian-huang/inkless/backend/internal/db"
	"github.com/yixian-huang/inkless/backend/pkg/config"
	"gorm.io/gorm/logger"
)

var (
	ErrEnvAlreadyConfigured = errors.New("environment file already configured")
	ErrEnvConfigNotAllowed  = errors.New("environment configuration not allowed in current mode")
)

// BootstrapInput is the payload for POST /setup/save-env.
type BootstrapInput struct {
	Port     int                  `json:"port"`
	Env      string               `json:"env"`
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
	if err := validateDatabaseInput(in); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

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

// SaveEnv writes .env for bootstrap installs.
func SaveEnv(allow bool, outputDir string, in BootstrapInput) (*BootstrapResult, error) {
	if !allow {
		return nil, ErrEnvConfigNotAllowed
	}
	if config.EnvFileConfigured() {
		return nil, ErrEnvAlreadyConfigured
	}

	port, err := config.ParsePort(in.Port)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	runtimeEnv, err := normalizeRuntimeEnv(in.Env)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	if err := validateDatabaseInput(in.Database); err != nil {
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
		Env:              runtimeEnv,
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

func validateDatabaseInput(in config.DatabaseInput) error {
	dbType := strings.ToLower(strings.TrimSpace(in.Type))
	if dbType == "postgres" || dbType == "postgresql" {
		if in.Postgres == nil {
			return errors.New("postgres connection details are required")
		}
		host := strings.TrimSpace(in.Postgres.Host)
		if host == "" {
			return errors.New("postgres host is required")
		}
		if err := validatePostgresHost(host); err != nil {
			return err
		}
	}
	return nil
}

func validatePostgresHost(host string) error {
	if strings.EqualFold(host, "localhost") || host == "127.0.0.1" || host == "::1" {
		return nil
	}
	ip := net.ParseIP(host)
	if ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() {
			return nil
		}
		return fmt.Errorf("postgres host must be localhost or a private-network address during setup")
	}
	return nil
}

func normalizeRuntimeEnv(raw string) (string, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return "development", nil
	}
	if raw != "development" && raw != "production" {
		return "", errors.New("env must be development or production")
	}
	return raw, nil
}

// WorkingDirectory returns the directory used for bootstrap file writes.
func WorkingDirectory() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}
