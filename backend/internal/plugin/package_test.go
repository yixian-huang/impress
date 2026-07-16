package plugin

import (
	"archive/zip"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"blotting-consultancy/internal/provider"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type recordingNotifier struct {
	events []provider.NotifyEvent
}

func (n *recordingNotifier) Notify(_ context.Context, event provider.NotifyEvent) error {
	n.events = append(n.events, event)
	return nil
}

func TestExternalPluginLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("builds and launches an external plugin process")
	}

	moduleRoot := pluginModuleRoot(t)
	tempDir := t.TempDir()
	packagePath := buildFileNotifierPackage(t, moduleRoot, tempDir)
	pluginDir := filepath.Join(tempDir, "plugins")
	dataDir := filepath.Join(tempDir, "data")

	db, err := gorm.Open(sqlite.Open(filepath.Join(tempDir, "plugins.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	store := NewStore(db)
	require.NoError(t, store.AutoMigrate())

	registry := provider.NewRegistry()
	builtIn := &recordingNotifier{}
	registry.Register("notifier", builtIn)

	manager := NewManager(ManagerConfig{PluginDir: pluginDir, DataDir: dataDir}, store, registry)
	meta, err := manager.InstallPackage(context.Background(), packagePath)
	require.NoError(t, err)
	assert.Equal(t, "file-notifier", meta.ID)

	installed, err := store.GetByPluginID(context.Background(), meta.ID)
	require.NoError(t, err)
	assert.Equal(t, "package", installed.Source)
	assert.FileExists(t, installed.BinaryPath)

	require.NoError(t, manager.EnablePlugin(context.Background(), meta.ID))
	assert.NotSame(t, builtIn, registry.Notifier())
	require.NoError(t, registry.Notifier().Notify(context.Background(), provider.NotifyEvent{
		Type:    "content.published",
		Subject: "First",
		Body:    "Plugin process handled this event",
	}))
	assertFileContains(t, filepath.Join(dataDir, meta.ID, "notifications.jsonl"), `"subject":"First"`)

	const updateFailureCallback = "test:fail_plugin_state_update"
	require.NoError(t, store.db.Callback().Update().Before("gorm:update").Register(
		updateFailureCallback,
		func(tx *gorm.DB) {
			if tx.Statement.Table == "plugins" {
				tx.AddError(errors.New("forced plugin state update failure"))
			}
		},
	))
	err = manager.DisablePlugin(context.Background(), meta.ID)
	require.Error(t, err)
	enabledRecord, enabledErr := store.GetByPluginID(context.Background(), meta.ID)
	require.NoError(t, enabledErr)
	assert.Equal(t, string(StateEnabled), enabledRecord.State)
	assert.NotSame(t, builtIn, registry.Notifier())
	require.NoError(t, registry.Notifier().Notify(context.Background(), provider.NotifyEvent{
		Type:    "content.published",
		Subject: "After failed disable",
	}))
	assertFileContains(
		t,
		filepath.Join(dataDir, meta.ID, "notifications.jsonl"),
		`"subject":"After failed disable"`,
	)
	store.db.Callback().Update().Remove(updateFailureCallback)

	require.NoError(t, manager.DisablePlugin(context.Background(), meta.ID))
	assert.Same(t, builtIn, registry.Notifier())

	require.NoError(t, manager.EnablePlugin(context.Background(), meta.ID))
	require.NoError(t, manager.StopAll())
	assert.Same(t, builtIn, registry.Notifier())

	restarted := NewManager(ManagerConfig{PluginDir: pluginDir, DataDir: dataDir}, store, registry)
	require.NoError(t, restarted.StartEnabledPlugins(context.Background()))
	restoredRecord, err := store.GetByPluginID(context.Background(), meta.ID)
	require.NoError(t, err)
	assert.Equal(t, string(StateEnabled), restoredRecord.State)
	require.NoError(t, registry.Notifier().Notify(context.Background(), provider.NotifyEvent{
		Type:    "content.published",
		Subject: "After restart",
		Body:    "Restored from persisted enabled state",
	}))
	assertFileContains(t, filepath.Join(dataDir, meta.ID, "notifications.jsonl"), `"subject":"After restart"`)

	secondManifest := filepath.Join(tempDir, "second-plugin.yaml")
	require.NoError(t, os.WriteFile(secondManifest, []byte(`id: second-notifier
name: Second Notifier
version: 1.0.0
permissions:
  - network:outbound
providers:
  - type: notifier
    name: notifier
`), 0o640))
	secondPackage := filepath.Join(tempDir, "second-notifier.zip")
	writeZipFiles(t, secondPackage, map[string]string{
		ManifestFileName: secondManifest,
		"plugin":         filepath.Join(tempDir, "file-notifier"),
	})
	_, err = restarted.InstallPackage(context.Background(), secondPackage)
	require.NoError(t, err)
	err = restarted.EnablePlugin(context.Background(), "second-notifier")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already owned")
	require.NoError(t, registry.Notifier().Notify(context.Background(), provider.NotifyEvent{
		Type:    "content.published",
		Subject: "Still first",
	}))
	assertFileContains(t, filepath.Join(dataDir, meta.ID, "notifications.jsonl"), `"subject":"Still first"`)
	require.NoError(t, restarted.UninstallPlugin(context.Background(), "second-notifier"))

	const deleteFailureCallback = "test:fail_enabled_plugin_delete"
	require.NoError(t, store.db.Callback().Delete().Before("gorm:delete").Register(
		deleteFailureCallback,
		func(tx *gorm.DB) {
			if tx.Statement.Table == "plugins" {
				tx.AddError(errors.New("forced enabled plugin delete failure"))
			}
		},
	))
	err = restarted.UninstallPlugin(context.Background(), meta.ID)
	require.Error(t, err)
	rollbackRecord, rollbackErr := store.GetByPluginID(context.Background(), meta.ID)
	require.NoError(t, rollbackErr)
	assert.Equal(t, string(StateEnabled), rollbackRecord.State)
	assert.NotSame(t, builtIn, registry.Notifier())
	require.NoError(t, registry.Notifier().Notify(context.Background(), provider.NotifyEvent{
		Type:    "content.published",
		Subject: "After uninstall rollback",
	}))
	assertFileContains(
		t,
		filepath.Join(dataDir, meta.ID, "notifications.jsonl"),
		`"subject":"After uninstall rollback"`,
	)
	store.db.Callback().Delete().Remove(deleteFailureCallback)

	require.NoError(t, restarted.UninstallPlugin(context.Background(), meta.ID))
	assert.Same(t, builtIn, registry.Notifier())
	assert.NoDirExists(t, filepath.Join(pluginDir, meta.ID))
	assert.NoDirExists(t, filepath.Join(dataDir, meta.ID))
	exists, err := store.Exists(context.Background(), meta.ID)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestInstallPackageSerializesDuplicateInstalls(t *testing.T) {
	manager, store, _ := setupTestManager(t)
	packagePath := filepath.Join(t.TempDir(), "duplicate.zip")
	writeZip(t, packagePath, map[string]string{
		ManifestFileName: `id: duplicate-plugin
name: Duplicate Plugin
version: 1.0.0
permissions:
  - network:outbound
providers:
  - type: notifier
    name: notifier
`,
		"plugin": "#!/bin/sh\nexit 0\n",
	})

	var wait sync.WaitGroup
	wait.Add(2)
	errors := make(chan error, 2)
	for range 2 {
		go func() {
			defer wait.Done()
			_, err := manager.InstallPackage(context.Background(), packagePath)
			errors <- err
		}()
	}
	wait.Wait()
	close(errors)

	successes := 0
	failures := 0
	for err := range errors {
		if err == nil {
			successes++
		} else {
			failures++
			assert.Contains(t, err.Error(), "already installed")
		}
	}
	assert.Equal(t, 1, successes)
	assert.Equal(t, 1, failures)
	exists, err := store.Exists(context.Background(), "duplicate-plugin")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestRestoreDoesNotReenableDisabledPlugin(t *testing.T) {
	manager, store, _ := setupTestManager(t)
	packagePath := filepath.Join(t.TempDir(), "disabled.zip")
	writeZip(t, packagePath, map[string]string{
		ManifestFileName: `id: disabled-plugin
name: Disabled Plugin
version: 1.0.0
permissions:
  - network:outbound
providers:
  - type: notifier
    name: notifier
`,
		"plugin": "#!/bin/sh\nexit 0\n",
	})
	_, err := manager.InstallPackage(context.Background(), packagePath)
	require.NoError(t, err)
	require.NoError(t, store.UpdateState(context.Background(), "disabled-plugin", StateDisabled, ""))

	err = manager.enablePlugin(context.Background(), "disabled-plugin", true)
	require.ErrorIs(t, err, errPluginNoLongerEnabled)
	record, err := store.GetByPluginID(context.Background(), "disabled-plugin")
	require.NoError(t, err)
	assert.Equal(t, string(StateDisabled), record.State)
}

func TestUninstallRestoresFilesWhenDatabaseDeleteFails(t *testing.T) {
	manager, store, pluginDir := setupTestManager(t)
	packagePath := filepath.Join(t.TempDir(), "rollback.zip")
	writeZip(t, packagePath, map[string]string{
		ManifestFileName: `id: rollback-plugin
name: Rollback Plugin
version: 1.0.0
permissions:
  - network:outbound
providers:
  - type: notifier
    name: notifier
`,
		"plugin": "#!/bin/sh\nexit 0\n",
	})
	_, err := manager.InstallPackage(context.Background(), packagePath)
	require.NoError(t, err)
	dataPath := filepath.Join(manager.config.DataDir, "rollback-plugin")
	require.NoError(t, os.MkdirAll(dataPath, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dataPath, "state.txt"), []byte("keep"), 0o640))

	require.NoError(t, store.db.Callback().Delete().Before("gorm:delete").Register(
		"test:fail_plugin_delete",
		func(tx *gorm.DB) {
			if tx.Statement.Table == "plugins" {
				tx.AddError(errors.New("forced delete failure"))
			}
		},
	))

	err = manager.UninstallPlugin(context.Background(), "rollback-plugin")
	require.Error(t, err)
	assert.DirExists(t, filepath.Join(pluginDir, "rollback-plugin"))
	assert.FileExists(t, filepath.Join(dataPath, "state.txt"))
	exists, existsErr := store.Exists(context.Background(), "rollback-plugin")
	require.NoError(t, existsErr)
	assert.True(t, exists)
}

func TestStopAllPreventsLaterPluginRestart(t *testing.T) {
	manager, _, _ := setupTestManager(t)
	manager.StartHealthMonitor(time.Hour)
	require.NoError(t, manager.StopAll())

	err := manager.enablePlugin(context.Background(), "any-plugin", true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "stopping")
}

func TestInstallPackageRejectsZipSlipAndCleansStaging(t *testing.T) {
	manager, _, pluginDir := setupTestManager(t)
	packagePath := filepath.Join(t.TempDir(), "malicious.zip")
	writeZip(t, packagePath, map[string]string{
		"../escaped.txt": "unsafe",
		"plugin.yaml":    "id: malicious-plugin\nname: Malicious\nversion: 1.0.0\n",
	})

	_, err := manager.InstallPackage(context.Background(), packagePath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsafe path")
	assert.NoFileExists(t, filepath.Join(filepath.Dir(pluginDir), "escaped.txt"))

	entries, readErr := os.ReadDir(pluginDir)
	require.NoError(t, readErr)
	for _, entry := range entries {
		assert.False(t, strings.HasPrefix(entry.Name(), ".install-"))
	}
}

func TestExtractPackageEntryEnforcesActualByteBudget(t *testing.T) {
	packagePath := filepath.Join(t.TempDir(), "large-entry.zip")
	writeZip(t, packagePath, map[string]string{"large.txt": "0123456789"})

	archive, err := zip.OpenReader(packagePath)
	require.NoError(t, err)
	defer archive.Close()
	require.Len(t, archive.File, 1)

	written, err := extractPackageEntry(t.TempDir(), archive.File[0], 5)
	require.Error(t, err)
	assert.Zero(t, written)
	assert.Contains(t, err.Error(), "expands beyond")
}

func TestReconcileStagedRemovalPreservesEnabledSourceOfTruth(t *testing.T) {
	manager, store, pluginDir := setupTestManager(t)
	packagePath := filepath.Join(t.TempDir(), "staged.zip")
	writeZip(t, packagePath, map[string]string{
		ManifestFileName: `id: staged-plugin
name: Staged Plugin
version: 1.0.0
permissions:
  - network:outbound
providers:
  - type: notifier
    name: notifier
`,
		"plugin": "#!/bin/sh\nexit 0\n",
	})
	_, err := manager.InstallPackage(context.Background(), packagePath)
	require.NoError(t, err)
	require.NoError(t, store.UpdateState(context.Background(), "staged-plugin", StateEnabled, ""))

	dataPath := filepath.Join(manager.config.DataDir, "staged-plugin")
	require.NoError(t, os.MkdirAll(dataPath, 0o750))
	suffix := ".uninstall-crash"
	packageRemoval, err := stageRemoval(filepath.Join(pluginDir, "staged-plugin"), suffix)
	require.NoError(t, err)
	dataRemoval, err := stageRemoval(dataPath, suffix)
	require.NoError(t, err)

	_ = NewManager(
		ManagerConfig{PluginDir: pluginDir, DataDir: manager.config.DataDir},
		store,
		provider.NewRegistry(),
	)

	assert.DirExists(t, filepath.Join(pluginDir, "staged-plugin"))
	assert.DirExists(t, dataPath)
	assert.NoDirExists(t, packageRemoval.stagedPath)
	assert.NoDirExists(t, dataRemoval.stagedPath)
	record, err := store.GetByPluginID(context.Background(), "staged-plugin")
	require.NoError(t, err)
	assert.Equal(t, string(StateEnabled), record.State)
}

func pluginModuleRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}

func buildFileNotifierPackage(t *testing.T, moduleRoot, tempDir string) string {
	t.Helper()
	binaryPath := filepath.Join(tempDir, "file-notifier")
	command := exec.Command("go", "build", "-o", binaryPath, "./examples/plugins/file-notifier")
	command.Dir = moduleRoot
	output, err := command.CombinedOutput()
	require.NoError(t, err, string(output))

	manifestPath := filepath.Join(moduleRoot, "examples", "plugins", "file-notifier", ManifestFileName)
	manifest, err := os.ReadFile(manifestPath)
	require.NoError(t, err)

	packagePath := filepath.Join(tempDir, "file-notifier.zip")
	writeZipFiles(t, packagePath, map[string]string{
		ManifestFileName: string(manifest),
		"file-notifier":  binaryPath,
	})
	return packagePath
}

func writeZip(t *testing.T, target string, files map[string]string) {
	t.Helper()
	archive, err := os.Create(target)
	require.NoError(t, err)
	writer := zip.NewWriter(archive)
	for name, contents := range files {
		entry, err := writer.Create(name)
		require.NoError(t, err)
		_, err = io.WriteString(entry, contents)
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())
	require.NoError(t, archive.Close())
}

func writeZipFiles(t *testing.T, target string, files map[string]string) {
	t.Helper()
	archive, err := os.Create(target)
	require.NoError(t, err)
	writer := zip.NewWriter(archive)
	for name, sourcePath := range files {
		entry, err := writer.Create(name)
		require.NoError(t, err)
		source, err := os.Open(sourcePath)
		if err != nil {
			_, writeErr := io.WriteString(entry, sourcePath)
			require.NoError(t, writeErr)
			continue
		}
		_, copyErr := io.Copy(entry, source)
		require.NoError(t, copyErr)
		require.NoError(t, source.Close())
	}
	require.NoError(t, writer.Close())
	require.NoError(t, archive.Close())
}

func assertFileContains(t *testing.T, path, needle string) {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), needle)
}
