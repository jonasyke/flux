package manager

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"flux/internal/database"
	"flux/internal/nexus"
)

type ModManager struct {
	db          *sql.DB
	nexusClient *nexus.Client
	gameDomain  string
}

type ModSeed struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	FileName       string `json:"file_name"`
	NexusModID     int    `json:"nexus_mod_id"`
	CurrentVersion string `json:"current_version"`
}

func NewModManager(db *sql.DB, client *nexus.Client, gameDomain string) *ModManager {
	return &ModManager{
		db:          db,
		nexusClient: client,
		gameDomain:  gameDomain,
	}
}

func (m *ModManager) CheckForModUpdates() error {
	mods, err := database.GetAllMods(m.db)
	if err != nil {
		return fmt.Errorf("failed to retrieve mods from DB: %w", err)
	}

	if len(mods) == 0 {
		log.Println("No mods tracked in the database yet.")
		return nil
	}

	log.Printf("Checking updates for %d tracked mod(s)...\n", len(mods))

	for _, mod := range mods {
		if mod.NexusModID == 0 {
			continue
		}

		details, err := m.nexusClient.GetModDetails(m.gameDomain, mod.NexusModID)
		if err != nil {
			log.Printf("Could not check updates for %s (Nexus ID %d): %v\n", mod.Name, mod.NexusModID, err)
			continue
		}

		err = database.UpdateLatestVersion(m.db, mod.ID, details.Version)
		if err != nil {
			log.Printf("Failed to record latest version in DB: %v\n", err)
		}

		if mod.CurrentVersion != details.Version {
			log.Printf("[UPDATE AVAILABLE] %s: Local (%s) -> Nexus (%s)\n", mod.Name, mod.CurrentVersion, details.Version)
		} else {
			log.Printf("[UP TO DATE] %s (%s)\n", mod.Name, mod.CurrentVersion)
		}
	}

	return nil
}

func (m *ModManager) ImportModsFromJSON(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	var seeds []ModSeed
	if err := json.Unmarshal(data, &seeds); err != nil {
		return fmt.Errorf("failed to parse %s: %w", filePath, err)
	}

	for _, seed := range seeds {
		mod := database.Mod{
			ID:             seed.ID,
			Name:           seed.Name,
			FileName:       seed.FileName,
			CurrentVersion: seed.CurrentVersion,
			NexusModID:     seed.NexusModID,
			IsActive:       true,
		}
		if err := database.UpsertMod(m.db, mod); err != nil {
			log.Printf("Failed to import %s: %v\n", seed.Name, err)
		} else {
			log.Printf("Imported / updated tracked mod: %s\n", seed.Name)
		}
	}

	return nil
}

func (m *ModManager) DiscoverAndRegister(folderPath string) error {
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return fmt.Errorf("could not read directory: %w", err)
	}

	trackedMods, err := database.GetAllMods(m.db)
	if err != nil {
		return fmt.Errorf("could not get tracked mods: %w", err)
	}

	trackedFiles := make(map[string]bool)
	for _, mod := range trackedMods {
		trackedFiles[mod.FileName] = true
	}

	reader := bufio.NewReader(os.Stdin)

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || filepath.Ext(name) != ".pak" || strings.HasSuffix(name, "-Windows.pak") {
			continue
		}

		if trackedFiles[name] {
			continue
		}

		fmt.Println("--------------------------------------------------")
		fmt.Printf("Found untracked mod file: %s\n", name)
		fmt.Print("Enter Nexus Mod ID (or press Enter to skip): ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		nexusID, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("Invalid ID number, skipping...")
			continue
		}

		details, err := m.nexusClient.GetModDetails(m.gameDomain, nexusID)
		if err != nil {
			log.Printf("Could not fetch details from Nexus for ID %d: %v\n", nexusID, err)
			continue
		}

		newMod := database.Mod{
			ID:             fmt.Sprintf("nexus-%d", nexusID),
			Name:           details.Name,
			FileName:       name,
			CurrentVersion: details.Version,
			LatestVersion:  details.Version,
			NexusModID:     nexusID,
			IsActive:       true,
		}

		err = database.UpsertMod(m.db, newMod)
		if err != nil {
			log.Printf("Failed to save mod to DB: %v\n", err)
		} else {
			fmt.Printf("Successfully registered: %s (v%s)\n", details.Name, details.Version)
		}
	}

	return nil
}
