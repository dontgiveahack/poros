// Package api exposes Póros data over HTTP.
package api

import (
	"fmt"
	"encoding/json"
	"log/slog"
	"math/big"
	"net/http"
	"sort"
	"strconv"

	"github.com/dontgiveahack/poros/internal/domain"
	"github.com/dontgiveahack/poros/internal/fire"
	"github.com/dontgiveahack/poros/internal/store"
)

// Server serves the Póros REST API from a data directory.
type Server struct {
	dataDir string
	mux     *http.ServeMux
}

// New creates a Server backed by dataDir.
func New(dataDir string) *Server {
	s := &Server{dataDir: dataDir, mux: http.NewServeMux()}
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /api/v1/accounts", s.handleAccounts)
	s.mux.HandleFunc("GET /api/v1/transactions", s.handleTransactions)
	s.mux.HandleFunc("GET /api/v1/balances", s.handleBalances)
	s.mux.HandleFunc("GET /api/v1/goals", s.handleGoals)
	s.mux.HandleFunc("GET /api/v1/fire", s.handleFire)
	s.mux.HandleFunc("GET /api/v1/portfolio", s.handlePortfolio)
	return s
}

// Handler returns the http.Handler (with CORS for web dev).
func (s *Server) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		s.mux.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	if err := writeJSON(w, map[string]string{"status": "ok"}); err != nil {
		slog.Error("encode health", "err", err)
	}
}

