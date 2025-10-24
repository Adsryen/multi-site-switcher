package migrate

import (
    "context"
    "sort"
    "strings"
    "time"
    "github.com/jmoiron/sqlx"
)

import (
    "embed"
    "io/fs"
)

// Using Go embed for SQL scripts keeps deployment simple.
//go:generate echo "SQL migrations are embedded via go:embed in this package"

//go:embed sql/*.sql
var sqlFS embed.FS

func versionFromName(name string) string {
    n := name
    if i := strings.IndexByte(n, '_'); i >= 0 { n = n[:i] }
    if j := strings.Index(n, ".sql"); j >= 0 { n = n[:j] }
    return n
}

func listSQLNames() ([]string, error) {
    entries, err := fs.ReadDir(sqlFS, "sql")
    if err != nil { return nil, err }
    names := make([]string, 0, len(entries))
    for _, e := range entries {
        if e.IsDir() { continue }
        if !strings.HasSuffix(e.Name(), ".sql") { continue }
        names = append(names, e.Name())
    }
    sort.Strings(names)
    return names, nil
}

func Apply(ctx context.Context, db *sqlx.DB) error {
    names, err := listSQLNames()
    if err != nil { return err }

    tx, err := db.BeginTxx(ctx, nil)
    if err != nil { return err }
    // ensure ledger
    if _, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations(
        version TEXT PRIMARY KEY,
        applied_at INTEGER NOT NULL
    )`); err != nil { _ = tx.Rollback(); return err }

    var applied []string
    if err := tx.SelectContext(ctx, &applied, `SELECT version FROM schema_migrations`); err != nil { _ = tx.Rollback(); return err }
    done := make(map[string]struct{}, len(applied))
    for _, v := range applied { done[v] = struct{}{} }

    for _, name := range names {
        v := versionFromName(name)
        if _, ok := done[v]; ok { continue }
        b, err := sqlFS.ReadFile("sql/" + name)
        if err != nil { _ = tx.Rollback(); return err }
        if _, err := tx.ExecContext(ctx, string(b)); err != nil { _ = tx.Rollback(); return err }
        if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations(version, applied_at) VALUES(?, ?)`, v, time.Now().Unix()); err != nil { _ = tx.Rollback(); return err }
    }
    if err := tx.Commit(); err != nil { return err }
    return nil
}

// Pending returns the list of migration versions that have not been applied yet.
func Pending(ctx context.Context, db *sqlx.DB) ([]string, error) {
    names, err := listSQLNames()
    if err != nil { return nil, err }
    // ensure ledger table existence check (do not create it here)
    // if ledger not exists, consider none applied
    var exists int
    if err := db.GetContext(ctx, &exists, `SELECT COUNT(1) FROM sqlite_master WHERE type='table' AND name='schema_migrations'`); err != nil {
        return nil, err
    }
    appliedSet := make(map[string]struct{})
    if exists > 0 {
        var applied []string
        if err := db.SelectContext(ctx, &applied, `SELECT version FROM schema_migrations`); err != nil { return nil, err }
        for _, v := range applied { appliedSet[v] = struct{}{} }
    }
    pend := make([]string, 0)
    for _, name := range names {
        v := versionFromName(name)
        if _, ok := appliedSet[v]; !ok { pend = append(pend, v) }
    }
    return pend, nil
}
