package setup

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"blotting-consultancy/internal/db"
	install "blotting-consultancy/internal/setup"
)

// Handler serves first-run web setup endpoints.
type Handler struct {
	svc          *install.Service
	databaseType string
}

// NewHandler creates a setup handler.
func NewHandler(svc *install.Service, dbDSN string) *Handler {
	dbType := "sqlite"
	if db.IsPostgresDSN(dbDSN) {
		dbType = "postgres"
	}
	return &Handler{svc: svc, databaseType: dbType}
}

// RegisterRoutes mounts public setup routes on the root router.
func (h *Handler) RegisterRoutes(router *gin.Engine, rateLimit gin.HandlerFunc) {
	group := router.Group("/setup")
	group.Use(rateLimit)
	{
		group.GET("/status", h.GetStatus)
		group.POST("/complete", h.Complete)
	}
}

// GetStatus returns whether the instance has completed setup.
func (h *Handler) GetStatus(c *gin.Context) {
	status, err := h.svc.GetStatus(c.Request.Context(), h.databaseType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "failed to read setup status"}})
		return
	}
	c.JSON(http.StatusOK, status)
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
