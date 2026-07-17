package checks

import (
	"path/filepath"
	"testing"

	"github.com/scottlogic/gemara-catalog-validator/internal/check"
	"github.com/scottlogic/gemara-catalog-validator/internal/config"
	"github.com/scottlogic/gemara-catalog-validator/internal/model"
)

func loadCtx(t *testing.T, dir string) *check.Context {
	t.Helper()
	idx, err := model.Load([]string{dir})
	if err != nil {
		t.Fatalf("load %s: %v", dir, err)
	}
	cfg := config.Default()
	cfg.IDPattern = `^EX\.[A-Za-z0-9]+\.(CP|CN|TH)[0-9]+$`
	if err := cfg.Compile(); err != nil {
		t.Fatal(err)
	}
	return &check.Context{Index: idx, Config: cfg}
}

func fixture(parts ...string) string {
	return filepath.Join(append([]string{"..", "..", "..", "fixtures"}, parts...)...)
}

func TestReferentialIntegrity_Valid(t *testing.T) {
	ctx := loadCtx(t, fixture("valid"))
	if got := (referentialIntegrity{}).Run(ctx); len(got) != 0 {
		t.Fatalf("expected no findings, got %d: %+v", len(got), got)
	}
}

func TestReferentialIntegrity_Dangling(t *testing.T) {
	ctx := loadCtx(t, fixture("invalid"))
	got := (referentialIntegrity{}).Run(ctx)
	if len(got) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(got), got)
	}
	f := got[0]
	if f.Line == 0 || f.File == "" {
		t.Errorf("finding missing position: %+v", f)
	}
	if f.RuleID != "reference-resolves" {
		t.Errorf("unexpected rule id %q", f.RuleID)
	}
}

func TestUniqueIDs_NoDuplicatesInValid(t *testing.T) {
	ctx := loadCtx(t, fixture("valid"))
	if got := (uniqueIDs{}).Run(ctx); len(got) != 0 {
		t.Fatalf("expected no duplicate-id findings, got %d: %+v", len(got), got)
	}
}
