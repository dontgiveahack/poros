// Package fire computes savings and FIRE metrics from a ledger.
// Net worth is at cost (buy prices). Market valuations arrives with prices.
package fire

import (
	"math"
	"math/big"
	"math/rand"
	"sort"
	"time"

	"github.com/dontgiveahack/poros/internal/domain"
	"github.com/dontgiveahack/poros/internal/store"
)

// Options tunes the calculation. Zero values get sane defaults.
type Options struct {
	Year           int     // calendar year for income/expenses (Default current)
	WithdrawalRate float64 // e.g. 0.04 (Default 0.04)
	ExpectedReturn float64 // e.g. 0.05 (Default 0.05)
	// Coast FIRE needs both ages; zero disables it.
	CurrentAge     int
	RetirementAge  int
	// Lean/Fat variants. nil disables each.
	LeanExpenses   *domain.Amount
	FatExpenses    *domain.Amount
	// Monte Carlo (simulate enables/disables it)
	Simulate       bool
	Runs           int     // default 10000
	Volatility     float64 // annual std-dev (Default 0.15)
	Seed           int64   // (Default 42)
	HorizonYears   int     // 0 = retire-current, fallback 10
}

// Summary is the FIRE report.
type Summary struct {
	Year           int            `json:"year"`
	NetWorth       domain.Amount  `json:"net_worth"`
	AnnualIncome   domain.Amount  `json:"annual_income"`
	AnnualExpenses domain.Amount  `json:"annual_expenses"`
	SavingsRate    float64        `json:"savings_rate"`  // 1 - expenses / income
	AnnualSavings  domain.Amount  `json:"annual_savings"`
	FireNumber     domain.Amount  `json:"fire_number"`   // expenses / withdrawal
	YearsToFire    float64        `json:"years_to_fire"` // -1 = never at current pace
	// Coast/Lean/Fat are omitempty: only present when configured
	CoastFire      *domain.Amount `json:"coast_fire,omitempty"`
	CoastProgress  float64        `json:"coast_progress,omitempty"` // networth / coast
	LeanFire       *domain.Amount `json:"lean_fire,omitempty"`
	FatFire        *domain.Amount `json:"fat_fire,omitempty"`
	Sim            *Simulation    `json:"simulation,omitempty"`
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
		if tx.Date.Year() != opts.Year {
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

	s := &Summary{
		Year: opts.Year, NetWorth: nw,
		AnnualIncome: income, AnnualExpenses: expenses,
		SavingsRate: rate, AnnualSavings: savings,
		FireNumber: fireNum, YearsToFire: years,
	}

	// Coast FIRE: capital needed today to reach the FIRE number at
	// retirement age with no further contributions.
	if opts.CurrentAge > 0 && opts.RetirementAge > opts.CurrentAge {
		yearsLeft := float64(opts.RetirementAge - opts.CurrentAge)
		divisor := ratFromFloat(math.Pow(1+opts.ExpectedReturn, yearsLeft))
		coast := domain.NewAmount(new(big.Rat).Quo(fireRat, divisor), cur)
		s.CoastFire = &coast
		if cf, _ := coast.Rat().Float64(); cf > 0 {
			if nwf, _ := nw.Rat().Float64(); true {
				s.CoastProgress = nwf / cf
			}
		}
	}

	// Lean/Fat: same rule applied to lean/fat expense levels.
	if opts.LeanExpenses != nil {
		lean := domain.NewAmount(new(big.Rat).Quo(
			opts.LeanExpenses.Rat(),
			ratFromFloat(opts.WithdrawalRate),
		), cur)
		s.LeanFire = &lean
	}

	if opts.FatExpenses != nil {
		fat := domain.NewAmount(new(big.Rat).Quo(
			opts.FatExpenses.Rat(),
			ratFromFloat(opts.WithdrawalRate),
		), cur)
		s.FatFire = &fat
	}

	if opts.Simulate {
		runs := opts.Runs
		if runs <= 0 {
			runs = 10000
		}

		years := opts.HorizonYears
		if years <= 0 {
			years = 10
			if opts.CurrentAge > 0 && opts.RetirementAge > opts.CurrentAge {
				years = opts.RetirementAge - opts.CurrentAge
			}
		}

		seed := opts.Seed
		if seed == 0 {
			seed = 42
		}

		sd := opts.Volatility
		if sd == 0 {
			sd = 0.15
		}

		nwf, _ := nw.Rat().Float64()
		svf, _ := savings.Rat().Float64()
		tf,  _ := fireRat.Float64()
		s.Sim = simulate(nwf, svf, tf, opts.ExpectedReturn, sd, years, runs, seed, cur)
	}

	return s, nil
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

// Simulation is a Monte Carlo projection of the portfolio.
type Simulation struct {
	Runs     int           `json:"runs"`
	Years    int           `json:"years"`
	Seed     int64         `json:"seed"`
	P10      domain.Amount `json:"p10"`
	P50      domain.Amount `json:"p50"`
	P90      domain.Amount `json:"p90"`
	ProbFire float64       `json:"prob_fire"`
}

// simulate runs the Monte Carlo accumulation: starting from nw, adding
// save each year, compounding at Normal(mean, sd) returns.
// All float: this is projection, not accounting.
func simulate(nw, save, target float64, mean, sd float64, years, runs int,
              seed int64, cur domain.Commodity,
) *Simulation {
	rng := rand.New(rand.NewSource(seed))
	finals := make([]float64, runs)
	for i := range finals {
		v := nw
		for y := 0; y < years; y++ {
			r := mean + sd*rng.NormFloat64()
			v = v*(1+r) + save
		}

		finals[i] = v
	}

	sort.Float64s(finals)
	hit := 0
	for _, v := range finals {
		if v >= target {
			hit++
		}
	}

	amt := func(f float64) domain.Amount {
		return domain.NewAmount(ratFromFloat(f), cur)
	}

	return &Simulation{
		Runs: runs, Years: years, Seed: seed,
		P10: amt(finals[runs/10]), P50: amt(finals[runs/2]), P90: amt(finals[runs*9/10]),
		ProbFire: float64(hit) / float64(runs),
	}
}
