package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
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
