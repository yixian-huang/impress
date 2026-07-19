package setup

import (
	"context"
	"testing"

	"github.com/yixian-huang/inkless/backend/internal/db"
	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/internal/repository"
	"github.com/yixian-huang/inkless/backend/internal/seed"
	"github.com/yixian-huang/inkless/backend/internal/service"
	"github.com/yixian-huang/inkless/backend/pkg/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	database, err := db.Init(db.InitOptions{
		DSN:      ":memory:",
		LogLevel: logger.Silent,
	})
	require.NoError(t, err)
	require.NoError(t, database.DB.AutoMigrate(
		&model.User{},
		&model.ContentDocument{},
		&model.SiteConfig{},
		&model.InstalledTheme{},
		&model.Page{},
		&model.UnifiedPage{},
		&model.PageTemplate{},
		&model.RBACRole{},
		&model.Permission{},
		&model.UserRole{},
	))
	return database.DB
}

func newTestService(t *testing.T, gormDB *gorm.DB, opts ServiceOptions) *Service {
	t.Helper()
	userRepo := repository.NewGormUserRepository(gormDB)
	siteCfgRepo := repository.NewGormSiteConfigRepository(gormDB)
	seederFactory := func(tx *gorm.DB) *seed.Seeder {
		pageRepo := repository.NewGormPageRepository(tx)
		return seed.NewSeeder(
			repository.NewGormUserRepository(tx),
			repository.NewGormContentDocumentRepository(tx),
			repository.NewGormInstalledThemeRepository(tx),
			service.NewThemePageService(pageRepo),
			repository.NewGormUnifiedPageRepository(tx),
			repository.NewGormPageTemplateRepository(tx),
			repository.NewGormSiteConfigRepository(tx),
		)
	}
	return NewService(gormDB, userRepo, siteCfgRepo, seederFactory, nil, opts)
}

func TestService_IsInstalled_FalseOnEmptyDB(t *testing.T) {
	gormDB := setupTestDB(t)
	svc := newTestService(t, gormDB, ServiceOptions{EnvSecretsLoaded: true, DatabaseType: "sqlite"})

	installed, err := svc.IsInstalled(context.Background())
	require.NoError(t, err)
	assert.False(t, installed)
}

func TestService_Complete_CreatesAdminAndMarksInstalled(t *testing.T) {
	gormDB := setupTestDB(t)
	svc := newTestService(t, gormDB, ServiceOptions{EnvSecretsLoaded: true, DatabaseType: "sqlite"})
	userRepo := repository.NewGormUserRepository(gormDB)
	siteCfgRepo := repository.NewGormSiteConfigRepository(gormDB)

	err := svc.Complete(context.Background(), CompleteInput{
		Admin: AdminInput{Username: "owner", Password: "securepass1"},
		Site: SiteInput{
			Name:          map[string]string{"zh": "测试站点", "en": "Test Site"},
			DefaultLocale: "zh",
		},
		SeedMode: "blank",
	})
	require.NoError(t, err)

	installed, err := svc.IsInstalled(context.Background())
	require.NoError(t, err)
	assert.True(t, installed)

	user, err := userRepo.FindByUsername(context.Background(), "owner")
	require.NoError(t, err)
	assert.True(t, user.IsSuperAdmin)
	assert.NoError(t, auth.VerifyPassword(user.PasswordHash, "securepass1"))

	sysCfg, err := siteCfgRepo.FindByKey(context.Background(), model.SiteConfigKeySystem)
	require.NoError(t, err)
	assert.Equal(t, true, sysCfg.PublishedConfig["installed"])
}

func TestService_Complete_RejectsWithoutEnvSecrets(t *testing.T) {
	gormDB := setupTestDB(t)
	svc := newTestService(t, gormDB, ServiceOptions{EnvSecretsLoaded: false, DatabaseType: "sqlite"})

	err := svc.Complete(context.Background(), CompleteInput{
		Admin:    AdminInput{Username: "owner", Password: "securepass1"},
		Site:     SiteInput{Name: map[string]string{"zh": "站点"}},
		SeedMode: "blank",
	})
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestService_Complete_RejectsWhenAlreadyInstalled(t *testing.T) {
	gormDB := setupTestDB(t)
	svc := newTestService(t, gormDB, ServiceOptions{EnvSecretsLoaded: true, DatabaseType: "sqlite"})

	input := CompleteInput{
		Admin:    AdminInput{Username: "owner", Password: "securepass1"},
		Site:     SiteInput{Name: map[string]string{"zh": "站点"}},
		SeedMode: "blank",
	}
	require.NoError(t, svc.Complete(context.Background(), input))

	err := svc.Complete(context.Background(), input)
	assert.ErrorIs(t, err, ErrAlreadyCompleted)
}

func TestService_NeedsEnvConfig(t *testing.T) {
	svc := &Service{opts: ServiceOptions{EnvSecretsLoaded: false}}
	assert.True(t, svc.NeedsEnvConfig())
	svc.opts.EnvSecretsLoaded = true
	assert.False(t, svc.NeedsEnvConfig())
}
