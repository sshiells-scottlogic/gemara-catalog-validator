// Command gemara-validate checks the data correctness of Gemara source
// catalogs (referential integrity, id uniqueness, …) — the semantic rules that
// schema validation does not cover — and reports findings for a PR to fix.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sshiells-scottlogic/gemara-catalog-validator/internal/check"
	// Side-effect import: registers all checks.
	_ "github.com/sshiells-scottlogic/gemara-catalog-validator/internal/check/checks"
	"github.com/sshiells-scottlogic/gemara-catalog-validator/internal/config"
	"github.com/sshiells-scottlogic/gemara-catalog-validator/internal/model"
)

// version is stamped at build time by GoReleaser (-X main.version=...).
var version = "dev"

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
	orphans := flag.Bool("orphans", false, "also report orphans (unmitigated threats, unreferenced capabilities) as warnings")
	failOn := flag.String("fail-on", "", "exit non-zero policy: error | never (overrides config; default from config)")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return 0, nil
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return 0, err
	}

	// A CLI -fail-on overrides the config. This is how the GitHub Action toggles
	// between gate mode (fail the build on errors) and monitor mode (report but
	// never fail) without editing the repo's config file.
	if *failOn != "" {
		switch *failOn {
		case "error", "never":
			cfg.FailOn = *failOn
		default:
			return 0, fmt.Errorf("invalid -fail-on %q (want: error | never)", *failOn)
		}
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

	checks := check.All()
	if *orphans {
		if c, ok := check.Optional("orphans"); ok {
			checks = append(checks, c)
		}
	}

	ctx := &check.Context{Index: idx, Config: cfg}
	var findings []check.Finding
	for _, c := range checks {
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
		len(idx.Files), len(checks), len(findings), errCount)

	if errCount > 0 && cfg.FailOn == "error" {
		return 1, nil
	}
	return 0, nil
}
