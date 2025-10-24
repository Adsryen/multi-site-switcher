package store

import (
	"context"
	"github.com/jmoiron/sqlx"
)

func CreateSite(ctx context.Context, db *sqlx.DB, s *Site) error {
	_, err := db.ExecContext(ctx, `INSERT INTO sites(key, name, login_url) VALUES(?,?,?)`, s.Key, s.Name, s.LoginURL)
	return err
}

func UpdateSite(ctx context.Context, db *sqlx.DB, s *Site) error {
	_, err := db.ExecContext(ctx, `UPDATE sites SET name = ?, login_url = ?, updated_at = CAST(strftime('%s','now') AS INTEGER) WHERE key = ?`, s.Name, s.LoginURL, s.Key)
	return err
}

func DeleteSite(ctx context.Context, db *sqlx.DB, key string) error {
	_, err := db.ExecContext(ctx, `DELETE FROM sites WHERE key = ?`, key)
	return err
}
