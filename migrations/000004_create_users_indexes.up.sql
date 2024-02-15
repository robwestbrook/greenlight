CREATE INDEX IF NOT EXISTS users_pkey_idx
ON users (id);
CREATE UNIQUE INDEX IF NOT EXISTS users_email_idx
ON users (email);