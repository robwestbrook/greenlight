CREATE INDEX IF NOT EXISTS event_title_idx
ON events (title);
CREATE INDEX IF NOT EXISTS event_description_idx
ON events (description);
CREATE INDEX IF NOT EXISTS event_tags_idx
ON events (tags);