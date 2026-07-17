package checks

import (
	"fmt"
	"strings"

	"github.com/sshiells-scottlogic/gemara-catalog-validator/internal/check"
)

// idPrefixMatchesCatalog verifies every locally-defined id carries its
// catalog's canonical prefix. The prefix is taken from the catalog's
// `metadata.id` (e.g. a catalog whose metadata.id is "CCC.Monitor" must define
// ids like "CCC.Monitor.CP01"), so no per-project configuration is needed.
type idPrefixMatchesCatalog struct{}

func (idPrefixMatchesCatalog) ID() string { return "id-prefix" }

func (idPrefixMatchesCatalog) Description() string {
	return "Every defined id must be prefixed with its catalog's metadata.id."
}

func (idPrefixMatchesCatalog) Run(ctx *check.Context) []check.Finding {
	var out []check.Finding
	for _, def := range ctx.Index.AllDefs {
		cat := ctx.Index.CatalogFor(def.Location.File)
		if cat == nil || cat.Prefix == "" {
			continue // no known catalog prefix (no metadata.id) — cannot validate
		}
		if strings.HasPrefix(def.ID, cat.Prefix+".") {
			continue
		}
		out = append(out, check.Finding{
			RuleID:   "id-prefix",
			Severity: check.Error,
			File:     def.Location.File,
			Line:     def.Location.Line,
			Col:      def.Location.Col,
			Message: fmt.Sprintf("id %q does not match its catalog prefix (expected %q.*, from %s metadata.id)",
				def.ID, cat.Prefix, cat.Dir),
		})
	}
	return out
}

func init() { check.Register(idPrefixMatchesCatalog{}) }
