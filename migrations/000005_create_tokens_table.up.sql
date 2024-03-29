CREATE TABLE IF NOT EXISTS tokens (
  hash BLOB PRIMARY KEY,
  user_id INTEGER,
  expiry DATETIME NOT NULL,
  scope TEXT NOT NULL,
  FOREIGN KEY (user_id)
  REFERENCES users(id)
  ON DELETE CASCADE
);