package main

import (
	"log"
	"path/filepath"

	"flux/config"
	"flux/internal/database"
	"flux/internal/file_management"
	"flux/internal/manager"
	"flux/internal/nexus"
)

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	dbPath := filepath.Join(".", "storage", "flux.db")
	db, err := database.InitDB(dbPath)
	if err != nil {
		log.Fatalf("DB error: %v", err)
	}
	defer db.Close()

	if cfg.NexusAPIKey == "" {
		log.Fatal("Nexus API key missing in config.json")
	}
	nexusClient := nexus.NewClient(cfg.NexusAPIKey)

	modManager := manager.NewModManager(db, nexusClient, nexus.GameDomainName)

	err = modManager.CheckForModUpdates()
	if err != nil {
		log.Fatalf("Update check error: %v", err)
	}

	downloadedFile := "/mnt/c/Users/bigsy/Downloads/Visceral Blood 3311 2.0.2 2026-06-21T09-37Z 6M9NITVLr.zip"
	nexusID := 1234

	err = modManager.UpdateModFromFile(downloadedFile, cfg.CacheDir, nexusID)
	if err != nil {
		log.Fatalf("Update mod error: %v", err)
	}

	err = file_management.DeployActiveMods(db, cfg.CacheDir, cfg.GameModDir)
}
