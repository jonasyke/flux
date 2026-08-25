package main

import (
	"fmt"
	"log"
	"path/filepath"

	"flux/config"
	fm "flux/file_management"
	"flux/internal/database"
	"flux/internal/steam"
)

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	if fm.DirExists(cfg.GameModDir) {
		fmt.Println("Game mod/pak folder found.")
	}

	dbPath := filepath.Join(".", "storage", "flux.db")
	db, err := database.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer db.Close()

	log.Println("Database initialized successfully!")

	currentBuildID, err := steam.GetCurrentBuildID(cfg.GameModDir, steam.ReadyOrNotAppID)
	if err != nil {
		log.Printf("Warning: Could not check Steam version: %v\n", err)
	} else {
		log.Printf("Current Game Build ID: %s\n", currentBuildID)

		lastBuildID, err := database.GetLastBuildID(db, steam.ReadyOrNotAppID)
		if err != nil {
			log.Fatalf("Failed to fetch last build ID: %v", err)
		}

		if lastBuildID == "" {
			log.Println("First run detected. Recording current game version...")
			_ = database.SaveBuildID(db, steam.ReadyOrNotAppID, currentBuildID)
		} else if lastBuildID != currentBuildID {
			log.Printf("GAME UPDATE DETECTED! Old build: %s -> New build: %s\n", lastBuildID, currentBuildID)
			log.Println("Safely caching mods to prevent crashes...")

			err := fm.CacheMods(cfg.GameModDir, cfg.CacheDir)
			if err != nil {
				log.Printf("Error caching mods: %v\n", err)
			} else {
				log.Println("Mods moved to cache successfully.")
			}
			_ = database.SaveBuildID(db, steam.ReadyOrNotAppID, currentBuildID)
		} else {
			log.Println("Game version matches last check. No update needed.")
		}
	}

}
