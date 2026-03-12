package sdk

import (
	"context"
	"errors"
	"testing"

	"blotting-consultancy/internal/eventbus"
	"blotting-consultancy/internal/plugin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestHookRegistrar(perms ...plugin.Permission) *HookRegistrar {
	return NewHookRegistrar("test-plugin", perms, eventbus.NewHookRegistry())
}

func TestHookRegistrar_Register_NoPerm(t *testing.T) {
	h := newTestHookRegistrar() // no perms
	err := h.Register(HookBeforePublish, func(_ context.Context, d interface{}) (interface{}, error) {
		return d, nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hook registration denied")
}

func TestHookRegistrar_Register_WithPerm(t *testing.T) {
	registry := eventbus.NewHookRegistry()
	h := NewHookRegistrar("test-plugin", []plugin.Permission{plugin.PermHookRegister}, registry)

	err := h.Register(HookBeforePublish, func(_ context.Context, d interface{}) (interface{}, error) {
		return d, nil
	})
	require.NoError(t, err)

	// Verify hook was registered by executing it
	result, err := registry.Execute(context.Background(), HookBeforePublish, "data")
	require.NoError(t, err)
	assert.Equal(t, "data", result)
}

func TestHookRegistrar_HookTransformsData(t *testing.T) {
	registry := eventbus.NewHookRegistry()
	h := NewHookRegistrar("test-plugin", []plugin.Permission{plugin.PermHookRegister}, registry)

	err := h.Register(HookBeforeRender, func(_ context.Context, d interface{}) (interface{}, error) {
		s, ok := d.(string)
		if !ok {
			return d, nil
		}
		return s + "_transformed", nil
	})
	require.NoError(t, err)

	result, err := registry.Execute(context.Background(), HookBeforeRender, "content")
	require.NoError(t, err)
	assert.Equal(t, "content_transformed", result)
}

func TestHookRegistrar_HookReturnsError(t *testing.T) {
	registry := eventbus.NewHookRegistry()
	h := NewHookRegistrar("test-plugin", []plugin.Permission{plugin.PermHookRegister}, registry)

	require.NoError(t, h.Register(HookBeforeContentDelete, func(_ context.Context, d interface{}) (interface{}, error) {
		return nil, errors.New("deletion not allowed")
	}))

	_, err := registry.Execute(context.Background(), HookBeforeContentDelete, "item")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "deletion not allowed")
}

func TestHookRegistrar_RegisterBeforeContentCreate(t *testing.T) {
	registry := eventbus.NewHookRegistry()
	h := NewHookRegistrar("test-plugin", []plugin.Permission{plugin.PermHookRegister}, registry)

	var called bool
	err := h.RegisterBeforeContentCreate(func(_ context.Context, d interface{}) (interface{}, error) {
		called = true
		return d, nil
	})
	require.NoError(t, err)

	_, err = registry.Execute(context.Background(), HookBeforeContentCreate, nil)
	require.NoError(t, err)
	assert.True(t, called)
}

func TestHookRegistrar_RegisterAfterContentCreate(t *testing.T) {
	registry := eventbus.NewHookRegistry()
	h := NewHookRegistrar("test-plugin", []plugin.Permission{plugin.PermHookRegister}, registry)

	require.NoError(t, h.RegisterAfterContentCreate(func(_ context.Context, d interface{}) (interface{}, error) {
		return d, nil
	}))
}

func TestHookRegistrar_RegisterBeforePublish(t *testing.T) {
	registry := eventbus.NewHookRegistry()
	h := NewHookRegistrar("test-plugin", []plugin.Permission{plugin.PermHookRegister}, registry)

	require.NoError(t, h.RegisterBeforePublish(func(_ context.Context, d interface{}) (interface{}, error) {
		return d, nil
	}))
}

func TestHookRegistrar_RegisterAfterPublish(t *testing.T) {
	registry := eventbus.NewHookRegistry()
	h := NewHookRegistrar("test-plugin", []plugin.Permission{plugin.PermHookRegister}, registry)

	require.NoError(t, h.RegisterAfterPublish(func(_ context.Context, d interface{}) (interface{}, error) {
		return d, nil
	}))
}

func TestHookRegistrar_RegisterBeforeRender(t *testing.T) {
	registry := eventbus.NewHookRegistry()
	h := NewHookRegistrar("test-plugin", []plugin.Permission{plugin.PermHookRegister}, registry)

	require.NoError(t, h.RegisterBeforeRender(func(_ context.Context, d interface{}) (interface{}, error) {
		return d, nil
	}))
}

func TestHookRegistrar_RegisterAfterMediaUpload(t *testing.T) {
	registry := eventbus.NewHookRegistry()
	h := NewHookRegistrar("test-plugin", []plugin.Permission{plugin.PermHookRegister}, registry)

	require.NoError(t, h.RegisterAfterMediaUpload(func(_ context.Context, d interface{}) (interface{}, error) {
		return d, nil
	}))
}

func TestHookRegistrar_RegisterBeforeMediaDelete(t *testing.T) {
	registry := eventbus.NewHookRegistry()
	h := NewHookRegistrar("test-plugin", []plugin.Permission{plugin.PermHookRegister}, registry)

	require.NoError(t, h.RegisterBeforeMediaDelete(func(_ context.Context, d interface{}) (interface{}, error) {
		return d, nil
	}))
}

func TestAllHookPoints(t *testing.T) {
	points := AllHookPoints()
	assert.NotEmpty(t, points)

	// Verify well-known constants are present
	assert.Contains(t, points, HookBeforePublish)
	assert.Contains(t, points, HookAfterPublish)
	assert.Contains(t, points, HookBeforeRender)
	assert.Contains(t, points, HookAfterMediaUpload)
	assert.Contains(t, points, HookAfterSettingsSave)
}

func TestPlugin_RegisterHook_ExecutesViaRegistry(t *testing.T) {
	meta := newTestMeta("test-plugin", plugin.PermHookRegister)
	bus := eventbus.New()
	hooks := eventbus.NewHookRegistry()
	p, err := NewPlugin(meta, nil, bus, hooks, nil)
	require.NoError(t, err)

	var invoked bool
	require.NoError(t, p.RegisterHook(HookBeforePublish, func(_ context.Context, d interface{}) (interface{}, error) {
		invoked = true
		return d, nil
	}))

	_, err = hooks.Execute(context.Background(), HookBeforePublish, nil)
	require.NoError(t, err)
	assert.True(t, invoked)
}
