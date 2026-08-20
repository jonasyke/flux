package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	SteamPakPath   string `json:"steam_pak_path"`
	AppDataPakPath string `json:"appdata_pak_path"`
	NexusAPIKey    string `json:"nexus_api_key"`
}

func LoadConfig(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	err = json.NewDecoder(file).Decode(&cfg)
	return &cfg, err
}

func SaveConfig(configPath string, cfg *Config) error {
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cfg)
}
