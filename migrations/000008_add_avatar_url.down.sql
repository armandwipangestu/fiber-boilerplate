DROP INDEX IF EXISTS idx_users_avatar_url;

ALTER TABLE users
    DROP COLUMN IF EXISTS avatar_url;