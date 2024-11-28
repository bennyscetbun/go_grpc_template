BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY NOT NULL,
    username TEXT UNIQUE NOT NULL,
    verified_email TEXT UNIQUE,
    new_email TEXT UNIQUE,
    is_verified BOOLEAN NOT NULL,
    pswhash BYTEA NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);


CREATE INDEX IF NOT EXISTS users_idx_username ON users (username);
CREATE INDEX IF NOT EXISTS users_idx_verified_email ON users (verified_email);
CREATE INDEX IF NOT EXISTS users_idx_new_email ON users (new_email);

COMMIT;