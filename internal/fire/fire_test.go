package fire

import (
	"testing"
	"time"

	"github.com/dontgiveahack/poros/internal/domain"
	"github.com/dontgiveahack/poros/internal/store"
)

func amt(t *testing.T, s string) *domain.Amount {
	t.Helper()
	a, err := domain.ParseAmount(s)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}

	return &a
}

func TestCalculateGolden(t *testing.T) {
	ledger := &store.Ledger{Transactions: []domain.Transaction{
		{ID: "i1", Date: dDate("2026-01-15"), Type: domain.TxIncome, Amount: amt(t, "60000 EUR"), Account: "bank/checking"},
		{ID: "e1", Date: dDate("2026-06-01"), Type: domain.TxExpense, Amount: amt(t, "30000 EUR"), Account: "bank/checking"},
	}}

	s, err := Calculate(ledger, Options{Year: 2026})
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}

	if s.SavingsRate != 0.5 {
		t.Errorf("rate = %v, want 0.5", s.SavingsRate)
	}

	if got := s.FireNumber.String(); got != "750000 EUR" { // 30000/0.04
		t.Errorf("fire = %q", got)
	}

	if s.YearsToFire <= 0 {
		t.Errorf("years = %v, want > 0", s.YearsToFire)
	}
}

func TestCalculateZeroIncome(t *testing.T) {
	ledger := &store.Ledger{}
	s, err := Calculate(ledger, Options{Year: 2026})
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}

	if s.SavingsRate != 0 || s.YearsToFire != 0 {
		t.Errorf("got rate=%v years=%v", s.SavingsRate, s.YearsToFire)
	}
}

func TestCalculateNegativeSavings(t *testing.T) {
	ledger := &store.Ledger{Transactions: []domain.Transaction{
		{ID: "i1", Date: dDate("2026-01-15"), Type: domain.TxIncome, Amount: amt(t, "10000 EUR"), Account: "bank/checking"},
		{ID: "e1", Date: dDate("2026-06-01"), Type: domain.TxExpense, Amount: amt(t, "20000 EUR"), Account: "bank/checking"},
	}}
	s, err := Calculate(ledger, Options{Year: 2026})
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}

	if s.YearsToFire != -1 {
		t.Errorf("years = %v, want -1 (unreachable)", s.YearsToFire)
	}
}

func dDate(s string) domain.Date {
	t, _ := time.Parse("2006-01-02", s)
	return domain.Date{Time: t}
}
