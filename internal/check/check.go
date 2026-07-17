// Package check defines the validation engine: the Finding type, the Check
// interface every rule implements, and a registry so rules self-register.
package check

import (
	"github.com/sshiells-scottlogic/gemara-catalog-validator/internal/config"
	"github.com/sshiells-scottlogic/gemara-catalog-validator/internal/model"
)

// Severity classifies a finding.
type Severity string

const (
	Error   Severity = "error"
	Warning Severity = "warning"
)

// Finding is a single data-correctness problem, positioned in source.
type Finding struct {
	RuleID   string
	Severity Severity
	File     string
	Line     int
	Col      int
	Message  string
}

// Context is handed to every check: the loaded catalog set plus config.
type Context struct {
	Index  *model.Index
	Config *config.Config
}

// Check is one validation rule. Keep each check small and single-purpose.
type Check interface {
	ID() string
	Description() string
	Run(*Context) []Finding
}

var (
	registry []Check
	optional = map[string]Check{}
)

// Register adds a check that runs by default. Call from an init() in the rule's file.
func Register(c Check) { registry = append(registry, c) }

// RegisterOptional adds an opt-in check, enabled by name via a CLI flag rather
// than run by default (e.g. noisy "stats" checks).
func RegisterOptional(name string, c Check) { optional[name] = c }

// All returns the default checks in registration order.
func All() []Check { return registry }

// Optional returns the opt-in check registered under name, if any.
func Optional(name string) (Check, bool) {
	c, ok := optional[name]
	return c, ok
}
