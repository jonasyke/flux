package main

import (
	"log"

	"flux/config"
	"flux/internal/nexus"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	if cfg.NexusAPIKey == "" {
		log.Println("No Nexus API key provided in config.json. Skipping Nexus check.")
		return
	}

	client := nexus.NewClient(cfg.NexusAPIKey)

	err = client.ValidateKey()
	if err != nil {
		log.Fatalf("Nexus API Key validation failed: %v", err)
	}
	log.Println("Nexus API key validated successfully!")

	sampleModID := 1
	details, err := client.GetModDetails(nexus.GameDomainName, sampleModID)
	if err != nil {
		log.Printf("Failed to get mod details: %v", err)
	} else {
		log.Printf("Fetched Mod: %s (Latest Version: %s)", details.Name, details.Version)
	}
}
