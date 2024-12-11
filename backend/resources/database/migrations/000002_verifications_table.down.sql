BEGIN;

DROP INDEX IF EXISTS email_verifications_idx_user_id;

DROP TABLE IF EXISTS email_verifications;

COMMIT;