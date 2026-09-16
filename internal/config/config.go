// Package config loads poros.toml (currency, data dir, future FIRE/budget).
package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/dontgiveahack/poros/internal/domain"
	"github.com/dontgiveahack/poros/internal/fire"
)

// Config mirrors poros.toml.
// Comments are preserved on write by not round-tripping through toml.
type Config struct {
	Currency string     `toml:"currency"`
	DataDir  string     `toml:"data_dir"`
	Locale   Locale     `toml:"locale"`
	Fire     FireConfig `toml:"fire"`
}

// Locale holds display conventions (language, decimal separator).
type Locale struct {
	Language string `toml:"language"`
	Decimal  string `toml:"decimal"`
}

// Default returns sane defaults for a new project.
func Default() Config {
	return Config{
		Currency: "EUR",
		DataDir:  "data",
		Locale:   Locale{Language: "es", Decimal: "."},
		Fire:     FireConfig{WithdrawalRate: 0.04, ExpectedReturn: 0.05},
	}
}

// Load reads poros.toml from path. If the file does not exist, it returns
// Default() without error (so `poros balance` works without init in tests).
func Load(path string) (Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}

		return Config{}, fmt.Errorf("read %s: %w", path, err)
	}

	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", path, err)
	}

	return cfg, nil
}

// Write creates poros.toml at path with defaults if it does not exist.
// It never overwrites an existing file.
func Write(path string, cfg Config) (err error) {
	if _, statErr := os.Stat(path); statErr == nil {
		return fmt.Errorf("%s already exists", path)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	return toml.NewEncoder(f).Encode(cfg)
}

// FireConfig mirrors the [fire] section of poros.toml.
type FireConfig struct {
	WithdrawalRate float64 `toml:"withdrawal_rate"`
	ExpectedReturn float64 `toml:"expected_return"`
	CurrentAge     int     `toml:"current_age"`
	RetirementAge  int     `toml:"retirement_age"`
	LeanExpenses   string  `toml:"lean_expenses"`
	FatExpenses    string  `toml:"fat_expenses"`
}

// FireOptions converts the [fire] section into fire.Options.
// Unset values stay zero and fire.Calculate applies its defaults.
func (c Config) FireOptions() (fire.Options, error) {
	o := fire.Options{
		WithdrawalRate: c.Fire.WithdrawalRate,
		ExpectedReturn: c.Fire.ExpectedReturn,
		CurrentAge:     c.Fire.CurrentAge,
		RetirementAge:  c.Fire.RetirementAge,
	}

	if c.Fire.LeanExpenses != "" {
		a, err := domain.ParseAmount(c.Fire.LeanExpenses)
		if err != nil {
			return fire.Options{}, fmt.Errorf("lean_expenses: %w", err)
		}

		o.LeanExpenses = &a
	}

	if c.Fire.FatExpenses != "" {
		a, err := domain.ParseAmount(c.Fire.FatExpenses)
		if err != nil {
			return fire.Options{}, fmt.Errorf("fat_expenses: %w", err)
		}

		o.FatExpenses = &a
	}

	return o, nil
}
