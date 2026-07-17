// Package check defines the validation engine: the Finding type, the Check
// interface every rule implements, and a registry so rules self-register.
package check

import (
	"github.com/scottlogic/gemara-catalog-validator/internal/config"
	"github.com/scottlogic/gemara-catalog-validator/internal/model"
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

var registry []Check

// Register adds a check to the registry. Call from an init() in the rule's file.
func Register(c Check) { registry = append(registry, c) }

// All returns the registered checks in registration order.
func All() []Check { return registry }
