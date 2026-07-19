package backup

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yixian-huang/inkless/backend/internal/model"

	"gorm.io/gorm"
)

// ExportManifest contains metadata about the export archive
type ExportManifest struct {
	Version    string                     `json:"version"`
	ExportedAt string                     `json:"exportedAt"`
	Tables     map[string]TableExportInfo `json:"tables"`
	MediaFiles int                        `json:"mediaFiles"`
}

// TableExportInfo describes exported table stats
type TableExportInfo struct {
	Count int `json:"count"`
}

// ArticleTagRow represents a row in the article_tags join table
type ArticleTagRow struct {
	ArticleID uint `json:"articleId"`
	TagID     uint `json:"tagId"`
}

// ValidationResult describes the result of validating an import archive
type ValidationResult struct {
	Valid      bool                       `json:"valid"`
	Version    string                     `json:"version"`
	ExportedAt string                     `json:"exportedAt"`
	Tables     map[string]TableExportInfo `json:"tables"`
	MediaFiles int                        `json:"mediaFiles"`
	Errors     []string                   `json:"errors,omitempty"`
}

// ExportRecord describes an export file on disk
type ExportRecord struct {
	Filename  string    `json:"filename"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"createdAt"`
}

// RunExport creates a full site export ZIP archive.
func (s *Service) RunExport() (*ExportRecord, error) {
	if err := os.MkdirAll(s.backupDir, 0755); err != nil {
		return nil, fmt.Errorf("create backup dir: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	tmpDir, err := os.MkdirTemp("", "site-export-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	dataDir := filepath.Join(tmpDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	manifest := ExportManifest{
		Version:    s.appVersion,
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Tables:     make(map[string]TableExportInfo),
	}

	// Export each table
	if err := s.exportUsers(dataDir, &manifest); err != nil {
		return nil, fmt.Errorf("export users: %w", err)
	}
	if err := s.exportContentDocuments(dataDir, &manifest); err != nil {
		return nil, fmt.Errorf("export content_documents: %w", err)
	}
	if err := s.exportContentVersions(dataDir, &manifest); err != nil {
		return nil, fmt.Errorf("export content_versions: %w", err)
	}
	if err := s.exportPages(dataDir, &manifest); err != nil {
		return nil, fmt.Errorf("export pages: %w", err)
	}
	if err := s.exportArticles(dataDir, &manifest); err != nil {
		return nil, fmt.Errorf("export articles: %w", err)
	}
	if err := s.exportArticleTags(dataDir, &manifest); err != nil {
		return nil, fmt.Errorf("export article_tags: %w", err)
	}
	if err := s.exportCategories(dataDir, &manifest); err != nil {
		return nil, fmt.Errorf("export categories: %w", err)
	}
	if err := s.exportTags(dataDir, &manifest); err != nil {
		return nil, fmt.Errorf("export tags: %w", err)
	}
	if err := s.exportMedia(dataDir, &manifest); err != nil {
		return nil, fmt.Errorf("export media: %w", err)
	}
	if err := s.exportAuditEvents(dataDir, &manifest); err != nil {
		return nil, fmt.Errorf("export audit_events: %w", err)
	}
	if err := s.exportPageViews(dataDir, &manifest); err != nil {
		return nil, fmt.Errorf("export page_views: %w", err)
	}
	if err := s.exportBackupRecords(dataDir, &manifest); err != nil {
		return nil, fmt.Errorf("export backup_records: %w", err)
	}

	// Copy uploads directory
	uploadsDir := filepath.Join(tmpDir, "uploads")
	mediaCount, err := copyDir(s.uploadDir, uploadsDir)
	if err != nil {
		// If upload dir doesn't exist, that's fine
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("copy uploads: %w", err)
		}
		mediaCount = 0
	}
	manifest.MediaFiles = mediaCount

	// Write manifest
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal manifest: %w", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "manifest.json"), manifestData, 0644); err != nil {
		return nil, fmt.Errorf("write manifest: %w", err)
	}

	// Create ZIP
	zipFilename := fmt.Sprintf("site-export-%s.zip", timestamp)
	zipPath := filepath.Join(s.backupDir, zipFilename)
	if err := createZipFromDir(tmpDir, zipPath); err != nil {
		return nil, fmt.Errorf("create zip: %w", err)
	}

	info, err := os.Stat(zipPath)
	if err != nil {
		return nil, fmt.Errorf("stat zip: %w", err)
	}

	return &ExportRecord{
		Filename:  zipFilename,
		Size:      info.Size(),
		CreatedAt: info.ModTime(),
	}, nil
}

// ValidateExportArchive opens a ZIP and checks its structure without importing.
func ValidateExportArchive(zipPath string) (*ValidationResult, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return &ValidationResult{Valid: false, Errors: []string{"无法打开 ZIP 文件: " + err.Error()}}, nil
	}
	defer r.Close()

	result := &ValidationResult{
		Tables: make(map[string]TableExportInfo),
	}

	hasManifest := false
	for _, f := range r.File {
		if !isSafePath(f.Name) {
			result.Errors = append(result.Errors, fmt.Sprintf("不安全的路径: %s", f.Name))
			continue
		}
		if f.Name == "manifest.json" {
			hasManifest = true
			rc, err := f.Open()
			if err != nil {
				result.Errors = append(result.Errors, "无法读取 manifest.json")
				continue
			}
			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				result.Errors = append(result.Errors, "读取 manifest.json 失败")
				continue
			}
			var manifest ExportManifest
			if err := json.Unmarshal(data, &manifest); err != nil {
				result.Errors = append(result.Errors, "manifest.json 格式错误")
				continue
			}
			result.Version = manifest.Version
			result.ExportedAt = manifest.ExportedAt
			result.Tables = manifest.Tables
			result.MediaFiles = manifest.MediaFiles
		}
	}

	if !hasManifest {
		result.Errors = append(result.Errors, "缺少 manifest.json")
	}

	// Check required data files
	requiredFiles := []string{
		"data/users.json",
		"data/content_documents.json",
		"data/content_versions.json",
		"data/pages.json",
		"data/articles.json",
		"data/article_tags.json",
		"data/categories.json",
		"data/tags.json",
		"data/media.json",
		"data/audit_events.json",
		"data/page_views.json",
		"data/backup_records.json",
	}

	fileSet := make(map[string]bool)
	for _, f := range r.File {
		fileSet[f.Name] = true
	}
	for _, req := range requiredFiles {
		if !fileSet[req] {
			result.Errors = append(result.Errors, fmt.Sprintf("缺少数据文件: %s", req))
		}
	}

	result.Valid = len(result.Errors) == 0
	return result, nil
}

