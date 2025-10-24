-- site field schemas for per-site props validation and UI rendering
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS site_field_schemas (
  site_key TEXT NOT NULL,
  field TEXT NOT NULL,
  type TEXT NOT NULL, -- string|number|boolean|datetime|json
  required INTEGER NOT NULL DEFAULT 0, -- 0/1
  default_value TEXT, -- JSON serialized
  regex TEXT,
  choices TEXT, -- JSON array
  secret INTEGER NOT NULL DEFAULT 0, -- 0/1
  "order" INTEGER DEFAULT 0,
  ui_hint TEXT,
  PRIMARY KEY(site_key, field),
  FOREIGN KEY(site_key) REFERENCES sites(key) ON DELETE CASCADE
);
