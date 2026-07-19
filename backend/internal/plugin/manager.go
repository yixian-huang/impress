package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/internal/provider"
	pb "github.com/yixian-huang/inkless/backend/pkg/pluginproto"
)

var errPluginNoLongerEnabled = errors.New("plugin is no longer enabled")

// ManagerConfig holds configuration for the plugin manager.
type ManagerConfig struct {
	PluginDir string // directory where plugins are installed (default: ./plugins)
	DataDir   string // directory for plugin data storage (default: ./data/plugins)
}

// Manager orchestrates plugin discovery, lifecycle, and provider registration.
type Manager struct {
	config         ManagerConfig
	store          *Store
	registry       *provider.Registry
	hosts          map[string]*GRPCHost // pluginID -> running host
	registrations  map[string][]providerRegistration
	providerOwners map[string]string
	mu             sync.RWMutex
	installMu      sync.Mutex
	settingsMu     sync.Mutex
	healthWG       sync.WaitGroup
	stopping       bool

	// healthStop signals the health monitor to stop
	healthStop chan struct{}
}

type providerRegistration struct {
	name     string
	previous interface{}
	current  interface{}
}

// NewManager creates a new plugin manager.
func NewManager(cfg ManagerConfig, store *Store, registry *provider.Registry) *Manager {
	if cfg.PluginDir == "" {
		cfg.PluginDir = "./plugins"
	}
	if cfg.DataDir == "" {
		cfg.DataDir = "./data/plugins"
	}
	manager := &Manager{
		config:         cfg,
		store:          store,
		registry:       registry,
		hosts:          make(map[string]*GRPCHost),
		registrations:  make(map[string][]providerRegistration),
		providerOwners: make(map[string]string),
		healthStop:     make(chan struct{}),
	}
	manager.reconcileStagedRemovals(context.Background())
	return manager
}

// DiscoverPlugins scans PluginDir for directories containing valid plugin.yaml files.
func (m *Manager) DiscoverPlugins(_ context.Context) ([]PluginMeta, error) {
	entries, err := os.ReadDir(m.config.PluginDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // no plugins directory is fine
		}
		return nil, fmt.Errorf("failed to read plugin directory %s: %w", m.config.PluginDir, err)
	}

	var discovered []PluginMeta
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(m.config.PluginDir, entry.Name())
		meta, err := LoadAndValidateManifest(dir)
		if err != nil {
			log.Printf("[PluginManager] skipping %s: %v", entry.Name(), err)
			continue
		}
		discovered = append(discovered, *meta)
	}
	return discovered, nil
}

// InstallPlugin installs a plugin from a directory path.
// It parses plugin.yaml, validates the manifest, checks for a binary, and
// creates a DB record with state=installed.
func (m *Manager) InstallPlugin(ctx context.Context, dir string) (*PluginMeta, error) {
	m.installMu.Lock()
	defer m.installMu.Unlock()

	meta, err := LoadAndValidateManifest(dir)
	if err != nil {
		return nil, fmt.Errorf("invalid plugin manifest: %w", err)
	}

	// Check permissions are sufficient for declared capabilities
	if err := ValidateManifestPermissions(meta); err != nil {
		return nil, err
	}
	if err := ValidateSupportedSettings(meta); err != nil {
		return nil, err
	}

	// Check if already installed
	exists, err := m.store.Exists(ctx, meta.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing plugin: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("plugin %q is already installed", meta.ID)
	}

	// Look for the plugin binary
	binaryPath := findBinary(dir, meta.ID)

	if err := m.createPluginRecord(ctx, meta, binaryPath, "local"); err != nil {
		return nil, fmt.Errorf("failed to store plugin record: %w", err)
	}

	log.Printf("[PluginManager] installed plugin %s v%s", meta.ID, meta.Version)
	return meta, nil
}

// EnablePlugin starts a plugin process and registers its providers.
func (m *Manager) EnablePlugin(ctx context.Context, pluginID string) error {
	return m.enablePlugin(ctx, pluginID, false)
}

func (m *Manager) enablePlugin(ctx context.Context, pluginID string, restoreEnabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.enablePluginLocked(ctx, pluginID, restoreEnabled)
}

