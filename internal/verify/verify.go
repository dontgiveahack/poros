// Package verify compares the file ledger with the JSONB mirror.
package verify

import (
	"encoding/json"
	"fmt"

	"github.com/dontgiveahack/poros/internal/store"
)

// Diff returns human-readable differences.
func Diff(a, b *store.Ledger) []string {
	var out []string
	out = append(out, diffTable("accounts", toMap(a.Accounts), toMap(b.Accounts))...)
	out = append(out, diffTable("assets", toMap(a.Assets), toMap(b.Assets))...)
	out = append(out, diffTable("transactions", toMap(a.Transactions), toMap(b.Transactions))...)
	out = append(out, diffTable("goals", toMap(a.Goals), toMap(b.Goals))...)
	return out
}

type withID interface{ GetId() string }

// Helper: convert slice of structs with ID field to map[id]rawJSON.
func toMap[T any](items []T) map[string]json.RawMessage {
	m := make(map[string]json.RawMessage, len(items))
	for _, it := range items {
		raw, _ := json.Marshal(it)
		var tmp struct{ ID string `json:"id"` }
		json.Unmarshal(raw, &tmp)
		if tmp.ID == "" {
			tmp.ID = fmt.Sprintf("%p", &it)
		}

		m[tmp.ID] = raw
	}

	return m
}

func diffTable(name string, a, b map[string]json.RawMessage) []string {
	var out []string
	for id, av := range a {
		bv, ok := b[id]
		if !ok {
			out = append(out, fmt.Sprintf("%s: missing in DB: %s", name, id))
			continue
		}

		if string(av) != string(bv) {
			out = append(out, fmt.Sprintf("%s: mismatch %s", name, id))
		}
	}

	for id := range b {
		if _, ok := a[id]; !ok {
			out = append(out, fmt.Sprintf("%s: extra in DB: %s", name, id))
		}
	}

	return out
}
