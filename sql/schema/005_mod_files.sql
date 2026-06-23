-- +goose Up
CREATE TABLE mod_files (
    id SERIAL PRIMARY KEY,
    mod_id INT NOT NULL REFERENCES mods(id) ON DELETE CASCADE,
    filename TEXT NOT NULL UNIQUE,
    file_path TEXT NOT NULL,
    current_version TEXT NOT NULL,
    latest_version TEXT,
    game_version_compat TEXT NOT NULL DEFAULT '1.0',
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    hash TEXT,
    installed_at TIMESTAMP NOT NULL DEFAULT NOW()
);
-- +goose Down
DROP TABLE mod_files;