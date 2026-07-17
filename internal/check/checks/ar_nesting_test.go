package checks

import (
	"testing"
)

func TestARNesting_FlagsMisnestedAndMalformed(t *testing.T) {
	ctx := loadCtx(t, fixture("ar"))
	got := (arNesting{}).Run(ctx)
	// Two of the three requirements are wrong: one nests under the wrong
	// control, one has a non-AR suffix.
	if len(got) != 2 {
		t.Fatalf("expected 2 findings, got %d: %+v", len(got), got)
	}
	for _, f := range got {
		if f.Line == 0 || f.File == "" {
			t.Errorf("finding missing position: %+v", f)
		}
		if f.RuleID != "ar-nesting" {
			t.Errorf("unexpected rule id %q", f.RuleID)
		}
	}
}

func TestARNesting_WellFormedIsSilent(t *testing.T) {
	// fixtures/valid has no controls/ARs at all — nothing to flag.
	ctx := loadCtx(t, fixture("valid"))
	if got := (arNesting{}).Run(ctx); len(got) != 0 {
		t.Fatalf("expected no findings, got %d: %+v", len(got), got)
	}
}
