package database

import (
	"database/sql"
	"fmt"
)

type Mod struct {
	ID             string
	Name           string
	FileName       string
	CurrentVersion string
	LatestVersion  string
	NexusModID     int
	IsActive       bool
}

func UpsertMod(db *sql.DB, mod Mod) error {
	query := `
	INSERT INTO mods (id, name, file_name, current_version, latest_version, nexus_mod_id, is_active)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		name = excluded.name,
		file_name = excluded.file_name,
		current_version = excluded.current_version,
		latest_version = excluded.latest_version,
		nexus_mod_id = excluded.nexus_mod_id,
		is_active = excluded.is_active,
		updated_at = CURRENT_TIMESTAMP;
	`
	_, err := db.Exec(query, mod.ID, mod.Name, mod.FileName, mod.CurrentVersion, mod.LatestVersion, mod.NexusModID, mod.IsActive)
	if err != nil {
		return fmt.Errorf("could not upsert mod: %w", err)
	}
	return nil
}

func GetAllMods(db *sql.DB) ([]Mod, error) {
	rows, err := db.Query("SELECT id, name, file_name, current_version, latest_version, nexus_mod_id, is_active FROM mods")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mods []Mod
	for rows.Next() {
		var m Mod
		var latestVersion sql.NullString

		err := rows.Scan(&m.ID, &m.Name, &m.FileName, &m.CurrentVersion, &latestVersion, &m.NexusModID, &m.IsActive)
		if err != nil {
			return nil, err
		}
		if latestVersion.Valid {
			m.LatestVersion = latestVersion.String
		}

		mods = append(mods, m)
	}

	return mods, nil
}
