// verify-ledger walks ledger/shear-chain.jsonl offline and prints a JSON
// summary. Exits 0 when the chain is intact, 1 on a mismatch, 2 on I/O
// error. No write side-effects: it never creates directories or files.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"entropy-shear/internal/ledger"
)

func main() {
	path := flag.String("ledger", "ledger/shear-chain.jsonl",
		"path to the append-only JSONL ledger")
	pretty := flag.Bool("pretty", true, "pretty-print the JSON output")
	flag.Parse()

	res, err := ledger.VerifyFile(*path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "verify-ledger: %v\n", err)
		os.Exit(2)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	if *pretty {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(res); err != nil {
		fmt.Fprintf(os.Stderr, "verify-ledger: encode: %v\n", err)
		os.Exit(2)
	}

	if !res.OK {
		os.Exit(1)
	}
}
