BEGIN;

DROP INDEX IF EXISTS users_idx_new_email;
DROP INDEX IF EXISTS users_idx_verified_email;
DROP INDEX IF EXISTS users_idx_username;

DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS "uuid-ossp";

COMMIT;