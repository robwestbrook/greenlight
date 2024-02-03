CREATE TABLE IF NOT EXISTS events(
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL,
  description TEXT NOT NULL,
  tags TEXT,
  all_day INTEGER NOT NULL,
  start STRING,
  end STRING,
  created_at STRING,
  updated_at STRING,
  version INTEGER NOT NULL
);