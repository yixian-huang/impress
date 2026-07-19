package migrations

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/pressly/goose/v3"

	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/internal/service"
)

func init() {
	goose.AddMigrationContext(upUnifiedPages, downUnifiedPages)
}

// pageKeyToID maps page keys to deterministic unified_pages IDs (1-7).
var pageKeyToID = map[string]uint{
	"home":          1,
	"about":         2,
	"advantages":    3,
	"core-services": 4,
	"cases":         5,
	"experts":       6,
	"contact":       7,
}

// pageKeyNames maps page keys to bilingual display names.
var pageKeyNames = map[string][2]string{
	"home":          {"首页", "Home"},
	"about":         {"关于我们", "About Us"},
	"advantages":    {"我们的优势", "Our Advantages"},
	"core-services": {"核心服务", "Core Services"},
	"cases":         {"案例展示", "Case Studies"},
	"experts":       {"专家团队", "Our Experts"},
	"contact":       {"联系我们", "Contact Us"},
}

func upUnifiedPages(ctx context.Context, tx *sql.Tx) error {
	if Dialect != "sqlite3" {
		log.Println("Migration 00008_unified_pages: skipping (non-SQLite dialect, no legacy data to migrate)")
		return nil
	}

	// Check if old tables exist (fresh DB won't have them)
	if !tableExists(ctx, tx, "content_documents") {
		log.Println("Migration 00008_unified_pages: no old tables found (fresh DB), skipping data migration")
		return nil
	}

	// Step 1: Migrate global + theme content docs → site_configs
	if err := migrateContentDocsToSiteConfigs(ctx, tx); err != nil {
		return fmt.Errorf("migrate site_configs: %w", err)
	}

	// Step 2: Migrate 7 page-type content docs → page_templates + unified_pages
	if err := migrateContentDocsToUnifiedPages(ctx, tx); err != nil {
		return fmt.Errorf("migrate content docs to unified pages: %w", err)
	}

	// Step 3: Migrate content_versions → page_versions
	if tableExists(ctx, tx, "content_versions") {
		if err := migrateContentVersions(ctx, tx); err != nil {
			return fmt.Errorf("migrate content versions: %w", err)
		}
	}

	// Step 4: Migrate block pages → unified_pages (IDs starting from 100)
	if tableExists(ctx, tx, "pages") {
		if err := migrateBlockPages(ctx, tx); err != nil {
			return fmt.Errorf("migrate block pages: %w", err)
		}
	}

	log.Println("Migration 00008_unified_pages: up complete")
	return nil
}

// tableExists checks if a table exists in the database.
func tableExists(ctx context.Context, tx *sql.Tx, name string) bool {
	var count int
	err := tx.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", name).Scan(&count)
	return err == nil && count > 0
}

func downUnifiedPages(ctx context.Context, tx *sql.Tx) error {
	if Dialect != "sqlite3" {
		log.Println("Migration 00008_unified_pages: skipping down (non-SQLite dialect)")
		return nil
	}

	// Delete migrated data from new tables; old tables remain intact
	for _, table := range []string{"page_versions", "unified_pages", "page_templates", "site_configs"} {
		if _, err := tx.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s", table)); err != nil {
			return fmt.Errorf("delete from %s: %w", table, err)
		}
	}
	log.Println("Migration 00008_unified_pages: down complete")
	return nil
}

