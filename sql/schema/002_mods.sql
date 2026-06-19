-- +goose up 

CREATE TABLE mods (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    nexus_id INT UNIQUE,
    author TEXT,
    description TEXT,
    source_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose down
DROP TABLE mods;