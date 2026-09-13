// Package api exposes Póros data over HTTP.
package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
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