// migrateContentDocsToSiteConfigs moves global and theme content docs into site_configs.
func migrateContentDocsToSiteConfigs(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx,
		`SELECT page_key, draft_config, draft_version, published_config, published_version
		 FROM content_documents WHERE page_key IN ('global', 'theme')`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var pageKey string
		var draftConfig, publishedConfig sql.NullString
		var draftVersion, publishedVersion int

		if err := rows.Scan(&pageKey, &draftConfig, &draftVersion, &publishedConfig, &publishedVersion); err != nil {
			return err
		}

		dc := "{}"
		if draftConfig.Valid && draftConfig.String != "" {
			dc = draftConfig.String
		}
		pc := "{}"
		if publishedConfig.Valid && publishedConfig.String != "" {
			pc = publishedConfig.String
		}

		_, err := tx.ExecContext(ctx,
			`INSERT INTO site_configs (key, draft_config, draft_version, published_config, published_version, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
			pageKey, dc, draftVersion, pc, publishedVersion)
		if err != nil {
			return fmt.Errorf("insert site_config %s: %w", pageKey, err)
		}
	}
	return rows.Err()
}

// migrateContentDocsToUnifiedPages converts 7 page-type content docs into
// page_templates + unified_pages with section-based configs.
func migrateContentDocsToUnifiedPages(ctx context.Context, tx *sql.Tx) error {
	pageKeys := []string{"home", "about", "advantages", "core-services", "cases", "experts", "contact"}

	for _, pk := range pageKeys {
		var draftConfigStr, publishedConfigStr sql.NullString
		var draftVersion, publishedVersion int

		err := tx.QueryRowContext(ctx,
			`SELECT draft_config, draft_version, published_config, published_version
			 FROM content_documents WHERE page_key = ?`, pk).
			Scan(&draftConfigStr, &draftVersion, &publishedConfigStr, &publishedVersion)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return fmt.Errorf("read content_document %s: %w", pk, err)
		}

		// Parse and convert draft config
		draftSections := model.JSONMap{"sections": []interface{}{}}
		if draftConfigStr.Valid && draftConfigStr.String != "" && draftConfigStr.String != "{}" {
			var raw model.JSONMap
			if err := json.Unmarshal([]byte(draftConfigStr.String), &raw); err == nil {
				draftSections = service.ConvertContentDocToSections(pk, raw)
			}
		}

		// Parse and convert published config
		var publishedSectionsJSON *string
		if publishedConfigStr.Valid && publishedConfigStr.String != "" && publishedConfigStr.String != "{}" {
			var raw model.JSONMap
			if err := json.Unmarshal([]byte(publishedConfigStr.String), &raw); err == nil {
				converted := service.ConvertContentDocToSections(pk, raw)
				b, _ := json.Marshal(converted)
				s := string(b)
				publishedSectionsJSON = &s
			}
		}

		draftJSON, _ := json.Marshal(draftSections)
		names := pageKeyNames[pk]
		pageID := pageKeyToID[pk]
		templateKey := fmt.Sprintf("builtin-%s", pk)

		// Create page_template
		_, err = tx.ExecContext(ctx,
			`INSERT INTO page_templates (id, key, name_zh, name_en, description_zh, description_en, category, config, created_at, updated_at)
			 VALUES (?, ?, ?, ?, '', '', 'builtin', ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
			pageID, templateKey, names[0], names[1], string(draftJSON))
		if err != nil {
			return fmt.Errorf("insert page_template %s: %w", pk, err)
		}

		// Determine status and published_at
		status := "draft"
		var publishedAt *string
		pubConfigVal := "NULL"
		if publishedSectionsJSON != nil {
			status = "published"
			now := "CURRENT_TIMESTAMP"
			publishedAt = &now
			pubConfigVal = *publishedSectionsJSON
		}

		if publishedAt != nil {
			_, err = tx.ExecContext(ctx,
				`INSERT INTO unified_pages (id, slug, zh_title, en_title, zh_description, en_description,
					mode, template_id, draft_config, draft_version, published_config, published_version,
					status, sort_order, show_in_nav, created_at, updated_at, published_at)
				 VALUES (?, ?, ?, ?, '', '', 'template', ?, ?, ?, ?, ?, ?, ?, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
				pageID, pk, names[0], names[1],
				pageID, string(draftJSON), draftVersion, pubConfigVal, publishedVersion,
				status, pageID)
		} else {
			_, err = tx.ExecContext(ctx,
				`INSERT INTO unified_pages (id, slug, zh_title, en_title, zh_description, en_description,
					mode, template_id, draft_config, draft_version, published_config, published_version,
					status, sort_order, show_in_nav, created_at, updated_at)
				 VALUES (?, ?, ?, ?, '', '', 'template', ?, ?, ?, NULL, ?, ?, ?, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
				pageID, pk, names[0], names[1],
				pageID, string(draftJSON), draftVersion, publishedVersion,
				status, pageID)
		}
		if err != nil {
			return fmt.Errorf("insert unified_page %s: %w", pk, err)
		}
	}
	return nil
}

// migrateContentVersions converts content_versions into page_versions,
// transforming configs via ConvertContentDocToSections.
func migrateContentVersions(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx,
		`SELECT id, page_key, version, config, created_by, created_at
		 FROM content_versions
		 WHERE page_key NOT IN ('global', 'theme')
		 ORDER BY id`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id uint
		var pageKey string
		var version int
		var configStr sql.NullString
		var createdBy uint
		var createdAt string

		if err := rows.Scan(&id, &pageKey, &version, &configStr, &createdBy, &createdAt); err != nil {
			return err
		}

		pageID, ok := pageKeyToID[pageKey]
		if !ok {
			continue
		}

		convertedJSON := "{}"
		if configStr.Valid && configStr.String != "" && configStr.String != "{}" {
			var raw model.JSONMap
			if err := json.Unmarshal([]byte(configStr.String), &raw); err == nil {
				converted := service.ConvertContentDocToSections(pageKey, raw)
				b, _ := json.Marshal(converted)
				convertedJSON = string(b)
			}
		}

		_, err := tx.ExecContext(ctx,
			`INSERT INTO page_versions (page_id, version, config, created_by, created_at)
			 VALUES (?, ?, ?, ?, ?)`,
			pageID, version, convertedJSON, createdBy, createdAt)
		if err != nil {
			return fmt.Errorf("insert page_version for %s v%d: %w", pageKey, version, err)
		}
	}
	return rows.Err()
}

