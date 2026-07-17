package checks

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sshiells-scottlogic/gemara-catalog-validator/internal/check"
)

// arSuffix is the expected tail of an assessment-requirement id, after the
// parent control id and a dot: "AR" followed by a number (e.g. AR01).
var arSuffix = regexp.MustCompile(`^AR[0-9]+$`)

// arNesting verifies every assessment-requirement id nests under its parent
// control id, i.e. it is "<controlId>.AR<n>" (e.g. control CCC.GenAI.CN01 ->
// CCC.GenAI.CN01.AR01). Catches copy-paste ARs pointing at the wrong control
// and malformed requirement ids.
type arNesting struct{}

func (arNesting) ID() string { return "ar-nesting" }

func (arNesting) Description() string {
	return "Every assessment-requirement id must nest under its parent control id."
}

func (arNesting) Run(ctx *check.Context) []check.Finding {
	var out []check.Finding
	for _, r := range ctx.Index.Requirements {
		if r.ControlID == "" {
			continue // control has no id — a separate defect, not this check's concern
		}
		if suffix, ok := strings.CutPrefix(r.ID, r.ControlID+"."); ok && arSuffix.MatchString(suffix) {
			continue
		}
		out = append(out, check.Finding{
			RuleID:   "ar-nesting",
			Severity: check.Error,
			File:     r.Location.File,
			Line:     r.Location.Line,
			Col:      r.Location.Col,
			Message: fmt.Sprintf("assessment-requirement id %q does not nest under its control %q (expected %s.AR<n>)",
				r.ID, r.ControlID, r.ControlID),
		})
	}
	return out
}

func init() { check.Register(arNesting{}) }
