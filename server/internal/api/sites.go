package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"mss/internal/store"
)

func (a *API) listSites(w http.ResponseWriter, r *http.Request) {
	sites, err := store.ListSites(r.Context(), a.db)
	if err != nil { fail(w, http.StatusInternalServerError, err); return }
	ok(w, sites)
}

func (a *API) getSite(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	s, err := store.GetSite(r.Context(), a.db, key)
	if err != nil { fail(w, http.StatusInternalServerError, err); return }
	if s == nil { fail(w, http.StatusNotFound, nil); return }
	ok(w, s)
}
