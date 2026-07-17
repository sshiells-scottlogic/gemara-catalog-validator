// Package config loads the project-specific validation settings that keep the
// validator decoupled from any one project's conventions (paths, id scheme).
package config

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// Config is the contents of .gemara-validate.yaml. Anything project-specific
// lives here so the engine and checks stay generic.
type Config struct {
	// Paths are files or directories to scan for catalog assets.
	Paths []string `yaml:"paths"`
	// IDPattern matches an *internal*, resolvable reference-id. References that
	// don't match (mapping-group shorthands like "CCC", external frameworks) are
	// ignored by referential-integrity checks.
	IDPattern string `yaml:"id-pattern"`
	// FailOn controls the exit code: "error" (default) exits non-zero on any
	// error-severity finding; "never" always exits zero.
	FailOn string `yaml:"fail-on"`

	idRE *regexp.Regexp
}

// Default returns the built-in configuration used when no config file exists.
func Default() *Config {
	return &Config{
		Paths:     []string{"catalogs"},
		IDPattern: `^[A-Za-z0-9]+\.[A-Za-z0-9]+\.(CP|CN|TH)[0-9]+$`,
		FailOn:    "error",
	}
}

// Load reads config from path, falling back to Default() when the file is
// absent. Values present in the file override the defaults.
func Load(path string) (*Config, error) {
	c := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return c, c.Compile()
		}
		return nil, err
	}
	if err := yaml.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("parsing %q: %w", path, err)
	}
	return c, c.Compile()
}

// Compile validates and caches the id-pattern regexp. Call after mutating
// IDPattern directly (e.g. in tests).
func (c *Config) Compile() error {
	re, err := regexp.Compile(c.IDPattern)
	if err != nil {
		return fmt.Errorf("invalid id-pattern %q: %w", c.IDPattern, err)
	}
	c.idRE = re
	return nil
}

// InternalRef reports whether a reference-id points at an id this catalog set
// is expected to define (and must therefore resolve).
func (c *Config) InternalRef(id string) bool {
	return c.idRE.MatchString(id)
}
