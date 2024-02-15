CREATE TABLE IF NOT EXISTS users (
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  email TEXT UNIQUE NOT NULL,
  password_hash BLOB NOT NULL,
  activated BOOLEAN NOT NULL,
  created_at DATETIME,
  updated_at DATETIME,
  version INTEGER NOT NULL
);