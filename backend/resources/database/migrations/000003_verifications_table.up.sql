BEGIN;

CREATE TABLE IF NOT EXISTS email_verifications (
    id TEXT PRIMARY KEY NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    expired_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS email_verifications_idx_user_id ON email_verifications (user_id);

COMMIT;