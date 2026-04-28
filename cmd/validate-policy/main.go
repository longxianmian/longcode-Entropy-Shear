// validate-policy reports whether a policy file conforms to the same
// schema the running /shear endpoint enforces. It always emits JSON
// to stdout — exit 0 when ok=true, 1 when ok=false (schema violation),
// 2 on I/O / setup error.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"entropy-shear/internal/policy"
)

func main() {
	file := flag.String("file", "", "path to a policy JSON file")
	pretty := flag.Bool("pretty", true, "pretty-print JSON output")
	flag.Parse()
	if *file == "" {
		fmt.Fprintln(os.Stderr, "validate-policy: --file is required")
		os.Exit(2)
	}

	report, err := policy.ValidateFile(*file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "validate-policy: %v\n", err)
		os.Exit(2)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	if *pretty {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(report); err != nil {
		fmt.Fprintf(os.Stderr, "validate-policy: encode: %v\n", err)
		os.Exit(2)
	}

	if !report.OK {
		os.Exit(1)
	}
}
