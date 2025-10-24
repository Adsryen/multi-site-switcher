package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"mss/internal/store"
)

func (a *API) getActiveAccount(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	id, err := store.GetActiveAccountID(r.Context(), a.db, key)
	if err != nil { fail(w, http.StatusInternalServerError, err); return }
	ok(w, map[string]interface{}{"accountId": id})
}

func (a *API) setActiveAccount(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	var body struct{ AccountID *string `json:"accountId"` }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil { fail(w, http.StatusBadRequest, err); return }
	if err := store.SetActiveAccountID(r.Context(), a.db, key, body.AccountID); err != nil { fail(w, http.StatusInternalServerError, err); return }
	ok(w, map[string]interface{}{"accountId": body.AccountID})
}
