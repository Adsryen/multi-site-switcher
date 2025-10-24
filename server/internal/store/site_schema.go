package store

import (
	"context"
	"github.com/jmoiron/sqlx"
)

func GetSiteFieldSchemas(ctx context.Context, db *sqlx.DB, siteKey string) ([]SiteFieldSchema, error) {
	var items []SiteFieldSchema
	err := db.SelectContext(ctx, &items, `SELECT site_key, field, type, required, default_value, regex, choices, secret, "order", ui_hint
		FROM site_field_schemas WHERE site_key = ? ORDER BY "order", field`, siteKey)
	if err != nil { return nil, err }
	return items, nil
}

func UpsertSiteFieldSchema(ctx context.Context, db *sqlx.DB, s *SiteFieldSchema) error {
	_, err := db.ExecContext(ctx, `INSERT INTO site_field_schemas(site_key, field, type, required, default_value, regex, choices, secret, "order", ui_hint)
		VALUES(?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(site_key, field) DO UPDATE SET
			type=excluded.type,
			required=excluded.required,
			default_value=excluded.default_value,
			regex=excluded.regex,
			choices=excluded.choices,
			secret=excluded.secret,
			"order"=excluded."order",
			ui_hint=excluded.ui_hint`,
		s.SiteKey, s.Field, s.Type, s.Required, s.DefaultValue, s.Regex, s.Choices, s.Secret, s.Order, s.UIHint)
	return err
}

func DeleteSiteFieldSchema(ctx context.Context, db *sqlx.DB, siteKey, field string) error {
	_, err := db.ExecContext(ctx, `DELETE FROM site_field_schemas WHERE site_key = ? AND field = ?`, siteKey, field)
	return err
}
