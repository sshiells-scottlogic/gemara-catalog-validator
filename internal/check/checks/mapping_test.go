package checks

import (
	"strings"
	"testing"
)

func TestMappingCompleteness_FlagsMissingMappings(t *testing.T) {
	ctx := loadCtx(t, fixture("mapping"))
	got := (mappingCompleteness{}).Run(ctx)
	// CN02 maps to no threat; TH02 maps to no capability.
	if len(got) != 2 {
		t.Fatalf("expected 2 findings, got %d: %+v", len(got), got)
	}
	var control, threat bool
	for _, f := range got {
		if f.RuleID != "mapping-completeness" {
			t.Errorf("unexpected rule id %q", f.RuleID)
		}
		if f.Line == 0 {
			t.Errorf("finding missing position: %+v", f)
		}
		if strings.Contains(f.Message, "EX.Svc.CN02") {
			control = true
		}
		if strings.Contains(f.Message, "EX.Svc.TH02") {
			threat = true
		}
	}
	if !control || !threat {
		t.Errorf("expected findings for CN02 and TH02, got: %+v", got)
	}
}

func TestMappingOrphans_FlagsUnreferenced(t *testing.T) {
	ctx := loadCtx(t, fixture("mapping"))
	got := (mappingOrphans{}).Run(ctx)
	// TH02 is mitigated by no control; CP02 is referenced by no threat.
	if len(got) != 2 {
		t.Fatalf("expected 2 findings, got %d: %+v", len(got), got)
	}
	for _, f := range got {
		if f.Severity != "warning" {
			t.Errorf("orphans should be warnings, got %q", f.Severity)
		}
	}
}
