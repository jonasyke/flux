package internal

import (
	"fmt"
	"os"
	"path/filepath"
)

type AppPaths struct {
	Root            string
	RawDownloadsDir string
	ActiveModsDir   string
}

func NewAppPath(root string) AppPaths {
	rawDir := os.Getenv("RAW_DOWNLOADS_DIR")
	if rawDir == "" {
		rawDir = filepath.Join(root, "downloads", "cache", "rawdownloads")
	}

	modsDir := os.Getenv("ACTIVE_MOD_DIR")
	if modsDir == "" {
		modsDir = filepath.Join(root, "downloads", "storage", "mods")
	}

	return AppPaths{
		Root:            root,
		RawDownloadsDir: rawDir,
		ActiveModsDir:   modsDir,
	}
}

func (p AppPaths) EnsureDirsExist() error {
	if err := os.MkdirAll(p.RawDownloadsDir, 0755); err != nil {
		return fmt.Errorf("failed to create raw downloads dir: %w", err)
	}

	if err := os.MkdirAll(p.ActiveModsDir, 0755); err != nil {
		return fmt.Errorf("failed to create active mods dir: %w", err)
	}

	return nil
}

func ResolveRootDir() string {
	if envRoot := os.Getenv("FLUX_ROOT_DIR"); envRoot != "" {
		return envRoot
	}

	userConfig, err := os.UserConfigDir()
	if err != nil {
		return "./flux"
	}

	return filepath.Join(userConfig, "Flux")
}
