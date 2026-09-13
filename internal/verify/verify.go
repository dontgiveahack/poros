// Package verify compares the file ledger with the JSONB mirror.
package verify

import (
	"encoding/json"
	"fmt"

	"github.com/dontgiveahack/poros/internal/store"
)

// Diff returns human-readable differences.
func Diff(a, b *store.Ledger) ([]string, error) {
	var out []string
	for _, part := range []struct {
		name string
		x, y map[string]json.RawMessage
	}{} {
		_ = part
	}

	accA, err := toMap(a.Accounts)
	if err != nil {
		return nil, fmt.Errorf("accounts: %w", err)
	}

	accB, err := toMap(b.Accounts)
	if err != nil {
		return nil, fmt.Errorf("accounts: %w", err)
	}

	out = append(out, diffTable("accounts", accA, accB)...)

	txA, err := toMap(a.Transactions)
	if err != nil {
		return nil, fmt.Errorf("transactions: %w", err)
	}

	txB, err := toMap(b.Transactions)
	if err != nil {
		return nil, fmt.Errorf("transactions: %w", err)
	}

	out = append(out, diffTable("transactions", txA, txB)...)

	goA, err := toMap(a.Goals)
	if err != nil {
		return nil, fmt.Errorf("goals: %w", err)
	}

	goB, err := toMap(b.Goals)
	if err != nil {
		return nil, fmt.Errorf("goals: %w", err)
	}

	out = append(out, diffTable("goals", goA, goB)...)

	asA, err := toMap(a.Assets)
	if err != nil {
		return nil, fmt.Errorf("assets: %w", err)
	}

	asB, err := toMap(b.Assets)
	if err != nil {
		return nil, fmt.Errorf("assets: %w", err)
	}

	out = append(out, diffTable("assets", asA, asB)...)

	return out, nil
}

// toMap indexes records by their JSON id field.
func toMap[T any](items []T) (map[string]json.RawMessage, error) {
	m := make(map[string]json.RawMessage, len(items))
	for i, it := range items {
		raw, err := json.Marshal(it)
		if err != nil {
			return nil, fmt.Errorf("[%d]: %w", i, err)
		}

		var tmp struct{
			ID string `json:"id"`
		}
		if err := json.Unmarshal(raw, &tmp); err != nil {
			return nil, fmt.Errorf("[%d]: %w", i, err)
		}

		if tmp.ID == "" {
			return nil, fmt.Errorf("[%d]: missing id", i)
		}

		m[tmp.ID] = raw
	}

	return m, nil
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
