package domain

import (
	"testing"
	"time"
)

func mustParse(t *testing.T, s string) Amount {
	t.Helper()
	a, err := ParseAmount(s)
	if err != nil {
		t.Fatalf("mustParse %q: %v", s, err)
	}

	return a
}

func date(s string) Date {
	t, _ := time.Parse("2006-01-02", s)
	return Date{Time: t}
}

func TestCalculateBalances(t *testing.T) {
	txs := []Transaction{
		{ID: "t1", Date: date("2026-08-02"), Type: TxExpense,  Amount: mustParse(t, "54.32 EUR"), Account: "bank/checking"},
		{ID: "t2", Date: date("2026-08-05"), Type: TxTransfer, Amount: mustParse(t, "500 EUR"),   From: "bank/checking", To: "bank/savings"},
		{ID: "t3", Date: date("2026-08-06"), Type: TxIncome,   Amount: mustParse(t, "2500 EUR"),  Account: "bank/checking"},
	}

	bals, err := CalculateBalances(txs)
	if err != nil {
		t.Fatalf("CalculateBalances: %v", err)
	}

	want := "1945.68 EUR" // -54.32 -500 +2500
	if got := bals["bank/checking"]["EUR"].String(); got != want {
		t.Errorf("checking = %q, want %q", got, want)
	}

	if got := bals["bank/savings"]["EUR"].String(); got != "500 EUR" {
		t.Errorf("savings = %q, want 500 EUR", got)
	}
}

func TestCalculateBalancesBuy(t *testing.T) {
	price, _ := ParseAmount("134.52 EUR")
	txs := []Transaction{
		{ID: "b1", Date: date("2026-08-10"), Type: TxBuy, Asset: "VWCE", Quantity: "5", Price: &price, Account: "broker/ibkr"},
	}

	bals, err := CalculateBalances(txs)
	if err != nil {
		t.Fatalf("buy: %v", err)
	}

	if got := bals["broker/ibkr"]["EUR"].String(); got != "-672.6 EUR" {
		t.Errorf("EUR = %q, want -672.6 EUR", got)
	}

	if got := bals["broker/ibkr"]["VWCE"].String(); got != "5 VWCE" {
		t.Errorf("VWCE = %q, want 5 VWCE", got)
	}
}

func TestCalculateBalancesEmpty(t *testing.T) {
	bals, err := CalculateBalances(nil)
	if err != nil || len(bals) != 0 {
		t.Fatalf("empty: err=%v len=%d", err, len(bals))
	}
}
