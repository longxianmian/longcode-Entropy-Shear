// hash-policy emits the canonical SHA-256 of a policy file. The hash
// is over the canonicalized JSON (sorted keys at every depth), so it
// is independent of source-file formatting.
//
// Exit 0 on success, 1 when the policy fails schema validation, 2 on
// I/O error.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	apperr "entropy-shear/internal/errors"
	"entropy-shear/internal/policy"
)

func main() {
	file := flag.String("file", "", "path to a policy JSON file")
	pretty := flag.Bool("pretty", true, "pretty-print JSON output")
	flag.Parse()
	if *file == "" {
		fmt.Fprintln(os.Stderr, "hash-policy: --file is required")
		os.Exit(2)
	}

	report, err := policy.HashFileReport(*file)
	if err != nil {
		var ae *apperr.APIError
		if errors.As(err, &ae) {
			fmt.Fprintf(os.Stderr, "hash-policy: %s: %s\n", ae.Code, ae.Detail)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "hash-policy: %v\n", err)
		os.Exit(2)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	if *pretty {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(report); err != nil {
		fmt.Fprintf(os.Stderr, "hash-policy: encode: %v\n", err)
		os.Exit(2)
	}
}
