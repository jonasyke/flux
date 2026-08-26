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
	"flux/internal/ingest"
	"flux/internal/nexus"
)

type ModManager struct {
	db          *sql.DB
	nexusClient *nexus.Client
	gameDomain  string
}

type ProfileFile struct {
	ProfileName string           `json:"profile_name"`
	GameDomain  string           `json:"game_domain"`
	Mods        []ProfileModItem `json:"mods"`
}

type ProfileModItem struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name"`
	FileName   string `json:"file_name"`
	NexusModID int    `json:"nexus_mod_id"`
	Version    string `json:"version"`
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

	var items []ProfileModItem
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("failed to parse %s: %w", filePath, err)
	}

	for _, item := range items {
		modID := item.ID
		if modID == "" {
			modID = fmt.Sprintf("nexus-%d", item.NexusModID)
		}

		mod := database.Mod{
			ID:             modID,
			Name:           item.Name,
			FileName:       item.FileName,
			CurrentVersion: item.Version,
			NexusModID:     item.NexusModID,
			IsActive:       true,
		}
		if err := database.UpsertMod(m.db, mod); err != nil {
			log.Printf("Failed to import %s: %v\n", item.Name, err)
		} else {
			log.Printf("Imported / updated tracked mod: %s\n", item.Name)
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

func (m *ModManager) UpdateModFromFile(downloadedFilePath, cacheDir string, nexusModID int) error {
	mod, err := database.GetModByNexusID(m.db, nexusModID)
	if err != nil {
		return fmt.Errorf("could not find mod with Nexus ID %d in database: %w", nexusModID, err)
	}

	extractedPaks, err := ingest.ProcessDownloadedFile(downloadedFilePath, cacheDir)
	if err != nil {
		return fmt.Errorf("failed to ingest update file: %w", err)
	}

	if len(extractedPaks) == 0 {
		return fmt.Errorf("no .pak file found inside %s", downloadedFilePath)
	}

	newFileName := extractedPaks[0]

	if mod.FileName != newFileName {
		oldCachePath := filepath.Join(cacheDir, mod.FileName)
		_ = os.Remove(oldCachePath)
	}

	newVersion := mod.LatestVersion
	details, err := m.nexusClient.GetModDetails(m.gameDomain, nexusModID)
	if err == nil && details.Version != "" {
		newVersion = details.Version
	}

	err = database.MarkModUpdated(m.db, mod.ID, newFileName, newVersion)
	if err != nil {
		return fmt.Errorf("failed to update mod record in database: %w", err)
	}

	log.Printf("Successfully updated %s to version %s (file: %s)!\n", mod.Name, newVersion, newFileName)
	return nil
}

func (m *ModManager) ExportProfileToFile(profileName, outputPath string) error {
	profile, err := database.ExportActiveProfile(m.db, profileName)
	if err != nil {
		return fmt.Errorf("failed to export profile: %w", err)
	}

	var items []ProfileModItem
	for _, mod := range profile.Mods {
		items = append(items, ProfileModItem{
			Name:       mod.Name,
			NexusModID: mod.NexusModID,
			Version:    mod.CurrentVersion,
			FileName:   mod.FileName,
		})
	}

	pf := ProfileFile{
		ProfileName: profileName,
		GameDomain:  m.gameDomain,
		Mods:        items,
	}

	data, err := json.MarshalIndent(pf, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	err = os.WriteFile(outputPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write profile file: %w", err)
	}

	log.Printf("Exported profile '%s' with %d mods to %s\n", profileName, len(items), outputPath)
	return nil
}

func (m *ModManager) InspectSharedProfile(profilePath string) error {
	data, err := os.ReadFile(profilePath)
	if err != nil {
		return fmt.Errorf("could not read profile file: %w", err)
	}

	var pf ProfileFile
	if err := json.Unmarshal(data, &pf); err != nil {
		return fmt.Errorf("invalid profile file: %w", err)
	}

	log.Printf("=== Inspecting Shared Profile: %s (%s) ===\n", pf.ProfileName, pf.GameDomain)

	localMods, err := database.GetAllMods(m.db)
	if err != nil {
		return fmt.Errorf("could not fetch local mods: %w", err)
	}

	localLookup := make(map[int]database.Mod)
	for _, lm := range localMods {
		if lm.NexusModID != 0 {
			localLookup[lm.NexusModID] = lm
		}
	}

	missingCount := 0
	for _, item := range pf.Mods {
		localMod, exists := localLookup[item.NexusModID]
		if !exists {
			missingCount++
			fmt.Printf("[MISSING] %s (v%s) -> Link: https://www.nexusmods.com/%s/mods/%d\n",
				item.Name, item.Version, pf.GameDomain, item.NexusModID)
		} else if localMod.CurrentVersion != item.Version {
			fmt.Printf("[VERSION MISMATCH] %s -> Friend has v%s, You have v%s\n",
				item.Name, item.Version, localMod.CurrentVersion)
		} else {
			fmt.Printf("[MATCH] %s (v%s)\n", item.Name, item.Version)
		}
	}

	if missingCount == 0 {
		log.Println("All mods in this profile match your local installation!")
	} else {
		log.Printf("%d mod(s) are missing from your setup.\n", missingCount)
	}

	return nil
}
