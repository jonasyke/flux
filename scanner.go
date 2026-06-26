package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jonasyke/flux/db"
)

type ModFileResponse struct {
	ID             int32  `json:"id"`
	ModID          int32  `json:"mod_id"`
	Filename       string `json:"filename"`
	FilePath       string `json:"file_path"`
	CurrentVersion string `json:"current_version"`
}

func (a *App) ScanLocalMods(targetDir string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	info, err := os.Stat(targetDir)
	if os.IsNotExist(err) {
		return 0, fmt.Errorf("target directory does not exist: %s", targetDir)
	}

	if !info.IsDir() {
		return 0, fmt.Errorf("provided path is a file, not a directory: %s", targetDir)
	}

	files, err := os.ReadDir(targetDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read directory: %w", err)
	}

	scannedCount := 0

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		if strings.ToLower(filepath.Ext(filename)) == ".pak" {
			fullPath := filepath.Join(targetDir, filename)

			log.Printf("found local mod file: %s", filename)

			_, err := a.db.Queries.SaveScannedModFile(ctx, db.SaveScannedModFileParams{
				ModID:          1,
				Filename:       filename,
				FilePath:       fullPath,
				CurrentVersion: "1.0.0",
				LatestVersion:  pgtype.Text{Valid: false},
			})
			if err != nil {
				log.Printf("Error saving %s to database: %v", filename, err)
				continue
			}

			scannedCount++
		}
	}

	log.Printf("Scan complete. Successfully tracked %d mod files.", scannedCount)
	return scannedCount, nil
}

func (a *App) GetScannedMods() ([]ModFileResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	targetDir := a.paths.ActiveModsDir
	files, err := os.ReadDir(targetDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read test directory: %w", err)
	}

	var mods []ModFileResponse
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".pak" {
			filename := file.Name()
			row, err := a.db.Queries.GetModByFilename(ctx, file.Name())
			if err != nil {
				continue
			}

			fullPath := filepath.Join(targetDir, filename)

			mods = append(mods, ModFileResponse{
				ID:             row.ID,
				ModID:          row.ID,
				Filename:       row.Filename,
				FilePath:       fullPath,
				CurrentVersion: row.CurrentVersion,
			})
		}
	}

	return mods, nil
}

func (a *App) GetRawDownloads() ([]string, error) {
	targetDir := a.paths.RawDownloadsDir
	if targetDir == "" {
		return nil, fmt.Errorf("raw downloads directory is not configured")
	}

	files, err := os.ReadDir(targetDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read raw downloads directory: %w", err)
	}

	var rawFiles []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if strings.ToLower(filepath.Ext(file.Name())) == ".pak" {
			rawFiles = append(rawFiles, file.Name())
		}
	}

	return rawFiles, nil
}
