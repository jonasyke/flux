package main

import (
	"fmt"
	"log"

	"flux/config"
	fm "flux/file_management"
)

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	exists := fm.DirExists(cfg.GameModDir)

	fmt.Printf("Did we find the folder: %v\n", exists)

	mods, err := fm.ReadModFiles(cfg.GameModDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, mod := range mods {
		fmt.Printf("mod: %s\n", mod)
	}

	err = fm.CacheMods(cfg.GameModDir, cfg.CacheDir)
	if err != nil {
		log.Fatal(err)
	}

}
