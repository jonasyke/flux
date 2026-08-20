package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	GameModDir string `json:"game_mod_dir"`
	CacheDir   string `json:"cache_dir"`
}

var commonSteamPaths = []string{
	`C:\Program Files (x86)\Steam\steamapps\common\Ready Or Not\ReadyOrNot\Content\Paks`,
	`D:\SteamLibrary\steamapps\common\Ready Or Not\ReadyOrNot\Content\Paks`,
	`/mnt/d/SteamLibrary/steamapps/common/Ready Or Not/ReadyOrNot/Content/Paks`,
}

func LoadConfig() (*Config, error) {
	configFile := "config.json"
	data, err := os.ReadFile(configFile)
	if err == nil {
		var cfg Config
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("corrupt config file: %w", err)
		}
		return &cfg, nil
	}

	detectedDir := ""
	for _, candidate := range commonSteamPaths {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			detectedDir = candidate
			break
		}
	}

	if detectedDir == "" {
		return nil, fmt.Errorf("could not find Ready Or Not install; please create config.json manually")
	}

	cfg := Config{
		GameModDir: detectedDir,
		CacheDir:   filepath.Join(".", "storage", "cache"),
	}

	data, err = json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, err
	}
	_ = os.WriteFile(configFile, data, 0644)

	return &cfg, nil
}
