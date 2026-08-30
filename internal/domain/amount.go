// Package domain holds the Póros core model:
// accounts, transacions and the value types
package domain

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
)

// Commodity identifies a currency or an asset unit: "EUR", "USD", "VWCE", "BTC".
type Commodity string

// Amount is a quantity of a commodity with exact rational precision.
// The zero value is NOT usabe. Build amounts with NewAmount or ParseAmount
type Amount struct {
	value *big.Rat
	commodity Commodity
}

// NewAmount build an Amount from an exact rational value.
// The receiver keeps its own copy of value.
func NewAmount(value *big.Rat, c Commodity) Amount {
	return Amount{
		value: new(big.Rat).Set(value),
		commodity: c,
	}
}

// ParseAmount parses strings like "54.32 EUR".
func ParseAmount(s string) (Amount, error) {
	fields := strings.Fields(s)
	if len(fields) != 2 {
		return Amount{}, fmt.Errorf("invalid amount %q: want \"<number> <commodity>\"", s)
	}

	value, ok := new(big.Rat).SetString(fields[0])
	if !ok {
		return Amount{}, fmt.Errorf("invalid number %q in amount %q", fields[0], s)
	}

	commodity := Commodity(fields[1])
	if !validCommodity(commodity) {
		return Amount{}, fmt.Errorf("invalid commodity %q in amount %q", fields[1], s)
	}

	return Amount{
		value: value,
		commodity: commodity,
	}, nil
}

// validCommodity accepts uppercase alphanumeric identifiers: EUR, VWCE, BTC...
func validCommodity(c Commodity) bool {
	s := string(c)
	if s == "" {
		return false
	}

	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		default: return false
		}
	}

	return true
}

// Commodity returns the amount's commodity.
func (a Amount) Commodity() Commodity { return a.commodity }

// Rat returns a copy of the underlying rational value, so callers cannot
// mutate the internal state.
func (a Amount) Rat() *big.Rat { return new(big.Rat).Set(a.value) }

// Add returns the sum of two amounts.
// It reports an error when commodities differ: adding EUR to VWCE is a caller
// bug.
func (a Amount) Add(b Amount) (Amount, error) {
	if a.commodity != b.commodity {
		return Amount{}, fmt.Errorf("cannot add %s to %s", a.commodity, b.commodity)
	}

	return Amount{
		value: new(big.Rat).Add(a.value, b.value),
		commodity: a.commodity,
	}, nil
}

// Neg returns the amount with inverted sign.
func (a Amount) Neg() Amount {
	return Amount{
		value: new(big.Rat).Neg(a.value),
		commodity: a.commodity,
	}
}

// String renders the amount as "<number> <COMMODITY>", e.g. "54.32 EUR".
// Trailing zeros are trimmed: "5 VWCE", "0.002 BTC".
func (a Amount) String() string {
	if a.value == nil {
		return "0 " + string(a.commodity)
	}
	return trimDecimal(a.value.FloatString(8)) + " " + string(a.commodity)
}

// trimDecimal removes trailing zeros from a fixed-precision decimal string.
func trimDecimal(s string) string {
	s = strings.TrimRight(s, "0")
	return strings.TrimSuffix(s, ".")
}

// MarshalJSON renders ("value":"54.32","commodity":"EUR").
func (a Amount) MarshalJSON() ([]byte, error) {
	v := a.value
	if v == nil {
		v = new(big.Rat)
	}

	return json.Marshal(struct {
		Value     string    `json:"value"`
		Commodity Commodity `json:"commodity"`
	}{
		Value: trimDecimal(v.FloatString(8)),
		Commodity: a.commodity,
	})
}

// UnmarshalJSON parses ("value":"54.32","commodity":"EUR").
func (a *Amount) UnmarshalJSON(b []byte) error {
	var raw struct {
		Value     string    `json:"value"`
		Commodity Commodity `json:"commodity"`
	}

	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	parsed, err := ParseAmount(raw.Value + " " + string(raw.Commodity))
	if err != nil {
		return err
	}

	*a = parsed
	return nil
}
