-- name: InsertMod :one
INSERT INTO mods (name, nexus_id, author, description, source_url)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
-- name: GetModByFilename :one
SELECT m.*,
    mf.filename,
    mf.current_version,
    mf.latest_version,
    mf.is_enabled
FROM mods m
    JOIN mod_files mf ON m.id = mf.mod_id
WHERE mf.filename = $1;
-- name: UpdateModVersionCheck :exec
UPDATE mod_files
SET latest_version = $2
WHERE mod_id = $1;
-- name: GetModlistWithMods :many
SELECT ml.name AS modlist_name,
    m.name AS mod_name,
    mf.filename,
    mf.hash
FROM modlists ml
    JOIN modlist_junction mlm ON ml.id = mlm.modlist_id -- Fixed table name here
    JOIN mods m ON mlm.mod_id = m.id
    JOIN mod_files mf ON m.id = mf.mod_id
WHERE ml.id = $1;
-- name: GetModlistFilesByListID :many
SELECT mf.id,
    mf.filename,
    mf.file_path,
    mf.is_enabled
FROM modlist_junction mlm -- Fixed table name here
    JOIN mod_files mf ON mlm.mod_id = mf.mod_id
WHERE mlm.modlist_id = $1;
-- name: GetOutdatedModFiles :many
SELECT id,
    filename,
    file_path
FROM mod_files
WHERE current_version <> latest_version;
-- name: SetModFileStatus :exec
UPDATE mod_files
SET is_enabled = $2
WHERE id = $1;
-- name: GetModsIncompatibleWithGameVersion :many
SELECT id,
    filename,
    file_path
FROM mod_files
WHERE game_version_compat <> $1
    AND is_enabled = TRUE;
-- name: SaveScannedModFile :one
INSERT INTO mod_files (
        mod_id,
        filename,
        file_path,
        current_version,
        latest_version
    )
VALUES ($1, $2, $3, $4, $5) ON CONFLICT (filename) DO
UPDATE
SET file_path = excluded.file_path
RETURNING *;
-- name: EnsureDefaultModProfile :exec
INSERT INTO mods (id, name, author, description, source_url)
VALUES (
        1,
        'Unassigned Local Mods',
        'System',
        'Placeholder profile for scanned local game files.',
        ''
    ) ON CONFLICT (id) DO NOTHING;
-- name: UpdateModFileStatus :exec
UPDATE mod_files
SET file_path = $2
WHERE id = $1;