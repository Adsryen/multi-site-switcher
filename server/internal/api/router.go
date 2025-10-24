package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

type Response struct {
	Ok    bool        `json:"ok"`
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func ok(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, Response{Ok: true, Data: data})
}

func fail(w http.ResponseWriter, status int, err error) {
	msg := ""
	if err != nil { msg = err.Error() }
	writeJSON(w, status, Response{Ok: false, Error: msg})
}

type API struct {
	db *sqlx.DB
}

func NewRouter(db *sqlx.DB) http.Handler {
	a := &API{db: db}
	r := chi.NewRouter()

	r.Get("/sites", a.listSites)
	r.Get("/sites/{key}", a.getSite)
	r.Post("/sites", a.createSite)
	r.Put("/sites/{key}", a.updateSite)
	r.Delete("/sites/{key}", a.deleteSite)

	// site field schemas
	r.Get("/sites/{key}/schema", a.getSchema)
	r.Post("/sites/{key}/schema", a.postSchema)
	r.Put("/sites/{key}/schema/{field}", a.putSchema)
	r.Delete("/sites/{key}/schema/{field}", a.deleteSchema)

	r.Get("/sites/{key}/accounts", a.listAccounts)
	r.Post("/sites/{key}/accounts", a.createAccount)
	r.Put("/sites/{key}/accounts/{id}", a.updateAccount)
	r.Delete("/sites/{key}/accounts/{id}", a.deleteAccount)

	r.Get("/sites/{key}/active-account", a.getActiveAccount)
	r.Put("/sites/{key}/active-account", a.setActiveAccount)

	r.Post("/sites/{key}/switch", a.switchAccount)

	return r
}
