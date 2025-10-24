package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"mss/internal/store"
	"mss/internal/validation"
)

type accountReq struct {
	ID       string                 `json:"id,omitempty"`
	Username string                 `json:"username"`
	Password string                 `json:"password,omitempty"`
	Props    map[string]interface{} `json:"props,omitempty"`
}

type accountResp struct {
	ID        string                 `json:"id"`
	SiteKey   string                 `json:"siteKey"`
	Username  string                 `json:"username"`
	Password  string                 `json:"password,omitempty"`
	Props     map[string]interface{} `json:"props,omitempty"`
	CreatedAt int64                  `json:"createdAt"`
	UpdatedAt int64                  `json:"updatedAt"`
}

func toAccountResp(a store.Account) accountResp {
	var props map[string]interface{}
	if a.Extra != "" {
		_ = json.Unmarshal([]byte(a.Extra), &props)
	}
	return accountResp{
		ID: a.ID, SiteKey: a.SiteKey, Username: a.Username, Password: a.Password,
		Props: props, CreatedAt: a.Created, UpdatedAt: a.Updated,
	}
}

func (a *API) listAccounts(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	accs, err := store.ListAccounts(r.Context(), a.db, key)
	if err != nil { fail(w, http.StatusInternalServerError, err); return }
	activeId, err := store.GetActiveAccountID(r.Context(), a.db, key)
	if err != nil { fail(w, http.StatusInternalServerError, err); return }
	res := make([]accountResp, 0, len(accs))
	for _, it := range accs {
		ar := toAccountResp(it)
		if ar.Props != nil {
			if masked, err := validation.MaskSecretProps(r.Context(), a.db, key, ar.Props); err == nil { ar.Props = masked }
		}
		res = append(res, ar)
	}
	ok(w, map[string]interface{}{"accounts": res, "activeId": activeId})
}

func (a *API) createAccount(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	var body accountReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil { fail(w, http.StatusBadRequest, err); return }
	if body.Props != nil {
		if err := validation.ValidateProps(r.Context(), a.db, key, body.Props); err != nil { fail(w, http.StatusBadRequest, err); return }
	}
	acc := store.Account{ ID: body.ID, SiteKey: key, Username: body.Username, Password: body.Password }
	if acc.ID == "" { acc.ID = store.GenerateID("acc") }
	if body.Props != nil {
		b, _ := json.Marshal(body.Props)
		acc.Extra = string(b)
	}
	if err := store.CreateAccount(r.Context(), a.db, &acc); err != nil { fail(w, http.StatusInternalServerError, err); return }
	resp := toAccountResp(acc)
	if resp.Props != nil {
		if masked, err := validation.MaskSecretProps(r.Context(), a.db, key, resp.Props); err == nil { resp.Props = masked }
	}
	ok(w, resp)
}

func (a *API) updateAccount(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	id := chi.URLParam(r, "id")
	var body accountReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil { fail(w, http.StatusBadRequest, err); return }
	if body.Props != nil {
		if err := validation.ValidateProps(r.Context(), a.db, key, body.Props); err != nil { fail(w, http.StatusBadRequest, err); return }
	}
	acc := store.Account{ ID: id, SiteKey: key, Username: body.Username, Password: body.Password }
	if body.Props != nil {
		b, _ := json.Marshal(body.Props)
		acc.Extra = string(b)
	}
	if err := store.UpdateAccount(r.Context(), a.db, &acc); err != nil { fail(w, http.StatusInternalServerError, err); return }
	resp := toAccountResp(acc)
	if resp.Props != nil {
		if masked, err := validation.MaskSecretProps(r.Context(), a.db, key, resp.Props); err == nil { resp.Props = masked }
	}
	ok(w, resp)
}

func (a *API) deleteAccount(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	id := chi.URLParam(r, "id")
	if err := store.DeleteAccount(r.Context(), a.db, key, id); err != nil { fail(w, http.StatusInternalServerError, err); return }
	ok(w, map[string]string{"status":"deleted"})
}
