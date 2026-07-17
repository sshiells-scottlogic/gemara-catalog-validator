package checks

import (
	"fmt"

	"github.com/scottlogic/gemara-catalog-validator/internal/check"
)

// uniqueIDs verifies each id is defined exactly once across the whole set.
type uniqueIDs struct{}

func (uniqueIDs) ID() string { return "unique-ids" }

func (uniqueIDs) Description() string {
	return "Every id must be defined exactly once across the catalog set."
}

func (uniqueIDs) Run(ctx *check.Context) []check.Finding {
	var out []check.Finding
	for _, dup := range ctx.Index.DupDefs {
		first := ctx.Index.Defs[dup.ID]
		out = append(out, check.Finding{
			RuleID:   "unique-ids",
			Severity: check.Error,
			File:     dup.Location.File,
			Line:     dup.Location.Line,
			Col:      dup.Location.Col,
			Message: fmt.Sprintf("duplicate id %q (first defined at %s:%d)",
				dup.ID, first.Location.File, first.Location.Line),
		})
	}
	return out
}

func init() { check.Register(uniqueIDs{}) }
