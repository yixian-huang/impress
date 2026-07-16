package plugin

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	maxPluginPackageSize = int64(100 << 20)
	maxPluginEntries     = 1024
)

// InstallPackage installs a zip package into the managed plugin directory.
// Extraction is isolated in a temporary directory and promoted atomically only
// after the manifest, permissions, and executable have been validated.
func (m *Manager) InstallPackage(ctx context.Context, packagePath string) (*PluginMeta, error) {
	m.installMu.Lock()
	defer m.installMu.Unlock()

	if err := os.MkdirAll(m.config.PluginDir, 0o750); err != nil {
		return nil, fmt.Errorf("create plugin directory: %w", err)
	}

	info, err := os.Stat(packagePath)
	if err != nil {
		return nil, fmt.Errorf("stat plugin package: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("plugin package must be a zip file")
	}
	if info.Size() > maxPluginPackageSize {
		return nil, fmt.Errorf("plugin package exceeds %d bytes", maxPluginPackageSize)
	}

	archive, err := zip.OpenReader(packagePath)
	if err != nil {
		return nil, fmt.Errorf("open plugin package: %w", err)
	}
	defer archive.Close()

	if len(archive.File) == 0 {
		return nil, fmt.Errorf("plugin package is empty")
	}
	if len(archive.File) > maxPluginEntries {
		return nil, fmt.Errorf("plugin package contains too many entries")
	}

	tempDir, err := os.MkdirTemp(m.config.PluginDir, ".install-*")
	if err != nil {
		return nil, fmt.Errorf("create install staging directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	var declaredExtracted int64
	var actualExtracted int64
	for _, file := range archive.File {
		if file.UncompressedSize64 > uint64(maxPluginPackageSize) {
			return nil, fmt.Errorf("plugin package entry %q is too large", file.Name)
		}
		declaredExtracted += int64(file.UncompressedSize64)
		if declaredExtracted > maxPluginPackageSize {
			return nil, fmt.Errorf("plugin package expands beyond %d bytes", maxPluginPackageSize)
		}
		written, err := extractPackageEntry(
			tempDir,
			file,
			maxPluginPackageSize-actualExtracted,
		)
		if err != nil {
			return nil, err
		}
		actualExtracted += written
	}

	packageRoot, err := locatePackageRoot(tempDir)
	if err != nil {
		return nil, err
	}
	meta, err := LoadAndValidateManifest(packageRoot)
	if err != nil {
		return nil, fmt.Errorf("invalid plugin manifest: %w", err)
	}
	if err := ValidateManifestPermissions(meta); err != nil {
		return nil, err
	}
	if err := ValidateSupportedSettings(meta); err != nil {
		return nil, err
	}
	if err := ValidateExternalRuntimeContract(meta); err != nil {
		return nil, err
	}

	exists, err := m.store.Exists(ctx, meta.ID)
	if err != nil {
		return nil, fmt.Errorf("check existing plugin: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("plugin %q is already installed", meta.ID)
	}

	binaryPath := findBinary(packageRoot, meta.ID)
	if binaryPath == "" {
		return nil, fmt.Errorf("plugin %q package does not contain an executable named %q or %q", meta.ID, meta.ID, "plugin")
	}
	if err := os.Chmod(binaryPath, 0o750); err != nil {
		return nil, fmt.Errorf("make plugin executable: %w", err)
	}

	finalDir := filepath.Join(m.config.PluginDir, meta.ID)
	if _, err := os.Stat(finalDir); err == nil {
		return nil, fmt.Errorf("plugin install directory %q already exists", finalDir)
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("check plugin install directory: %w", err)
	}
	if err := os.Rename(packageRoot, finalDir); err != nil {
		return nil, fmt.Errorf("promote plugin package: %w", err)
	}

	finalBinaryPath := findBinary(finalDir, meta.ID)
	if err := m.createPluginRecord(ctx, meta, finalBinaryPath, "package"); err != nil {
		_ = os.RemoveAll(finalDir)
		return nil, fmt.Errorf("store plugin record: %w", err)
	}

	return meta, nil
}

func extractPackageEntry(root string, file *zip.File, remainingBytes int64) (int64, error) {
	name := strings.ReplaceAll(file.Name, "\\", "/")
	cleanName := path.Clean(name)
	if cleanName == "." || path.IsAbs(cleanName) || cleanName == ".." || strings.HasPrefix(cleanName, "../") {
		return 0, fmt.Errorf("plugin package contains unsafe path %q", file.Name)
	}
	if file.Mode()&os.ModeSymlink != 0 {
		return 0, fmt.Errorf("plugin package contains unsupported symlink %q", file.Name)
	}

	target := filepath.Join(root, filepath.FromSlash(cleanName))
	relative, err := filepath.Rel(root, target)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return 0, fmt.Errorf("plugin package path escapes install root: %q", file.Name)
	}

	if file.FileInfo().IsDir() {
		if err := os.MkdirAll(target, 0o750); err != nil {
			return 0, fmt.Errorf("create package directory %q: %w", file.Name, err)
		}
		return 0, nil
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
		return 0, fmt.Errorf("create package parent directory %q: %w", file.Name, err)
	}

	source, err := file.Open()
	if err != nil {
		return 0, fmt.Errorf("open package entry %q: %w", file.Name, err)
	}
	defer source.Close()

	destination, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o640)
	if err != nil {
		return 0, fmt.Errorf("create package entry %q: %w", file.Name, err)
	}
	written, copyErr := io.Copy(destination, io.LimitReader(source, remainingBytes+1))
	closeErr := destination.Close()
	if copyErr != nil {
		return 0, fmt.Errorf("extract package entry %q: %w", file.Name, copyErr)
	}
	if closeErr != nil {
		return 0, fmt.Errorf("close package entry %q: %w", file.Name, closeErr)
	}
	if written > remainingBytes {
		return 0, fmt.Errorf("plugin package expands beyond %d bytes", maxPluginPackageSize)
	}
	return written, nil
}

func locatePackageRoot(tempDir string) (string, error) {
	if _, err := os.Stat(filepath.Join(tempDir, ManifestFileName)); err == nil {
		return tempDir, nil
	}

	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return "", fmt.Errorf("read package root: %w", err)
	}
	if len(entries) == 1 && entries[0].IsDir() {
		nested := filepath.Join(tempDir, entries[0].Name())
		if _, err := os.Stat(filepath.Join(nested, ManifestFileName)); err == nil {
			return nested, nil
		}
	}
	return "", fmt.Errorf("plugin package must contain %s at its root", ManifestFileName)
}
