// Package store persistence helpers: append transactions to data/*.json.
package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dontgiveahack/poros/internal/domain"
)

// SaveTransactions rewrites transactions.json atomically:
// write to temp file in the same dir, then rename over the original,
// so a crash mid-write never leaves a half-written ledger.
func SaveTransactions(dir string, txs []domain.Transaction) error {
	for i, tx := range txs {
		if err := tx.Validate(); err != nil {
			return fmt.Errorf("transactions[%d]: %w", i, err)
		}
	}

	data, err := json.MarshalIndent(txs, "", "  ")
	if err != nil {
		return fmt.Errorf("encode transactions: %w", err)
	}

	data = append(data, '\n')
	path := filepath.Join(dir, "transactions.json")
	tmp, err := os.CreateTemp(dir, "transactions-*.json")
	if err != nil {
		return fmt.Errorf("temp file: %w", err)
	}

	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write temp: %w", err)
	}

	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close temp: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("replace %s: %w", path, err)
	}

	return nil // placeholder, replaced below
}

// AppendTransaction loads the ledger, appends one validated transaction
// and saves it back. It returns the assigned ID.
func AppendTransaction(dir string, tx domain.Transaction) (string, error) {
	l, err := LoadDir(dir)
	if err != nil {
		return "", err
	}

	// Check if the transaction already exists.
	for _, existing := range l.Transactions {
		if existing.ID == tx.ID {
			return "", fmt.Errorf("duplicate transaction id %q", tx.ID)
		}
	}

	if err := tx.Validate(); err != nil {
		return "", err
	}

	l.Transactions = append(l.Transactions, tx)
	if err := SaveTransactions(dir, l.Transactions); err != nil {
		return "", err
	}

	return tx.ID, nil
}

// NextID proposes "txn-YYYY-MM-DD-NNN" based on existing transactions
// for that date (deterministic, no randomness, no clock dependency beyond the
// date itself).
func NextID(txs []domain.Transaction, dateStr string) string {
	n := 0
	for _, tx := range txs {
		if tx.Date.Format("2006-01-02") == dateStr {
			n++
		}
	}

	return fmt.Sprintf("txn-%s-%03d", dateStr, n+1)
}
