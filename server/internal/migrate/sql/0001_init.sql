-- schema initialization (idempotent)
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS sites (
  key TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  login_url TEXT,
  created_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s','now') AS INTEGER)),
  updated_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s','now') AS INTEGER))
);

CREATE TABLE IF NOT EXISTS accounts (
  id TEXT PRIMARY KEY,
  site_key TEXT NOT NULL,
  username TEXT NOT NULL,
  password TEXT,
  extra TEXT,
  created_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s','now') AS INTEGER)),
  updated_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s','now') AS INTEGER)),
  FOREIGN KEY(site_key) REFERENCES sites(key) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_accounts_site_key_username ON accounts(site_key, username);

CREATE TABLE IF NOT EXISTS active_accounts (
  site_key TEXT PRIMARY KEY,
  account_id TEXT NULL,
  updated_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s','now') AS INTEGER)),
  FOREIGN KEY(site_key) REFERENCES sites(key) ON DELETE CASCADE,
  FOREIGN KEY(account_id) REFERENCES accounts(id) ON DELETE SET NULL
);
