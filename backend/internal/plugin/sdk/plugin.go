// Package sdk provides the Plugin SDK for Inkless CMS plugins.
// It wraps context, configuration, event subscriptions, hook registrations,
// and database/route access into a single cohesive API surface.
package sdk

import (
	"context"
	"fmt"
	"sync"

	"github.com/yixian-huang/inkless/backend/internal/eventbus"
	"github.com/yixian-huang/inkless/backend/internal/plugin"

	"gorm.io/gorm"
)

// Plugin is the main SDK entry point given to each plugin at initialisation.
// It exposes scoped sub-systems (DB, Routes, Hooks, Events) while carrying the
// plugin's identity and current configuration.
type Plugin struct {
	id       string
	meta     *plugin.PluginMeta
	config   map[string]string
	sandbox  *plugin.Sandbox
	bus      eventbus.EventBus
	hooks    *eventbus.HookRegistry
	db       *gorm.DB
	router   *Router
	settings *SettingsManager

	mu   sync.RWMutex
	subs []subEntry // subscriptions to unregister on shutdown
}

// subEntry tracks an event subscription so it can be cleaned up.
type subEntry struct {
	eventType string
	id        uint64
}

// NewPlugin constructs a Plugin SDK instance for the given plugin metadata.
// db may be nil if the plugin does not declare database permissions.
func NewPlugin(meta *plugin.PluginMeta, config map[string]string, bus eventbus.EventBus, hooks *eventbus.HookRegistry, db *gorm.DB) (*Plugin, error) {
	if meta == nil {
		return nil, fmt.Errorf("plugin meta must not be nil")
	}

	sandbox := plugin.NewSandbox(meta.ID, meta.Permissions)

	p := &Plugin{
		id:      meta.ID,
		meta:    meta,
		config:  config,
		sandbox: sandbox,
		bus:     bus,
		hooks:   hooks,
		db:      db,
	}

	p.router = newRouter(meta.ID, sandbox)
	p.settings = newSettingsManager(meta.SettingsSchema, config)

	return p, nil
}

// ID returns the plugin's unique identifier.
func (p *Plugin) ID() string { return p.id }

// Meta returns the plugin's manifest metadata.
func (p *Plugin) Meta() *plugin.PluginMeta { return p.meta }

// Config returns the current configuration map passed at initialisation.
func (p *Plugin) Config() map[string]string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make(map[string]string, len(p.config))
	for k, v := range p.config {
		out[k] = v
	}
	return out
}

// ConfigValue returns a single configuration value by key, plus an ok flag.
func (p *Plugin) ConfigValue(key string) (string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	v, ok := p.config[key]
	return v, ok
}

// DB returns a plugin-scoped database accessor that automatically prefixes
// table names with "plg_<pluginID>_".
// Returns nil if the plugin has no database permissions.
func (p *Plugin) DB(ctx context.Context) (*PluginDB, error) {
	if err := p.sandbox.Check(plugin.PermDatabaseRead); err != nil {
		return nil, fmt.Errorf("plugin %q: database access denied: %w", p.id, err)
	}
	if p.db == nil {
		return nil, fmt.Errorf("plugin %q: no database connection available", p.id)
	}
	return newPluginDB(p.id, p.db.WithContext(ctx), p.sandbox), nil
}

// Routes returns the plugin's route registration helper.
func (p *Plugin) Routes() *Router {
	return p.router
}

// Settings returns the plugin's settings manager.
func (p *Plugin) Settings() *SettingsManager {
	return p.settings
}

// Subscribe registers an async event handler for the given event type.
// The subscription is tracked for automatic cleanup on Shutdown.
// Returns an error if the plugin lacks PermEventSubscribe.
func (p *Plugin) Subscribe(eventType string, fn eventbus.Handler) (uint64, error) {
	if err := p.sandbox.Check(plugin.PermEventSubscribe); err != nil {
		return 0, err
	}
	id := p.bus.Subscribe(eventType, eventbus.AsyncHandler(fn))
	p.mu.Lock()
	p.subs = append(p.subs, subEntry{eventType: eventType, id: id})
	p.mu.Unlock()
	return id, nil
}

// Publish publishes a domain event on the bus.
// Returns an error if the plugin lacks PermEventPublish.
func (p *Plugin) Publish(ctx context.Context, eventType string, payload interface{}) error {
	if err := p.sandbox.Check(plugin.PermEventPublish); err != nil {
		return err
	}
	p.bus.Publish(eventbus.Event{Type: eventType, Payload: payload})
	return nil
}

// RegisterHook attaches a hook function to a named hook point.
// Returns an error if the plugin lacks PermHookRegister.
func (p *Plugin) RegisterHook(hookPoint string, fn eventbus.HookFunc) error {
	if err := p.sandbox.Check(plugin.PermHookRegister); err != nil {
		return err
	}
	// Name the hook by plugin ID for traceability in logs.
	name := fmt.Sprintf("%s/%s", p.id, hookPoint)
	p.hooks.Register(hookPoint, name, fn)
	return nil
}

// Shutdown unregisters all event subscriptions. Call during plugin teardown.
func (p *Plugin) Shutdown() {
	p.mu.Lock()
	subs := p.subs
	p.subs = nil
	p.mu.Unlock()

	for _, s := range subs {
		p.bus.Unsubscribe(s.eventType, s.id)
	}
}
