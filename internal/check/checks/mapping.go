package checks

import (
	"fmt"

	"github.com/sshiells-scottlogic/gemara-catalog-validator/internal/check"
)

// mappingCompleteness verifies the core relationships are present: every
// control mitigates at least one threat, and every threat is scoped to at
// least one capability. A missing mapping is a data-completeness defect that
// schema validation does not catch.
type mappingCompleteness struct{}

func (mappingCompleteness) ID() string { return "mapping-completeness" }

func (mappingCompleteness) Description() string {
	return "Every control must map to a threat, and every threat to a capability."
}

func (mappingCompleteness) Run(ctx *check.Context) []check.Finding {
	var out []check.Finding
	for _, it := range ctx.Index.Items {
		switch it.Kind {
		case "control":
			if len(it.Mappings["threats"]) == 0 {
				out = append(out, check.Finding{
					RuleID:   "mapping-completeness",
					Severity: check.Error,
					File:     it.Location.File,
					Line:     it.Location.Line,
					Col:      it.Location.Col,
					Message:  fmt.Sprintf("control %q maps to no threat", it.ID),
				})
			}
		case "threat":
			if len(it.Mappings["capabilities"]) == 0 {
				out = append(out, check.Finding{
					RuleID:   "mapping-completeness",
					Severity: check.Error,
					File:     it.Location.File,
					Line:     it.Location.Line,
					Col:      it.Location.Col,
					Message:  fmt.Sprintf("threat %q maps to no capability", it.ID),
				})
			}
		}
	}
	return out
}

// mappingOrphans flags entries on the far end of a missing relationship: a
// threat no control mitigates, or a capability no threat references. These are
// reported as warnings — they can be legitimate gaps rather than errors.
type mappingOrphans struct{}

func (mappingOrphans) ID() string { return "mapping-orphans" }

func (mappingOrphans) Description() string {
	return "Warn on threats not referenced by any control, and capabilities not referenced by any threat."
}

func (mappingOrphans) Run(ctx *check.Context) []check.Finding {
	referencedThreats := map[string]bool{}
	referencedCaps := map[string]bool{}
	for _, it := range ctx.Index.Items {
		for _, m := range it.Mappings["threats"] {
			referencedThreats[m.RefID] = true
		}
		for _, m := range it.Mappings["capabilities"] {
			referencedCaps[m.RefID] = true
		}
	}

	var out []check.Finding
	for _, it := range ctx.Index.Items {
		switch it.Kind {
		case "threat":
			if !referencedThreats[it.ID] {
				out = append(out, check.Finding{
					RuleID:   "mapping-orphans",
					Severity: check.Warning,
					File:     it.Location.File,
					Line:     it.Location.Line,
					Col:      it.Location.Col,
					Message:  fmt.Sprintf("threat %q is not referenced by any control", it.ID),
				})
			}
		case "capability":
			if !referencedCaps[it.ID] {
				out = append(out, check.Finding{
					RuleID:   "mapping-orphans",
					Severity: check.Warning,
					File:     it.Location.File,
					Line:     it.Location.Line,
					Col:      it.Location.Col,
					Message:  fmt.Sprintf("capability %q is not referenced by any threat", it.ID),
				})
			}
		}
	}
	return out
}

func init() {
	check.Register(mappingCompleteness{})
	// Orphan detection is opt-in (see -orphans): useful as a build-out stat,
	// but noisy while catalogs are still being populated.
	check.RegisterOptional("orphans", mappingOrphans{})
}
