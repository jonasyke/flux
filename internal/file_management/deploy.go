package file_management

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"flux/internal/database"
)

func DeployActiveMods(db *sql.DB, cacheDir, gameModDir string) error {
	mods, err := database.GetAllMods(db)
	if err != nil {
		return fmt.Errorf("failed to retrieve mods from database: %w", err)
	}

	deployedCount := 0
	for _, mod := range mods {
		if !mod.IsActive {
			continue
		}

		srcFile := filepath.Join(cacheDir, mod.FileName)
		dstFile := filepath.Join(gameModDir, mod.FileName)

		if _, err := os.Stat(srcFile); os.IsNotExist(err) {
			log.Printf("Warning: Mod '%s' is marked active but missing in cache: %s\n", mod.Name, srcFile)
			continue
		}

		err = CopyMods(srcFile, dstFile)
		if err != nil {
			return fmt.Errorf("failed to deploy mod %s: %w", mod.Name, err)
		}

		deployedCount++
		log.Printf("Deployed: %s -> %s\n", mod.Name, mod.FileName)
	}

	log.Printf("Deployment complete: %d active mod(s) installed into game.\n", deployedCount)
	return nil
}

func PurgeGameMods(db *sql.DB, gameModDir string) error {
	mods, err := database.GetAllMods(db)
	if err != nil {
		return fmt.Errorf("failed to retrieve mods from database: %w", err)
	}

	purgedCount := 0
	for _, mod := range mods {
		gameFilePath := filepath.Join(gameModDir, mod.FileName)

		if _, err := os.Stat(gameFilePath); err == nil {
			err = os.Remove(gameFilePath)
			if err != nil {
				log.Printf("Warning: Could not remove %s: %v\n", mod.FileName, err)
				continue
			}
			purgedCount++
			log.Printf("Removed from game: %s\n", mod.FileName)
		}
	}

	log.Printf("Purge complete: %d mod(s) removed from game directory.\n", purgedCount)
	return nil
}
