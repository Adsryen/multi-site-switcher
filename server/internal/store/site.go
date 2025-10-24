package store

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
)

func ListSites(ctx context.Context, db *sqlx.DB) ([]Site, error) {
	var items []Site
	err := db.SelectContext(ctx, &items, `SELECT key, name, login_url, created_at, updated_at FROM sites ORDER BY key`)
	if err != nil { return nil, err }
	return items, nil
}

func GetSite(ctx context.Context, db *sqlx.DB, key string) (*Site, error) {
	var s Site
	err := db.GetContext(ctx, &s, `SELECT key, name, login_url, created_at, updated_at FROM sites WHERE key = ?`, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { return nil, nil }
		return nil, err
	}
	return &s, nil
}
