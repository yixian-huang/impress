package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yixian-huang/inkless/backend/pkg/apierror"

	"github.com/yixian-huang/inkless/backend/internal/repository"
	"github.com/yixian-huang/inkless/backend/internal/service"
)

// Handler handles storage configuration HTTP requests
type Handler struct {
	runtime *service.StorageRuntimeService
}

// NewHandler creates a new storage handler
func NewHandler(repo repository.StorageConfigRepository) *Handler {
	runtime := service.ConfigureDefaultStorageRuntime(repo, "")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := runtime.RestoreStartupConfig(ctx); err != nil {
		log.Printf("[storage] restore startup config: %v", err)
	}
	return NewHandlerWithRuntime(runtime)
}

// NewHandlerWithRuntime creates a storage handler with an explicit storage runtime.
func NewHandlerWithRuntime(runtime *service.StorageRuntimeService) *Handler {
	if runtime == nil {
		runtime = service.DefaultStorageRuntime()
	}
	return &Handler{runtime: runtime}
}

// GetConfig returns the current storage configuration
// GET /admin/storage/config
func (h *Handler) GetConfig(c *gin.Context) {
	config, err := h.runtime.GetConfig(c.Request.Context())
	if err != nil {
		apierror.Message(c, http.StatusInternalServerError, "获取存储配置失败")
		return
	}

	c.JSON(http.StatusOK, config)
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
	req, err := decodeUpdateConfigRequest(c)
	if err != nil {
		apierror.Message(c, http.StatusBadRequest, "无效的请求数据")
		return
	}

	config, err := h.runtime.UpdateConfig(c.Request.Context(), service.StorageConfigRequest{
		Strategy:  req.Strategy,
		Bucket:    req.Bucket,
		Region:    req.Region,
		Endpoint:  req.Endpoint,
		AccessKey: req.AccessKey,
		SecretKey: req.SecretKey,
		BasePath:  req.BasePath,
	})
	if err != nil {
		apierror.Message(c, http.StatusBadRequest, err.Error())
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
	if err := h.runtime.TestConnection(c.Request.Context()); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "存储连接正常",
	})
}

func decodeUpdateConfigRequest(c *gin.Context) (*UpdateConfigRequest, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	allowed := map[string]struct{}{
		"strategy":  {},
		"bucket":    {},
		"region":    {},
		"endpoint":  {},
		"accessKey": {},
		"secretKey": {},
		"basePath":  {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return nil, fmt.Errorf("unknown field %q", key)
		}
	}
	if _, ok := raw["strategy"]; !ok {
		return nil, errors.New("strategy is required")
	}

	var req UpdateConfigRequest
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Strategy) == "" {
		return nil, errors.New("strategy is required")
	}
	return &req, nil
}
