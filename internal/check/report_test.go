package check

import (
	"strings"
	"testing"
)

func TestReportTextIncludesLocation(t *testing.T) {
	var b strings.Builder
	err := Report(&b, "text", []Finding{{
		RuleID: "id-prefix", Severity: Error,
		File: "catalogs/x.yaml", Line: 12, Col: 9,
		Message: "id \"X\" bad",
	}})
	if err != nil {
		t.Fatal(err)
	}
	got := b.String()
	want := "catalogs/x.yaml:12:9: [error] id-prefix: id \"X\" bad\n"
	if got != want {
		t.Errorf("text report\n got: %q\nwant: %q", got, want)
	}
}

// The GitHub raw job log shows only an annotation's message, not its file/line
// metadata, so the location must also appear in the message text.
func TestReportGitHubMessageIncludesLocation(t *testing.T) {
	var b strings.Builder
	err := Report(&b, "github", []Finding{{
		RuleID: "reference-resolves", Severity: Error,
		File: "catalogs/mgmt/threats.yaml", Line: 40, Col: 7,
		Message: "reference-id \"CCC.MLDE.TH01\" does not resolve to any defined id",
	}})
	if err != nil {
		t.Fatal(err)
	}
	got := b.String()
	// Annotation metadata is still present for the GitHub UI.
	if !strings.Contains(got, "file=catalogs/mgmt/threats.yaml,line=40,col=7") {
		t.Errorf("missing annotation metadata in %q", got)
	}
	// ...and the location is folded into the visible message.
	if !strings.Contains(got, "::catalogs/mgmt/threats.yaml:40:7: reference-id") {
		t.Errorf("location missing from visible message in %q", got)
	}
}

func TestFindingLocationUnknown(t *testing.T) {
	if got := (Finding{}).location(); got != "<unknown>" {
		t.Errorf("location() = %q, want %q", got, "<unknown>")
	}
	if got := (Finding{File: "a.yaml"}).location(); got != "a.yaml" {
		t.Errorf("location() = %q, want %q", got, "a.yaml")
	}
}
