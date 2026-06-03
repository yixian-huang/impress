package main

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	analyticsHandler "blotting-consultancy/internal/handler/analytics"
	articleHandler "blotting-consultancy/internal/handler/article"
	auditlogHandler "blotting-consultancy/internal/handler/auditlog"
	authHandler "blotting-consultancy/internal/handler/auth"
	bootstrapHandler "blotting-consultancy/internal/handler/bootstrap"
	categoryHandler "blotting-consultancy/internal/handler/category"
	chunkedUploadHandler "blotting-consultancy/internal/handler/chunked_upload"
	emailSettingsHandler "blotting-consultancy/internal/handler/email_settings"
	feedHandler "blotting-consultancy/internal/handler/feed"
	featuresHandler "blotting-consultancy/internal/handler/features"
	globalConfigHandler "blotting-consultancy/internal/handler/global_config"
	installedThemeHandler "blotting-consultancy/internal/handler/installed_theme"
	marketplaceHandler "blotting-consultancy/internal/handler/marketplace"
	mediaHandler "blotting-consultancy/internal/handler/media"
	mediaFolderHandler "blotting-consultancy/internal/handler/media_folder"
	menuHandler "blotting-consultancy/internal/handler/menu"
	migrationHandler "blotting-consultancy/internal/handler/migration"
	pageTemplateHandler "blotting-consultancy/internal/handler/page_template"
	publicHandler "blotting-consultancy/internal/handler/public"
	roleHandler "blotting-consultancy/internal/handler/role"
	searchhandler "blotting-consultancy/internal/handler/search"
	seoHandler "blotting-consultancy/internal/handler/seo"
	siteHandler "blotting-consultancy/internal/handler/site"
	sitemapHandler "blotting-consultancy/internal/handler/sitemap"
	storageHandler "blotting-consultancy/internal/handler/storage"
	systemHandler "blotting-consultancy/internal/handler/system"
	tagHandler "blotting-consultancy/internal/handler/tag"
	themeHandler "blotting-consultancy/internal/handler/theme"
	themeExportHandler "blotting-consultancy/internal/handler/theme_export"
	translationHandler "blotting-consultancy/internal/handler/translation"
	unifiedPageHandler "blotting-consultancy/internal/handler/unified_page"
	wizardHandler "blotting-consultancy/internal/handler/wizard"
	userHandler "blotting-consultancy/internal/handler/user"
	aiHandler "blotting-consultancy/internal/handler/ai"

	"blotting-consultancy/internal/cache"
	"blotting-consultancy/internal/db"
	"blotting-consultancy/internal/middleware"
	"blotting-consultancy/internal/module"
	"blotting-consultancy/internal/repository"
	"blotting-consultancy/internal/seo"
	"blotting-consultancy/pkg/config"
	"blotting-consultancy/pkg/metrics"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Handlers holds all initialized HTTP handlers.
type Handlers struct {
	Auth            *authHandler.Handler
	Article         *articleHandler.Handler
	Public          *publicHandler.Handler
	Bootstrap       *bootstrapHandler.Handler
	Media           *mediaHandler.Handler
	Analytics       *analyticsHandler.Handler
	Category        *categoryHandler.Handler
	Tag             *tagHandler.Handler
	Menu            *menuHandler.Handler
	AuditLog        *auditlogHandler.Handler
	Sitemap         *sitemapHandler.Handler
	Feed            *feedHandler.Handler
	Theme           *themeHandler.Handler
	InstalledTheme  *installedThemeHandler.Handler
	EmailSettings   *emailSettingsHandler.Handler
	Features        *featuresHandler.Handler
	GlobalConfig    *globalConfigHandler.Handler
	User            *userHandler.Handler
	SEO             *seoHandler.Handler
	Search          *searchhandler.Handler
	Role            *roleHandler.Handler
	Marketplace     *marketplaceHandler.Handler
	Wizard          *wizardHandler.Handler
	AI              *aiHandler.Handler
	ChunkedUpload   *chunkedUploadHandler.Handler
	MediaFolder     *mediaFolderHandler.Handler
	Migration       *migrationHandler.Handler
	Site            *siteHandler.Handler
	Storage         *storageHandler.Handler
	System          *systemHandler.Handler
	Translation     *translationHandler.Handler
	UnifiedPage     *unifiedPageHandler.Handler
	PageTemplate    *pageTemplateHandler.Handler
	ThemeExport     *themeExportHandler.Handler
}

