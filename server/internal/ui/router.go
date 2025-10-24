package ui

import (
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"

	"mss/internal/store"
)

type UI struct {
	db *sqlx.DB
	t *template.Template
}

func NewRouter(db *sqlx.DB) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	ui := &UI{db: db}
	ui.t = template.Must(template.ParseFiles(
		"internal/ui/templates/layout.html",
		"internal/ui/templates/sites.html",
	))
	r.Get("/", ui.sitesPage)
	r.Post("/sites", ui.createSite)
	return r
}

func (u *UI) sitesPage(w http.ResponseWriter, r *http.Request) {
	sites, err := store.ListSites(r.Context(), u.db)
	if err != nil { http.Error(w, err.Error(), 500); return }
	data := map[string]interface{}{
		"Sites": sites,
	}
	_ = u.t.ExecuteTemplate(w, "layout", data)
}

func (u *UI) createSite(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil { http.Error(w, err.Error(), 400); return }
	s := &store.Site{ Key: r.FormValue("key"), Name: r.FormValue("name"), LoginURL: r.FormValue("loginUrl") }
	if s.Key == "" || s.Name == "" { http.Error(w, "key and name required", 400); return }
	if err := store.CreateSite(r.Context(), u.db, s); err != nil { http.Error(w, err.Error(), 500); return }
	http.Redirect(w, r, "/ui/", http.StatusSeeOther)
}
