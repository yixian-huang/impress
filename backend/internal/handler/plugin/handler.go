package plugin

import (
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"blotting-consultancy/internal/model"
	pluginruntime "blotting-consultancy/internal/plugin"
	"blotting-consultancy/internal/provider"
)

const maxUploadSize = int64(101 << 20)

// Handler exposes the managed external plugin lifecycle.
type Handler struct {
	manager  *pluginruntime.Manager
	registry *provider.Registry
	enabled  bool
}

// NewHandler creates a plugin lifecycle handler.
func NewHandler(manager *pluginruntime.Manager, registry *provider.Registry, enabled bool) *Handler {
	return &Handler{manager: manager, registry: registry, enabled: enabled}
}

// List returns all installed external plugins and their persisted state.
func (h *Handler) List(c *gin.Context) {
	plugins, err := h.manager.ListPlugins(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	items := make([]gin.H, 0, len(plugins))
	for _, installed := range plugins {
		items = append(items, pluginResponse(installed))
	}
	c.JSON(http.StatusOK, gin.H{
		"plugins":                items,
		"externalPluginsEnabled": h.enabled,
	})
}

// Install accepts a zip package in multipart field "package".
func (h *Handler) Install(c *gin.Context) {
	if !h.requireEnabled(c) {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)
	source, _, err := c.Request.FormFile("package")
	if err != nil {
		writeError(c, http.StatusBadRequest, "multipart field \"package\" is required")
		return
	}
	defer source.Close()

	tempFile, err := os.CreateTemp("", "impress-plugin-*.zip")
	if err != nil {
		writeError(c, http.StatusInternalServerError, "failed to stage plugin package")
		return
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	written, copyErr := io.Copy(tempFile, io.LimitReader(source, maxUploadSize+1))
	closeErr := tempFile.Close()
	if copyErr != nil || closeErr != nil {
		writeError(c, http.StatusBadRequest, "failed to read plugin package")
		return
	}
	if written > maxUploadSize {
		writeError(c, http.StatusRequestEntityTooLarge, "plugin package is too large")
		return
	}

	meta, err := h.manager.InstallPackage(c.Request.Context(), tempPath)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "already installed") {
			status = http.StatusConflict
		}
		writeError(c, status, err.Error())
		return
	}
	installed, err := h.manager.GetPlugin(c.Request.Context(), meta.ID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, pluginResponse(*installed))
}

// Enable launches a plugin process and activates its providers.
func (h *Handler) Enable(c *gin.Context) {
	if !h.requireEnabled(c) {
		return
	}
	if err := h.manager.EnablePlugin(c.Request.Context(), c.Param("id")); err != nil {
		writeLifecycleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "plugin enabled"})
}

// Disable stops a plugin process and restores replaced providers.
func (h *Handler) Disable(c *gin.Context) {
	if !h.requireEnabled(c) {
		return
	}
	if err := h.manager.DisablePlugin(c.Request.Context(), c.Param("id")); err != nil {
		writeLifecycleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "plugin disabled"})
}

// Uninstall stops and permanently removes a managed plugin package and data.
func (h *Handler) Uninstall(c *gin.Context) {
	if !h.requireEnabled(c) {
		return
	}
	if err := h.manager.UninstallPlugin(c.Request.Context(), c.Param("id")); err != nil {
		writeLifecycleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "plugin uninstalled"})
}

// UpdateSettings persists settings and re-initializes a running plugin.
func (h *Handler) UpdateSettings(c *gin.Context) {
	if !h.requireEnabled(c) {
		return
	}
	var settings map[string]any
	if err := c.ShouldBindJSON(&settings); err != nil {
		writeError(c, http.StatusBadRequest, "invalid settings payload")
		return
	}
	if err := h.manager.UpdateSettings(c.Request.Context(), c.Param("id"), settings); err != nil {
		writeLifecycleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "plugin settings updated"})
}

// TestNotification proves the currently active notifier provider is callable.
func (h *Handler) TestNotification(c *gin.Context) {
	if !h.requireEnabled(c) {
		return
	}
	notifier := h.registry.Notifier()
	if notifier == nil {
		writeError(c, http.StatusServiceUnavailable, "no notifier provider is active")
		return
	}

	var event provider.NotifyEvent
	if err := c.ShouldBindJSON(&event); err != nil && err != io.EOF {
		writeError(c, http.StatusBadRequest, "invalid notification payload")
		return
	}
	if event.Type == "" {
		event.Type = "plugin.test"
	}
	if event.Subject == "" {
		event.Subject = "Impress plugin test"
	}
	if event.Body == "" {
		event.Body = "The active notifier provider handled this test event."
	}
	if err := notifier.Notify(c.Request.Context(), event); err != nil {
		writeError(c, http.StatusBadGateway, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) requireEnabled(c *gin.Context) bool {
	if h.enabled {
		return true
	}
	writeError(
		c,
		http.StatusServiceUnavailable,
		"external plugins are disabled; set ENABLE_EXTERNAL_PLUGINS=true to enable trusted plugin code",
	)
	return false
}

func pluginResponse(installed model.Plugin) gin.H {
	return gin.H{
		"id":          installed.ID,
		"pluginId":    installed.PluginID,
		"name":        installed.Name,
		"nameZh":      installed.NameZh,
		"version":     installed.Version,
		"description": installed.Description,
		"author":      installed.Author,
		"license":     installed.License,
		"homepage":    installed.Homepage,
		"state":       installed.State,
		"source":      installed.Source,
		"permissions": installed.Permissions,
		"hasSettings": len(installed.Settings) > 0,
		"errorMsg":    installed.ErrorMsg,
		"createdAt":   installed.CreatedAt,
		"updatedAt":   installed.UpdatedAt,
	}
}

func writeLifecycleError(c *gin.Context, err error) {
	status := http.StatusBadRequest
	if strings.Contains(err.Error(), "not found") {
		status = http.StatusNotFound
	}
	writeError(c, status, err.Error())
}

func writeError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": gin.H{"message": message}})
}
