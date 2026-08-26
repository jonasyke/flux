package main

import (
	"log"
	"path/filepath"

	"flux/config"
	"flux/internal/database"
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

		testMod := database.Mod{
			ID:             "mod-menu-sample",
			Name:           "In-Game Menu Sample",
			FileName:       "pakchunk99-Menu_P.pak",
			CurrentVersion: "0.9.0",
			NexusModID:     1,
			IsActive:       true,
		}

		err = database.UpsertMod(db, testMod)
		if err != nil {
			log.Fatalf("Failed to seed mod: %v", err)
		}

		modManager := manager.NewModManager(db, nexusClient, nexus.GameDomainName)

		err = modManager.DiscoverAndRegister(cfg.CacheDir)
		if err != nil {
			log.Printf("Discovery error: %v\n", err)
		}

		err = modManager.CheckForModUpdates()
		if err != nil {
			log.Fatalf("Update check error: %v", err)
		}

}
