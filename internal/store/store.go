// Package store loads Póros JSON documents from disk into the domain model.
package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dontgiveahack/poros/internal/domain"
)

// Ledger is the in-memory aggregate of all Póros data documents.
type Ledger struct {
	Accounts     []domain.Account
	Transactions []domain.Transaction
	Assets       []domain.Asset
	Goals        []domain.Goal
}

// LoadDir reads accounts.json, transactions.json, assets.json, goals.json
// from dir and validates every record.
func LoadDir(dir string) (*Ledger, error) {
	l := &Ledger{}

	// Accounts
	if err := loadFile(filepath.Join(dir, "accounts.json"),
	                   &l.Accounts,
		           func(a domain.Account) error { return a.Validate() },
	); err != nil {
		return nil, err
	}

	// Transactions
	if err := loadFile(filepath.Join(dir, "transactions.json"),
	                   &l.Transactions,
			   func(t domain.Transaction) error { return t.Validate() },
	); err != nil {
		return nil, err
	}

	// Assets
	if err := loadFile(filepath.Join(dir, "assets.json"),
	                   &l.Assets,
			   func(a domain.Asset) error { return a.Validate() },
	); err != nil {
		return nil, err
	}

	// Goals
	if err := loadFile(filepath.Join(dir, "goals.json"),
	                   &l.Goals,
			   func(g domain.Goal) error { return g.Validate() },
	); err != nil {
		return nil, err
	}

	return l, nil
}

func loadFile[T any](path string, dst *[]T, validate func(T) error ) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("read %s: %w", path, err)
	}

	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}

	for i, item := range *dst {
		if err := validate(item); err != nil {
			return fmt.Errorf("%s[%d]: %w", path, i, err)
		}
	}

	return nil
}
