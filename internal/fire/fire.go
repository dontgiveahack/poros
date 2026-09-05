// Package fire computes savings and FIRE metrics from a ledger.
// Net worth is at cost (buy prices). Market valuations arrives with prices.
package fire

import (
	"math"
	"math/big"
	"time"

	"github.com/dontgiveahack/poros/internal/domain"
	"github.com/dontgiveahack/poros/internal/store"
)

// Options tunes the calculation. Zero values get sane defaults.
type Options struct {
	Year           int     // calendar year for income/expenses (Default current)
	WithdrawalRate float64 // e.g. 0.04 (Default 0.04)
	ExpectedReturn float64 // e.g. 0.05 (Default 0.05)
}

// Summary is the FIRE report.
type Summary struct {
	Year           int           `json:"year"`
	NetWorth       domain.Amount `json:"net_worth"`
	AnnualIncome   domain.Amount `json:"annual_income"`
	AnnualExpenses domain.Amount `json:"annual_expenses"`
	SavingsRate    float64       `json:"savings_rate"`  // 1 - expenses / income
	AnnualSavings  domain.Amount `json:"annual_savings"`
	FireNumber     domain.Amount `json:"fire_number"`   // expenses / withdrawal
	YearsToFire    float64       `json:"years_to_fire"` // -1 = never at current pace
}

func (o *Options) withDefaults() {
	if o.Year == 0 {
		o.Year = time.Now().Year()
	}

	if o.WithdrawalRate == 0 {
		o.WithdrawalRate = 0.04
	}

	if o.ExpectedReturn == 0 {
		o.ExpectedReturn = 0.05
	}
}

// Calculate builds the FIRE summary from the ledger.
// All money math uses domain.Amount (exact rational). Only ratios are float.
func Calculate(l *store.Ledger, opts Options) (*Summary, error) {
	opts.withDefaults()
	cur := domain.Commodity("EUR") // TODO: picks ledger currency from config

	balances, err := domain.CalculateBalances(l.Transactions)
	if err != nil {
		return nil, err
	}

	nw := sumCommodity(balances, cur)

	var income, expenses domain.Amount
	income = mustZero(cur)
	expenses = mustZero(cur)
	for _, tx := range l.Transactions {
		if tx.Date.Time.Year() != opts.Year {
			continue
		}

		if tx.Amount == nil || tx.Amount.Commodity() != cur {
			continue
		}

		switch tx.Type {
		case domain.TxIncome:
			income, _ = income.Add(*tx.Amount)
		case domain.TxExpense, domain.TxFee:
			expenses, _ = expenses.Add(*tx.Amount)
		}
	}

	savings, _ := income.Add(expenses.Neg())

	var rate float64
	if !isZeroRat(income) {
		r := new(big.Rat).Quo(expenses.Rat(), income.Rat())
		f, _ := r.Float64()
		rate = 1 - f
	}

	// FIRE number = expenses / withdrawal
	fireRat := new(big.Rat).Quo(expenses.Rat(), ratFromFloat(opts.WithdrawalRate))
	fireNum := domain.NewAmount(fireRat, cur)

	// Years to FIRE: solve nw*(1+r)^n + save*((1+r)^n -1)/r = target
	years := yearsToFire(nw.Rat(), savings.Rat(), fireRat, opts.ExpectedReturn)

	return &Summary{
		Year: opts.Year, NetWorth: nw,
		AnnualIncome: income, AnnualExpenses: expenses,
		SavingsRate: rate, AnnualSavings: savings,
		FireNumber: fireNum, YearsToFire: years,
	}, nil
}

func mustZero(c domain.Commodity) domain.Amount {
	a, _ := domain.ParseAmount("0 " + string(c))
	return a
}

func isZeroRat(a domain.Amount) bool {
	return a.Rat().Sign() == 0
}

func ratFromFloat(f float64) *big.Rat {
	return new(big.Rat).SetFloat64(f)
}

func sumCommodity(balances domain.Balances, c domain.Commodity) domain.Amount {
	total := mustZero(c)
	for _, byComm := range balances {
		if amt, ok := byComm[c]; ok {
			total, _ = total.Add(amt)
		}
	}

	return total
}

func yearsToFire(nw, save, target *big.Rat, r float64) float64 {
	t, _ := target.Float64()
	n, _ := nw.Float64()
	s, _ := save.Float64()

	if n >= t {
		return 0
	}

	if s <= 0 || r <= 0 {
		return -1
	}

	// Closed form: n = ln((t*r + s) / (n*r + s)) / ln(1 + r)
	num := t*r + s
	den := n*r + s
	if num <= 0 || den <= 0 {
		return -1
	}

	return math.Log(num/den) / math.Log(1+r)
}
