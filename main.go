package main

import (
	fm "flux/file_management"
	"fmt"
	"log"
)

func main() {

	pakDir := "/mnt/d/SteamLibrary/steamapps/common/Ready Or Not/ReadyOrNot/Content/Paks"

	exists := fm.DirExists(pakDir)

	fmt.Printf("Did we find the folder: %v\n", exists)

	mods, err := fm.ReadModFiles(pakDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, mod := range mods {
		fmt.Printf("mod: %s\n", mod)
	}

}
