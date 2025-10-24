package migrate

import (
	"context"
	"github.com/jmoiron/sqlx"
)

import (
    _ "embed"
)

// Using Go embed for SQL scripts keeps deployment simple.
//go:generate echo "SQL migrations are embedded via go:embed in this package"

//go:embed sql/0001_init.sql
var initSQL string

func Apply(ctx context.Context, db *sqlx.DB) error {
	if _, err := db.ExecContext(ctx, initSQL); err != nil {
		return err
	}
	return nil
}
