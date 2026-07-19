package sdk

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/yixian-huang/inkless/backend/internal/plugin"
)

// RouteHandler is the HTTP handler signature used by plugin routes.
// Plugins receive a plain http.Handler instead of a Gin-specific context
// to keep the SDK decoupled from the web framework.
type RouteHandler = http.HandlerFunc

// Route describes a single HTTP route registered by a plugin.
type Route struct {
	Method  string       // HTTP method (GET, POST, PUT, PATCH, DELETE)
	Path    string       // path relative to the plugin's base, e.g. "/items/:id"
	Handler RouteHandler // handler function
}

// Router manages route registration for a single plugin.
// All routes are automatically prefixed with /api/plugins/<pluginID>/.
type Router struct {
	pluginID string
	sandbox  *plugin.Sandbox

	mu     sync.RWMutex
	routes []Route
}

// newRouter creates a new Router for the given plugin.
func newRouter(pluginID string, sandbox *plugin.Sandbox) *Router {
	return &Router{pluginID: pluginID, sandbox: sandbox}
}

// basePath returns the base URL prefix for this plugin's routes.
func (r *Router) basePath() string {
	return "/api/plugins/" + r.pluginID
}

// register adds a route after checking permissions and normalising the path.
func (r *Router) register(method, path string, handler RouteHandler) error {
	if err := r.sandbox.Check(plugin.PermRouteRegister); err != nil {
		return fmt.Errorf("route registration denied: %w", err)
	}
	if handler == nil {
		return fmt.Errorf("route handler must not be nil")
	}
	method = strings.ToUpper(strings.TrimSpace(method))
	if !isValidHTTPMethod(method) {
		return fmt.Errorf("invalid HTTP method %q", method)
	}
	// Normalise leading slash
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	r.mu.Lock()
	r.routes = append(r.routes, Route{Method: method, Path: path, Handler: handler})
	r.mu.Unlock()
	return nil
}

// GET registers a GET route at the given relative path.
func (r *Router) GET(path string, handler RouteHandler) error {
	return r.register(http.MethodGet, path, handler)
}

// POST registers a POST route at the given relative path.
func (r *Router) POST(path string, handler RouteHandler) error {
	return r.register(http.MethodPost, path, handler)
}

// PUT registers a PUT route at the given relative path.
func (r *Router) PUT(path string, handler RouteHandler) error {
	return r.register(http.MethodPut, path, handler)
}

// PATCH registers a PATCH route at the given relative path.
func (r *Router) PATCH(path string, handler RouteHandler) error {
	return r.register(http.MethodPatch, path, handler)
}

// DELETE registers a DELETE route at the given relative path.
func (r *Router) DELETE(path string, handler RouteHandler) error {
	return r.register(http.MethodDelete, path, handler)
}

// Handle registers a route with an arbitrary HTTP method.
func (r *Router) Handle(method, path string, handler RouteHandler) error {
	return r.register(method, path, handler)
}

// Routes returns a snapshot of all registered routes with their full absolute paths.
// The returned slice is a copy; mutations do not affect the router state.
func (r *Router) Routes() []Route {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Route, len(r.routes))
	base := r.basePath()
	for i, rt := range r.routes {
		out[i] = Route{
			Method:  rt.Method,
			Path:    base + rt.Path,
			Handler: rt.Handler,
		}
	}
	return out
}

// RelativeRoutes returns registered routes with their original relative paths,
// without the /api/plugins/<id> prefix.
func (r *Router) RelativeRoutes() []Route {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Route, len(r.routes))
	copy(out, r.routes)
	return out
}

// Len returns the number of registered routes.
func (r *Router) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.routes)
}

// BasePath returns the base URL prefix for this plugin (/api/plugins/<pluginID>).
func (r *Router) BasePath() string {
	return r.basePath()
}

// ServeHTTP implements http.Handler, routing requests to the correct plugin handler.
// The request path must be relative to the plugin base path.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, rt := range r.routes {
		if rt.Method == req.Method && pathMatches(rt.Path, req.URL.Path) {
			rt.Handler(w, req)
			return
		}
	}
	http.NotFound(w, req)
}

// isValidHTTPMethod reports whether m is a known HTTP method.
func isValidHTTPMethod(m string) bool {
	switch m {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodDelete, http.MethodHead, http.MethodOptions:
		return true
	}
	return false
}

// pathMatches performs a simple prefix/exact match.
// For simplicity this implementation uses exact string equality; real use-cases
// that need path parameters should wrap with a more capable mux.
func pathMatches(pattern, actual string) bool {
	return pattern == actual
}
