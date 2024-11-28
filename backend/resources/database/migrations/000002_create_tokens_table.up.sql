BEGIN;

CREATE TABLE IF NOT EXISTS user_tokens (
    id TEXT PRIMARY KEY NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expired_at TIMESTAMP NOT NULL,
    
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS user_tokens_idx_user_id ON user_tokens (user_id);

COMMIT;