package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"mss/internal/store"
)

type schemaFieldReq struct {
	Field    string      `json:"field"`
	Type     string      `json:"type"`
	Required bool        `json:"required"`
	Default  interface{} `json:"default"`
	Regex    string      `json:"regex"`
	Choices  interface{} `json:"choices"`
	Secret   bool        `json:"secret"`
	Order    int         `json:"order"`
	UIHint   string      `json:"uiHint"`
}

type schemaFieldResp struct {
	Field    string      `json:"field"`
	Type     string      `json:"type"`
	Required bool        `json:"required"`
	Default  interface{} `json:"default"`
	Regex    string      `json:"regex"`
	Choices  interface{} `json:"choices"`
	Secret   bool        `json:"secret"`
	Order    int         `json:"order"`
	UIHint   string      `json:"uiHint"`
}

func toStoreSchema(siteKey string, f schemaFieldReq) store.SiteFieldSchema {
	var defStr string
	if f.Default != nil {
		if b, err := json.Marshal(f.Default); err == nil { defStr = string(b) }
	}
	var choicesStr string
	if f.Choices != nil {
		if b, err := json.Marshal(f.Choices); err == nil { choicesStr = string(b) }
	}
	req := 0
	if f.Required { req = 1 }
	sec := 0
	if f.Secret { sec = 1 }
	return store.SiteFieldSchema{
		SiteKey:      siteKey,
		Field:        f.Field,
		Type:         f.Type,
		Required:     req,
		DefaultValue: defStr,
		Regex:        f.Regex,
		Choices:      choicesStr,
		Secret:       sec,
		Order:        f.Order,
		UIHint:       f.UIHint,
	}
}

func toRespSchema(s store.SiteFieldSchema) schemaFieldResp {
	var def interface{}
	if s.DefaultValue != "" { _ = json.Unmarshal([]byte(s.DefaultValue), &def) }
	var ch interface{}
	if s.Choices != "" { _ = json.Unmarshal([]byte(s.Choices), &ch) }
	return schemaFieldResp{
		Field: s.Field,
		Type: s.Type,
		Required: s.Required != 0,
		Default: def,
		Regex: s.Regex,
		Choices: ch,
		Secret: s.Secret != 0,
		Order: s.Order,
		UIHint: s.UIHint,
	}
}

func (a *API) getSchema(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	items, err := store.GetSiteFieldSchemas(r.Context(), a.db, key)
	if err != nil { fail(w, http.StatusInternalServerError, err); return }
	out := make([]schemaFieldResp, 0, len(items))
	for _, it := range items { out = append(out, toRespSchema(it)) }
	ok(w, out)
}

func (a *API) postSchema(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	var body struct{ Fields []schemaFieldReq `json:"fields"` }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil { fail(w, http.StatusBadRequest, err); return }
	for _, f := range body.Fields {
		if f.Field == "" || f.Type == "" { fail(w, http.StatusBadRequest, nil); return }
		m := toStoreSchema(key, f)
		if err := store.UpsertSiteFieldSchema(r.Context(), a.db, &m); err != nil { fail(w, http.StatusInternalServerError, err); return }
	}
	items, err := store.GetSiteFieldSchemas(r.Context(), a.db, key)
	if err != nil { fail(w, http.StatusInternalServerError, err); return }
	out := make([]schemaFieldResp, 0, len(items))
	for _, it := range items { out = append(out, toRespSchema(it)) }
	ok(w, out)
}

func (a *API) putSchema(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	field := chi.URLParam(r, "field")
	var f schemaFieldReq
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil { fail(w, http.StatusBadRequest, err); return }
	if f.Field == "" { f.Field = field }
	if f.Field != field { fail(w, http.StatusBadRequest, nil); return }
	if f.Field == "" || f.Type == "" { fail(w, http.StatusBadRequest, nil); return }
	m := toStoreSchema(key, f)
	if err := store.UpsertSiteFieldSchema(r.Context(), a.db, &m); err != nil { fail(w, http.StatusInternalServerError, err); return }
	items, err := store.GetSiteFieldSchemas(r.Context(), a.db, key)
	if err != nil { fail(w, http.StatusInternalServerError, err); return }
	out := make([]schemaFieldResp, 0, len(items))
	for _, it := range items { out = append(out, toRespSchema(it)) }
	ok(w, out)
}

func (a *API) deleteSchema(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	field := chi.URLParam(r, "field")
	if field == "" { fail(w, http.StatusBadRequest, nil); return }
	if err := store.DeleteSiteFieldSchema(r.Context(), a.db, key, field); err != nil { fail(w, http.StatusInternalServerError, err); return }
	ok(w, map[string]string{"status":"deleted"})
}