// migrateBlockPages converts non-theme block pages into unified_pages (IDs starting from 100).
func migrateBlockPages(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx,
		`SELECT id, slug, title, config, status, sort_order, parent_id,
			seo_title, seo_description, keywords, created_at, updated_at, published_at
		 FROM pages
		 WHERE is_theme_page = 0 OR is_theme_page IS NULL
		 ORDER BY id`)
	if err != nil {
		return err
	}
	defer rows.Close()

	nextID := uint(100)
	for rows.Next() {
		var (
			oldID                                   uint
			slug                                    string
			titleJSON, configJSON                   sql.NullString
			status                                  string
			sortOrder                               int
			parentID                                sql.NullInt64
			seoTitleJSON, seoDescJSON, keywordsJSON sql.NullString
			createdAt, updatedAt                    string
			publishedAt                             sql.NullString
		)

		if err := rows.Scan(&oldID, &slug, &titleJSON, &configJSON, &status,
			&sortOrder, &parentID, &seoTitleJSON, &seoDescJSON, &keywordsJSON,
			&createdAt, &updatedAt, &publishedAt); err != nil {
			return fmt.Errorf("scan block page: %w", err)
		}

		// Parse bilingual title
		zhTitle, enTitle := "", ""
		if titleJSON.Valid && titleJSON.String != "" {
			var titleMap map[string]interface{}
			if err := json.Unmarshal([]byte(titleJSON.String), &titleMap); err == nil {
				if zh, ok := titleMap["zh"].(string); ok {
					zhTitle = zh
				}
				if en, ok := titleMap["en"].(string); ok {
					enTitle = en
				}
			}
		}

		// Parse and convert config
		draftConfigJSON := `{"sections":[]}`
		if configJSON.Valid && configJSON.String != "" && configJSON.String != "{}" {
			var raw model.JSONMap
			if err := json.Unmarshal([]byte(configJSON.String), &raw); err == nil {
				converted := service.ConvertBlockPageToUnified(raw)
				b, _ := json.Marshal(converted)
				draftConfigJSON = string(b)
			}
		}

		// Parse SEO fields
		zhMetaTitle, enMetaTitle := extractBilingual(seoTitleJSON)
		zhMetaDesc, enMetaDesc := extractBilingual(seoDescJSON)
		zhKeywords, enKeywords := extractBilingual(keywordsJSON)

		// Determine published config
		var pubConfig interface{}
		if status == "published" {
			pubConfig = draftConfigJSON
		}

		pageID := nextID
		nextID++

		if pubConfig != nil {
			_, err = tx.ExecContext(ctx,
				`INSERT INTO unified_pages (id, slug, zh_title, en_title, zh_description, en_description,
					mode, draft_config, draft_version, published_config, published_version,
					status, sort_order, show_in_nav,
					zh_meta_title, en_meta_title, zh_meta_description, en_meta_description,
					zh_meta_keywords, en_meta_keywords,
					created_at, updated_at, published_at)
				 VALUES (?, ?, ?, ?, '', '', 'composable', ?, 1, ?, 1, ?, ?, false,
					?, ?, ?, ?, ?, ?,
					?, ?, ?)`,
				pageID, slug, zhTitle, enTitle,
				draftConfigJSON, pubConfig, status, sortOrder,
				zhMetaTitle, enMetaTitle, zhMetaDesc, enMetaDesc, zhKeywords, enKeywords,
				createdAt, updatedAt, publishedAt.String)
		} else {
			_, err = tx.ExecContext(ctx,
				`INSERT INTO unified_pages (id, slug, zh_title, en_title, zh_description, en_description,
					mode, draft_config, draft_version, published_config, published_version,
					status, sort_order, show_in_nav,
					zh_meta_title, en_meta_title, zh_meta_description, en_meta_description,
					zh_meta_keywords, en_meta_keywords,
					created_at, updated_at)
				 VALUES (?, ?, ?, ?, '', '', 'composable', ?, 1, NULL, 0, ?, ?, false,
					?, ?, ?, ?, ?, ?,
					?, ?)`,
				pageID, slug, zhTitle, enTitle,
				draftConfigJSON, status, sortOrder,
				zhMetaTitle, enMetaTitle, zhMetaDesc, enMetaDesc, zhKeywords, enKeywords,
				createdAt, updatedAt)
		}
		if err != nil {
			return fmt.Errorf("insert unified_page from block page %d (%s): %w", oldID, slug, err)
		}
	}
	return rows.Err()
}

// extractBilingual parses a nullable JSON string into zh/en string pair.
func extractBilingual(ns sql.NullString) (string, string) {
	if !ns.Valid || ns.String == "" {
		return "", ""
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(ns.String), &m); err != nil {
		return "", ""
	}
	zh, _ := m["zh"].(string)
	en, _ := m["en"].(string)
	return zh, en
}
