package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/yixian-huang/inkless/backend/internal/cache"
	"github.com/yixian-huang/inkless/backend/internal/db"
	"github.com/yixian-huang/inkless/backend/internal/eventbus"
	aiHandler "github.com/yixian-huang/inkless/backend/internal/handler/ai"
	analyticsHandler "github.com/yixian-huang/inkless/backend/internal/handler/analytics"
	dashboardHandler "github.com/yixian-huang/inkless/backend/internal/handler/dashboard"
	articleHandler "github.com/yixian-huang/inkless/backend/internal/handler/article"
	auditlogHandler "github.com/yixian-huang/inkless/backend/internal/handler/auditlog"
	authHandler "github.com/yixian-huang/inkless/backend/internal/handler/auth"
	bootstrapHandler "github.com/yixian-huang/inkless/backend/internal/handler/bootstrap"
	categoryHandler "github.com/yixian-huang/inkless/backend/internal/handler/category"
	chunkedUploadHandler "github.com/yixian-huang/inkless/backend/internal/handler/chunked_upload"
	emailSettingsHandler "github.com/yixian-huang/inkless/backend/internal/handler/email_settings"
	featuresHandler "github.com/yixian-huang/inkless/backend/internal/handler/features"
	feedHandler "github.com/yixian-huang/inkless/backend/internal/handler/feed"
	globalConfigHandler "github.com/yixian-huang/inkless/backend/internal/handler/global_config"
	installedThemeHandler "github.com/yixian-huang/inkless/backend/internal/handler/installed_theme"
	marketplaceHandler "github.com/yixian-huang/inkless/backend/internal/handler/marketplace"
	mediaHandler "github.com/yixian-huang/inkless/backend/internal/handler/media"
	mediaFolderHandler "github.com/yixian-huang/inkless/backend/internal/handler/media_folder"
	menuHandler "github.com/yixian-huang/inkless/backend/internal/handler/menu"
	migrationHandler "github.com/yixian-huang/inkless/backend/internal/handler/migration"
	pageTemplateHandler "github.com/yixian-huang/inkless/backend/internal/handler/page_template"
	pluginHandler "github.com/yixian-huang/inkless/backend/internal/handler/plugin"
	publicHandler "github.com/yixian-huang/inkless/backend/internal/handler/public"
	roleHandler "github.com/yixian-huang/inkless/backend/internal/handler/role"
	schedulerHandler "github.com/yixian-huang/inkless/backend/internal/handler/scheduler"
	searchhandler "github.com/yixian-huang/inkless/backend/internal/handler/search"
	seoHandler "github.com/yixian-huang/inkless/backend/internal/handler/seo"
	setupHandler "github.com/yixian-huang/inkless/backend/internal/handler/setup"
	sitemapHandler "github.com/yixian-huang/inkless/backend/internal/handler/sitemap"
	storageHandler "github.com/yixian-huang/inkless/backend/internal/handler/storage"
	systemHandler "github.com/yixian-huang/inkless/backend/internal/handler/system"
	tagHandler "github.com/yixian-huang/inkless/backend/internal/handler/tag"
	themeHandler "github.com/yixian-huang/inkless/backend/internal/handler/theme"
	themeExportHandler "github.com/yixian-huang/inkless/backend/internal/handler/theme_export"
	translationHandler "github.com/yixian-huang/inkless/backend/internal/handler/translation"
	unifiedPageHandler "github.com/yixian-huang/inkless/backend/internal/handler/unified_page"
	userHandler "github.com/yixian-huang/inkless/backend/internal/handler/user"
	wizardHandler "github.com/yixian-huang/inkless/backend/internal/handler/wizard"
	"github.com/yixian-huang/inkless/backend/internal/middleware"
	"github.com/yixian-huang/inkless/backend/internal/migration"
	"github.com/yixian-huang/inkless/backend/internal/module"
	backupMod "github.com/yixian-huang/inkless/backend/internal/modules/backup"
	commentMod "github.com/yixian-huang/inkless/backend/internal/modules/comment"
	formSubmissionMod "github.com/yixian-huang/inkless/backend/internal/modules/form_submission"
	qa "github.com/yixian-huang/inkless/backend/internal/modules/qa"
	pluginruntime "github.com/yixian-huang/inkless/backend/internal/plugin"
	"github.com/yixian-huang/inkless/backend/internal/provider"
	"github.com/yixian-huang/inkless/backend/internal/repository"
	"github.com/yixian-huang/inkless/backend/internal/seed"
	"github.com/yixian-huang/inkless/backend/internal/service"
	install "github.com/yixian-huang/inkless/backend/internal/setup"
	"github.com/yixian-huang/inkless/backend/pkg/apierror"
	"github.com/yixian-huang/inkless/backend/pkg/audit"
	"github.com/yixian-huang/inkless/backend/pkg/brand"
	"github.com/yixian-huang/inkless/backend/pkg/config"
	appLogger "github.com/yixian-huang/inkless/backend/pkg/logger"
	"github.com/yixian-huang/inkless/backend/pkg/secretcipher"
)

