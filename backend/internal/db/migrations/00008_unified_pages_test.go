package migrations_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"

	"github.com/yixian-huang/inkless/backend/internal/db/migrations"
)

func TestMigrationUnifiedPages(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create old tables with seed data
	setupOldTables(t, db)

	// Create new tables (normally created by GORM AutoMigrate)
	setupNewTables(t, db)

	// Set up goose — use Go migrations only (no filesystem)
	goose.SetBaseFS(nil)
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("set dialect: %v", err)
	}
	migrations.Dialect = "sqlite3"

	// Create goose version table
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS goose_db_version (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		version_id INTEGER NOT NULL,
		is_applied INTEGER NOT NULL,
		tstamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		t.Fatalf("create goose table: %v", err)
	}

	// Mark versions 1-7 as applied so goose only runs migration 8
	for i := 1; i <= 7; i++ {
		if _, err := db.ExecContext(ctx,
			`INSERT INTO goose_db_version (version_id, is_applied) VALUES (?, 1)`, i); err != nil {
			t.Fatalf("mark version %d: %v", i, err)
		}
	}

	// Run migration up
	if err := goose.UpByOne(db, "."); err != nil {
		t.Fatalf("goose up: %v", err)
	}

	// Verify site_configs: should have 2 rows (global + theme)
	var scCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM site_configs`).Scan(&scCount); err != nil {
		t.Fatalf("count site_configs: %v", err)
	}
	if scCount != 2 {
		t.Errorf("expected 2 site_configs, got %d", scCount)
	}

	// Verify unified_pages from content docs: should have 7 template pages
	var templateCount int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM unified_pages WHERE mode = 'template'`).Scan(&templateCount); err != nil {
		t.Fatalf("count template pages: %v", err)
	}
	if templateCount != 7 {
		t.Errorf("expected 7 template pages, got %d", templateCount)
	}

	// Verify page_templates
	var ptCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM page_templates`).Scan(&ptCount); err != nil {
		t.Fatalf("count page_templates: %v", err)
	}
	if ptCount != 7 {
		t.Errorf("expected 7 page_templates, got %d", ptCount)
	}

	// Verify page_versions are populated
	var pvCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM page_versions`).Scan(&pvCount); err != nil {
		t.Fatalf("count page_versions: %v", err)
	}
	if pvCount != 2 {
		t.Errorf("expected 2 page_versions, got %d", pvCount)
	}

	// Verify home page has sections array in draft_config
	var homeConfig string
	if err := db.QueryRowContext(ctx,
		`SELECT draft_config FROM unified_pages WHERE slug = 'home'`).Scan(&homeConfig); err != nil {
		t.Fatalf("read home draft_config: %v", err)
	}

	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(homeConfig), &configMap); err != nil {
		t.Fatalf("parse home config: %v", err)
	}

	sections, ok := configMap["sections"].([]interface{})
	if !ok {
		t.Fatal("home config should have sections array")
	}
	if len(sections) == 0 {
		t.Error("home page should have at least one section")
	}

	// First section should be hero
	firstSection, ok := sections[0].(map[string]interface{})
	if !ok {
		t.Fatal("first section should be a map")
	}
	if firstSection["type"] != "hero" {
		t.Errorf("expected first section type 'hero', got %v", firstSection["type"])
	}

	// Verify block pages were migrated (1 non-theme page)
	var blockCount int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM unified_pages WHERE mode = 'composable'`).Scan(&blockCount); err != nil {
		t.Fatalf("count composable pages: %v", err)
	}
	if blockCount != 1 {
		t.Errorf("expected 1 composable page from block pages, got %d", blockCount)
	}

	// Test down migration
	if err := goose.Down(db, "."); err != nil {
		t.Fatalf("goose down: %v", err)
	}

	// All new tables should be empty after down
	for _, table := range []string{"page_versions", "unified_pages", "page_templates", "site_configs"} {
		var count int
		if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil {
			t.Fatalf("count %s after down: %v", table, err)
		}
		if count != 0 {
			t.Errorf("expected 0 rows in %s after down, got %d", table, count)
		}
	}
}

func setupOldTables(t *testing.T, db *sql.DB) {
	t.Helper()

	// content_documents table
	_, err := db.Exec(`CREATE TABLE content_documents (
		page_key TEXT PRIMARY KEY,
		draft_config TEXT,
		draft_version INTEGER NOT NULL DEFAULT 0,
		published_config TEXT,
		published_version INTEGER NOT NULL DEFAULT 0,
		updated_at DATETIME
	)`)
	if err != nil {
		t.Fatalf("create content_documents: %v", err)
	}

	// Seed content documents
	homeConfig := `{"hero":{"title":{"zh":"欢迎","en":"Welcome"},"subtitle":{"zh":"专业服务","en":"Professional Services"},"backgroundImage":{"url":"/images/hero.jpg"}},"advantages":{"title":{"zh":"优势","en":"Advantages"},"cards":[]}}`

	contentDocs := []struct {
		key     string
		config  string
		version int
	}{
		{"home", homeConfig, 2},
		{"about", `{"hero":{"title":{"zh":"关于我们","en":"About Us"}},"companyProfile":{"title":{"zh":"公司简介","en":"Company Profile"}}}`, 1},
		{"advantages", `{"hero":{"title":{"zh":"优势","en":"Advantages"}}}`, 1},
		{"core-services", `{"hero":{"title":{"zh":"服务","en":"Services"}},"services":[]}`, 1},
		{"cases", `{"hero":{"title":{"zh":"案例","en":"Cases"}},"cases":[]}`, 1},
		{"experts", `{"hero":{"title":{"zh":"专家","en":"Experts"}},"experts":[]}`, 1},
		{"contact", `{"hero":{"title":{"zh":"联系","en":"Contact"}},"contactInfo":{"email":"test@example.com"}}`, 1},
		{"global", `{"header":{"logo":{"url":"/logo.svg"}},"footer":{"links":[]}}`, 1},
		{"theme", `{"activeTheme":"corporate-classic"}`, 1},
	}

	for _, doc := range contentDocs {
		_, err := db.Exec(
			`INSERT INTO content_documents (page_key, draft_config, draft_version, published_config, published_version, updated_at)
			 VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
			doc.key, doc.config, doc.version, doc.config, doc.version)
		if err != nil {
			t.Fatalf("seed content_document %s: %v", doc.key, err)
		}
	}

	// content_versions table
	_, err = db.Exec(`CREATE TABLE content_versions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		page_key TEXT NOT NULL,
		version INTEGER NOT NULL,
		config TEXT NOT NULL,
		published_at DATETIME NOT NULL,
		created_by INTEGER NOT NULL,
		created_at DATETIME
	)`)
	if err != nil {
		t.Fatalf("create content_versions: %v", err)
	}

	// Seed versions for home page
	_, err = db.Exec(
		`INSERT INTO content_versions (page_key, version, config, published_at, created_by, created_at)
		 VALUES ('home', 1, ?, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP)`, homeConfig)
	if err != nil {
		t.Fatalf("seed content_version 1: %v", err)
	}
	_, err = db.Exec(
		`INSERT INTO content_versions (page_key, version, config, published_at, created_by, created_at)
		 VALUES ('home', 2, ?, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP)`, homeConfig)
	if err != nil {
		t.Fatalf("seed content_version 2: %v", err)
	}

	// pages table (block pages)
	_, err = db.Exec(`CREATE TABLE pages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		slug TEXT NOT NULL UNIQUE,
		parent_id INTEGER,
		title TEXT,
		template TEXT DEFAULT 'default',
		config TEXT,
		status TEXT DEFAULT 'draft',
		sort_order INTEGER DEFAULT 0,
		seo_title TEXT,
		seo_description TEXT,
		keywords TEXT,
		theme_id TEXT,
		content_key TEXT,
		render_mode TEXT DEFAULT 'dynamic',
		is_theme_page INTEGER DEFAULT 0,
		nav_config TEXT,
		cover_image TEXT,
		auto_summary INTEGER DEFAULT 0,
		allow_comments INTEGER DEFAULT 1,
		pinned INTEGER DEFAULT 0,
		visibility TEXT DEFAULT 'public',
		scheduled_at DATETIME,
		published_at DATETIME,
		metadata TEXT,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`)
	if err != nil {
		t.Fatalf("create pages: %v", err)
	}

	// Seed a non-theme block page
	blockConfig := `{"sections":[{"id":"s1","type":"hero","props":{"title":"Block Hero","backgroundImage":"/img/bg.jpg"}},{"id":"s2","type":"card-grid","props":{"title":"Cards","cards":[{"title":"Card A","description":"Desc A"}]}}]}`
	_, err = db.Exec(
		`INSERT INTO pages (slug, title, config, status, sort_order, is_theme_page, seo_title, seo_description, keywords, created_at, updated_at, published_at)
		 VALUES ('custom-page', '{"zh":"自定义页","en":"Custom Page"}', ?, 'published', 5, 0, '{"zh":"SEO标题","en":"SEO Title"}', '{"zh":"SEO描述","en":"SEO Desc"}', '{"zh":"关键词","en":"keywords"}', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		blockConfig)
	if err != nil {
		t.Fatalf("seed block page: %v", err)
	}

	// Seed a theme page (should NOT be migrated)
	_, err = db.Exec(
		`INSERT INTO pages (slug, title, config, status, sort_order, is_theme_page, created_at, updated_at)
		 VALUES ('theme-page', '{"zh":"主题页","en":"Theme Page"}', '{}', 'draft', 0, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`)
	if err != nil {
		t.Fatalf("seed theme page: %v", err)
	}
}

func setupNewTables(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec(`CREATE TABLE site_configs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT NOT NULL UNIQUE,
		draft_config TEXT,
		draft_version INTEGER NOT NULL DEFAULT 1,
		published_config TEXT,
		published_version INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME,
		updated_at DATETIME
	)`)
	if err != nil {
		t.Fatalf("create site_configs: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE page_templates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT NOT NULL UNIQUE,
		name_zh TEXT NOT NULL,
		name_en TEXT NOT NULL DEFAULT '',
		description_zh TEXT NOT NULL DEFAULT '',
		description_en TEXT NOT NULL DEFAULT '',
		category TEXT NOT NULL DEFAULT 'custom',
		config TEXT NOT NULL,
		thumbnail TEXT,
		created_at DATETIME,
		updated_at DATETIME
	)`)
	if err != nil {
		t.Fatalf("create page_templates: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE unified_pages (
		id INTEGER PRIMARY KEY,
		slug TEXT NOT NULL UNIQUE,
		zh_title TEXT NOT NULL DEFAULT '',
		en_title TEXT NOT NULL DEFAULT '',
		zh_description TEXT NOT NULL DEFAULT '',
		en_description TEXT NOT NULL DEFAULT '',
		mode TEXT NOT NULL DEFAULT 'composable',
		template_id INTEGER,
		draft_config TEXT,
		draft_version INTEGER NOT NULL DEFAULT 1,
		published_config TEXT,
		published_version INTEGER NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'draft',
		scheduled_at DATETIME,
		translation_status TEXT,
		zh_meta_title TEXT NOT NULL DEFAULT '',
		en_meta_title TEXT NOT NULL DEFAULT '',
		zh_meta_description TEXT NOT NULL DEFAULT '',
		en_meta_description TEXT NOT NULL DEFAULT '',
		zh_meta_keywords TEXT NOT NULL DEFAULT '',
		en_meta_keywords TEXT NOT NULL DEFAULT '',
		sort_order INTEGER NOT NULL DEFAULT 0,
		show_in_nav INTEGER NOT NULL DEFAULT 0,
		parent_id INTEGER,
		created_at DATETIME,
		updated_at DATETIME,
		published_at DATETIME,
		deleted_at DATETIME
	)`)
	if err != nil {
		t.Fatalf("create unified_pages: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE page_versions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		page_id INTEGER NOT NULL,
		version INTEGER NOT NULL,
		config TEXT NOT NULL,
		created_by INTEGER NOT NULL,
		created_at DATETIME
	)`)
	if err != nil {
		t.Fatalf("create page_versions: %v", err)
	}
}
