package setup

import (
	"context"
	"testing"

	"blotting-consultancy/internal/db"
	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/repository"
	"blotting-consultancy/internal/seed"
	"blotting-consultancy/internal/service"
	"blotting-consultancy/pkg/auth"

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
	))
	return database.DB
}

func newTestService(t *testing.T, gormDB *gorm.DB) *Service {
	t.Helper()
	userRepo := repository.NewGormUserRepository(gormDB)
	contentRepo := repository.NewGormContentDocumentRepository(gormDB)
	siteCfgRepo := repository.NewGormSiteConfigRepository(gormDB)
	installedThemeRepo := repository.NewGormInstalledThemeRepository(gormDB)
	pageRepo := repository.NewGormPageRepository(gormDB)
	themePageSvc := service.NewThemePageService(pageRepo)
	unifiedPageRepo := repository.NewGormUnifiedPageRepository(gormDB)
	templateRepo := repository.NewGormPageTemplateRepository(gormDB)
	seeder := seed.NewSeeder(userRepo, contentRepo, installedThemeRepo, themePageSvc, unifiedPageRepo, templateRepo, siteCfgRepo)
	return NewService(userRepo, siteCfgRepo, contentRepo, seeder, func(ctx context.Context) error { return nil }, ServiceOptions{
		DatabaseType: "sqlite",
	})
}

func TestService_IsInstalled_FalseOnEmptyDB(t *testing.T) {
	gormDB := setupTestDB(t)
	svc := newTestService(t, gormDB)

	installed, err := svc.IsInstalled(context.Background())
	require.NoError(t, err)
	assert.False(t, installed)
}

func TestService_Complete_CreatesAdminAndMarksInstalled(t *testing.T) {
	gormDB := setupTestDB(t)
	svc := newTestService(t, gormDB)
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

func TestService_Complete_RejectsWhenAlreadyInstalled(t *testing.T) {
	gormDB := setupTestDB(t)
	svc := newTestService(t, gormDB)

	input := CompleteInput{
		Admin:    AdminInput{Username: "owner", Password: "securepass1"},
		Site:     SiteInput{Name: map[string]string{"zh": "站点"}},
		SeedMode: "blank",
	}
	require.NoError(t, svc.Complete(context.Background(), input))

	err := svc.Complete(context.Background(), input)
	assert.ErrorIs(t, err, ErrAlreadyCompleted)
}

func TestService_Complete_ValidatesPassword(t *testing.T) {
	gormDB := setupTestDB(t)
	svc := newTestService(t, gormDB)

	err := svc.Complete(context.Background(), CompleteInput{
		Admin:    AdminInput{Username: "owner", Password: "short"},
		Site:     SiteInput{Name: map[string]string{"zh": "站点"}},
		SeedMode: "blank",
	})
	assert.ErrorIs(t, err, ErrInvalidInput)
}