// App is the fully wired Inkless CMS process (DB, HTTP, background workers).
type App struct {
	Build BuildInfo
	Cfg   *config.Config
	Log   *appLogger.Logger

	database         *db.DB
	router           *gin.Engine
	schedulerService *service.SchedulerService
	pageViewRecorder *service.PageViewRecorder
	commentModule    *commentMod.Module
	pluginManager    *pluginruntime.Manager
	publicCache      *cache.Cache
	rbacCache        *cache.Cache
}

// Options configures application bootstrap.
type Options struct {
	Build BuildInfo
}

// New loads infrastructure, runs migrations/seed, wires handlers, and builds the HTTP router.
// It does not start listening; call Run.
func New(loadResult *config.LoadResult, opts Options) (*App, error) {
	if loadResult == nil || loadResult.Config == nil {
		return nil, fmt.Errorf("app: nil config load result")
	}
	build := opts.Build
	if build.Version == "" {
		build = DefaultBuildInfo()
	}
	cfg := loadResult.Config

	log := appLogger.New(cfg.Env, map[string]interface{}{
		"service": brand.APIService,
		"version": build.Version,
	})
	log.Info("Starting server",
		"env", cfg.Env,
		"port", cfg.Port,
		"version", build.Version,
		"buildTime", build.BuildTime,
		"gitCommit", build.GitCommit,
		"gitBranch", build.GitBranch,
		"bootstrapMode", loadResult.BootstrapMode,
	)
	if loadResult.BootstrapMode {
		log.Warn("Setup bootstrap mode active — use /setup to persist .env and restart")
	}

	maxOpenConn := 25
	maxIdleConn := 5
	maxLifetime := 5 * time.Minute
	if !db.IsPostgresDSN(cfg.DBDSN) {
		maxOpenConn = 4
		maxIdleConn = 2
		maxLifetime = 0
	}

	database, err := db.Init(db.InitOptions{
		DSN:         cfg.DBDSN,
		MaxOpenConn: maxOpenConn,
		MaxIdleConn: maxIdleConn,
		MaxLifetime: maxLifetime,
		LogLevel:    gormLogLevel(cfg.Env),
	})
	if err != nil {
		return nil, fmt.Errorf("initialize database: %w", err)
	}
	log.Info("Database connection established")

	if err := migrateSchema(database, log); err != nil {
		_ = database.Close()
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := database.HealthCheck(ctx); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("database health check: %w", err)
	}
	log.Info("Database health check passed")

	// --- repositories ---
	userRepo := repository.NewGormUserRepository(database.DB)
	refreshTokenRepo := repository.NewGormRefreshTokenRepository(database.DB)
	contentDocRepo := repository.NewGormContentDocumentRepository(database.DB)
	mediaRepo := repository.NewGormMediaRepository(database.DB)
	pageViewRepo := repository.NewGormPageViewRepository(database.DB)
	categoryRepo := repository.NewGormCategoryRepository(database.DB)
	tagRepo := repository.NewGormTagRepository(database.DB)
	articleRepo := repository.NewGormArticleRepository(database.DB)
	articleVersionRepo := repository.NewGormArticleVersionRepository(database.DB)
	auditEventRepo := repository.NewGormAuditEventRepository(database.DB)
	pageRepo := repository.NewGormPageRepository(database.DB)
	installedThemeRepo := repository.NewGormInstalledThemeRepository(database.DB)
	menuRepo := repository.NewGormMenuRepository(database.DB)
	roleRepo := repository.NewGormRoleRepository(database.DB)
	marketplaceRepo := repository.NewGormMarketplaceRepository(database.DB)
	mediaFolderRepo := repository.NewGormMediaFolderRepository(database.DB)
	chunkedUploadRepo := repository.NewGormChunkedUploadRepository(database.DB)
	glossaryRepo := repository.NewGormGlossaryRepository(database.DB)
	storageConfigRepo := repository.NewGormStorageConfigRepository(database.DB)
	unifiedPageRepo := repository.NewGormUnifiedPageRepository(database.DB)
	pageVersionRepo := repository.NewGormPageVersionRepository(database.DB)
	scheduledPublishJobRepo := repository.NewGormScheduledPublishJobRepository(database.DB)
	pageTemplateRepo := repository.NewGormPageTemplateRepository(database.DB)
	siteConfigRepo := repository.NewGormSiteConfigRepository(database.DB)
	log.Info("Repositories initialized")

	themePageService := service.NewThemePageService(pageRepo)

	seeder := seed.NewSeeder(userRepo, contentDocRepo, installedThemeRepo, themePageService, unifiedPageRepo, pageTemplateRepo, siteConfigRepo)
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
		userRepo,
		siteConfigRepo,
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
		_ = database.Close()
		return nil, fmt.Errorf("check install status: %w", err)
	}

	seedMode := os.Getenv("SEED_MODE")
	if err := seed.RunStartupSeed(seedCtx, installed, seedMode, seeder, func(ctx context.Context) error {
		return seedRBAC(ctx, roleRepo)
	}); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("startup seed: %w", err)
	}
	log.Info("Seed data initialized", "installed", installed, "seedMode", seedMode)
	log.Info("Services initialized")

	auditDbWriter := audit.NewDbWriter(auditEventRepo, log)
	log.Info("Audit logger initialized")

	searchService := service.NewSearchService(database.DB, db.IsPostgresDSN(cfg.DBDSN))

	registry := provider.NewRegistry()
	registry.Register("notifier", service.NewLogNotifier())
	registry.Register("captcha", &provider.NoopCaptchaProvider{})

	secretCipher, err := secretcipher.New(cfg.JWTSecret)
	if err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("secret cipher: %w", err)
	}

	aiConfigSvc := service.NewAIConfigService(
		repository.NewGormAIConfigRepository(database.DB),
		secretCipher,
		registry,
	)
	if err := aiConfigSvc.Restore(context.Background()); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("restore AI config: %w", err)
	}

	storageRuntime := service.NewStorageRuntimeService(
		storageConfigRepo,
		registry,
		service.NewLocalStorage(cfg.UploadDir),
		secretCipher,
	)
	if err := storageRuntime.RestoreStartupConfig(context.Background()); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("restore storage config: %w", err)
	}

	pluginStore := pluginruntime.NewStore(database.DB)
	pluginManager := pluginruntime.NewManager(pluginruntime.ManagerConfig{
		PluginDir: cfg.PluginDir,
		DataDir:   cfg.PluginDataDir,
	}, pluginStore, registry)
	if cfg.ExternalPlugins {
		if err := pluginManager.StartEnabledPlugins(context.Background()); err != nil {
			_ = database.Close()
			return nil, fmt.Errorf("start plugins: %w", err)
		}
		pluginManager.StartHealthMonitor(30 * time.Second)
	}
	log.Info("Provider registry initialized", "providers", registry.List())

	chunkedUploadSvc := service.NewChunkedUploadServiceWithStorage(
		chunkedUploadRepo,
		mediaRepo,
		"./tmp/uploads",
		cfg.UploadDir,
		"",
		storageRuntime,
	)

	migrationSvc := migration.NewService(articleRepo, categoryRepo, tagRepo)

	bus := eventbus.New()
	bus.Subscribe(eventbus.ContentCreated, eventbus.AsyncHandler(func(e eventbus.Event) {
		log.Info("Content event", "type", e.Type)
	}))
	bus.Subscribe(eventbus.ContentUpdated, eventbus.AsyncHandler(func(e eventbus.Event) {
		log.Info("Content event", "type", e.Type)
	}))
	bus.Subscribe(eventbus.ContentDeleted, eventbus.AsyncHandler(func(e eventbus.Event) {
		log.Info("Content event", "type", e.Type)
	}))
	log.Info("Event bus initialized")

	publicCache := cache.New(60 * time.Second)
	rbacCache := cache.New(30 * time.Second)

	mgr := module.NewManager()
	commentModule := commentMod.New()
	mgr.Register(qa.New())
	mgr.Register(commentModule)
	mgr.Register(formSubmissionMod.New())
	backupModule := backupMod.New()
	mgr.Register(backupModule)
	if err := mgr.InitAll(module.Dependencies{
		DB:       database.DB,
		Registry: registry,
		Repos: &module.SharedRepos{
			ContentDoc: contentDocRepo,
			Article:    articleRepo,
		},
		SiteCfg:    siteConfigRepo,
		UserRepo:   userRepo,
		RBACCache:  rbacCache,
		UploadDir:  cfg.UploadDir,
		BackupDir:  cfg.BackupDir,
		AppVersion: build.Version,
	}); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("init modules: %w", err)
	}

	invalidateFromEvent := func(e eventbus.Event) {
		contentType, slug := "", ""
		if p, ok := e.Payload.(eventbus.ContentEventPayload); ok {
			contentType, slug = p.ContentType, p.Slug
		} else if p, ok := e.Payload.(*eventbus.ContentEventPayload); ok && p != nil {
			contentType, slug = p.ContentType, p.Slug
		}
		cache.InvalidatePublicFromContentEvent(publicCache, contentType, slug)
	}
	for _, evt := range []string{
		eventbus.ContentCreated,
		eventbus.ContentUpdated,
		eventbus.ContentDeleted,
		eventbus.ContentPublished,
		eventbus.ContentUnpublished,
		eventbus.ContentRolledBack,
	} {
		bus.Subscribe(evt, eventbus.AsyncHandler(invalidateFromEvent))
	}

	authHandlerInst := authHandler.NewHandler(userRepo, refreshTokenRepo, cfg)
	pageViewRecorder := service.NewPageViewRecorder(pageViewRepo)
	pageViewRecorder.Start()

	publicHandlerInst := publicHandler.NewHandler(contentDocRepo, pageViewRepo, unifiedPageRepo, publicCache).
		WithViewTracker(pageViewRecorder)
	mediaHandlerInst := mediaHandler.NewHandlerWithStorage(mediaRepo, cfg.UploadDir, "", storageRuntime)
	analyticsHandlerInst := analyticsHandler.NewHandler(pageViewRepo).WithCache(publicCache)
	dashboardHandlerInst := dashboardHandler.NewHandler(articleRepo, unifiedPageRepo, mediaRepo, pageViewRepo).WithCache(publicCache)
	categoryHandlerInst := categoryHandler.NewHandler(categoryRepo, articleRepo)
	tagHandlerInst := tagHandler.NewHandler(tagRepo, articleRepo)
	menuHandlerInst := menuHandler.NewHandler(menuRepo)
	articleHandlerInst := articleHandler.NewHandler(articleRepo, categoryRepo, tagRepo, searchService, bus, publicCache).
		WithPageViews(pageViewRepo).
		WithViewTracker(pageViewRecorder).
		WithVersionRepo(articleVersionRepo)
	auditlogHandlerInst := auditlogHandler.NewHandler(auditEventRepo)
	sitemapHandlerInst := sitemapHandler.NewHandler(contentDocRepo, articleRepo, cfg.BaseURL)
	feedHandlerInst := feedHandler.NewHandler(articleRepo, siteConfigRepo, cfg.BaseURL, "Blog", "Latest posts")
	themeHandlerInst := themeHandler.NewHandler(siteConfigRepo, publicCache)
	installedThemeHandlerInst := installedThemeHandler.NewHandler(installedThemeRepo, themePageService, publicCache)
	bootstrapHandlerInst := bootstrapHandler.NewHandler(contentDocRepo, installedThemeRepo, pageRepo, unifiedPageRepo, siteConfigRepo, publicCache)
	globalConfigHandlerInst := globalConfigHandler.NewHandler(contentDocRepo, publicCache)
	featuresHandlerInst := featuresHandler.NewHandler(siteConfigRepo, publicCache)
	emailSvc := service.NewEmailService(siteConfigRepo)
	emailSettingsHandlerInst := emailSettingsHandler.NewHandler(siteConfigRepo, emailSvc)
	userHandlerInst := userHandler.NewHandler(userRepo)
	seoHandlerInst := seoHandler.NewHandler(database.DB)
	searchHandlerInst := searchhandler.NewHandler(searchService)
	roleHandlerInst := roleHandler.NewHandler(roleRepo, userRepo)
	marketplaceSvc := service.NewMarketplaceService(marketplaceRepo)
	marketplaceHandlerInst := marketplaceHandler.NewHandler(marketplaceSvc)
	pluginHandlerInst := pluginHandler.NewHandler(pluginManager, registry, cfg.ExternalPlugins)
	wizardSvc := service.NewWizardServiceWithRegistry(registry, unifiedPageRepo)
	wizardHandlerInst := wizardHandler.NewHandler(wizardSvc)
	aiHandlerInst := aiHandler.NewHandler(registry, aiConfigSvc)
	chunkedUploadHandlerInst := chunkedUploadHandler.NewHandler(chunkedUploadSvc)
	mediaFolderHandlerInst := mediaFolderHandler.NewHandler(mediaFolderRepo, mediaRepo)
	migrationHandlerInst := migrationHandler.NewHandler(migrationSvc)
	storageHandlerInst := storageHandler.NewHandlerWithRuntime(storageRuntime)
	systemHandlerInst := systemHandler.NewHandler(database.DB, cfg.UploadDir, build.Version)
	translationHandlerInst := translationHandler.NewHandlerWithRegistry(registry, glossaryRepo, articleRepo)
	unifiedPageSvc := service.NewUnifiedPageService(unifiedPageRepo, pageVersionRepo, bus).
		WithAuditWriter(auditDbWriter)
	unifiedPageHdl := unifiedPageHandler.NewHandler(unifiedPageRepo, pageVersionRepo, unifiedPageSvc, publicCache, bus)
	articlePublicationSvc := service.NewArticlePublicationService(articleRepo, searchService, bus).
		WithTaxonomyRepositories(categoryRepo, tagRepo).
		WithAuditWriter(auditDbWriter)
	schedulerService := service.NewSchedulerService(scheduledPublishJobRepo, articlePublicationSvc, unifiedPageSvc)
	schedulerService.Start()
	schedulerHdl := schedulerHandler.NewHandler(schedulerService)
	pageTemplateHdl := pageTemplateHandler.NewHandler(pageTemplateRepo)
	themeExportSvc := service.NewThemeExportService(pageTemplateRepo, siteConfigRepo)
	themeExportHdl := themeExportHandler.NewHandler(themeExportSvc)
	setupHandlerInst := setupHandler.NewHandler(setupSvc)
	log.Info("Handlers initialized")

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger(log, middleware.RequestLoggerOptions{}))
	// Skip gzip on health/metrics and already-compressed assets.
	router.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPaths([]string{
		"/health",
		"/healthz",
		"/ready",
		"/metrics",
	}), gzip.WithExcludedExtensions([]string{
		".png", ".jpg", ".jpeg", ".gif", ".webp", ".woff", ".woff2", ".zip", ".gz",
	})))
	router.Use(middleware.AuditContext())

	corsConfig := cors.Config{
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization", "If-Match"},
		MaxAge:       10 * time.Minute,
	}
	if len(cfg.CORSAllowedOrigins) > 0 {
		corsConfig.AllowOrigins = cfg.CORSAllowedOrigins
	} else {
		corsConfig.AllowAllOrigins = true
		log.Warn("CORS allowed origins not configured; falling back to allow all origins")
	}
	router.Use(cors.New(corsConfig))
	router.Use(apierror.ErrorHandler())

	handlers := &Handlers{
		Auth:           authHandlerInst,
		Article:        articleHandlerInst,
		Public:         publicHandlerInst,
		Bootstrap:      bootstrapHandlerInst,
		Media:          mediaHandlerInst,
		Analytics:      analyticsHandlerInst,
		Dashboard:      dashboardHandlerInst,
		Category:       categoryHandlerInst,
		Tag:            tagHandlerInst,
		Menu:           menuHandlerInst,
		AuditLog:       auditlogHandlerInst,
		Sitemap:        sitemapHandlerInst,
		Feed:           feedHandlerInst,
		Theme:          themeHandlerInst,
		InstalledTheme: installedThemeHandlerInst,
		EmailSettings:  emailSettingsHandlerInst,
		Features:       featuresHandlerInst,
		GlobalConfig:   globalConfigHandlerInst,
		User:           userHandlerInst,
		SEO:            seoHandlerInst,
		Search:         searchHandlerInst,
		Role:           roleHandlerInst,
		Marketplace:    marketplaceHandlerInst,
		Plugin:         pluginHandlerInst,
		Wizard:         wizardHandlerInst,
		AI:             aiHandlerInst,
		ChunkedUpload:  chunkedUploadHandlerInst,
		MediaFolder:    mediaFolderHandlerInst,
		Migration:      migrationHandlerInst,
		Storage:        storageHandlerInst,
		System:         systemHandlerInst,
		Translation:    translationHandlerInst,
		UnifiedPage:    unifiedPageHdl,
		Scheduler:      schedulerHdl,
		PageTemplate:   pageTemplateHdl,
		ThemeExport:    themeExportHdl,
	}
	routeDeps := &RouteDeps{
		UserRepo:       userRepo,
		RBACCache:      rbacCache,
		Cfg:            cfg,
		Database:       database,
		ModuleMgr:      mgr,
		ContentDocRepo: contentDocRepo,
		AuditWriter:    auditDbWriter,
		Build:          build,
	}
	registerRoutes(router, handlers, routeDeps)
	setupHandlerInst.RegisterRoutes(router, middleware.LoginRateLimit())

	log.Info("Router configured with all routes")

	return &App{
		Build:            build,
		Cfg:              cfg,
		Log:              log,
		database:         database,
		router:           router,
		schedulerService: schedulerService,
		pageViewRecorder: pageViewRecorder,
		commentModule:    commentModule,
		pluginManager:    pluginManager,
		publicCache:      publicCache,
		rbacCache:        rbacCache,
	}, nil
}

