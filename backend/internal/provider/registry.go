package provider

import (
	"log"
	"reflect"
	"strings"
	"sync"
)

// Registry is a centralized store for Provider instances.
// Same-type providers follow last-registration-wins semantics.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]interface{}
}

// NewRegistry creates a new Provider Registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]interface{}),
	}
}

// Register adds or replaces a provider by name.
// If a provider with the same name already exists, it is replaced and a log entry is emitted.
func (r *Registry) Register(name string, provider interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if existing, ok := r.providers[name]; ok {
		log.Printf("[Registry] replacing provider %q: %T -> %T", name, existing, provider)
	}
	r.providers[name] = provider
	log.Printf("[Registry] registered provider %q (%T)", name, provider)
}

// Unregister removes a provider by name.
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.providers, name)
	log.Printf("[Registry] unregistered provider %q", name)
}

// ReplaceIf atomically replaces a provider only when the current value is the
// expected instance. Passing nil as replacement removes the provider.
func (r *Registry) ReplaceIf(name string, expected, replacement interface{}) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, ok := r.providers[name]
	if !ok || !sameProviderInstance(current, expected) {
		return false
	}
	if replacement == nil {
		delete(r.providers, name)
		log.Printf("[Registry] unregistered provider %q", name)
		return true
	}
	r.providers[name] = replacement
	log.Printf("[Registry] restored provider %q (%T)", name, replacement)
	return true
}

func sameProviderInstance(left, right interface{}) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	leftValue := reflect.ValueOf(left)
	rightValue := reflect.ValueOf(right)
	if leftValue.Type() != rightValue.Type() || !leftValue.Type().Comparable() {
		return false
	}
	return leftValue.Interface() == rightValue.Interface()
}

// Get retrieves a provider by name. Returns nil if not found.
func (r *Registry) Get(name string) interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.providers[name]
}

// List returns all registered provider names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// MustGet retrieves a provider by name and panics if not found.
// Use in startup code where missing providers are fatal.
func (r *Registry) MustGet(name string) interface{} {
	p := r.Get(name)
	if p == nil {
		panic("required provider not registered: " + name)
	}
	return p
}

// AI returns the registered AIProvider, or a noop provider if none is registered.
func (r *Registry) AI() AIProvider {
	p := r.Get("ai")
	if p == nil {
		return nil
	}
	if ai, ok := p.(AIProvider); ok {
		return ai
	}
	return nil
}

// SetAI registers (or replaces) the AIProvider.
func (r *Registry) SetAI(ai AIProvider) {
	if ai == nil {
		r.Unregister("ai")
		return
	}
	r.Register("ai", ai)
}

// Storage returns the active StorageProvider, or nil if none is registered.
func (r *Registry) Storage() StorageProvider {
	p := r.Get("storage")
	if storage, ok := p.(StorageProvider); ok {
		return storage
	}
	return nil
}

// SetStorage registers or clears the active StorageProvider.
func (r *Registry) SetStorage(storage StorageProvider) {
	if storage == nil {
		r.Unregister("storage")
		return
	}
	r.Register("storage", storage)
}

// StorageByName returns a retained storage provider by logical strategy name.
func (r *Registry) StorageByName(name string) (StorageProvider, bool) {
	p := r.Get("storage:" + strings.TrimSpace(name))
	storage, ok := p.(StorageProvider)
	return storage, ok
}

// SetStorageByName retains a storage provider for media created before a
// runtime strategy switch.
func (r *Registry) SetStorageByName(name string, storage StorageProvider) {
	key := "storage:" + strings.TrimSpace(name)
	if storage == nil {
		r.Unregister(key)
		return
	}
	r.Register(key, storage)
}

// Notifier returns the active NotifierProvider, or nil if none is registered.
func (r *Registry) Notifier() NotifierProvider {
	p := r.Get("notifier")
	if notifier, ok := p.(NotifierProvider); ok {
		return notifier
	}
	return nil
}
