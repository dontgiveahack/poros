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
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["status"] != "ok" {
		t.Errorf("body = %v", body)
	}
}

func TestBalances(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "transactions.json"), []byte(`[
	  {"id":"t1","date":"2026-08-02","type":"expense","amount":{"value":"10","commodity":"EUR"},"account":"bank/checking"}
	]`), 0o644)

	s := New(dir)
	req := httptest.NewRequest("GET", "/api/v1/balances", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("balances = %d: %s", rec.Code, rec.Body.String())
	}
}
