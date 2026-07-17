package check

import (
	"fmt"
	"io"
	"strings"
)

// Report writes findings to w in the requested format. Format is one of
// "text" (human/local) or "github" (Actions inline annotations).
func Report(w io.Writer, format string, findings []Finding) error {
	switch format {
	case "", "text":
		return reportText(w, findings)
	case "github":
		return reportGitHub(w, findings)
	default:
		return fmt.Errorf("unknown format %q (want: text, github)", format)
	}
}

func reportText(w io.Writer, findings []Finding) error {
	for _, f := range findings {
		loc := f.File
		if f.Line > 0 {
			loc = fmt.Sprintf("%s:%d", f.File, f.Line)
		}
		if _, err := fmt.Fprintf(w, "%s: [%s] %s: %s\n", loc, f.Severity, f.RuleID, f.Message); err != nil {
			return err
		}
	}
	return nil
}

func reportGitHub(w io.Writer, findings []Finding) error {
	for _, f := range findings {
		level := "error"
		if f.Severity == Warning {
			level = "warning"
		}
		var b strings.Builder
		fmt.Fprintf(&b, "::%s title=%s", level, escapeProp(f.RuleID))
		if f.File != "" {
			fmt.Fprintf(&b, ",file=%s", escapeProp(f.File))
		}
		if f.Line > 0 {
			fmt.Fprintf(&b, ",line=%d", f.Line)
		}
		if f.Col > 0 {
			fmt.Fprintf(&b, ",col=%d", f.Col)
		}
		fmt.Fprintf(&b, "::%s\n", escapeData(f.Message))
		if _, err := io.WriteString(w, b.String()); err != nil {
			return err
		}
	}
	return nil
}

// GitHub workflow-command escaping.
func escapeData(s string) string {
	r := strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A")
	return r.Replace(s)
}

func escapeProp(s string) string {
	r := strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A", ":", "%3A", ",", "%2C")
	return r.Replace(s)
}
