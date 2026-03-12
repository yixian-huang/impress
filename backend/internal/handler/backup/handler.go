package backup

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	backupService "blotting-consultancy/internal/backup"
)

// Handler handles backup-related HTTP requests
type Handler struct {
	service *backupService.Service
}

// NewHandler creates a new backup handler
func NewHandler(service *backupService.Service) *Handler {
	return &Handler{service: service}
}

// List returns all backup records.
// @Summary      List backups
// @Description  Returns all backup records
// @Tags         Backup
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} object{items=[]object}
// @Router       /admin/backups [get]
func (h *Handler) List(c *gin.Context) {
	records, err := h.service.ListBackups(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "查询备份记录失败"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": records,
	})
}

// Trigger manually triggers a database backup.
// @Summary      Trigger backup
// @Description  Manually trigger a database backup
// @Tags         Backup
// @Produce      json
// @Security     BearerAuth
// @Success      201 {object} object
// @Router       /admin/backups/trigger [post]
func (h *Handler) Trigger(c *gin.Context) {
	record, err := h.service.RunBackup(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "备份失败: " + err.Error()}})
		return
	}

	c.JSON(http.StatusCreated, record)
}

// Export generates a full site export ZIP archive.
// @Summary      Export site
// @Description  Generate a full site export as a ZIP archive
// @Tags         Backup
// @Produce      json
// @Security     BearerAuth
// @Success      201 {object} object
// @Router       /admin/backups/export [post]
func (h *Handler) Export(c *gin.Context) {
	record, err := h.service.RunExport()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "导出失败: " + err.Error()}})
		return
	}

	c.JSON(http.StatusCreated, record)
}

// DownloadExport serves an export ZIP file for download.
// @Summary      Download export
// @Description  Download a site export ZIP file
// @Tags         Backup
// @Produce      application/zip
// @Security     BearerAuth
// @Param        filename path string true "Export filename"
// @Success      200 {file} file
// @Failure      404 {object} object{error=string}
// @Router       /admin/backups/export/{filename} [get]
func (h *Handler) DownloadExport(c *gin.Context) {
	filename := c.Param("filename")

	// Path traversal check
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "非法文件名"}})
		return
	}

	if !strings.HasPrefix(filename, "site-export-") || !strings.HasSuffix(filename, ".zip") {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "非法文件名"}})
		return
	}

	filePath := filepath.Join(h.service.BackupDir(), filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "文件不存在"}})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.File(filePath)
}

// Import uploads a ZIP and restores the full site.
// @Summary      Import site
// @Description  Upload a ZIP archive and restore the full site
// @Tags         Backup
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file formData file true "Import ZIP file"
// @Success      200 {object} object{message=string}
// @Router       /admin/backups/import [post]
func (h *Handler) Import(c *gin.Context) {
	// Limit request body to 500MB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 500<<20)

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "请上传文件: " + err.Error()}})
		return
	}

	tmpPath, err := h.saveTempUpload(c, file, "site-import-*.zip")
	if err != nil {
		return
	}
	defer os.Remove(tmpPath)

	if err := h.service.RunImport(tmpPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "导入失败: " + err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "导入成功，请重新登录"})
}

// ValidateImport validates an import archive without applying it.
// @Summary      Validate import
// @Description  Upload a ZIP and validate its structure without importing
// @Tags         Backup
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file formData file true "Import ZIP file to validate"
// @Success      200 {object} object
// @Router       /admin/backups/import/validate [post]
func (h *Handler) ValidateImport(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 500<<20)

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "请上传文件: " + err.Error()}})
		return
	}

	tmpPath, err := h.saveTempUpload(c, file, "site-validate-*.zip")
	if err != nil {
		return
	}
	defer os.Remove(tmpPath)

	if err := c.SaveUploadedFile(file, tmpPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "保存文件失败: " + err.Error()}})
		return
	}

	result, err := backupService.ValidateExportArchive(tmpPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "验证失败: " + err.Error()}})
		return
	}

	c.JSON(http.StatusOK, result)
}

// saveTempUpload saves a multipart file to a temp file inside the backup directory
// and returns the path. On error it writes the JSON response and returns err != nil.
func (h *Handler) saveTempUpload(c *gin.Context, fh *multipart.FileHeader, pattern string) (string, error) {
	tmpDir := filepath.Join(h.service.BackupDir(), "tmp")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "创建临时目录失败"}})
		return "", err
	}

	tmpFile, err := os.CreateTemp(tmpDir, pattern)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "创建临时文件失败"}})
		return "", err
	}
	tmpPath := tmpFile.Name()

	src, err := fh.Open()
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "读取上传文件失败"}})
		return "", err
	}
	defer src.Close()

	if _, err := io.Copy(tmpFile, src); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "保存文件失败: " + err.Error()}})
		return "", err
	}
	tmpFile.Close()

	return tmpPath, nil
}
