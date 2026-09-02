package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sort"

	"github.com/dontgiveahack/poros/internal/api"
	"github.com/dontgiveahack/poros/internal/config"
	"github.com/dontgiveahack/poros/internal/domain"
	"github.com/dontgiveahack/poros/internal/pgstore"
	"github.com/dontgiveahack/poros/internal/store"
	"github.com/dontgiveahack/poros/internal/verify"
)

const usage = `poros - personal finance manager

Usage:
  poros <command> [args]

Commands:
  version    Print the poros version
  init       Initialise a new poros project
  balance    Show balances per account
  serve      Start the HTTP API server
  verify     Compare data/*.json with the DB mirror
  help       Show this help

Run 'poros balance -h' for balance options.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}

	var code int
	switch os.Args[1] {
	case "version":
		fmt.Println("poros v0.0.1")
	case "balance":
		code = runBalance(os.Args[2:])
	case "serve":
		code = runServe(os.Args[2:])
	case "init":
		code = runInit(os.Args[2:])
	case "verify":
		code = runVerify(os.Args[2:])
	case "help", "-h", "--help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", os.Args[1], usage)
		os.Exit(2)
	}

	os.Exit(code)
}

func runBalance(args []string) int {
	fs := flag.NewFlagSet("balance", flag.ContinueOnError)
	dataDir := fs.String("data", "data", "data directory containing accounts.json, transactions.json, ...")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	ledger, err := store.LoadDir(*dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load %s: %v\n", *dataDir, err)
		return 1
	}

	balances, err := domain.CalculateBalances(ledger.Transactions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "calculate: %v\n", err)
		return 1
	}

	if len(balances) == 0 {
		fmt.Println("no transactions")
		return 0
	}

	// Deterministic order: sort account names, then commodities.
	accounts := make([]string, 0, len(balances))
	for a := range balances {
		accounts = append(accounts, a)
	}
	sort.Strings(accounts)

	for _, account := range accounts {
		comms := make([]string, 0, len(balances[account]))
		for c := range balances[account] {
			comms = append(comms, string(c))
		}

		sort.Strings(comms)
		for _, c := range comms {
			amount := balances[account][domain.Commodity(c)]
			fmt.Printf("%-24s %12s\n", account, amount.String())
		}
	}

	return 0
}

func runServe(args []string) int {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	addr := fs.String("addr", ":8080", "listen address")
	dataDir := fs.String("data", "data", "data directory")
	dbURL := fs.String("db", "", "postgres DSN (e.g. postgres://poros:poros@localhost:5432/poros?sslmode=disable)")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	// Optional JSONB sync on startup if --db given
	if *dbURL != "" {
		ctx := context.Background()
		ps, err := pgstore.New(ctx, *dbURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "db connect: %v\n", err)
			return 1
		}
		defer ps.Close()

		if err := ps.Migrate(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
			return 1
		}

		ledger, err := store.LoadDir(*dataDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "load %s: %v\n", *dataDir, err)
			return 1
		}

		if err := ps.Sync(ctx, ledger); err != nil {
			fmt.Fprintf(os.Stderr, "sync: %v\n", err)
			return 1
		}

		fmt.Printf("synced %d transactions to postgres\n", len(ledger.Transactions))
	}

	srv := api.New(*dataDir)
	fmt.Printf("poros serve on http://localhost%s (data=%s)\n", *addr, *dataDir)
	if err := http.ListenAndServe(*addr, srv.Handler()); err != nil {
		fmt.Fprintf(os.Stderr, "serve: %v\n", err)
		return 1
	}

	return 0
}

func runInit(args []string) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	dir := fs.String("dir", ".", "project directory to initialise")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	tomlPath := fmt.Sprintf("%s/poros.toml", *dir)
	dataDir := fmt.Sprintf("%s/data", *dir)

	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", dataDir, err)
		return 1
	}

	cfg := config.Default()
	if err := config.Write(tomlPath, cfg); err != nil {
		// data dir already created, so no fatal for re-run
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	// Ensure empty JSON arrays exist so LoadDir sees valid files.
	for _, name := range[]string{"accounts.json", "transactions.json", "assets.json", "goals.json"} {
		p := fmt.Sprintf("%s/%s", dataDir, name)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			os.WriteFile(p, []byte("[]\n"), 0o644)
		}
	}

	fmt.Printf("initialised %s and %s\n", tomlPath, dataDir)

	return 0
}

func runVerify(args []string) int {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	dataDir := fs.String("data", "data", "data directory")
	dbURL := fs.String("db", "", "postgres DSN")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *dbURL == "" {
		fmt.Fprintln(os.Stderr, "verify: --db is required")
		return 2
	}

	fileLedger, err := store.LoadDir(*dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load %s: %v\n", *dataDir, err)
		return 1
	}

	ctx := context.Background()
	ps, err := pgstore.New(ctx, *dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db connect: %v\n", err)
		return 1
	}

	defer ps.Close()

	dbLedger, err := ps.LoadLedger(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load db: %v\n", err)
		return 1
	}

	diffs := verify.Diff(fileLedger, dbLedger)
	if len(diffs) == 0 {
		fmt.Println("verify: OK - file and DB are identical")
		return 0
	}

	fmt.Println("verify: differences found:")
	for _, d := range diffs {
		fmt.Printf("  - %s\n", d)
	}

	return 1
}
