package store

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
)

func ListAccounts(ctx context.Context, db *sqlx.DB, siteKey string) ([]Account, error) {
	var items []Account
	err := db.SelectContext(ctx, &items, `SELECT id, site_key, username, password, extra, created_at, updated_at FROM accounts WHERE site_key = ? ORDER BY username`, siteKey)
	if err != nil { return nil, err }
	return items, nil
}

func CreateAccount(ctx context.Context, db *sqlx.DB, a *Account) error {
	_, err := db.ExecContext(ctx, `INSERT INTO accounts(id, site_key, username, password, extra) VALUES(?,?,?,?,?)`, a.ID, a.SiteKey, a.Username, a.Password, a.Extra)
	return err
}

func UpdateAccount(ctx context.Context, db *sqlx.DB, a *Account) error {
	_, err := db.ExecContext(ctx, `UPDATE accounts SET username = ?, password = ?, extra = ?, updated_at = CAST(strftime('%s','now') AS INTEGER) WHERE id = ? AND site_key = ?`, a.Username, a.Password, a.Extra, a.ID, a.SiteKey)
	return err
}

func DeleteAccount(ctx context.Context, db *sqlx.DB, siteKey string, id string) error {
	_, err := db.ExecContext(ctx, `DELETE FROM accounts WHERE id = ? AND site_key = ?`, id, siteKey)
	return err
}

func GetActiveAccountID(ctx context.Context, db *sqlx.DB, siteKey string) (*string, error) {
	var id sql.NullString
	err := db.GetContext(ctx, &id, `SELECT account_id FROM active_accounts WHERE site_key = ?`, siteKey)
	if err != nil {
		if err == sql.ErrNoRows { return nil, nil }
		return nil, err
	}
	if !id.Valid { return nil, nil }
	v := id.String
	return &v, nil
}

func SetActiveAccountID(ctx context.Context, db *sqlx.DB, siteKey string, accountID *string) error {
    var v interface{}
    if accountID != nil { v = *accountID } else { v = nil }
    _, err := db.ExecContext(ctx, `INSERT INTO active_accounts(site_key, account_id, updated_at)
        VALUES(?, ?, CAST(strftime('%s','now') AS INTEGER))
        ON CONFLICT(site_key) DO UPDATE SET account_id = excluded.account_id, updated_at = excluded.updated_at`, siteKey, v)
    return err
}
