package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestHealth(t *testing.T) {
	s := New(t.TempDir())
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("health = %d", rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal health: %v", err)
	}

	if body["status"] != "ok" {
		t.Errorf("body = %v", body)
	}
}

func TestBalances(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "transactions.json"), []byte(`[
	  {"id":"t1","date":"2026-08-02","type":"expense","amount":{"value":"10","commodity":"EUR"},"account":"bank/checking"}
	]`), 0o644); err != nil {
		t.Fatalf("setup WriteFile: %v", err)
	}

	s := New(dir)
	req := httptest.NewRequest("GET", "/api/v1/balances", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("balances = %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPortfolio(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "transactions.json"), []byte(`[
		{"id":"b1","date":"2026-08-10","type":"buy","asset":"VWCE","quantity":"5","price":{"value":"130","commodity":"EUR"},"account":"broker/ibkr"},
		{"id":"b2","date":"2026-09-10","type":"buy","asset":"VWCE","quantity":"2","price":{"value":"134.52","commodity":"EUR"},"account":"broker/ibkr"},
		{"id":"s1","date":"2026-09-12","type":"sell","asset":"VWCE","quantity":"4","price":{"value":"140","commodity":"EUR"},"account":"broker/ibkr"}
	]`), 0o644); err != nil {
		t.Fatalf("setup WriteFile: %v", err)
	}

	s := New(dir)
	req := httptest.NewRequest("GET", "/api/v1/portfolio", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("portfolio = %d: %s", rec.Code, rec.Body.String())
	}

	var pf Portfolio
	if err := json.Unmarshal(rec.Body.Bytes(), &pf); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(pf.Positions) != 1 {
		t.Fatalf("positions = %d, want 1", len(pf.Positions))
	}

	p := pf.Positions[0]
	if got := p.Quantity.String(); got != "3 VWCE" { // 5+2-4
		t.Errorf("qty = %q, want 3 VWCE", got)
	}

	if got := p.Price.String(); got != "134.52 EUR" { // last buy price
		t.Errorf("price = %q, want 134.52 EUR", got)
	}

	if got := p.CostValue.String(); got != "403.56 EUR" { // 3 × 134.52
		t.Errorf("cost = %q, want 403.56 EUR", got)
	}
}

func TestPortfolioMarket(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "transactions.json"), []byte(`[
		{"id":"b1","date":"2026-08-10","type":"buy","asset":"VWCE","quantity":"5","price":{"value":"130","commodity":"EUR"},"account":"broker/ibkr"}
	]`), 0o644); err != nil {
		t.Fatalf("setup WriteFile: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "prices.json"), []byte(`[
		{"asset":"VWCE","date":"2026-08-20","price":{"value":"135","commodity":"EUR"}},
		{"asset":"VWCE","date":"2026-08-24","price":{"value":"140.10","commodity":"EUR"}}
	]`), 0o644); err != nil {
		t.Fatalf("setup WriteFile: %v", err)
	}

	s := New(dir)
	req := httptest.NewRequest("GET", "/api/v1/portfolio?basis=market", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("portfolio = %d: %s", rec.Code, rec.Body.String())
	}

	var pf Portfolio
	if err := json.Unmarshal(rec.Body.Bytes(), &pf); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if pf.Basis != "market" {
		t.Errorf("basis = %q", pf.Basis)
	}

	p := pf.Positions[0]
	if p.MarketValue == nil || p.MarketValue.String() != "700.5 EUR" { // 5 × 140.10 (latest quote)
		t.Errorf("market = %v, want 700.5 EUR", p.MarketValue)
	}

	if got := p.CostValue.String(); got != "650 EUR" { // 5 × 130 untouched
		t.Errorf("cost = %q, want 650 EUR", got)
	}
}
