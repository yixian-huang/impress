package sdk

import (
	"context"
	"fmt"

	"github.com/yixian-huang/inkless/backend/internal/eventbus"
	"github.com/yixian-huang/inkless/backend/internal/plugin"
)

// Well-known hook point names that plugins can attach to.
// These mirror the constants in eventbus/hooks.go and extend them with
// additional plugin-specific hook points.
const (
	// Content lifecycle hooks — fired around article/page CRUD.
	HookBeforeContentCreate = eventbus.HookBeforeCreate
	HookAfterContentCreate  = eventbus.HookAfterCreate
	HookBeforeContentDelete = eventbus.HookBeforeDelete
	HookAfterContentDelete  = eventbus.HookAfterDelete
	HookBeforePublish       = eventbus.HookBeforePublish
	HookAfterPublish        = eventbus.HookAfterPublish
	HookBeforeRender        = eventbus.HookBeforeRender

	// Request/response hooks.
	HookBeforeAPIResponse = "hook.before_api_response"

	// Media hooks.
	HookAfterMediaUpload  = "hook.after_media_upload"
	HookBeforeMediaDelete = "hook.before_media_delete"

	// Settings hooks.
	HookAfterSettingsSave = "hook.after_settings_save"
)

// HookRegistrar provides a focused interface for attaching hooks from plugin
// initialisation code without requiring the full Plugin instance.
type HookRegistrar struct {
	pluginID string
	sandbox  *plugin.Sandbox
	registry *eventbus.HookRegistry
}

// NewHookRegistrar creates a standalone HookRegistrar backed by the given
// HookRegistry. This is useful for testing or for plugins that prefer to
// register hooks separately from the main Plugin lifecycle.
func NewHookRegistrar(pluginID string, perms []plugin.Permission, registry *eventbus.HookRegistry) *HookRegistrar {
	return &HookRegistrar{
		pluginID: pluginID,
		sandbox:  plugin.NewSandbox(pluginID, perms),
		registry: registry,
	}
}

// Register attaches fn to the named hook point.
// Returns an error if the plugin does not have PermHookRegister.
func (h *HookRegistrar) Register(hookPoint string, fn eventbus.HookFunc) error {
	if err := h.sandbox.Check(plugin.PermHookRegister); err != nil {
		return fmt.Errorf("hook registration denied: %w", err)
	}
	name := fmt.Sprintf("%s/%s", h.pluginID, hookPoint)
	h.registry.Register(hookPoint, name, fn)
	return nil
}

// RegisterBeforeContentCreate attaches fn to the before-content-create hook.
func (h *HookRegistrar) RegisterBeforeContentCreate(fn func(ctx context.Context, content interface{}) (interface{}, error)) error {
	return h.Register(HookBeforeContentCreate, eventbus.HookFunc(fn))
}

// RegisterAfterContentCreate attaches fn to the after-content-create hook.
func (h *HookRegistrar) RegisterAfterContentCreate(fn func(ctx context.Context, content interface{}) (interface{}, error)) error {
	return h.Register(HookAfterContentCreate, eventbus.HookFunc(fn))
}

// RegisterBeforeContentDelete attaches fn to the before-content-delete hook.
func (h *HookRegistrar) RegisterBeforeContentDelete(fn func(ctx context.Context, content interface{}) (interface{}, error)) error {
	return h.Register(HookBeforeContentDelete, eventbus.HookFunc(fn))
}

// RegisterAfterContentDelete attaches fn to the after-content-delete hook.
func (h *HookRegistrar) RegisterAfterContentDelete(fn func(ctx context.Context, content interface{}) (interface{}, error)) error {
	return h.Register(HookAfterContentDelete, eventbus.HookFunc(fn))
}

// RegisterBeforePublish attaches fn to the before-publish hook.
func (h *HookRegistrar) RegisterBeforePublish(fn func(ctx context.Context, content interface{}) (interface{}, error)) error {
	return h.Register(HookBeforePublish, eventbus.HookFunc(fn))
}

// RegisterAfterPublish attaches fn to the after-publish hook.
func (h *HookRegistrar) RegisterAfterPublish(fn func(ctx context.Context, content interface{}) (interface{}, error)) error {
	return h.Register(HookAfterPublish, eventbus.HookFunc(fn))
}

// RegisterBeforeRender attaches fn to the before-render hook.
func (h *HookRegistrar) RegisterBeforeRender(fn func(ctx context.Context, content interface{}) (interface{}, error)) error {
	return h.Register(HookBeforeRender, eventbus.HookFunc(fn))
}

// RegisterAfterMediaUpload attaches fn to the after-media-upload hook.
func (h *HookRegistrar) RegisterAfterMediaUpload(fn func(ctx context.Context, media interface{}) (interface{}, error)) error {
	return h.Register(HookAfterMediaUpload, eventbus.HookFunc(fn))
}

// RegisterBeforeMediaDelete attaches fn to the before-media-delete hook.
func (h *HookRegistrar) RegisterBeforeMediaDelete(fn func(ctx context.Context, media interface{}) (interface{}, error)) error {
	return h.Register(HookBeforeMediaDelete, eventbus.HookFunc(fn))
}

// AllHookPoints returns the full list of well-known hook point names.
// This is useful for documentation generation and settings UI.
func AllHookPoints() []string {
	return []string{
		HookBeforeContentCreate,
		HookAfterContentCreate,
		HookBeforeContentDelete,
		HookAfterContentDelete,
		HookBeforePublish,
		HookAfterPublish,
		HookBeforeRender,
		HookBeforeAPIResponse,
		HookAfterMediaUpload,
		HookBeforeMediaDelete,
		HookAfterSettingsSave,
	}
}
