package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/jmoiron/sqlx"

	"mss/internal/store"
)

func ValidateProps(ctx context.Context, db *sqlx.DB, siteKey string, props map[string]interface{}) error {
	schemas, err := store.GetSiteFieldSchemas(ctx, db, siteKey)
	if err != nil { return err }
	// build map for quick lookup
	sm := make(map[string]store.SiteFieldSchema, len(schemas))
	for _, s := range schemas { sm[s.Field] = s }

	// required check
	for _, s := range schemas {
		if s.Required != 0 {
			v, ok := props[s.Field]
			if !ok || isEmptyForType(v, s.Type) {
				return fmt.Errorf("invalid props: field '%s' required", s.Field)
			}
		}
	}
	// type/regex/choices check on present fields
	for k, v := range props {
		s, ok := sm[k]
		if !ok { continue } // allow extra fields not defined? choose to allow silently
		if !typeMatches(v, s.Type) {
			return fmt.Errorf("invalid props: field '%s' type mismatch, expect %s", k, s.Type)
		}
		if s.Regex != "" {
			str, ok := v.(string)
			if ok {
				re, err := regexp.Compile(s.Regex)
				if err != nil { return fmt.Errorf("invalid schema regex for field '%s'", k) }
				if !re.MatchString(str) { return fmt.Errorf("invalid props: field '%s' does not match regex", k) }
			}
		}
		if s.Choices != "" {
			var arr []interface{}
			if err := json.Unmarshal([]byte(s.Choices), &arr); err == nil {
				if !inChoices(v, arr) { return fmt.Errorf("invalid props: field '%s' not in choices", k) }
			}
		}
	}
	return nil
}

func MaskSecretProps(ctx context.Context, db *sqlx.DB, siteKey string, props map[string]interface{}) (map[string]interface{}, error) {
	schemas, err := store.GetSiteFieldSchemas(ctx, db, siteKey)
	if err != nil { return nil, err }
	return maskWithSchemas(schemas, props), nil
}

func maskWithSchemas(schemas []store.SiteFieldSchema, props map[string]interface{}) map[string]interface{} {
	pm := make(map[string]interface{}, len(props))
	for k, v := range props { pm[k] = v }
	for _, s := range schemas {
		if s.Secret != 0 {
			if _, ok := pm[s.Field]; ok { pm[s.Field] = "***" }
		}
	}
	return pm
}

func isEmptyForType(v interface{}, typ string) bool {
	switch typ {
	case "string", "datetime":
		if s, ok := v.(string); ok { return s == "" }
		return v == nil
	case "number":
		return v == nil
	case "boolean":
		return v == nil
	case "json":
		return v == nil
	default:
		return v == nil
	}
}

func typeMatches(v interface{}, typ string) bool {
	switch typ {
	case "string":
		_, ok := v.(string); return ok
	case "number":
		// encoding/json decodes numbers as float64
		_, ok := v.(float64); return ok
	case "boolean":
		_, ok := v.(bool); return ok
	case "datetime":
		s, ok := v.(string); if !ok { return false }
		if s == "" { return false }
		_, err := time.Parse(time.RFC3339, s)
		return err == nil
	case "json":
		// allow object or array
		if v == nil { return false }
		switch v.(type) {
		case map[string]interface{}, []interface{}:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func inChoices(v interface{}, arr []interface{}) bool {
	for _, c := range arr {
		if equalJSONValue(v, c) { return true }
	}
	return false
}

func equalJSONValue(a, b interface{}) bool {
	switch av := a.(type) {
	case string:
		bv, ok := b.(string); return ok && av == bv
	case float64:
		bv, ok := b.(float64); return ok && av == bv
	case bool:
		bv, ok := b.(bool); return ok && av == bv
	case nil:
		return b == nil
	default:
		return false
	}
}
