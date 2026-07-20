package setup

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/yixian-huang/inkless/backend/pkg/apierror"

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
		apierror.Message(c, http.StatusInternalServerError, "failed to read setup status")
		return
	}
	c.JSON(http.StatusOK, status)
}

// TestDatabase checks database connectivity for wizard-provided settings.
func (h *Handler) TestDatabase(c *gin.Context) {
	allowed, err := h.svc.AllowsEnvConfiguration(c.Request.Context())
	if err != nil {
		apierror.Message(c, http.StatusInternalServerError, "failed to read setup status")
		return
	}
	if !allowed {
		apierror.Message(c, http.StatusForbidden, "database configuration is not required")
		return
	}

	var in config.DatabaseInput
	if err := c.ShouldBindJSON(&in); err != nil {
		apierror.Message(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := install.TestDatabase(c.Request.Context(), in); err != nil {
		apierror.Message(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// SaveEnv persists .env during bootstrap setup.
func (h *Handler) SaveEnv(c *gin.Context) {
	allowed, err := h.svc.AllowsEnvConfiguration(c.Request.Context())
	if err != nil {
		apierror.Message(c, http.StatusInternalServerError, "failed to read setup status")
		return
	}

	var in install.BootstrapInput
	if err := c.ShouldBindJSON(&in); err != nil {
		apierror.Message(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if in.Port == 0 {
		in.Port = h.svc.ServerPort()
	}

	result, err := install.SaveEnv(allowed, install.WorkingDirectory(), in)
	if err != nil {
		if errors.Is(err, install.ErrEnvAlreadyConfigured) {
			apierror.Write(c, apierror.Conflict("environment file already exists").WithDetails(map[string]any{
				"code": "ENV_ALREADY_CONFIGURED",
			}))
			return
		}
		if errors.Is(err, install.ErrEnvConfigNotAllowed) {
			apierror.Message(c, http.StatusForbidden, "environment configuration is not required")
			return
		}
		if errors.Is(err, install.ErrInvalidInput) {
			apierror.Message(c, http.StatusBadRequest, strings.TrimPrefix(err.Error(), "invalid setup input: "))
			return
		}
		apierror.Message(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, result)
}

// Complete finishes first-run installation.
func (h *Handler) Complete(c *gin.Context) {
	var in install.CompleteInput
	if err := c.ShouldBindJSON(&in); err != nil {
		apierror.Message(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.svc.Complete(c.Request.Context(), in); err != nil {
		if errors.Is(err, install.ErrAlreadyCompleted) {
			apierror.Write(c, apierror.Forbidden("setup already completed").WithDetails(map[string]any{
				"code": "SETUP_ALREADY_COMPLETED",
			}))
			return
		}
		if errors.Is(err, install.ErrInvalidInput) {
			apierror.Message(c, http.StatusBadRequest, strings.TrimPrefix(err.Error(), "invalid setup input: "))
			return
		}
		apierror.Message(c, http.StatusInternalServerError, "setup failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
