package setup

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	install "github.com/yixian-huang/inkless/backend/internal/setup"
	"github.com/yixian-huang/inkless/backend/pkg/config"
)

// Handler serves first-run web setup endpoints.
type Handler struct {
	svc *install.Service
}

// NewHandler creates a setup handler.
func NewHandler(svc *install.Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts public setup routes on the root router.
func (h *Handler) RegisterRoutes(router *gin.Engine, rateLimit gin.HandlerFunc) {
	group := router.Group("/setup")
	group.Use(rateLimit)
	{
		group.GET("/status", h.GetStatus)
		group.POST("/test-database", h.TestDatabase)
		group.POST("/save-env", h.SaveEnv)
		group.POST("/complete", h.Complete)
	}
}

// GetStatus returns whether the instance has completed setup.
func (h *Handler) GetStatus(c *gin.Context) {
	status, err := h.svc.GetStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "failed to read setup status"}})
		return
	}
	c.JSON(http.StatusOK, status)
}

// TestDatabase checks database connectivity for wizard-provided settings.
func (h *Handler) TestDatabase(c *gin.Context) {
	allowed, err := h.svc.AllowsEnvConfiguration(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "failed to read setup status"}})
		return
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": gin.H{"message": "database configuration is not required"}})
		return
	}

	var in config.DatabaseInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid request body"}})
		return
	}
	if err := install.TestDatabase(c.Request.Context(), in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// SaveEnv persists .env during bootstrap setup.
func (h *Handler) SaveEnv(c *gin.Context) {
	allowed, err := h.svc.AllowsEnvConfiguration(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "failed to read setup status"}})
		return
	}

	var in install.BootstrapInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid request body"}})
		return
	}
	if in.Port == 0 {
		in.Port = h.svc.ServerPort()
	}

	result, err := install.SaveEnv(allowed, install.WorkingDirectory(), in)
	if err != nil {
		if errors.Is(err, install.ErrEnvAlreadyConfigured) {
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{"message": "environment file already exists", "code": "ENV_ALREADY_CONFIGURED"}})
			return
		}
		if errors.Is(err, install.ErrEnvConfigNotAllowed) {
			c.JSON(http.StatusForbidden, gin.H{"error": gin.H{"message": "environment configuration is not required"}})
			return
		}
		if errors.Is(err, install.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": strings.TrimPrefix(err.Error(), "invalid setup input: ")}})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Complete finishes first-run installation.
func (h *Handler) Complete(c *gin.Context) {
	var in install.CompleteInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid request body"}})
		return
	}

	if err := h.svc.Complete(c.Request.Context(), in); err != nil {
		if errors.Is(err, install.ErrAlreadyCompleted) {
			c.JSON(http.StatusForbidden, gin.H{"error": gin.H{"message": "setup already completed", "code": "SETUP_ALREADY_COMPLETED"}})
			return
		}
		if errors.Is(err, install.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": strings.TrimPrefix(err.Error(), "invalid setup input: ")}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "setup failed"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
