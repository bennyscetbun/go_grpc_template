BEGIN;

DROP INDEX IF EXISTS user_tokens_idx_user_id;

DROP TABLE IF EXISTS user_tokens;

COMMIT;