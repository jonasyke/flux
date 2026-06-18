package main

import (
	"fmt"
	"log"

	"github.com/jonasyke/flux/internal"
)

func main() {
	rootDir := internal.ResolveRootDir()
	paths := internal.NewAppPath(rootDir)

	if err := paths.EnsureDirsExist(); err != nil {
		log.Fatalf("FilePath could not be created: %v", err)
	}

	testURL := "https://github.com/octocat/Spoon-Knife/archive/refs/heads/main.zip"
	fmt.Println("Starting download...")

	savedPath, err := internal.DownloadModfile(testURL, paths)
	if err != nil {
		fmt.Printf("Download failed: %v\n", err)
		return
	}

	fmt.Printf("Success! Raw mod archive stored at: %s\n", savedPath)

	fmt.Println("Extracting mod archive...")
	err = internal.ExtractMod(savedPath, paths)
	if err != nil {
		fmt.Printf("Extraction failed: %v\n", err)
		return
	}

	fmt.Println("Extraction complete!")
}
