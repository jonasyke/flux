-- +goose up
CREATE TABLE modlists (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    description text,
    created_by TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
-- +goose down
DROP TABLE modlists;