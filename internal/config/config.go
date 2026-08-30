// Package config loads poros.toml (currency, data dir, future FIRE/budget).
package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Config mirrors poros.toml.
// Comments are preserved on write by not round-tripping through toml.
type Config struct {
	Currency string `toml:"currency"`
	DataDir  string `toml:"data_dir"`
	Locale   Locale `toml:"locale"`
}

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
func Write(path string, cfg Config) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists", path)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := toml.NewEncoder(f)
	return enc.Encode(cfg)
}
