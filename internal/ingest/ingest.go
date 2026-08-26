package ingest

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ProcessDownloadedFile(filePath, cacheDir string) ([]string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	var extractedPaks []string

	switch ext {
	case ".pak":
		fileName := filepath.Base(filePath)
		destPath := filepath.Join(cacheDir, fileName)
		if err := copyFile(filePath, destPath); err != nil {
			return nil, fmt.Errorf("failed to copy pak file: %w", err)
		}
		extractedPaks = append(extractedPaks, fileName)

	case ".zip":
		paks, err := extractPaksFromZip(filePath, cacheDir)
		if err != nil {
			return nil, fmt.Errorf("failed to extract zip: %w", err)
		}
		extractedPaks = append(extractedPaks, paks...)

	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	return extractedPaks, nil
}

func extractPaksFromZip(zipPath, destDir string) ([]string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var extracted []string

	for _, f := range r.File {
		if f.FileInfo().IsDir() || !strings.HasSuffix(strings.ToLower(f.Name), ".pak") {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, err
		}

		baseName := filepath.Base(f.Name)
		targetPath := filepath.Join(destDir, baseName)

		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return nil, err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return nil, err
		}

		extracted = append(extracted, baseName)
	}

	return extracted, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
