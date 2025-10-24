package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"mss/internal/api"
	"mss/internal/migrate"
	"mss/internal/ui"
	"mss/internal/store"
)

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" { return def }
	return v
}

func main() {
	addr := getenv("MSS_LISTEN_ADDR", ":8080")
	dbPath := getenv("MSS_DB_PATH", "./data/mss.db")
	autoMigrate := getenv("MSS_AUTO_MIGRATE", "0") == "1"

	// Check DB file existence BEFORE opening sqlite (which would create the file).
	needInit := false
	if _, err := os.Stat(dbPath); err != nil {
		if os.IsNotExist(err) {
			needInit = true
		} else {
			log.Fatalf("stat db: %v", err)
		}
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		log.Fatalf("mkdir data: %v", err)
	}

	db, err := store.Open(dbPath)
	if err != nil { log.Fatalf("open db: %v", err) }
	defer func() { _ = db.Close() }()

	if needInit {
		log.Printf("database not found at %s; initializing schema (auto-migrate forced)", dbPath)
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		if err := migrate.Apply(ctx, db); err != nil { log.Fatalf("migrate: %v", err) }
	} else if autoMigrate {
		log.Printf("existing database found; MSS_AUTO_MIGRATE=1 -> applying migrations")
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		if err := migrate.Apply(ctx, db); err != nil { log.Fatalf("migrate: %v", err) }
	} else {
		log.Printf("existing database found; skipping migrations (MSS_AUTO_MIGRATE=0). If schema changes are required, enable MSS_AUTO_MIGRATE=1 and restart.")
		// Log pending migrations for manual operation visibility
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if pend, err := migrate.Pending(ctx, db); err != nil {
			log.Printf("migrate: pending check failed: %v", err)
		} else if len(pend) > 0 {
			log.Printf("migrate: pending versions: %s", strings.Join(pend, ","))
		} else {
			log.Printf("migrate: no pending migrations")
		}
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) }))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, "<!doctype html><html><head><meta charset=\"utf-8\"><title>Multi Site Switcher Server</title></head><body><h1>Multi Site Switcher Server</h1><p>API at <a href=\"/api/sites\">/api/sites</a>. Health at <a href=\"/healthz\">/healthz</a>.</p><p>UI at <a href=\"/ui/\">/ui/</a>.</p></body></html>")
	})

	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	apiRouter := api.NewRouter(db)
	r.Mount("/api", apiRouter)

	// minimal server-side rendered UI
	r.Mount("/ui", ui.NewRouter(db))

	srv := &http.Server{ Addr: addr, Handler: r }
	log.Printf("mss-server listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %v", err)
	}
}
