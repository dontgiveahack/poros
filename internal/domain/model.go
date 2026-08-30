// Package domain holds the Póros core model: accounts, transactions,
// assets, goals and the value types they are built from.
package domain

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"
)

type Date struct {
	time.Time
}

func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Time.Format("2006-01-02"))
}

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

type AccountType string

const (
	AccountBank AccountType = "bank"
	AccountBroker AccountType = "broker"
	AccountCrypto AccountType = "crypto"
	AccountCash AccountType = "cash"
)

type Account struct {
	ID       string      `json:"id"`
	Type     AccountType `json:"type"`
	Name     string      `json:"name"`
	Currency Commodity   `json:"currency"`
}

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

type TxType string

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

type Transaction struct {
	ID       string    `json:"id"`
	Date     Date      `json:"date"`
	Type     TxType    `json:"type"`
	Title    string    `json:"title,omitempty"`
	Amount   Amount    `json:"amount"`
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

func (t Transaction) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("transaction: empty id")
	}

	if t.Date.IsZero() {
		return fmt.Errorf("tx %q: empty date", t.ID)
	}

	switch t.Type {
	case TxTransfer:
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

type AssetClass string

const (
	AssetStock  AssetClass = "stock"
	AssetETF    AssetClass = "etf"
	AssetBond   AssetClass = "bond"
	AssetCrypto AssetClass = "brypto"
	AssetCash   AssetClass = "cash"
	AssetOther  AssetClass = "other"
)

type Asset struct {
	ID    string `json:"id"`
	Class string `json:"class"`
	Name  string `json:"name,omitempty"`
}

func (a Asset) Validate() error {
	if a.ID == "" {
		return fmt.Errorf("asset: empty id")
	}

	return nil
}

// --- Goal ---

type GoalState string

const (
	GoalOpen      GoalState = "open"
	GoalDone      GoalState = "done"
	GoalCancelled GoalState = "cancelled"
)

type Goal struct {
	ID     string    `json:"id"`
	Title  string    `json:"title"`
	State  GoalState `json:"state"`
	Target Amount    `json:"target"`
	Date   string    `json:"date,omitempty"`
}

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
