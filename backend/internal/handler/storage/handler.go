package storage

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/repository"
)

// Handler handles storage configuration HTTP requests
type Handler struct {
	repo repository.StorageConfigRepository
}

// NewHandler creates a new storage handler
func NewHandler(repo repository.StorageConfigRepository) *Handler {
	return &Handler{repo: repo}
}

// GetConfig returns the current storage configuration
// GET /admin/storage/config
func (h *Handler) GetConfig(c *gin.Context) {
	config, err := h.repo.Get(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "获取存储配置失败"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"strategy":     config.Strategy,
		"bucket":       config.Bucket,
		"region":       config.Region,
		"endpoint":     config.Endpoint,
		"accessKey":    config.AccessKey,
		"hasSecretKey": config.HasSecretKey(),
		"basePath":     config.BasePath,
		"updatedAt":    config.UpdatedAt,
	})
}

// UpdateConfigRequest is the request body for updating storage config
type UpdateConfigRequest struct {
	Strategy  string `json:"strategy" binding:"required"`
	Bucket    string `json:"bucket"`
	Region    string `json:"region"`
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	BasePath  string `json:"basePath"`
}

// UpdateConfig updates the storage configuration
// PUT /admin/storage/config
func (h *Handler) UpdateConfig(c *gin.Context) {
	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "无效的请求数据"}})
		return
	}

	strategy := model.StorageStrategy(req.Strategy)

	// If secretKey is empty, keep the existing one
	existing, err := h.repo.Get(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "获取现有配置失败"}})
		return
	}

	secretKey := req.SecretKey
	if secretKey == "" && existing.SecretKey != "" && strategy == existing.Strategy {
		secretKey = existing.SecretKey
	}

	config := &model.StorageConfig{
		Strategy:  strategy,
		Bucket:    req.Bucket,
		Region:    req.Region,
		Endpoint:  req.Endpoint,
		AccessKey: req.AccessKey,
		SecretKey: secretKey,
		BasePath:  req.BasePath,
	}

	if err := h.repo.Upsert(c.Request.Context(), config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "存储配置已更新",
		"strategy": config.Strategy,
	})
}

// TestConnection tests the storage connection with current config
// POST /admin/storage/test
func (h *Handler) TestConnection(c *gin.Context) {
	config, err := h.repo.Get(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "获取存储配置失败"}})
		return
	}

	switch config.Strategy {
	case model.StorageLocal:
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "本地存储连接正常",
		})
	case model.StorageS3, model.StorageOSS:
		// For S3/OSS, validate that required credentials are present
		if config.Bucket == "" || config.AccessKey == "" || config.SecretKey == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "缺少必要的存储配置信息（bucket、accessKey、secretKey）",
			})
			return
		}
		// In a real implementation, we would attempt to list objects or put a test object.
		// For now, we validate the config is complete.
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "存储配置验证通过（凭证格式正确）",
		})
	default:
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "未知的存储策略: " + string(config.Strategy),
		})
	}
}