func (m *Manager) enablePluginLocked(ctx context.Context, pluginID string, restoreEnabled bool) error {
	if m.stopping {
		return errors.New("plugin manager is stopping")
	}

	// Get plugin from DB
	p, err := m.store.GetByPluginID(ctx, pluginID)
	if err != nil {
		return err
	}

	// Validate state transition
	currentState := PluginState(p.State)
	if restoreEnabled {
		if currentState != StateEnabled {
			return fmt.Errorf("%w: %s is %s", errPluginNoLongerEnabled, pluginID, currentState)
		}
		if host, ok := m.hosts[pluginID]; ok && host.IsRunning() {
			return nil
		}
	} else {
		if err := Transition(currentState, StateEnabled); err != nil {
			return err
		}
	}

	// Check if binary exists
	if p.BinaryPath == "" {
		return fmt.Errorf("plugin %q has no binary path configured", pluginID)
	}

	// Load manifest from the plugin directory to get provider declarations
	pluginDir := filepath.Dir(p.BinaryPath)
	meta, err := LoadAndValidateManifest(pluginDir)
	if err != nil {
		_ = m.store.UpdateState(ctx, pluginID, StateFailed, err.Error())
		return fmt.Errorf("failed to load manifest for plugin %s: %w", pluginID, err)
	}

	// Validate permissions for declared providers
	if err := ValidateManifestPermissions(meta); err != nil {
		_ = m.store.UpdateState(ctx, pluginID, StateFailed, err.Error())
		return err
	}
	if err := ValidateSupportedSettings(meta); err != nil {
		_ = m.store.UpdateState(ctx, pluginID, StateFailed, err.Error())
		return err
	}
	if err := ValidateExternalRuntimeContract(meta); err != nil {
		_ = m.store.UpdateState(ctx, pluginID, StateFailed, err.Error())
		return err
	}
	for _, prov := range meta.Providers {
		if owner := m.providerOwners[prov.Name]; owner != "" && owner != pluginID {
			err := fmt.Errorf(
				"provider %q is already owned by enabled plugin %q",
				prov.Name,
				owner,
			)
			_ = m.store.UpdateState(ctx, pluginID, StateFailed, err.Error())
			return err
		}
	}

	settings := settingsToStringMap(p.Settings)

	// Start plugin process
	host := NewGRPCHost(meta, p.BinaryPath)
	dataDir := filepath.Join(m.config.DataDir, pluginID)
	if err := os.MkdirAll(dataDir, 0o750); err != nil {
		_ = m.store.UpdateState(ctx, pluginID, StateFailed, err.Error())
		return fmt.Errorf("failed to create data directory for plugin %s: %w", pluginID, err)
	}
	if err := host.Start(settings, dataDir); err != nil {
		_ = m.store.UpdateState(ctx, pluginID, StateFailed, err.Error())
		return fmt.Errorf("failed to start plugin %s: %w", pluginID, err)
	}

	// Register providers
	registrations := make([]providerRegistration, 0, len(meta.Providers))
	for _, prov := range meta.Providers {
		previous := m.registry.Get(prov.Name)
		var current interface{}
		switch prov.Type {
		case "storage":
			current = host.AsStorageProvider()
		case "search":
			current = host.AsSearchProvider()
		case "notifier":
			current = host.AsNotifierProvider()
		case "captcha":
			current = host.AsCaptchaProvider()
		}
		if current != nil {
			m.registry.Register(prov.Name, current)
			m.providerOwners[prov.Name] = pluginID
			registrations = append(registrations, providerRegistration{
				name:     prov.Name,
				previous: previous,
				current:  current,
			})
		}
	}

	m.hosts[pluginID] = host
	m.registrations[pluginID] = registrations

	// Update DB state
	if err := m.store.UpdateState(ctx, pluginID, StateEnabled, ""); err != nil {
		m.restoreRegistrationsLocked(pluginID)
		delete(m.hosts, pluginID)
		_ = host.Stop()
		return err
	}

	log.Printf("[PluginManager] enabled plugin %s", pluginID)
	return nil
}

// DisablePlugin stops a plugin process and unregisters its providers.
func (m *Manager) DisablePlugin(ctx context.Context, pluginID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.disablePluginLocked(ctx, pluginID)
}

func (m *Manager) disablePluginLocked(ctx context.Context, pluginID string) error {
	// Get plugin from DB
	p, err := m.store.GetByPluginID(ctx, pluginID)
	if err != nil {
		return err
	}

	// Validate state transition
	currentState := PluginState(p.State)
	if err := Transition(currentState, StateDisabled); err != nil {
		return err
	}

	// Persist the desired state before mutating the live runtime. If the write
	// fails, the enabled process and provider remain untouched.
	if err := m.store.UpdateState(ctx, pluginID, StateDisabled, ""); err != nil {
		return err
	}

	m.stopPluginLocked(pluginID)

	log.Printf("[PluginManager] disabled plugin %s", pluginID)
	return nil
}

