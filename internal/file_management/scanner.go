package file_management

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func ReadModFiles(pakDir string) ([]string, error) {
	modList := []string{}
	entries, err := os.ReadDir(pakDir)
	if err != nil {
		return nil, fmt.Errorf("could not read the paks file: %w", err)
	}

	for _, entry := range entries {
		name := entry.Name()

		if entry.IsDir() || filepath.Ext(name) != ".pak" {
			continue
		}

		if strings.HasSuffix(name, "-Windows.pak") {
			continue
		}

		modList = append(modList, name)
	}
	return modList, nil
}
