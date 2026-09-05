ALTER TABLE users
    ADD COLUMN IF NOT EXISTS avatar_url TEXT;

CREATE INDEX IF NOT EXISTS idx_users_avatar_url ON users(avatar_url)
    WHERE avatar_url IS NOT NULL;