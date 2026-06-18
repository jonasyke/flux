package internal

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ExtractMod(zipPath string, paths AppPaths) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}

	defer reader.Close()

	for _, file := range reader.File {
    targetPath := filepath.Join(paths.ActiveModsDir, file.Name)

    if !strings.HasPrefix(targetPath, filepath.Clean(paths.ActiveModsDir)+string(os.PathSeparator)) {
        return fmt.Errorf("illegal file path detected in zip (Zip Slip attempt): %s", file.Name)
    }

    if file.FileInfo().IsDir() {
        if err := os.MkdirAll(targetPath, 0755); err != nil {
            return fmt.Errorf("failed to create directory entry: %w", err)
        }
        continue
    }

    parentDir := filepath.Dir(targetPath)
    if err := os.MkdirAll(parentDir, 0755); err != nil {
        return fmt.Errorf("failed to create parent directory for file %s: %w", file.Name, err)
    }

    if err := extractFile(file, targetPath); err != nil {
        return err
    }
}

	return nil
}

func extractFile(file *zip.File, targetPath string) error {
	srcFile, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file inside zip: %w", err)
	}

	defer srcFile.Close()

	destFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return fmt.Errorf("failed to create target file on disk: %w", err)
	}

	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("failed to write file content to disk: %w", err)
	}

	return nil
}