func (m *Manager) stopPluginLocked(pluginID string) {
	if host, ok := m.hosts[pluginID]; ok {
		m.restoreRegistrationsLocked(pluginID)
		if err := host.Stop(); err != nil {
			log.Printf("[PluginManager] warning: error stopping plugin %s: %v", pluginID, err)
		}
		delete(m.hosts, pluginID)
	} else {
		m.restoreRegistrationsLocked(pluginID)
	}
}

// UninstallPlugin stops (if running) and removes a plugin from the system.
func (m *Manager) UninstallPlugin(ctx context.Context, pluginID string) error {
	m.installMu.Lock()
	defer m.installMu.Unlock()
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get plugin from DB
	p, err := m.store.GetByPluginID(ctx, pluginID)
	if err != nil {
		return err
	}

	currentState := PluginState(p.State)
	wasEnabled := currentState == StateEnabled

	// Keep the persisted enabled state until the delete commits. If the server
	// crashes during staging, startup reconciliation restores the package and
	// StartEnabledPlugins restarts it from the DB source of truth.
	if wasEnabled {
		m.stopPluginLocked(pluginID)
	}

	// Check if uninstall is allowed
	if currentState != StateEnabled && !CanUninstall(currentState) {
		return fmt.Errorf("cannot uninstall plugin in state %s", currentState)
	}

	var stagedPaths []stagedRemoval
	if p.Source == "package" {
		suffix := fmt.Sprintf(".uninstall-%d", time.Now().UnixNano())
		packageRemoval, err := stageRemoval(filepath.Join(m.config.PluginDir, pluginID), suffix)
		if err != nil {
			return m.rollbackUninstallLocked(
				ctx,
				pluginID,
				wasEnabled,
				stagedPaths,
				fmt.Errorf("stage package cleanup: %w", err),
			)
		}
		if packageRemoval.stagedPath != "" {
			stagedPaths = append(stagedPaths, packageRemoval)
		}
		dataRemoval, err := stageRemoval(filepath.Join(m.config.DataDir, pluginID), suffix)
		if err != nil {
			return m.rollbackUninstallLocked(
				ctx,
				pluginID,
				wasEnabled,
				stagedPaths,
				fmt.Errorf("stage data cleanup: %w", err),
			)
		}
		if dataRemoval.stagedPath != "" {
			stagedPaths = append(stagedPaths, dataRemoval)
		}
	}
	if err := m.store.Delete(ctx, pluginID); err != nil {
		return m.rollbackUninstallLocked(ctx, pluginID, wasEnabled, stagedPaths, err)
	}
	for _, staged := range stagedPaths {
		if err := os.RemoveAll(staged.stagedPath); err != nil {
			log.Printf("[PluginManager] warning: deferred cleanup required for %s: %v", staged.stagedPath, err)
		}
	}

	log.Printf("[PluginManager] uninstalled plugin %s", pluginID)
	return nil
}

func (m *Manager) rollbackUninstallLocked(
	ctx context.Context,
	pluginID string,
	wasEnabled bool,
	stagedPaths []stagedRemoval,
	uninstallErr error,
) error {
	if restoreErr := restoreStagedRemovals(stagedPaths); restoreErr != nil {
		return fmt.Errorf("uninstall plugin: %v; restore files: %w", uninstallErr, restoreErr)
	}
	if wasEnabled {
		if enableErr := m.enablePluginLocked(ctx, pluginID, true); enableErr != nil {
			return fmt.Errorf("uninstall plugin: %v; restore enabled state: %w", uninstallErr, enableErr)
		}
	}
	return uninstallErr
}

type stagedRemoval struct {
	originalPath string
	stagedPath   string
}

func stageRemoval(originalPath, suffix string) (stagedRemoval, error) {
	if _, err := os.Stat(originalPath); err != nil {
		if os.IsNotExist(err) {
			return stagedRemoval{}, nil
		}
		return stagedRemoval{}, err
	}
	stagedPath := originalPath + suffix
	if err := os.Rename(originalPath, stagedPath); err != nil {
		return stagedRemoval{}, err
	}
	return stagedRemoval{originalPath: originalPath, stagedPath: stagedPath}, nil
}

func restoreStagedRemovals(stagedPaths []stagedRemoval) error {
	var errs []string
	for i := len(stagedPaths) - 1; i >= 0; i-- {
		staged := stagedPaths[i]
		if err := os.Rename(staged.stagedPath, staged.originalPath); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return errors.New(join(errs, "; "))
	}
	return nil
}

