package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (a *API) switchAccount(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "key")
	// TODO: integrate chromedp automation here (logout->login) using service layer
	ok(w, map[string]string{"status":"switch_triggered"})
}
