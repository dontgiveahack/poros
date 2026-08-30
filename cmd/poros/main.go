package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/dontgiveahack/poros/internal/domain"
	"github.com/dontgiveahack/poros/internal/store"
)

const usage = `poros - personal finance manager

Usage:
  poros <command> [args]

Commands:
  version    Print the poros version
  balance    Show balances per account
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