// RunImport restores a full site from an export ZIP archive.
func (s *Service) RunImport(zipPath string) error {
	s.importMu.Lock()
	defer s.importMu.Unlock()

	// Validate first
	vr, err := ValidateExportArchive(zipPath)
	if err != nil {
		return fmt.Errorf("validate archive: %w", err)
	}
	if !vr.Valid {
		return fmt.Errorf("invalid archive: %s", strings.Join(vr.Errors, "; "))
	}

	// Extract to temp dir
	tmpDir, err := os.MkdirTemp("", "site-import-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := extractZip(zipPath, tmpDir); err != nil {
		return fmt.Errorf("extract zip: %w", err)
	}

	dataDir := filepath.Join(tmpDir, "data")

	// Run import in a transaction
	err = s.db.Transaction(func(tx *gorm.DB) error {
		isSQLite := tx.Dialector.Name() == "sqlite"

		// Disable foreign keys for SQLite
		if isSQLite {
			if err := tx.Exec("PRAGMA foreign_keys = OFF").Error; err != nil {
				return fmt.Errorf("disable foreign keys: %w", err)
			}
		}

		// Clear tables in reverse dependency order
		clearOrder := []string{
			"article_tags",
			"articles",
			"content_versions",
			"pages",
			"page_views",
			"audit_events",
			"backup_records",
			"media",
			"tags",
			"categories",
			"content_documents",
			"users",
		}
		for _, table := range clearOrder {
			if err := tx.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error; err != nil {
				return fmt.Errorf("clear table %s: %w", table, err)
			}
		}

		// Import tables in forward dependency order
		if err := importTable[model.User](tx, dataDir, "users.json"); err != nil {
			return fmt.Errorf("import users: %w", err)
		}
		if err := importTable[model.ContentDocument](tx, dataDir, "content_documents.json"); err != nil {
			return fmt.Errorf("import content_documents: %w", err)
		}
		if err := importTable[model.Category](tx, dataDir, "categories.json"); err != nil {
			return fmt.Errorf("import categories: %w", err)
		}
		if err := importTable[model.Tag](tx, dataDir, "tags.json"); err != nil {
			return fmt.Errorf("import tags: %w", err)
		}
		if err := importTable[model.Media](tx, dataDir, "media.json"); err != nil {
			return fmt.Errorf("import media: %w", err)
		}
		if err := importTable[model.BackupRecord](tx, dataDir, "backup_records.json"); err != nil {
			return fmt.Errorf("import backup_records: %w", err)
		}
		if err := importTable[model.AuditEvent](tx, dataDir, "audit_events.json"); err != nil {
			return fmt.Errorf("import audit_events: %w", err)
		}
		if err := importTable[model.PageView](tx, dataDir, "page_views.json"); err != nil {
			return fmt.Errorf("import page_views: %w", err)
		}
		if err := importTable[model.Page](tx, dataDir, "pages.json"); err != nil {
			return fmt.Errorf("import pages: %w", err)
		}
		if err := importTable[model.ContentVersion](tx, dataDir, "content_versions.json"); err != nil {
			return fmt.Errorf("import content_versions: %w", err)
		}
		if err := importTable[model.Article](tx, dataDir, "articles.json"); err != nil {
			return fmt.Errorf("import articles: %w", err)
		}
		if err := s.importArticleTags(tx, dataDir); err != nil {
			return fmt.Errorf("import article_tags: %w", err)
		}

		// Reset SQLite auto-increment sequences
		if isSQLite {
			tables := []string{
				"users", "content_versions", "media", "page_views",
				"categories", "tags", "articles", "backup_records",
				"audit_events", "pages",
			}
			for _, table := range tables {
				tx.Exec("DELETE FROM sqlite_sequence WHERE name=?", table)
			}
			if err := tx.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
				return fmt.Errorf("enable foreign keys: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("import transaction: %w", err)
	}

	// Replace uploads directory
	uploadsSource := filepath.Join(tmpDir, "uploads")
	if _, statErr := os.Stat(uploadsSource); statErr == nil {
		// Clear existing uploads
		if err := os.RemoveAll(s.uploadDir); err != nil {
			return fmt.Errorf("clear uploads dir: %w", err)
		}
		if _, err := copyDir(uploadsSource, s.uploadDir); err != nil {
			return fmt.Errorf("restore uploads: %w", err)
		}
	}

	return nil
}

// ListExports scans the backup directory for site-export-*.zip files.
func (s *Service) ListExports() ([]ExportRecord, error) {
	entries, err := os.ReadDir(s.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var exports []ExportRecord
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "site-export-") || !strings.HasSuffix(entry.Name(), ".zip") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		exports = append(exports, ExportRecord{
			Filename:  entry.Name(),
			Size:      info.Size(),
			CreatedAt: info.ModTime(),
		})
	}

	return exports, nil
}

// --- table export helpers ---

func (s *Service) exportUsers(dataDir string, m *ExportManifest) error {
	var rows []model.User
	if err := s.db.Find(&rows).Error; err != nil {
		return err
	}
	m.Tables["users"] = TableExportInfo{Count: len(rows)}
	return writeJSON(filepath.Join(dataDir, "users.json"), rows)
}

func (s *Service) exportContentDocuments(dataDir string, m *ExportManifest) error {
	var rows []model.ContentDocument
	if err := s.db.Find(&rows).Error; err != nil {
		return err
	}
	m.Tables["content_documents"] = TableExportInfo{Count: len(rows)}
	return writeJSON(filepath.Join(dataDir, "content_documents.json"), rows)
}

func (s *Service) exportContentVersions(dataDir string, m *ExportManifest) error {
	var rows []model.ContentVersion
	if err := s.db.Find(&rows).Error; err != nil {
		return err
	}
	m.Tables["content_versions"] = TableExportInfo{Count: len(rows)}
	return writeJSON(filepath.Join(dataDir, "content_versions.json"), rows)
}

func (s *Service) exportPages(dataDir string, m *ExportManifest) error {
	var rows []model.Page
	if err := s.db.Unscoped().Find(&rows).Error; err != nil {
		return err
	}
	m.Tables["pages"] = TableExportInfo{Count: len(rows)}
	return writeJSON(filepath.Join(dataDir, "pages.json"), rows)
}

func (s *Service) exportArticles(dataDir string, m *ExportManifest) error {
	var rows []model.Article
	if err := s.db.Find(&rows).Error; err != nil {
		return err
	}
	m.Tables["articles"] = TableExportInfo{Count: len(rows)}
	return writeJSON(filepath.Join(dataDir, "articles.json"), rows)
}

func (s *Service) exportArticleTags(dataDir string, m *ExportManifest) error {
	var rows []ArticleTagRow
	if err := s.db.Raw("SELECT article_id, tag_id FROM article_tags").Scan(&rows).Error; err != nil {
		return err
	}
	m.Tables["article_tags"] = TableExportInfo{Count: len(rows)}
	return writeJSON(filepath.Join(dataDir, "article_tags.json"), rows)
}

func (s *Service) exportCategories(dataDir string, m *ExportManifest) error {
	var rows []model.Category
	if err := s.db.Find(&rows).Error; err != nil {
		return err
	}
	m.Tables["categories"] = TableExportInfo{Count: len(rows)}
	return writeJSON(filepath.Join(dataDir, "categories.json"), rows)
}

func (s *Service) exportTags(dataDir string, m *ExportManifest) error {
	var rows []model.Tag
	if err := s.db.Find(&rows).Error; err != nil {
		return err
	}
	m.Tables["tags"] = TableExportInfo{Count: len(rows)}
	return writeJSON(filepath.Join(dataDir, "tags.json"), rows)
}

func (s *Service) exportMedia(dataDir string, m *ExportManifest) error {
	var rows []model.Media
	if err := s.db.Find(&rows).Error; err != nil {
		return err
	}
	m.Tables["media"] = TableExportInfo{Count: len(rows)}
	return writeJSON(filepath.Join(dataDir, "media.json"), rows)
}

func (s *Service) exportAuditEvents(dataDir string, m *ExportManifest) error {
	var rows []model.AuditEvent
	if err := s.db.Find(&rows).Error; err != nil {
		return err
	}
	m.Tables["audit_events"] = TableExportInfo{Count: len(rows)}
	return writeJSON(filepath.Join(dataDir, "audit_events.json"), rows)
}

func (s *Service) exportPageViews(dataDir string, m *ExportManifest) error {
	var rows []model.PageView
	if err := s.db.Find(&rows).Error; err != nil {
		return err
	}
	m.Tables["page_views"] = TableExportInfo{Count: len(rows)}
	return writeJSON(filepath.Join(dataDir, "page_views.json"), rows)
}

func (s *Service) exportBackupRecords(dataDir string, m *ExportManifest) error {
	var rows []model.BackupRecord
	if err := s.db.Find(&rows).Error; err != nil {
		return err
	}
	m.Tables["backup_records"] = TableExportInfo{Count: len(rows)}
	return writeJSON(filepath.Join(dataDir, "backup_records.json"), rows)
}

// --- import helpers ---

func importTable[T any](tx *gorm.DB, dataDir, filename string) error {
	data, err := os.ReadFile(filepath.Join(dataDir, filename))
	if err != nil {
		return err
	}
	var rows []T
	if err := json.Unmarshal(data, &rows); err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	// Batch create to avoid huge single inserts
	batchSize := 100
	for i := 0; i < len(rows); i += batchSize {
		end := i + batchSize
		if end > len(rows) {
			end = len(rows)
		}
		if err := tx.Create(rows[i:end]).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) importArticleTags(tx *gorm.DB, dataDir string) error {
	data, err := os.ReadFile(filepath.Join(dataDir, "article_tags.json"))
	if err != nil {
		return err
	}
	var rows []ArticleTagRow
	if err := json.Unmarshal(data, &rows); err != nil {
		return err
	}
	for _, row := range rows {
		if err := tx.Exec("INSERT INTO article_tags (article_id, tag_id) VALUES (?, ?)", row.ArticleID, row.TagID).Error; err != nil {
			return err
		}
	}
	return nil
}

// --- file helpers ---

func writeJSON(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// copyDir recursively copies src to dst and returns the number of files copied.
func copyDir(src, dst string) (int, error) {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return 0, err
	}
	if !srcInfo.IsDir() {
		return 0, fmt.Errorf("%s is not a directory", src)
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return 0, err
	}

	count := 0
	entries, err := os.ReadDir(src)
	if err != nil {
		return 0, err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			n, err := copyDir(srcPath, dstPath)
			if err != nil {
				return count, err
			}
			count += n
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return count, err
			}
			count++
		}
	}
	return count, nil
}

func createZipFromDir(srcDir, zipPath string) error {
	outFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	w := zip.NewWriter(outFile)
	defer w.Close()

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		// Normalize to forward slashes for ZIP
		relPath = filepath.ToSlash(relPath)

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = relPath
		header.Method = zip.Deflate

		writer, err := w.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
}

func extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if !isSafePath(f.Name) {
			return fmt.Errorf("unsafe path in zip: %s", f.Name)
		}

		fpath := filepath.Join(destDir, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}

		outFile, err := os.Create(fpath)
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// isSafePath checks that a zip entry path doesn't contain directory traversal.
func isSafePath(name string) bool {
	if strings.Contains(name, "..") {
		return false
	}
	if filepath.IsAbs(name) {
		return false
	}
	return true
}
