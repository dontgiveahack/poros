// Package api exposes Póros data over HTTP.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/dontgiveahack/poros/internal/domain"
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
	writeJSON(w, map[string]string{"status": "ok"})
}

func (s *Server) handleAccounts(w http.ResponseWriter, _ *http.Request) {
	l, err := store.LoadDir(s.dataDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, l.Accounts)
}

func (s *Server) handleTransactions(w http.ResponseWriter, _ *http.Request) {
	l, err := store.LoadDir(s.dataDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, l.Transactions)
}

func (s *Server) handleGoals(w http.ResponseWriter, _ *http.Request) {
	l, err := store.LoadDir(s.dataDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, l.Goals)
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

	writeJSON(w, out)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
