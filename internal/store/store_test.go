package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dontgiveahack/poros/internal/domain"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func TestLoadDir(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "accounts.json", `[
	  {"id":"bank/checking","type":"bank","name":"Checking","currency":"EUR"}
	]`)

	writeFile(t, dir, "transactions.json", `[
	  {
	    "id":"txn-1","date":"2026-08-02","type":"expense","title":"Mercadona",
	    "amount":{"value":"54.32","commodity":"EUR"},
	    "account":"bank/checking","category":"food","tags":["food"]
	  },
	  {
	    "id":"txn-2","date":"2026-08-05","type":"transfer",
	    "amount":{"value":"500","commodity":"EUR"},
	    "from":"bank/checking","to":"bank/savings"
	  }
	]`)

	writeFile(t, dir, "goals.json", `[
	  {"id":"fi","title":"FI","state":"open","target":{"value":"750000","commodity":"EUR"},"date":"2035-01-01"}
	]`)

	l, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}

	if len(l.Accounts) != 1 || len(l.Transactions) != 2 || len(l.Goals) != 1 {
		t.Fatalf("unexpected counts: %+v", l)
	}

	if l.Transactions[0].Amount.String() != "54.32 EUR" {
		t.Errorf("amount = %q, want 54.32 EUR", l.Transactions[0].Amount.String())
	}
}

func TestLoadDirInvalidTransfer(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "transactions.json", `[
	  {"id":"bad","date":"2026-08-05","type":"transfer","amount":{"value":"500","commodity":"EUR"}}
	]`)

	if _, err := LoadDir(dir); err == nil {
		t.Fatal("transfer without from/to should fail")
	}
}

func TestLoadDirMissingFileOK(t *testing.T) {
	dir := t.TempDir()
	l, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("empty dir should load: %v", err)
	}

	if l == nil || len(l.Transactions) != 0 {
		t.Fatal("expected empty ledger")
	}
}

var _ = domain.TxExpense // keep domain imported if tree-shaken
