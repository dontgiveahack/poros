package domain

import (
	"testing"

	"math/big"
)

func TestParseAmount(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string // canonical String() output
	}{
		{"expensive", "54.32 EUR", "54.32 EUR"},
		{"negative", "-1200 USD", "-1200 USD"},
		{"fractional asset", "0.002 BTC", "0.002 BTC"},
		{"integer quantity", "5 VWCE", "5 VWCE"},
		{"trailing zeros trimmed", "5.00 EUR", "5 EUR"},
		{"plus sign", "+2500 EUR", "2500 EUR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAmount(tt.in)
			if err != nil {
				t.Fatalf("ParseAmount(%q) unexpected error: %v", tt.in, err)
			}

			if got.String() != tt.want {
				t.Errorf("ParseAmount(%q) = %q, want %q", tt.in, got.String(), tt.want)
			}
		})
	}
}

func TestParseAmountInvalid(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"missing commodity", "54.32"},
		{"missing number", "EUR"},
		{"too many fields", "54.32 EUR extra"},
		{"not a number", "abc EUR"},
		{"lowercase commodity", "54.32 eur"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseAmount(tt.in); err == nil {
				t.Errorf("ParseAmount(%q) = nil error, want error", tt.in)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	a, _ := ParseAmount("10.10 EUR")
	b, _ := ParseAmount("0.20 EUR")

	sum, err := a.Add(b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 10.10 + 0.20 is exactly 10.3. Trailing zeros are trimmed by design.
	if want := "10.3 EUR"; sum.String() != want {
		t.Errorf("sum = %q, want %q", sum.String(), want)
	}
}

func TestAddMismatchedCommodities(t *testing.T) {
	eur, _ := ParseAmount("10 EUR")
	vwce, _ := ParseAmount("5 VWCE")

	if _, err := eur.Add(vwce); err == nil {
		t.Error("adding EUR to VWCE should fail, got nil error")
	}
}

func TestNeg(t *testing.T) {
	a, _ := ParseAmount("54.32 EUR")

	if got := a.Neg().String(); got != "-54.32 EUR" {
		t.Errorf("Neg() = %q, want %q", got, "-54.32 EUR")
	}
}

func TestNewAmountCopiesValue(t *testing.T) {
	v := big.NewRat(100, 1)
	a := NewAmount(v, "EUR")

	v.SetInt64(999) // mutate the original

	if got := a.String(); got != "100 EUR" {
		t.Errorf("NewAmount did not copy: got %q, want %q", got, "100 EUR")
	}
}
