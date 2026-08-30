package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingReturnsDefault(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "no.toml"))
	if err != nil {
		t.Fatalf("Load missing: %v", err)
	}

	if cfg.Currency != "EUR" || cfg.DataDir != "data" {
		t.Fatalf("default = %+v", cfg)
	}
}

func TestWriteAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "poros.toml")

	if err := Write(path, Default()); err != nil {
		t.Fatalf("Write: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Currency != "EUR" {
		t.Errorf("Currency = %q", cfg.Currency)
	}
}

func TestWriteDoesNotOverwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "poros.toml")
	os.WriteFile(path, []byte(`currency = "USD"`), 0o644)

	if err := Write(path, Default()); err == nil {
		t.Fatal("second Write should fail")
	}
}