// Handler returns the HTTP handler for tests or custom servers.
func (a *App) Handler() http.Handler {
	return a.router
}

// Run starts the HTTP server and blocks until SIGINT/SIGTERM, then shuts down cleanly.
func (a *App) Run() error {
	addr := fmt.Sprintf(":%d", a.Cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      a.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		a.Log.Info("Server listening", "address", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		a.Log.Info("Server shutting down gracefully...", "signal", sig.String())
	case err := <-errCh:
		if err != nil {
			_ = a.shutdownWorkers()
			return fmt.Errorf("server listen: %w", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		a.Log.Error("Server forced to shutdown", "error", err)
	}
	return a.shutdownWorkers()
}

func (a *App) shutdownWorkers() error {
	if a.schedulerService != nil {
		a.schedulerService.Stop()
	}
	if a.pageViewRecorder != nil {
		a.pageViewRecorder.Stop(3 * time.Second)
	}
	if a.commentModule != nil {
		a.commentModule.Stop()
	}
	if a.pluginManager != nil {
		if err := a.pluginManager.StopAll(); err != nil {
			a.Log.Error("Failed to stop plugins cleanly", "error", err)
		}
	}
	if a.publicCache != nil {
		a.publicCache.Stop()
	}
	if a.rbacCache != nil {
		a.rbacCache.Stop()
	}
	if a.database != nil {
		if err := a.database.Close(); err != nil {
			a.Log.Error("Failed to close database connection", "error", err)
			return err
		}
	}
	a.Log.Info("Server stopped")
	return nil
}