func (m *Manager) reconcileStagedRemovals(ctx context.Context) {
	for _, parent := range []string{m.config.PluginDir, m.config.DataDir} {
		entries, err := os.ReadDir(parent)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Printf("[PluginManager] warning: reconcile staged removals in %s: %v", parent, err)
			}
			continue
		}
		for _, entry := range entries {
			suffix := stagedRemovalSuffix(entry.Name())
			if suffix == "" {
				continue
			}
			pluginID, found := strings.CutSuffix(entry.Name(), suffix)
			if !found || pluginID == "" {
				continue
			}
			stagedPath := filepath.Join(parent, entry.Name())
			originalPath := filepath.Join(parent, pluginID)
			exists, err := m.store.Exists(ctx, pluginID)
			if err != nil {
				log.Printf("[PluginManager] warning: reconcile staged removal %s: %v", stagedPath, err)
				continue
			}
			if exists {
				if _, statErr := os.Stat(originalPath); os.IsNotExist(statErr) {
					if err := os.Rename(stagedPath, originalPath); err != nil {
						log.Printf("[PluginManager] warning: restore staged removal %s: %v", stagedPath, err)
					}
				} else if statErr == nil {
					if err := os.RemoveAll(stagedPath); err != nil {
						log.Printf("[PluginManager] warning: clean duplicate staged removal %s: %v", stagedPath, err)
					}
				}
				continue
			}
			if err := os.RemoveAll(stagedPath); err != nil {
				log.Printf("[PluginManager] warning: clean staged removal %s: %v", stagedPath, err)
			}
		}
	}
}

func stagedRemovalSuffix(name string) string {
	index := strings.LastIndex(name, ".uninstall-")
	if index < 0 {
		return ""
	}
	return name[index:]
}

// GetPlugin returns the current state of a plugin.
func (m *Manager) GetPlugin(ctx context.Context, pluginID string) (*model.Plugin, error) {
	return m.store.GetByPluginID(ctx, pluginID)
}

// ListPlugins returns all installed plugins.
func (m *Manager) ListPlugins(ctx context.Context) ([]model.Plugin, error) {
	return m.store.List(ctx)
}

// UpdateSettings updates a plugin's settings and re-initializes if enabled.
func (m *Manager) UpdateSettings(ctx context.Context, pluginID string, settings map[string]any) error {
	m.settingsMu.Lock()
	defer m.settingsMu.Unlock()

	m.mu.RLock()
	defer m.mu.RUnlock()

	pluginRecord, err := m.store.GetByPluginID(ctx, pluginID)
	if err != nil {
		return err
	}

	host, running := m.hosts[pluginID]

	if running && host.IsRunning() {
		if err := host.Reinitialize(
			ctx,
			settingsToStringMap(model.JSONMap(settings)),
			filepath.Join(m.config.DataDir, pluginID),
		); err != nil {
			return err
		}
	}

	if err := m.store.UpdateSettings(ctx, pluginID, settings); err != nil {
		if running && host.IsRunning() {
			rollbackErr := host.Reinitialize(
				context.Background(),
				settingsToStringMap(pluginRecord.Settings),
				filepath.Join(m.config.DataDir, pluginID),
			)
			if rollbackErr != nil {
				log.Printf("[PluginManager] failed to roll back settings for %s: %v", pluginID, rollbackErr)
			}
		}
		return err
	}
	return nil
}

// HandlePluginHTTP routes an HTTP request to the correct plugin.
func (m *Manager) HandlePluginHTTP(ctx context.Context, pluginID string, req *pb.HTTPRequest) (*pb.HTTPResponse, error) {
	m.mu.RLock()
	host, ok := m.hosts[pluginID]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("plugin %q is not running", pluginID)
	}
	return host.HandleHTTP(ctx, req)
}

// StartEnabledPlugins starts all plugins that were previously enabled.
// Called during server startup.
func (m *Manager) StartEnabledPlugins(ctx context.Context) error {
	plugins, err := m.store.ListByState(ctx, StateEnabled)
	if err != nil {
		return fmt.Errorf("failed to list enabled plugins: %w", err)
	}

	for _, p := range plugins {
		if err := m.enablePlugin(ctx, p.PluginID, true); err != nil {
			if errors.Is(err, errPluginNoLongerEnabled) {
				continue
			}
			log.Printf("[PluginManager] failed to start plugin %s: %v", p.PluginID, err)
			_ = m.store.UpdateState(ctx, p.PluginID, StateFailed, err.Error())
			continue
		}
	}

	return nil
}

