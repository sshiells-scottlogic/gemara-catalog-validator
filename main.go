// Command gemara-validate checks the data correctness of Gemara source
// catalogs (referential integrity, id uniqueness, …) — the semantic rules that
// schema validation does not cover — and reports findings for a PR to fix.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/scottlogic/gemara-catalog-validator/internal/check"
	// Side-effect import: registers all checks.
	_ "github.com/scottlogic/gemara-catalog-validator/internal/check/checks"
	"github.com/scottlogic/gemara-catalog-validator/internal/config"
	"github.com/scottlogic/gemara-catalog-validator/internal/model"
)

func main() {
	code, err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}
	os.Exit(code)
}

func run() (int, error) {
	configPath := flag.String("config", ".gemara-validate.yaml", "path to config file")
	format := flag.String("format", "text", "output format: text | github")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		return 0, err
	}

	// Positional args override the configured paths (handy for local runs).
	paths := cfg.Paths
	if args := flag.Args(); len(args) > 0 {
		paths = args
	}

	idx, err := model.Load(paths)
	if err != nil {
		return 0, err
	}

	ctx := &check.Context{Index: idx, Config: cfg}
	var findings []check.Finding
	for _, c := range check.All() {
		findings = append(findings, c.Run(ctx)...)
	}

	if err := check.Report(os.Stdout, *format, findings); err != nil {
		return 0, err
	}

	errCount := 0
	for _, f := range findings {
		if f.Severity == check.Error {
			errCount++
		}
	}
	fmt.Fprintf(os.Stderr, "\n%d file(s), %d check(s), %d finding(s) (%d error).\n",
		len(idx.Files), len(check.All()), len(findings), errCount)

	if errCount > 0 && cfg.FailOn == "error" {
		return 1, nil
	}
	return 0, nil
}
