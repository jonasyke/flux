package file_management

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func CacheMods(modDir string) error {
	newPath := "/home/jonasyke/workspace/github.com/jonasyke/flux/storage/cache"

	mods, err := os.ReadDir(modDir)
	if err != nil {
		return fmt.Errorf("could not read folder: %w", err)
	}
	for _, mod := range mods {
		name := mod.Name()
		if mod.IsDir() || filepath.Ext(name) != ".pak" {
			continue
		}
		if strings.HasSuffix(name, "-Windows.pak") {
			continue
		}

		oldFilePath := filepath.Join(modDir, name)
		newFilePath := filepath.Join(newPath, name)
		err = os.Rename(oldFilePath, newFilePath)
		if err != nil {
			copyErr := CopyMods(oldFilePath, newFilePath)
			if copyErr != nil {
				return fmt.Errorf("could not copy %s across devices: %w (rename failed: %v)", name, copyErr, err)
			}
		}
	}
	return nil
}

func CopyMods(srcPath, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}
	return nil
}
