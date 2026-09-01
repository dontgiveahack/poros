// Package pgstroe mirrors Póros JSON documents into PostgreSQL JSONB.
package pgstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/dontgiveahack/poros/internal/domain"
	"github.com/dontgiveahack/poros/internal/store"
)

// Store is a JSONB mirror of the file ledger.
type Store struct {
	db *sql.DB
}

// New opens a connection and pings the database.
func New(ctx context.Context, connStr string) (*Store, error) {
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

// Migrate creates the JSONB tables if they do not exist.
func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
	  CREATE TABLE IF NOT EXISTS accounts     (id TEXT PRIMARY KEY, data JSONB NOT NULL);
	  CREATE TABLE IF NOT EXISTS assets       (id TEXT PRIMARY KEY, data JSONB NOT NULL);
	  CREATE TABLE IF NOT EXISTS transactions (id TEXT PRIMARY KEY, data JSONB NOT NULL);
	  CREATE TABLE IF NOT EXISTS goals        (id TEXT PRIMARY KEY, data JSONB NOT NULL);
	`)
	return err
}

// Sync upserts every document from the file ledger into JSONB.
func (s *Store) Sync(ctx context.Context, l *store.Ledger) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, a := range l.Accounts {
		if err := upsert(ctx, tx, "accounts", a.ID, a); err != nil {
			return err
		}
	}

	for _, a := range l.Assets {
		if err := upsert(ctx, tx, "assets", a.ID, a); err != nil {
			return err
		}
	}

	for _, t := range l.Transactions {
		if err := upsert(ctx, tx, "transactions", t.ID, t); err != nil {
			return err
		}
	}

	for _, g := range l.Goals {
		if err := upsert(ctx, tx, "goals", g.ID, g); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// LoadLedger reads the JSONB mirror back into a Ledger (for verify/tests).
func (s *Store) LoadLedger(ctx context.Context) (*store.Ledger, error) {
	l := &store.Ledger{}
	var err error
	if l.Accounts, err = loadTable[domain.Account](ctx, s.db, "accounts"); err != nil {
		return nil, err
	}

	if l.Assets, err = loadTable[domain.Asset](ctx, s.db, "assets"); err != nil {
		return nil, err
	}

	if l.Transactions, err = loadTable[domain.Transaction](ctx, s.db, "transactions"); err != nil {
		return nil, err
	}

	if l.Goals, err = loadTable[domain.Goal](ctx, s.db, "goals"); err != nil {
		return nil, err
	}

	return l, nil
}

func upsert(ctx context.Context, tx *sql.Tx, table, id string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	q := fmt.Sprintf(`
	  INSERT INTO %s (id, data) VALUES ($1, $2::jsonb)
	    ON CONFLICT (id) DO UPDATE SET data = EXCLUDED.data`, table)
	_, err = tx.ExecContext(ctx, q, id, data)
	return err
}

func loadTable[T any](ctx context.Context, db *sql.DB, table string) ([]T, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`
	  SELECT data FROM %s ORDER BY id`, table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []T
	for rows.Next() {
		var raw json.RawMessage
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}

		var v T
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, err
		}

		out = append(out, v)
	}

	return out, rows.Err()
}
