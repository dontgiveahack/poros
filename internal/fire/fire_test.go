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

func TestCoastFire(t *testing.T) {
	ledger := &store.Ledger{Transactions: []domain.Transaction{
		{ID: "e1", Date: dDate("2026-06-01"), Type: domain.TxExpense, Amount: amt(t, "30000 EUR"), Account: "bank/checking"},
	}}
	s, err := Calculate(ledger, Options{Year: 2026, CurrentAge: 35, RetirementAge: 67})
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	if s.CoastFire == nil {
		t.Fatal("coast should be set")
	}
	// 750000 / 1.05^32 ≈ 157k — assert roughly, not exact cents.
	f, _ := s.CoastFire.Rat().Float64()
	if f < 150000 || f > 165000 {
		t.Errorf("coast = %v, want ~157k", f)
	}
}

func TestLeanFat(t *testing.T) {
	ledger := &store.Ledger{}
	lean := mustAmt(t, "20000 EUR")
	fat := mustAmt(t, "45000 EUR")
	s, err := Calculate(ledger, Options{Year: 2026, LeanExpenses: &lean, FatExpenses: &fat})
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	if s.LeanFire == nil || s.LeanFire.String() != "500000 EUR" { // 20000/0.04
		t.Errorf("lean = %v", s.LeanFire)
	}
	if s.FatFire == nil || s.FatFire.String() != "1125000 EUR" { // 45000/0.04
		t.Errorf("fat = %v", s.FatFire)
	}
}

func mustAmt(t *testing.T, s string) domain.Amount {
	t.Helper()
	a, err := domain.ParseAmount(s)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return a
}

func dDate(s string) domain.Date {
	t, _ := time.Parse("2006-01-02", s)
	return domain.Date{Time: t}
}

func TestSimulateZeroVolatility(t *testing.T) {
	got := simulate(0, 10000, 0, 0.05, 0, 10, 50, 7, "EUR")
	f, _ := got.P50.Rat().Float64()
	if f < 125700 || f > 125900 {
		t.Errorf("P50 = %v, want ~125779", f)
	}
	if got.P10.Rat().Cmp(got.P90.Rat()) != 0 {
		t.Error("zero volatility must give identical percentiles")
	}
}

func TestSimulateOrdering(t *testing.T) {
	got := simulate(10000, 5000, 0, 0.05, 0.15, 20, 1000, 42, "EUR")
	if got.P10.Rat().Cmp(got.P50.Rat()) > 0 || got.P50.Rat().Cmp(got.P90.Rat()) > 0 {
		t.Error("percentiles out of order")
	}
	if got.ProbFire < 0 || got.ProbFire > 1 {
		t.Errorf("prob = %v", got.ProbFire)
	}
}