func (s *Server) handleAccounts(w http.ResponseWriter, _ *http.Request) {
	l, err := store.LoadDir(s.dataDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := writeJSON(w, l.Accounts); err != nil {
		slog.Error("encode accounts", "err", err)
	}
}

func (s *Server) handleTransactions(w http.ResponseWriter, _ *http.Request) {
	l, err := store.LoadDir(s.dataDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := writeJSON(w, l.Transactions); err != nil {
		slog.Error("encode transactions", "err", err)
	}
}

func (s *Server) handleGoals(w http.ResponseWriter, _ *http.Request) {
	l, err := store.LoadDir(s.dataDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := writeJSON(w, l.Goals); err != nil {
		slog.Error("encode goals", "err", err)
	}
}

func (s *Server) handleBalances(w http.ResponseWriter, _ *http.Request) {
	l, err := store.LoadDir(s.dataDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	balances, err := domain.CalculateBalances(l.Transactions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Flatten for JSON:
	// [{"account":"bank/checking","commodity":"EUR","amount":{"value":"...","commodity":"EUR"}}]
	type row struct {
		Account   string           `json:"account"`
		Commodity domain.Commodity `json:"commodity"`
		Amount    domain.Amount    `json:"amount"`
	}

	var out []row
	for account, byCommodity := range balances {
		for commodity, amount := range byCommodity {
			out = append(out, row{
				Account: account,
				Commodity: commodity,
				Amount: amount,
			})
		}
	}

	if err := writeJSON(w, out); err != nil {
		slog.Error("encode balances", "err", err)
	}
}

func (s *Server) handleFire(w http.ResponseWriter, r *http.Request) {
	year := atoiQuery(r, "year", 0)
	l, err := store.LoadDir(s.dataDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sum, err := fire.Calculate(l, fire.Options{Year: year})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := writeJSON(w, sum); err != nil {
		slog.Error("encode fire", "err", err)
	}
}

func atoiQuery(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}

	return n
}

func writeJSON(w http.ResponseWriter, v any) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

// Position is a holding valved at last buy price (cost basis)
type Position struct {
	Asset       string         `json:"asset"`
	Account     string         `json:"account"`
	Quantity    domain.Amount  `json:"quantity"`
	Price       domain.Amount  `json:"price"`
	CostValue   domain.Amount  `json:"cost_value"`
	MarketValue *domain.Amount `json:"market_value,omitempty"`
}

// Portfolio is the set of open positions plus their total cost.
// Single-currency assumption: positions in other currencies are listed
// but excluded from TotalCost (multi-currency comes with prices).
type Portfolio struct {
	Positions   []Position     `json:"positions"`
	TotalCost   domain.Amount  `json:"total_cost"`
	TotalMarket *domain.Amount `json:"total_market,omitempty"`
	Basis       string         `json:"basis"`
}

func (s *Server) handlePortfolio(w http.ResponseWriter, r *http.Request) {
	basis := r.URL.Query().Get("basis")
	if basis == "" {
		basis = "cost"
	}

	if basis != "cost" && basis != "market" {
		http.Error(w, "basis must be cost or market", http.StatusBadRequest)
		return
	}

	l, err := store.LoadDir(s.dataDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pf, err := buildPortfolio(l.Transactions, l.Prices, basis)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := writeJSON(w, pf); err != nil {
		slog.Error("encode portfolio", "err", err)
	}
}

// buildPortfolio nets buy/sell quantities per account+asset and values
// each open position at ost (last buy) and, when basis is "market",
// at the latest known quote (falling back to cost without quote).
func buildPortfolio(txs []domain.Transaction,
                    prices []domain.Price,
	            basis string,
) (Portfolio, error) {
	type key struct{ account, asset string
}
	qty := map[key]*big.Rat{}
	lastPrice := map[key]domain.Amount{}

	for _, tx := range txs {
		switch tx.Type {
		case domain.TxBuy, domain.TxSell:
			q, err := tx.QuantityRat()
			if err != nil {
				return Portfolio{}, err
			}

			if tx.Price == nil {
				return Portfolio{}, fmt.Errorf("tx %s: trade needs price", tx.ID)
			}

			k := key{tx.Account, tx.Asset}
			if _, ok := qty[k]; !ok {
				qty[k] = new(big.Rat)
			}

			if tx.Type == domain.TxBuy {
				qty[k].Add(qty[k], q)
				lastPrice[k] = *tx.Price
			} else {
				qty[k].Sub(qty[k], q)
			}
		}
	}

	quotes := latestPrices(prices)

	var out []Position
	var totalCost, totalMarket domain.Amount
	totalCost, _ = domain.ParseAmount("0 EUR")
	totalMarket, _ = domain.ParseAmount("0 EUR")
	for k, q := range qty {
		if q.Sign() <= 0 {
			continue
		}

		px := lastPrice[k]
		cost := domain.NewAmount(new(big.Rat).Mul(q, px.Rat()), px.Commodity())
		pos := Position{
			Asset: k.asset, Account: k.account,
			Quantity: domain.NewAmount(q, domain.Commodity(k.asset)),
			Price: px, CostValue: cost,
		}
		if sum, err := totalCost.Add(cost); err == nil {
			totalCost = sum
		}

		if basis == "market" {
			mv := cost // fallback: no quote -> cost
			if quote, ok := quotes[k.asset]; ok && quote.Commodity() == px.Commodity() {
				mv = domain.NewAmount(new(big.Rat).Mul(q, quote.Rat()), quote.Commodity())
			}

			pos.MarketValue = &mv
			if sum, err := totalMarket.Add(mv); err == nil {
				totalMarket = sum
			}
		}

		out = append(out, pos)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].CostValue.Rat().Cmp(out[j].CostValue.Rat()) > 0
	})

	pf := Portfolio{Positions: out, TotalCost: totalCost, Basis: basis}
	if basis == "market" {
		pf.TotalMarket = &totalMarket
	}

	return pf, nil
}

// latestPrices keeps the newest quote per asset.
func latestPrices(prices []domain.Price) map[string]domain.Amount {
	out := map[string]domain.Amount{}
	seen := map[string]domain.Date{}

	for _, p := range prices {
		if d, ok := seen[p.Asset]; !ok || p.Date.After(d.Time) {
			seen[p.Asset] = p.Date
			out[p.Asset] = p.Price
		}
	}

	return out
}
