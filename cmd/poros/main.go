package main

import (
	"fmt"
	"os"
)

const usage = `poros - personal finance manager

Usage:
  poros <command> [args]

Commands:
  version    Print the poros version
  help       Show this help
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}

	switch os.Args[1] {
	case "version":
		fmt.Println("poros v0.0.1")
	case "help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", os.Args[1], usage)
		os.Exit(2)
	}
}
