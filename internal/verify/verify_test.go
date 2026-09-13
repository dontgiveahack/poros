package verify

import (
	"testing"

	"github.com/dontgiveahack/poros/internal/domain"
	"github.com/dontgiveahack/poros/internal/store"
)

func TestDiffEmpty(t *testing.T) {
	a := &store.Ledger{}
	b := &store.Ledger{}

	diffs, err := Diff(a, b)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}

	if len(diffs) != 0 {
		t.Fatalf("empty diff = %v", diffs)
	}
}

func TestDiffMissing(t *testing.T) {
	a := &store.Ledger{Accounts: []domain.Account{{ID: "bank/checking", Type: "bank", Currency: "EUR"}}}
	b := &store.Ledger{}

	diffs, err := Diff(a, b)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}

	if len(diffs) == 0 {
		t.Fatal("expected missing diff")
	}
}
