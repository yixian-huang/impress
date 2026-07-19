package sdk

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/yixian-huang/inkless/backend/internal/eventbus"
	"github.com/yixian-huang/inkless/backend/internal/plugin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestMeta(id string, perms ...plugin.Permission) *plugin.PluginMeta {
	return &plugin.PluginMeta{
		ID:          id,
		Name:        "Test Plugin",
		Version:     "1.0.0",
		Permissions: perms,
	}
}

func newTestPlugin(t *testing.T, perms ...plugin.Permission) *Plugin {
	t.Helper()
	meta := newTestMeta("test-plugin", perms...)
	bus := eventbus.New()
	hooks := eventbus.NewHookRegistry()
	p, err := NewPlugin(meta, map[string]string{"key": "value"}, bus, hooks, nil)
	require.NoError(t, err)
	return p
}

func TestNewPlugin_NilMeta(t *testing.T) {
	_, err := NewPlugin(nil, nil, eventbus.New(), eventbus.NewHookRegistry(), nil)
	assert.Error(t, err)
}

func TestPlugin_ID(t *testing.T) {
	p := newTestPlugin(t)
	assert.Equal(t, "test-plugin", p.ID())
}

func TestPlugin_Config(t *testing.T) {
	p := newTestPlugin(t)
	cfg := p.Config()
	assert.Equal(t, "value", cfg["key"])

	v, ok := p.ConfigValue("key")
	assert.True(t, ok)
	assert.Equal(t, "value", v)

	_, ok = p.ConfigValue("missing")
	assert.False(t, ok)
}

func TestPlugin_DB_NoPerm(t *testing.T) {
	p := newTestPlugin(t) // no database perms
	_, err := p.DB(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database access denied")
}

func TestPlugin_DB_WithPerm_NoConn(t *testing.T) {
	p := newTestPlugin(t, plugin.PermDatabaseRead)
	_, err := p.DB(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no database connection")
}

func TestPlugin_Subscribe_NoPerm(t *testing.T) {
	p := newTestPlugin(t)
	_, err := p.Subscribe("content.created", func(eventbus.Event) {})
	assert.Error(t, err)
}

func TestPlugin_Subscribe_WithPerm(t *testing.T) {
	p := newTestPlugin(t, plugin.PermEventSubscribe)
	var called atomic.Int32
	id, err := p.Subscribe("content.created", func(eventbus.Event) {
		called.Add(1)
	})
	require.NoError(t, err)
	assert.Greater(t, id, uint64(0))
}

func TestPlugin_Publish_NoPerm(t *testing.T) {
	p := newTestPlugin(t)
	err := p.Publish(context.Background(), "content.created", nil)
	assert.Error(t, err)
}

func TestPlugin_Publish_WithPerm(t *testing.T) {
	p := newTestPlugin(t, plugin.PermEventPublish, plugin.PermEventSubscribe)

	var called atomic.Int32
	_, err := p.Subscribe("test.event", func(eventbus.Event) {
		called.Add(1)
	})
	require.NoError(t, err)

	err = p.Publish(context.Background(), "test.event", "payload")
	require.NoError(t, err)
}

func TestPlugin_RegisterHook_NoPerm(t *testing.T) {
	p := newTestPlugin(t)
	err := p.RegisterHook(HookBeforePublish, func(ctx context.Context, data interface{}) (interface{}, error) {
		return data, nil
	})
	assert.Error(t, err)
}

func TestPlugin_RegisterHook_WithPerm(t *testing.T) {
	p := newTestPlugin(t, plugin.PermHookRegister)
	err := p.RegisterHook(HookBeforePublish, func(ctx context.Context, data interface{}) (interface{}, error) {
		return data, nil
	})
	assert.NoError(t, err)
}

func TestPlugin_Shutdown_UnregistersSubscriptions(t *testing.T) {
	bus := eventbus.New()
	meta := newTestMeta("test-plugin", plugin.PermEventSubscribe, plugin.PermEventPublish)
	p, err := NewPlugin(meta, nil, bus, eventbus.NewHookRegistry(), nil)
	require.NoError(t, err)

	var count atomic.Int32
	_, err = p.Subscribe("evt", func(eventbus.Event) { count.Add(1) })
	require.NoError(t, err)

	// Fire once before shutdown
	bus.Publish(eventbus.Event{Type: "evt"})
	// Give async handler time to run
	// (handler is async; give a tiny window — in unit tests we just verify cleanup)

	p.Shutdown()

	// After shutdown subscriptions should be removed from the bus.
	// Verifying indirectly: no panic and subs slice is cleared.
	p.mu.RLock()
	assert.Empty(t, p.subs)
	p.mu.RUnlock()
}

func TestPlugin_Routes(t *testing.T) {
	p := newTestPlugin(t, plugin.PermRouteRegister)
	err := p.Routes().GET("/items", func(w http.ResponseWriter, r *http.Request) {})
	assert.NoError(t, err)
	assert.Equal(t, 1, p.Routes().Len())
}
