// Package checks holds the individual validation rules. Each file registers
// its check via init(); add new rules here and they run automatically.
package checks

import (
	"fmt"

	"github.com/scottlogic/gemara-catalog-validator/internal/check"
)

// referentialIntegrity verifies every internal reference-id resolves to a
// defined id somewhere in the catalog set. This is the core cross-catalog
// data-correctness check (imports and threat/control mappings all rely on it).
type referentialIntegrity struct{}

func (referentialIntegrity) ID() string { return "reference-resolves" }

func (referentialIntegrity) Description() string {
	return "Every internal reference-id must resolve to a defined id in the catalog set."
}

func (referentialIntegrity) Run(ctx *check.Context) []check.Finding {
	var out []check.Finding
	for _, ref := range ctx.Index.Refs {
		if !ctx.Config.InternalRef(ref.ID) {
			continue // mapping-group shorthand or external framework reference
		}
		if _, ok := ctx.Index.Defs[ref.ID]; ok {
			continue
		}
		out = append(out, check.Finding{
			RuleID:   "reference-resolves",
			Severity: check.Error,
			File:     ref.Location.File,
			Line:     ref.Location.Line,
			Col:      ref.Location.Col,
			Message:  fmt.Sprintf("reference-id %q does not resolve to any defined id", ref.ID),
		})
	}
	return out
}

func init() { check.Register(referentialIntegrity{}) }
