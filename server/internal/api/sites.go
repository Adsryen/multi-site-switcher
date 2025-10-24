package api

import (
	"encoding/json"
	"net/http"
	"strings"

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

type siteReq struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	LoginURL string `json:"loginUrl"`
}

func (a *API) createSite(w http.ResponseWriter, r *http.Request) {
	var body siteReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil { fail(w, http.StatusBadRequest, err); return }
	body.Key = strings.TrimSpace(body.Key)
	body.Name = strings.TrimSpace(body.Name)
	if body.Key == "" || body.Name == "" { fail(w, http.StatusBadRequest, nil); return }
	s := &store.Site{ Key: body.Key, Name: body.Name, LoginURL: body.LoginURL }
	if err := store.CreateSite(r.Context(), a.db, s); err != nil { fail(w, http.StatusInternalServerError, err); return }
	ok(w, s)
}

func (a *API) updateSite(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	var body siteReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil { fail(w, http.StatusBadRequest, err); return }
	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" { fail(w, http.StatusBadRequest, nil); return }
	s := &store.Site{ Key: key, Name: body.Name, LoginURL: body.LoginURL }
	if err := store.UpdateSite(r.Context(), a.db, s); err != nil { fail(w, http.StatusInternalServerError, err); return }
	ok(w, s)
}

func (a *API) deleteSite(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" { fail(w, http.StatusBadRequest, nil); return }
	if err := store.DeleteSite(r.Context(), a.db, key); err != nil { fail(w, http.StatusInternalServerError, err); return }
	ok(w, map[string]string{"status":"deleted"})
}
