package store

import (
	"testing"

	"github.com/dontgiveahack/poros/internal/domain"
)

func TestAppendTransaction(t *testing.T) {
	dir := t.TempDir()
	tx := mustTx(t, "txn-2026-10-04-001", "2026-10-04", domain.TxExpense, "54.32 EUR", "bank/checking")
	id, err := AppendTransaction(dir, tx)
	if err != nil {
		t.Fatalf("Append: %v", err)
	}

	if id != tx.ID {
		t.Fatalf("id = %q, want %q", id, tx.ID)
	}

	l, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}

	if len(l.Transactions) != 1 || l.Transactions[0].Amount.String() != "54.32 EUR" {
		t.Fatalf("ledger = %+v", l.Transactions)
	}
}

func TestAppendDuplicateID(t *testing.T) {
	dir := t.TempDir()
	tx := mustTx(t, "dup", "2026-10-04", domain.TxExpense, "10 EUR", "bank/checking")
	if _, err := AppendTransaction(dir, tx); err != nil {
		t.Fatalf("first Append: %v", err)
	}

	if _, err := AppendTransaction(dir, tx); err == nil {
		t.Fatal("duplicate ID should fail")
	}
}

func TestNextID(t *testing.T) {
	txs := []domain.Transaction{
		{ID: "txn-2026-10-04-001", Date: mustDate("2026-10-04")},
		{ID: "txn-2026-10-04-002", Date: mustDate("2026-10-04")},
		{ID: "txn-2026-10-03-001", Date: mustDate("2026-10-03")},
	}

	if got := NextID(txs, "2026-10-04"); got != "txn-2026-10-04-003" {
		t.Errorf("NextID = %q", got)
	}
}

func mustTx(t *testing.T, id, date string, typ domain.TxType, amount, account string) domain.Transaction {
	t.Helper()
	a, err := domain.ParseAmount(amount)
	if err != nil {
		t.Fatalf("parse amount: %v", err)
	}

	d, err := domain.ParseDate(date)
	if err != nil {
		t.Fatalf("parse date: %v", err)
	}

	return domain.Transaction{ID: id, Date: d, Type: typ, Amount: &a, Account: account}
}

func mustDate(s string) domain.Date {
	d, err := domain.ParseDate(s)
	if err != nil {
		panic(err)
	}

	return d
}
