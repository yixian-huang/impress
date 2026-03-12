package sdk

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"blotting-consultancy/internal/plugin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRouter(t *testing.T, perms ...plugin.Permission) *Router {
	t.Helper()
	sandbox := plugin.NewSandbox("test-plugin", perms)
	return newRouter("test-plugin", sandbox)
}

func TestRouter_BasePath(t *testing.T) {
	r := newTestRouter(t, plugin.PermRouteRegister)
	assert.Equal(t, "/api/plugins/test-plugin", r.BasePath())
}

func TestRouter_RegisterMethods(t *testing.T) {
	r := newTestRouter(t, plugin.PermRouteRegister)
	noop := func(w http.ResponseWriter, req *http.Request) {}

	require.NoError(t, r.GET("/items", noop))
	require.NoError(t, r.POST("/items", noop))
	require.NoError(t, r.PUT("/items/1", noop))
	require.NoError(t, r.PATCH("/items/1", noop))
	require.NoError(t, r.DELETE("/items/1", noop))

	assert.Equal(t, 5, r.Len())
}

func TestRouter_Register_NoPerm(t *testing.T) {
	r := newTestRouter(t) // no permissions
	err := r.GET("/items", func(w http.ResponseWriter, req *http.Request) {})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "route registration denied")
}

func TestRouter_Register_NilHandler(t *testing.T) {
	r := newTestRouter(t, plugin.PermRouteRegister)
	err := r.GET("/items", nil)
	assert.Error(t, err)
}

func TestRouter_Register_InvalidMethod(t *testing.T) {
	r := newTestRouter(t, plugin.PermRouteRegister)
	err := r.Handle("BREW", "/coffee", func(w http.ResponseWriter, req *http.Request) {})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid HTTP method")
}

func TestRouter_Routes_AbsolutePaths(t *testing.T) {
	r := newTestRouter(t, plugin.PermRouteRegister)
	require.NoError(t, r.GET("/items", func(w http.ResponseWriter, req *http.Request) {}))

	routes := r.Routes()
	require.Len(t, routes, 1)
	assert.Equal(t, "GET", routes[0].Method)
	assert.Equal(t, "/api/plugins/test-plugin/items", routes[0].Path)
}

func TestRouter_RelativeRoutes(t *testing.T) {
	r := newTestRouter(t, plugin.PermRouteRegister)
	require.NoError(t, r.POST("/orders", func(w http.ResponseWriter, req *http.Request) {}))

	routes := r.RelativeRoutes()
	require.Len(t, routes, 1)
	assert.Equal(t, "/orders", routes[0].Path)
}

func TestRouter_LeadingSlashNormalized(t *testing.T) {
	r := newTestRouter(t, plugin.PermRouteRegister)
	require.NoError(t, r.GET("no-slash", func(w http.ResponseWriter, req *http.Request) {}))

	routes := r.RelativeRoutes()
	assert.Equal(t, "/no-slash", routes[0].Path)
}

func TestRouter_ServeHTTP_Match(t *testing.T) {
	r := newTestRouter(t, plugin.PermRouteRegister)
	require.NoError(t, r.GET("/ping", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "pong", rec.Body.String())
}

func TestRouter_ServeHTTP_NotFound(t *testing.T) {
	r := newTestRouter(t, plugin.PermRouteRegister)
	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestRouter_Handle_CustomMethod(t *testing.T) {
	r := newTestRouter(t, plugin.PermRouteRegister)
	require.NoError(t, r.Handle("OPTIONS", "/meta", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	assert.Equal(t, 1, r.Len())
}
