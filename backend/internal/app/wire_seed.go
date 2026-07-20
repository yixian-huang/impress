package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"gorm.io/gorm"

	"github.com/yixian-huang/inkless/backend/internal/db"
	"github.com/yixian-huang/inkless/backend/internal/repository"
	"github.com/yixian-huang/inkless/backend/internal/seed"
	"github.com/yixian-huang/inkless/backend/internal/service"
	install "github.com/yixian-huang/inkless/backend/internal/setup"
	"github.com/yixian-huang/inkless/backend/pkg/config"
	appLogger "github.com/yixian-huang/inkless/backend/pkg/logger"
)

// runSeedAndSetup builds the setup service, runs startup seed, and returns setupSvc.
func runSeedAndSetup(
	database *db.DB,
	r *repos,
	cfg *config.Config,
	loadResult *config.LoadResult,
	log *appLogger.Logger,
) (*install.Service, error) {
	themePageService := service.NewThemePageService(r.page)
	seeder := seed.NewSeeder(
		r.user,
		r.contentDoc,
		r.installedTheme,
		themePageService,
		r.unifiedPage,
		r.pageTemplate,
		r.siteConfig,
	)
	seedRBAC := func(ctx context.Context, roleRepo repository.RoleRepository) error {
		return seed.SeedRBAC(ctx, roleRepo)
	}
	seederFactory := func(tx *gorm.DB) *seed.Seeder {
		txPageRepo := repository.NewGormPageRepository(tx)
		return seed.NewSeeder(
			repository.NewGormUserRepository(tx),
			repository.NewGormContentDocumentRepository(tx),
			repository.NewGormInstalledThemeRepository(tx),
			service.NewThemePageService(txPageRepo),
			repository.NewGormUnifiedPageRepository(tx),
			repository.NewGormPageTemplateRepository(tx),
			repository.NewGormSiteConfigRepository(tx),
		)
	}
	setupSvc := install.NewService(
		database.DB,
		r.user,
		r.siteConfig,
		seederFactory,
		seedRBAC,
		install.ServiceOptions{
			BootstrapMode:    loadResult.BootstrapMode,
			EnvSecretsLoaded: loadResult.EnvSecretsLoaded,
			DatabaseType:     config.DatabaseTypeFromDSN(cfg.DBDSN),
			EnvFilePath:      config.DefaultEnvFilePath(),
			ServerPort:       cfg.Port,
		},
	)

	seedCtx, seedCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer seedCancel()

	installed, err := setupSvc.IsInstalled(seedCtx)
	if err != nil {
		return nil, fmt.Errorf("check install status: %w", err)
	}

	seedMode := os.Getenv("SEED_MODE")
	if err := seed.RunStartupSeed(seedCtx, installed, seedMode, seeder, func(ctx context.Context) error {
		return seedRBAC(ctx, r.role)
	}); err != nil {
		return nil, fmt.Errorf("startup seed: %w", err)
	}
	log.Info("Seed data initialized", "installed", installed, "seedMode", seedMode)
	return setupSvc, nil
}
