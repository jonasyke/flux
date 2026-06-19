-- +goose up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- +goose down
DROP EXTENSION IF EXISTS "uuid-ossp";