// RouteDeps holds dependencies needed by route middleware.
type RouteDeps struct {
	UserRepo       repository.UserRepository
	RBACCache      *cache.Cache
	Cfg            *config.Config
	Database       *db.DB
	ModuleMgr      *module.Manager
	ContentDocRepo repository.ContentDocumentRepository
}

// registerRoutes sets up all route groups, middleware, and endpoint registrations
// on the provided Gin engine.
func registerRoutes(router *gin.Engine, handlers *Handlers, deps *RouteDeps) {
	cfg := deps.Cfg

	// Version endpoint (public, no auth required)
	router.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"version":   Version,
			"buildTime": BuildTime,
			"gitCommit": GitCommit,
			"gitBranch": GitBranch,
		})
	})

	// Health endpoint (no auth required)
	router.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		// Check database connection
		if err := deps.Database.HealthCheck(ctx); err != nil {
			c.JSON(503, gin.H{
				"status": "unhealthy",
				"error":  "database connection failed",
			})
			return
		}

		c.JSON(200, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   Version,
			"buildTime": BuildTime,
			"gitCommit": GitCommit,
		})
	})

	// Metrics endpoint (no auth required, for operations dashboards)
	router.GET("/metrics", func(c *gin.Context) {
		m := metrics.Global()
		publishTotal, publishSuccess, publishFailure := m.GetPublishMetrics()
		validationTotal, validationFailures := m.GetValidationMetrics()
		rollbackTotal, rollbackSuccess, rollbackFailure, rollbackP95 := m.GetRollbackMetrics()
		publicGetTotal, publicGetSuccess, publicGetFailure, publicGetP95 := m.GetPublicGetMetrics()

		c.JSON(200, gin.H{
			"publish": gin.H{
				"total":   publishTotal,
				"success": publishSuccess,
				"failure": publishFailure,
			},
			"validation": gin.H{
				"total":    validationTotal,
				"failures": validationFailures,
			},
			"rollback": gin.H{
				"total":       rollbackTotal,
				"success":     rollbackSuccess,
				"failure":     rollbackFailure,
				"latency_p95": rollbackP95.Milliseconds(),
			},
			"public_get": gin.H{
				"total":       publicGetTotal,
				"success":     publicGetSuccess,
				"failure":     publicGetFailure,
				"latency_p95": publicGetP95.Milliseconds(),
			},
		})
	})

	// Swagger API documentation (no auth required)
	router.GET("/api-docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Sitemap (no auth required)
	router.GET("/sitemap.xml", handlers.Sitemap.GetSitemap)

	// RSS feed (no auth required)
	if handlers.Feed != nil {
		router.GET("/feed.xml", handlers.Feed.GetFeed)
	}

	// Robots.txt (no auth required)
	router.GET("/robots.txt", handlers.SEO.GetRobotsTxt)

	// Public routes (no auth required)
	publicGroup := router.Group("/public")
	publicGroup.Use(middleware.PublicRateLimit())
	{
		publicGroup.GET("/bootstrap", handlers.Bootstrap.PublicBootstrap)
		publicGroup.GET("/content/:pageKey", handlers.Public.GetPublicContent)

		// Public article routes
		publicGroup.GET("/articles", handlers.Article.PublicList)
		publicGroup.GET("/articles/:slug", handlers.Article.PublicGetBySlug)

		// Public category routes
		publicGroup.GET("/categories", handlers.Category.PublicList)
		publicGroup.GET("/categories/:slug", handlers.Category.PublicGetBySlug)

		// Public tag routes
		publicGroup.GET("/tags", handlers.Tag.PublicList)
		publicGroup.GET("/tags/:slug", handlers.Tag.PublicGetBySlug)

		// Public menu route
		publicGroup.GET("/menu", handlers.Menu.PublicGetPrimary)

		// Public theme route
		publicGroup.GET("/theme", handlers.Theme.PublicGet)

		// Public active theme route
		publicGroup.GET("/active-theme", handlers.InstalledTheme.PublicGetActive)

		// Unified pages (replaces old page routes)
		publicGroup.GET("/pages", handlers.UnifiedPage.PublicList)
		publicGroup.GET("/pages/:slug", handlers.UnifiedPage.PublicGetBySlug)
	}

	// Auth routes (no auth middleware, but handlers validate credentials)
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", middleware.LoginRateLimit(), handlers.Auth.Login)
		authGroup.POST("/refresh", handlers.Auth.Refresh)
		authGroup.POST("/logout", handlers.Auth.Logout)

		// Protected auth routes
		authProtected := authGroup.Group("")
		authProtected.Use(middleware.Auth(cfg.JWTSecret))
		{
			authProtected.GET("/me", handlers.Auth.Me)
		}
	}

	// Initialize SEO renderer when FRONTEND_DIR is configured
	var seoRenderer *seo.Renderer
	if cfg.FrontendDir != "" {
		seoIndexPath := filepath.Join(cfg.FrontendDir, "index.html")
		var err error
		seoRenderer, err = seo.NewRenderer(seoIndexPath)
		if err != nil {
			panic(fmt.Sprintf("Failed to create SEO renderer: %v", err))
		}
	}

	// Admin routes (require authentication and authorization)
	adminGroup := router.Group("/admin")
	// SPA fallback: if FRONTEND_DIR is set and the browser asks for HTML,
	// serve index.html instead of requiring auth (the SPA handles its own auth).
	if cfg.FrontendDir != "" {
		indexPath := filepath.Join(cfg.FrontendDir, "index.html")
		adminGroup.Use(func(c *gin.Context) {
			accept := c.GetHeader("Accept")
			if c.Request.Method == "GET" && strings.Contains(accept, "text/html") {
				if !serveSPAWithMeta(c, seoRenderer, cfg.BaseURL, deps.ContentDocRepo) {
					c.File(indexPath)
					c.Abort()
				}
				return
			}
			c.Next()
		})
	}
	adminGroup.Use(middleware.Auth(cfg.JWTSecret))
	// Legacy middleware kept for backward compatibility with existing JWT tokens.
	// New RBAC permission checks are applied at the route-group level.
	adminGroup.Use(middleware.RequireAdminOrEditor())

	// Register module routes
	deps.ModuleMgr.RegisterAllRoutes(publicGroup, adminGroup)

	{
		// Media management
		adminGroup.POST("/media/upload", handlers.Media.Upload)
		adminGroup.GET("/media", handlers.Media.List)
		adminGroup.DELETE("/media/:id", handlers.Media.Delete)
		adminGroup.PUT("/media/:id/crop", handlers.Media.Recrop)
		adminGroup.PUT("/media/:id", handlers.Media.Rename)
		adminGroup.GET("/media/:id/usages", handlers.Media.GetUsages)

		// Analytics (requires analytics:read via RBAC)
		adminAnalytics := adminGroup.Group("")
		adminAnalytics.Use(middleware.RequirePermission("analytics", "read", deps.UserRepo, deps.RBACCache))
		{
			adminAnalytics.GET("/analytics/summary", handlers.Analytics.GetSummary)
		}

		// Article management
		adminGroup.GET("/articles", handlers.Article.AdminList)
		adminGroup.GET("/articles/:id", handlers.Article.AdminGetByID)
		adminGroup.POST("/articles", handlers.Article.AdminCreate)
		adminGroup.PUT("/articles/:id", handlers.Article.AdminUpdate)
		adminGroup.DELETE("/articles/:id", handlers.Article.AdminDelete)
		adminGroup.GET("/articles/:id/export", handlers.Article.AdminExportMarkdown)
		adminGroup.POST("/articles/import", handlers.Article.AdminImportMarkdown)

		// Category management
		adminGroup.GET("/categories", handlers.Category.List)
		adminGroup.GET("/categories/tree", handlers.Category.ListTree)
		adminGroup.GET("/categories/:id", handlers.Category.GetByID)
		adminGroup.POST("/categories", handlers.Category.Create)
		adminGroup.PUT("/categories/:id", handlers.Category.Update)
		adminGroup.DELETE("/categories/:id", handlers.Category.Delete)

		// Tag management
		adminGroup.GET("/tags", handlers.Tag.List)
		adminGroup.POST("/tags", handlers.Tag.Create)
		adminGroup.PUT("/tags/:id", handlers.Tag.Update)
		adminGroup.DELETE("/tags/:id", handlers.Tag.Delete)

		// Menu management
		adminGroup.GET("/menus", handlers.Menu.ListGroups)
		adminGroup.POST("/menus", handlers.Menu.CreateGroup)
		adminGroup.GET("/menus/:id", handlers.Menu.GetGroup)
		adminGroup.PUT("/menus/:id", handlers.Menu.UpdateGroup)
		adminGroup.DELETE("/menus/:id", handlers.Menu.DeleteGroup)
		adminGroup.PUT("/menus/:id/primary", handlers.Menu.SetPrimary)
		adminGroup.POST("/menus/:id/items", handlers.Menu.CreateItem)
		adminGroup.PUT("/menus/:id/items/:itemId", handlers.Menu.UpdateItem)
		adminGroup.DELETE("/menus/:id/items/:itemId", handlers.Menu.DeleteItem)
		adminGroup.PUT("/menus/:id/items/reorder", handlers.Menu.ReorderItems)

		// Audit logs (requires audit_logs:read via RBAC)
		adminAudit := adminGroup.Group("")
		adminAudit.Use(middleware.RequirePermission("audit_logs", "read", deps.UserRepo, deps.RBACCache))
		{
			adminAudit.GET("/audit-logs", handlers.AuditLog.List)
		}

		// Theme token management (existing)
		adminGroup.GET("/theme", handlers.Theme.AdminGet)
		adminGroup.PUT("/theme", handlers.Theme.AdminUpdate)

		// Installed theme management
		adminGroup.GET("/themes", handlers.InstalledTheme.AdminList)
		adminGroup.GET("/themes/:id", handlers.InstalledTheme.AdminGetByID)
		adminGroup.POST("/themes", handlers.InstalledTheme.AdminCreate)
		adminGroup.PUT("/themes/:id", handlers.InstalledTheme.AdminUpdate)
		adminGroup.DELETE("/themes/:id", handlers.InstalledTheme.AdminDelete)
		adminGroup.PUT("/themes/:id/activate", handlers.InstalledTheme.AdminActivate)

		// Email settings management
		adminGroup.GET("/email-settings", handlers.EmailSettings.HandleGet)
		adminGroup.PUT("/email-settings", handlers.EmailSettings.HandleUpdate)
		adminGroup.POST("/email-settings/test", handlers.EmailSettings.HandleTest)

		// Global config (branding / identity / SEO defaults)
		handlers.GlobalConfig.RegisterRoutes(adminGroup)

		// Features (route gates, blog comments/rss toggles)
		handlers.Features.RegisterRoutes(adminGroup)

		// User management (requires users:manage via RBAC)
		adminUsers := adminGroup.Group("/users")
		adminUsers.Use(middleware.RequirePermission("users", "manage", deps.UserRepo, deps.RBACCache))
		{
			adminUsers.GET("", handlers.User.List)
			adminUsers.GET("/:id", handlers.User.GetByID)
			adminUsers.POST("", handlers.User.Create)
			adminUsers.PUT("/:id", handlers.User.Update)
			adminUsers.DELETE("/:id", handlers.User.Delete)
		}

		// RBAC Role management (requires roles:manage via RBAC)
		adminRoles := adminGroup.Group("/roles")
		adminRoles.Use(middleware.RequirePermission("roles", "manage", deps.UserRepo, deps.RBACCache))
		{
			adminRoles.GET("", handlers.Role.List)
			adminRoles.GET("/:id", handlers.Role.GetByID)
			adminRoles.POST("", handlers.Role.Create)
			adminRoles.PUT("/:id", handlers.Role.Update)
			adminRoles.DELETE("/:id", handlers.Role.Delete)
			adminRoles.POST("/assign", handlers.Role.AssignRole)
			adminRoles.POST("/unassign", handlers.Role.UnassignRole)
			adminRoles.GET("/user/:userId", handlers.Role.GetUserRoles)
		}

		// Permission listing (requires roles:read via RBAC)
		adminGroup.GET("/permissions", handlers.Role.ListPermissions)

		// AI Site Building Wizard
		adminGroup.POST("/wizard/generate-plan", handlers.Wizard.GeneratePlan)
		adminGroup.POST("/wizard/apply-plan", handlers.Wizard.ApplyPlan)
		adminGroup.POST("/wizard/suggest-colors", handlers.Wizard.SuggestColors)
		adminGroup.POST("/wizard/generate-content", handlers.Wizard.GenerateContent)

		// Marketplace (plugin/theme registry)
		adminGroup.GET("/marketplace/items", handlers.Marketplace.AdminListItems)
		adminGroup.GET("/marketplace/installed", handlers.Marketplace.AdminListInstalled)
		adminGroup.POST("/marketplace/items", handlers.Marketplace.AdminRegisterItem)
		adminGroup.GET("/marketplace/items/:slug", handlers.Marketplace.AdminGetItem)
		adminGroup.POST("/marketplace/items/:slug/install", handlers.Marketplace.AdminInstallItem)
		adminGroup.PUT("/marketplace/items/:slug/update", handlers.Marketplace.AdminUpdateItem)
		adminGroup.DELETE("/marketplace/items/:slug", handlers.Marketplace.AdminUninstallItem)
		adminGroup.POST("/marketplace/items/:slug/versions", handlers.Marketplace.AdminAddVersion)

		// AI provider management
		adminGroup.POST("/ai/chat", handlers.AI.Chat)
		adminGroup.POST("/ai/summarize", handlers.AI.Summarize)
		adminGroup.POST("/ai/suggest-titles", handlers.AI.SuggestTitles)
		adminGroup.POST("/ai/suggest-tags", handlers.AI.SuggestTags)
		adminGroup.POST("/ai/complete", handlers.AI.Complete)
		adminGroup.GET("/ai/config", handlers.AI.GetConfig)
		adminGroup.PUT("/ai/config", handlers.AI.UpdateConfig)

		// Chunked upload
		adminGroup.POST("/media/upload/init", handlers.ChunkedUpload.InitUpload)
		adminGroup.POST("/media/upload/:uploadId/chunk", handlers.ChunkedUpload.UploadChunk)
		adminGroup.POST("/media/upload/:uploadId/complete", handlers.ChunkedUpload.CompleteUpload)

		// Media folders
		adminGroup.GET("/media/folders", handlers.MediaFolder.ListTree)
		adminGroup.POST("/media/folders", handlers.MediaFolder.Create)
		adminGroup.PUT("/media/folders/:id", handlers.MediaFolder.Rename)
		adminGroup.DELETE("/media/folders/:id", handlers.MediaFolder.Delete)
		adminGroup.PUT("/media/:id/move", handlers.MediaFolder.MoveMedia)

		// Data migration (import from WordPress, Halo, Markdown)
		adminGroup.POST("/migration/import", handlers.Migration.Import)
		adminGroup.GET("/migration/jobs", handlers.Migration.ListJobs)
		adminGroup.GET("/migration/jobs/:jobId", handlers.Migration.GetJob)
		adminGroup.GET("/migration/jobs/:jobId/stream", handlers.Migration.StreamProgress)

		// Site management
		adminGroup.GET("/sites", handlers.Site.AdminList)
		adminGroup.GET("/sites/:id", handlers.Site.AdminGetByID)
		adminGroup.POST("/sites", handlers.Site.AdminCreate)
		adminGroup.PUT("/sites/:id", handlers.Site.AdminUpdate)
		adminGroup.DELETE("/sites/:id", handlers.Site.AdminDelete)
		adminGroup.GET("/sites/:id/users", handlers.Site.AdminListUsers)
		adminGroup.POST("/sites/:id/users", handlers.Site.AdminAssignUser)
		adminGroup.DELETE("/sites/:id/users/:userId", handlers.Site.AdminUnassignUser)
		adminGroup.GET("/sites/:id/export", handlers.Site.AdminExport)
		adminGroup.POST("/sites/import", handlers.Site.AdminImport)

		// Storage configuration
		adminGroup.GET("/storage/config", handlers.Storage.GetConfig)
		adminGroup.PUT("/storage/config", handlers.Storage.UpdateConfig)
		adminGroup.POST("/storage/test", handlers.Storage.TestConnection)

		// System status
		adminGroup.GET("/system/status", handlers.System.GetStatus)

		// Translation & glossary
		adminGroup.POST("/translate", handlers.Translation.Translate)
		adminGroup.POST("/translate/batch", handlers.Translation.BatchTranslate)
		adminGroup.POST("/translate/article/:id", handlers.Translation.TranslateArticle)
		adminGroup.GET("/glossary", handlers.Translation.GlossaryList)
		adminGroup.POST("/glossary", handlers.Translation.GlossaryCreate)
		adminGroup.PUT("/glossary/:id", handlers.Translation.GlossaryUpdate)
		adminGroup.DELETE("/glossary/:id", handlers.Translation.GlossaryDelete)

		// Page management (unified page system)
		adminGroup.GET("/pages", handlers.UnifiedPage.AdminList)
		adminGroup.GET("/pages/:id", handlers.UnifiedPage.AdminGetByID)
		adminGroup.POST("/pages", handlers.UnifiedPage.AdminCreate)
		adminGroup.GET("/pages/:id/draft", handlers.UnifiedPage.AdminGetDraft)
		adminGroup.PUT("/pages/:id/draft", handlers.UnifiedPage.AdminUpdateDraft)
		adminGroup.POST("/pages/:id/publish", handlers.UnifiedPage.AdminPublish)
		adminGroup.POST("/pages/:id/unpublish", handlers.UnifiedPage.AdminUnpublish)
		adminGroup.POST("/pages/:id/rollback", handlers.UnifiedPage.AdminRollback)
		adminGroup.GET("/pages/:id/versions", handlers.UnifiedPage.AdminListVersions)
		adminGroup.GET("/pages/:id/versions/:version", handlers.UnifiedPage.AdminGetVersionDetail)
		adminGroup.DELETE("/pages/:id", handlers.UnifiedPage.AdminDelete)

		// Page template management
		adminGroup.GET("/templates", handlers.PageTemplate.List)
		adminGroup.POST("/templates", handlers.PageTemplate.Create)
		adminGroup.PUT("/templates/:id", handlers.PageTemplate.Update)
		adminGroup.DELETE("/templates/:id", handlers.PageTemplate.Delete)
		adminGroup.POST("/templates/:id/duplicate", handlers.PageTemplate.Duplicate)

		// Theme export/import
		adminGroup.POST("/theme-packages/export", handlers.ThemeExport.Export)
		adminGroup.POST("/theme-packages/import", handlers.ThemeExport.Import)
		adminGroup.GET("/theme-packages", handlers.ThemeExport.List)
		adminGroup.PUT("/theme-packages/:id/apply", handlers.ThemeExport.Apply)
	}

	// SEO routes (public + admin)
	handlers.SEO.RegisterRoutes(publicGroup, adminGroup)

	// Search routes (public + admin)
	handlers.Search.RegisterRoutes(publicGroup, adminGroup)

	// Serve uploaded files statically
	router.Static("/uploads", cfg.UploadDir)

	// Serve frontend static assets when FRONTEND_DIR is configured
	if cfg.FrontendDir != "" {
		router.Static("/assets", filepath.Join(cfg.FrontendDir, "assets"))
		router.Static("/images", filepath.Join(cfg.FrontendDir, "images"))
		router.StaticFile("/favicon.ico", filepath.Join(cfg.FrontendDir, "favicon.ico"))

		// SPA fallback: non-API GET requests return index.html with SEO meta
		indexHTML := filepath.Join(cfg.FrontendDir, "index.html")
		router.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			if c.Request.Method == "GET" &&
				!strings.HasPrefix(path, "/public/") &&
				!strings.HasPrefix(path, "/auth/") &&
				!strings.HasPrefix(path, "/uploads/") &&
				path != "/health" &&
				path != "/version" &&
				path != "/metrics" &&
				path != "/sitemap.xml" &&
				path != "/robots.txt" {
				if !serveSPAWithMeta(c, seoRenderer, cfg.BaseURL, deps.ContentDocRepo) {
					http.ServeFile(c.Writer, c.Request, indexHTML)
					c.Abort()
				}
				return
			}
			c.JSON(404, gin.H{"error": "not found"})
		})
	}
}
