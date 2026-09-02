package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"flux/config"
	"flux/internal/database"
	"flux/internal/file_management"
	"flux/internal/manager"
	"flux/internal/nexus"
	"flux/internal/protocol"
	"flux/internal/steam"
)

func printUsage() {
	fmt.Println(`"
Flux - Ready Or Not Mod Manager

Usage:
  flux <command> [arguments]

Commands:
  register-nxm              Register Flux as the nxm:// protocol handler
  unregister-nxm            Remove the nxm:// registration
  check                     Check for game updates & mod updates
  deploy                    Deploy all active mods from cache to the game
  purge                     Safely remove all custom mods from the game
  discover                  Scan cache for new .pak files and link Nexus IDs
  update <file> <nexus_id>  Ingest a downloaded .zip/.pak and update mod in DB
  export-profile <name>     Export active mod profile to a JSON file
  inspect-profile <file>    Compare a friend's profile with your installed mods
"`)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

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

	nexusClient := nexus.NewClient(cfg.NexusAPIKey)
	modManager := manager.NewModManager(db, nexusClient, nexus.GameDomainName)

	if len(os.Args) >= 2 && strings.HasPrefix(os.Args[1], "nxm://") {
		err := modManager.HandleNXMLink(os.Args[1], cfg.CacheDir)
		if err != nil {
			log.Fatalf("NXM download failed: %v", err)
		}
		fmt.Println("Download + registration complete.")
		return
	}

	command := os.Args[1]

	switch command {
	case "register-nxm":
		if err := protocol.RegisterNXM(); err != nil {
			log.Fatalf("Failed to register nxm handler: %v", err)
		}
		fmt.Println("Registered Flux as the nxm:// protocol handler.")

	case "unregister-nxm":
		if err := protocol.UnregisterNXM(); err != nil {
			log.Fatalf("Failed to unregister: %v", err)
		}
		fmt.Println("Unregistered nxm:// handler.")
	case "check":
		currentBuild, err := steam.GetCurrentBuildID(cfg.GameModDir, steam.ReadyOrNotAppID)
		if err == nil {
			lastBuild, _ := database.GetLastBuildID(db, steam.ReadyOrNotAppID)
			if lastBuild != "" && lastBuild != currentBuild {
				fmt.Printf("[ALERT] Game update detected (Build %s -> %s)!\n", lastBuild, currentBuild)
				fmt.Println("Consider running 'flux purge' before playing.")
			} else {
				_ = database.SaveBuildID(db, steam.ReadyOrNotAppID, currentBuild)
			}
		}

		fmt.Println("Checking for mod updates...")
		err = modManager.CheckForModUpdates()
		if err != nil {
			log.Fatalf("Update check failed: %v", err)
		}

	case "deploy":
		fmt.Println("Deploying active mods to game...")
		err = file_management.DeployActiveMods(db, cfg.CacheDir, cfg.GameModDir)
		if err != nil {
			log.Fatalf("Deployment failed: %v", err)
		}

	case "purge":
		fmt.Println("Purging custom mods from game folder...")
		err = file_management.PurgeGameMods(db, cfg.GameModDir)
		if err != nil {
			log.Fatalf("Purge failed: %v", err)
		}

	case "discover":
		fmt.Println("Scanning for untracked mod files...")
		err = modManager.DiscoverAndRegister(cfg.CacheDir)
		if err != nil {
			log.Fatalf("Discovery failed: %v", err)
		}

	case "update":
		if len(os.Args) < 4 {
			fmt.Println("Usage: flux update <path/to/download.zip> <nexus_mod_id>")
			return
		}
		filePath := os.Args[2]
		var nexusID int
		_, err := fmt.Sscanf(os.Args[3], "%d", &nexusID)
		if err != nil {
			log.Fatalf("Invalid Nexus Mod ID: %v", err)
		}

		err = modManager.UpdateModFromFile(filePath, cfg.CacheDir, nexusID)
		if err != nil {
			log.Fatalf("Update failed: %v", err)
		}

	case "export-profile":
		profileName := "My Mod Profile"
		if len(os.Args) >= 3 {
			profileName = os.Args[2]
		}
		outFile := "profile.json"
		err = modManager.ExportProfileToFile(profileName, outFile)
		if err != nil {
			log.Fatalf("Export failed: %v", err)
		}

	case "inspect-profile":
		if len(os.Args) < 3 {
			fmt.Println("Usage: flux inspect-profile <path/to/profile.json>")
			return
		}
		profilePath := os.Args[2]
		err = modManager.InspectSharedProfile(profilePath)
		if err != nil {
			log.Fatalf("Inspect failed: %v", err)
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}
