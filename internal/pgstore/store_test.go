package pgstore

import (
	"context"
	"os"
	"testing"

	"github.com/dontgiveahack/poros/internal/domain"
	"github.com/dontgiveahack/poros/internal/store"
)

func dsn() string {
	if v := os.Getenv("POROS_DB"); v != "" {
		return v
	}

	return "postgres://poros:poros@localhost:5432/poros?sslmode=disable"
}

func TestSyncAndLoad(t *testing.T) {
	ctx := context.Background()
	s, err := New(ctx, dsn())
	if err != nil {
		t.Skipf("no db: %v (run: docker compose up -d)", err)
	}
	defer s.Close()

	if err := s.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// Clean slate
	s.db.ExecContext(ctx, `TRUNCATE accounts, assets, transactions, goals`)

	price, _ := domain.ParseAmount("10 EUR")
	ledger := &store.Ledger{
		Accounts: []domain.Account{{ID: "bank/checking", Type: "bank", Name: "Checking", Currency: "EUR"}},
		Transactions: []domain.Transaction{
			{ID: "t1", Type: "expense", Amount: mustParsePG(t, "10 EUR"), Account: "bank/checking"},
			{ID: "b1", Type: "buy", Asset: "VWCE", Quantity: "2", Price: &price, Account: "bank/checking"},
		},
	}

	if err := s.Sync(ctx, ledger); err != nil {
		t.Fatalf("sync: %v", err)
	}

	loaded, err := s.LoadLedger(ctx)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if len(loaded.Accounts) != 1 || len(loaded.Transactions) != 2 {
		t.Fatalf("loaded = %+v", loaded)
	}
}

func mustParsePG(t *testing.T, s string) domain.Amount {
	t.Helper()
	a, err := domain.ParseAmount(s)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}

	return a
}
