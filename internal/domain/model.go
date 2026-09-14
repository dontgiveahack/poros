// Package domain holds the Póros core model: accounts, transactions,
// assets, goals and the value types they are built from.
package domain

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"
)

// Date is a calendar date (YYYY-MM-DD) without time zone.
type Date struct {
	time.Time
}

// MarshalJSON encodes the date as "YYYY-MM-DD"
func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format("2006-01-02"))
}

// UnmarshalJSON decodes a "YYYY-MM-DD" date string.
func (d *Date) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}

	d.Time = t
	return nil
}

// --- Account ---

// AccountType classifies an account: bank, broker, crypto or cash.
type AccountType string

// Supported account types.
const (
	AccountBank AccountType = "bank"
	AccountBroker AccountType = "broker"
	AccountCrypto AccountType = "crypto"
	AccountCash AccountType = "cash"
)

// Account is a money holder (bank account, broker, wallet...).
// Its ID is a hierarchical path like "bank/checking".
type Account struct {
	ID       string      `json:"id"`
	Type     AccountType `json:"type"`
	Name     string      `json:"name"`
	Currency Commodity   `json:"currency"`
}

// Validate checks the account invariants.
func (a Account) Validate() error {
	if a.ID == "" {
		return fmt.Errorf("account: empty id")
	}

	if a.Currency == "" {
		return fmt.Errorf("account %q: empty currency", a.ID)
	}

	return nil
}

// --- Transaction ---

// TxType classifies a transaction.
type TxType string

// Supported transaction types.
const (
	TxIncome   TxType = "income"
	TxExpense  TxType = "expense"
	TxTransfer TxType = "transfer"
	TxBuy      TxType = "buy"
	TxSell     TxType = "sell"
	TxDividend TxType = "dividend"
	TxInterest TxType = "interest"
	TxFee      TxType = "fee"
)

// Transaction is a single financial event: income, expense, transfer,
// asset buy/sell, dividend, interest or fee.
type Transaction struct {
	ID       string    `json:"id"`
	Date     Date      `json:"date"`
	Type     TxType    `json:"type"`
	Title    string    `json:"title,omitempty"`
	Amount   *Amount   `json:"amount,omitempty"`
	Account  string    `json:"account,omitempty"`
	Category string    `json:"category,omitempty"`
	Tags     []string  `json:"tags,omitempty"`
	Note     string    `json:"note,omitempty"`

	// Double-entry / asset fields (optional per type).
	From     string    `json:"from,omitempty"`
	To       string    `json:"to,omitempty"`
	Asset    string    `json:"asset,omitempty"`
	Quantity string    `json:"quantity,omitempty"`
	Price    *Amount   `json:"price,omitempty"`
	Tax      *Amount   `json:"tax,omitempty"`
}

// QuantityRat parses the exact rational quantity (e.g. "5", "0.002").
func (t Transaction) QuantityRat() (*big.Rat, error) {
	if t.Quantity == "" {
		return nil, nil
	}

	r := new(big.Rat)
	if _, ok := r.SetString(t.Quantity); !ok {
		return nil, fmt.Errorf("tx %q: invalid quantity %q", t.ID, t.Quantity)
	}

	return r, nil
}

// Validate checks the transaction invariants for its type
// (amount required, from/to for transfers, asset/quantity/price for trades).
func (t Transaction) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("transaction: empty id")
	}

	if t.Date.IsZero() {
		return fmt.Errorf("tx %q: empty date", t.ID)
	}

	switch t.Type {
	case TxIncome, TxExpense, TxFee, TxDividend, TxInterest:
		if t.Amount == nil {
			return fmt.Errorf("tx %q: %s needs amount", t.ID, t.Type)
		}

	case TxTransfer:
		if t.Amount == nil {
			return fmt.Errorf("tx %q: transfer needs amount", t.ID)
		}

		if t.From == "" || t.To == "" {
			return fmt.Errorf("tx %q: transfer needs from and to", t.ID)
		}

	case TxBuy, TxSell:
		if t.Asset == "" {
			return fmt.Errorf("tx %q: %s needs asset", t.ID, t.Type)
		}

		if _, err := t.QuantityRat(); err != nil {
			return err
		}

		if t.Price == nil {
			return fmt.Errorf("tx %q: %s needs price", t.ID, t.Type)
		}
	}

	return nil
}

// --- Asset ---

// AssetClass classifies an asset: stock, ETF, bond, crypto, cash or other.
type AssetClass string

// Supported asset classes.
const (
	AssetStock  AssetClass = "stock"
	AssetETF    AssetClass = "etf"
	AssetBond   AssetClass = "bond"
	AssetCrypto AssetClass = "brypto"
	AssetCash   AssetClass = "cash"
	AssetOther  AssetClass = "other"
)

// Asset is an investable instrument (stock, ETF, crypto...).
type Asset struct {
	ID    string `json:"id"`
	Class string `json:"class"`
	Name  string `json:"name,omitempty"`
}

// Validate checks the asset invariants.
func (a Asset) Validate() error {
	if a.ID == "" {
		return fmt.Errorf("asset: empty id")
	}

	return nil
}

// --- Goal ---

// GoalState is the lifecycle state of a goal.
type GoalState string

// Supported goal states.
const (
	GoalOpen      GoalState = "open"
	GoalDone      GoalState = "done"
	GoalCancelled GoalState = "cancelled"
)

// Goal is a financial target with an optional deadline.
type Goal struct {
	ID     string    `json:"id"`
	Title  string    `json:"title"`
	State  GoalState `json:"state"`
	Target Amount    `json:"target"`
	Date   string    `json:"date,omitempty"`
}

// Validate checks the goal invariants.
func (g Goal) Validate() error {
	if g.ID == "" {
		return fmt.Errorf("goal: empty id")
	}

	switch g.State {
	case GoalOpen, GoalDone, GoalCancelled:
	default:
		return fmt.Errorf("goal %q: invalid state %q", g.ID, g.State)
	}

	return nil
}

// Price is a market quote for an asset on a date.
type Price struct {
	Asset string `json:"asset"`
	Date  Date   `json:"date"`
	Price Amount `json:"price"`
}

// Validate checks the price invariants.
func (p Price) Validate() error {
	if p.Asset == "" {
		return fmt.Errorf("price: empty asset")
	}

	if p.Date.IsZero() {
		return fmt.Errorf("price %q: empty date", p.Asset)
	}

	return nil
}
