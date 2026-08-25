package database

import (
	"database/sql"
	"fmt"
)

func GetLastBuildID(db *sql.DB, appID int) (string, error) {
	var buildID string
	query := "SELECT last_build_id FROM game_state WHERE app_id = ?"
	err := db.QueryRow(query, appID).Scan(&buildID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("error querying game_state: %w", err)
	}
	return buildID, nil
}

func SaveBuildID(db *sql.DB, appID int, buildID string) error {
	query := `
	INSERT INTO game_state (app_id, last_build_id, last_checked_at)
	VALUES (?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(app_id) DO UPDATE SET
		last_build_id = excluded.last_build_id,
		last_checked_at = CURRENT_TIMESTAMP;
	`
	_, err := db.Exec(query, appID, buildID)
	if err != nil {
		return fmt.Errorf("could not save build id: %w", err)
	}
	return nil
}