// StopAll gracefully stops all running plugins.
// Called during server shutdown.
func (m *Manager) StopAll() error {
	m.mu.Lock()
	m.stopping = true
	select {
	case <-m.healthStop:
		// already closed
	default:
		close(m.healthStop)
	}
	m.mu.Unlock()
	m.healthWG.Wait()

	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []string
	for id, host := range m.hosts {
		m.restoreRegistrationsLocked(id)
		if err := host.Stop(); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", id, err))
		}
	}
	m.hosts = make(map[string]*GRPCHost)
	m.registrations = make(map[string][]providerRegistration)
	m.providerOwners = make(map[string]string)

	if len(errs) > 0 {
		return fmt.Errorf("errors stopping plugins: %s", join(errs, "; "))
	}
	return nil
}

// StartHealthMonitor starts a background goroutine that periodically checks
// plugin health and restarts crashed plugins.
func (m *Manager) StartHealthMonitor(interval time.Duration) {
	m.mu.Lock()
	if m.stopping {
		m.mu.Unlock()
		return
	}
	m.healthWG.Add(1)
	m.mu.Unlock()

	go func() {
		defer m.healthWG.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-m.healthStop:
				return
			case <-ticker.C:
				m.checkHealth()
			}
		}
	}()
}

// checkHealth inspects all running plugin hosts and restarts any that have crashed.
func (m *Manager) checkHealth() {
	m.mu.RLock()
	if m.stopping {
		m.mu.RUnlock()
		return
	}
	toRestart := make(map[string]*GRPCHost)
	for id, host := range m.hosts {
		if !host.IsRunning() {
			toRestart[id] = host
		}
	}
	m.mu.RUnlock()

	for id, observedHost := range toRestart {
		log.Printf("[PluginManager] plugin %s crashed, attempting restart", id)
		ctx := context.Background()

		m.mu.Lock()
		if m.stopping {
			m.mu.Unlock()
			return
		}
		currentHost, exists := m.hosts[id]
		if !exists || currentHost != observedHost || currentHost.IsRunning() {
			m.mu.Unlock()
			continue
		}
		record, err := m.store.GetByPluginID(ctx, id)
		if err != nil || PluginState(record.State) != StateEnabled {
			m.mu.Unlock()
			continue
		}
		m.restoreRegistrationsLocked(id)
		delete(m.hosts, id)
		m.mu.Unlock()

		if err := m.enablePlugin(ctx, id, true); err != nil {
			if errors.Is(err, errPluginNoLongerEnabled) {
				continue
			}
			log.Printf("[PluginManager] failed to restart plugin %s: %v", id, err)
			_ = m.store.UpdateState(ctx, id, StateFailed, err.Error())
		} else {
			log.Printf("[PluginManager] plugin %s restarted successfully", id)
		}
	}
}

func (m *Manager) restoreRegistrationsLocked(pluginID string) {
	registrations := m.registrations[pluginID]
	for i := len(registrations) - 1; i >= 0; i-- {
		registration := registrations[i]
		m.registry.ReplaceIf(registration.name, registration.current, registration.previous)
		if m.providerOwners[registration.name] == pluginID {
			delete(m.providerOwners, registration.name)
		}
	}
	delete(m.registrations, pluginID)
}

func (m *Manager) createPluginRecord(
	ctx context.Context,
	meta *PluginMeta,
	binaryPath string,
	source string,
) error {
	permStrings := make(model.JSONStringSlice, len(meta.Permissions))
	for i, permission := range meta.Permissions {
		permStrings[i] = string(permission)
	}

	return m.store.Create(ctx, &model.Plugin{
		PluginID:    meta.ID,
		Name:        meta.Name,
		NameZh:      meta.NameZh,
		Version:     meta.Version,
		Description: meta.Description,
		Author:      meta.Author,
		License:     meta.License,
		Homepage:    meta.Homepage,
		State:       string(StateInstalled),
		Source:      source,
		BinaryPath:  binaryPath,
		Permissions: permStrings,
		Settings:    make(model.JSONMap),
	})
}

// findBinary locates the plugin binary in the given directory.
// It looks for a binary named after the plugin ID or a "plugin" binary.
func findBinary(dir, pluginID string) string {
	// Check for binary with plugin ID name
	candidate := filepath.Join(dir, pluginID)
	if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
		return candidate
	}
	// Check for generic "plugin" binary
	candidate = filepath.Join(dir, "plugin")
	if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
		return candidate
	}
	return ""
}

// join concatenates strings with a separator.
func join(parts []string, sep string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += sep
		}
		result += p
	}
	return result
}

// settingsToStringMap converts a JSON map to a string map for gRPC.
func settingsToStringMap(settings model.JSONMap) map[string]string {
	result := make(map[string]string, len(settings))
	for k, v := range settings {
		switch val := v.(type) {
		case string:
			result[k] = val
		default:
			data, _ := json.Marshal(val)
			result[k] = string(data)
		}
	}
	return result
}
