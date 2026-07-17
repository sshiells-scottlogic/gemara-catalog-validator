package checks

import (
	"strings"
	"testing"
)

func TestIDPrefix_Mismatch(t *testing.T) {
	ctx := loadCtx(t, fixture("prefix"))
	got := (idPrefixMatchesCatalog{}).Run(ctx)
	if len(got) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(got), got)
	}
	f := got[0]
	if !strings.Contains(f.Message, "EX.Wrong.CP01") {
		t.Errorf("finding should name the offending id, got: %s", f.Message)
	}
	if f.Line == 0 || f.File == "" {
		t.Errorf("finding missing position: %+v", f)
	}
	if f.RuleID != "id-prefix" {
		t.Errorf("unexpected rule id %q", f.RuleID)
	}
}

func TestIDPrefix_NoMetadataSkips(t *testing.T) {
	// fixtures/valid has no metadata.yaml, so there is no prefix to validate
	// against — the check must stay silent rather than false-positive.
	ctx := loadCtx(t, fixture("valid"))
	if got := (idPrefixMatchesCatalog{}).Run(ctx); len(got) != 0 {
		t.Fatalf("expected no findings without metadata.id, got %d: %+v", len(got), got)
	}
}
