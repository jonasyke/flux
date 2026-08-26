package database

import (
	"database/sql"
	"fmt"
)

type Profile struct {
	ID        string
	Name      string
	CreatedAt string
	Mods      []Mod
}

func ExportActiveProfile(db *sql.DB, profileName string) (*Profile, error) {
	mods, err := GetAllMods(db)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch mods: %w", err)
	}

	var activeMods []Mod
	for _, m := range mods {
		if m.IsActive {
			activeMods = append(activeMods, m)
		}
	}

	return &Profile{
		Name: profileName,
		Mods: activeMods,
	}, nil
}
