-- +goose up
CREATE TABLE modlist_junction (
    modlist_id UUID NOT NULL REFERENCES modlists(id) ON DELETE CASCADE,
    mod_id INT NOT NULL REFERENCES mods(id) ON DELETE CASCADE,
    PRIMARY KEY (modlist_id, mod_id)
);
-- +goose down
DELETE TABLE modlist_junction;