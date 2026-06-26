package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jonasyke/flux/db"
)

func (a *App) EnableMod(modFileID int32, filename string, gamePaksDir string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	quarantinePath := filepath.Join(a.paths.ActiveModsDir, filename)
	targetGamePath := filepath.Join(gamePaksDir, filename)

	if _, err := os.Stat(quarantinePath); os.IsNotExist(err) {
		return fmt.Errorf("mod file not found in quarantine staging: %s", quarantinePath)
	}

	log.Printf("Enabling mod: Moving %s to game directory...", filename)
	err := os.Rename(quarantinePath, targetGamePath)
	if err != nil {
		return fmt.Errorf("failed to move file to game folder: %w", err)
	}

	err = a.db.Queries.UpdateModFileStatus(ctx, db.UpdateModFileStatusParams{
		ID:       modFileID,
		FilePath: targetGamePath,
	})
	if err != nil {
		_ = os.Rename(targetGamePath, quarantinePath)
		return fmt.Errorf("database sync failed, rolling back file move: %w", err)
	}

	log.Printf("Successfully enabled %s", filename)
	return nil
}

func (a *App) DisableMod(modfileID int32, filename string, gamePaksDir string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	currentGamePath := filepath.Join(gamePaksDir, filename)
	targetQuarantinePath := filepath.Join(a.paths.ActiveModsDir, filename)

	if _, err := os.Stat(currentGamePath); os.IsNotExist(err) {
		return fmt.Errorf("mod file not found in game directory: %s", currentGamePath)
	}
	log.Printf("Disabling mod: Quaranting %s", filename)
	err := os.Rename(currentGamePath, targetQuarantinePath)
	if err != nil {
		return fmt.Errorf("failed to quarantine file: %w", err)
	}

	err = a.db.Queries.UpdateModFileStatus(ctx, db.UpdateModFileStatusParams{
		ID:       modfileID,
		FilePath: targetQuarantinePath,
	})
	if err != nil {
		_ = os.Rename(targetQuarantinePath, currentGamePath)
		return fmt.Errorf("database sync failed, rolling back quarantine: %w", err)
	}

	log.Printf("Successfully quarantined %s", filename)
	return nil
}

func (a *App) ImportModFile(sourceFilePath string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	srcInfo, err := os.Stat(sourceFilePath)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("source file does not exist: %s", sourceFilePath)
	}
	if srcInfo.IsDir() {
		return "", fmt.Errorf("selected path is a directory, please select a file")
	}

	filename := srcInfo.Name()
	if strings.ToLower(filepath.Ext(filename)) != ".pak" {
		return "", fmt.Errorf("invalid file type: only .pak files are supported")
	}

	if err := os.MkdirAll(a.paths.ActiveModsDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create quarantine directory structure: %w", err)
	}

	targetPath := filepath.Join(a.paths.ActiveModsDir, filename)

	srcFile, err := os.Open(sourceFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to create staging file: %w", err)
	}
	defer destFile.Close()

	log.Printf("Importing mod: Copying %s to staging quarantine...", filename)

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return "", fmt.Errorf("failed during file streaming copy: %w", err)
	}

	_, err = a.db.Queries.SaveScannedModFile(ctx, db.SaveScannedModFileParams{
		ModID:          1,
		Filename:       filename,
		FilePath:       targetPath,
		CurrentVersion: "1.0.0",
	})
	if err != nil {
		_ = os.Remove(targetPath)
		return "", fmt.Errorf("failed to log imported file to database: %w", err)
	}

	log.Printf("Successfully imported and staged %s", filename)
	return filename, nil
}
