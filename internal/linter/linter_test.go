package linter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeSpec(t *testing.T, filename string, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), filename)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func issueExists(issues []Issue, severity Severity, contains string) bool {
	for _, issue := range issues {
		if issue.Severity == severity && strings.Contains(issue.Message, contains) {
			return true
		}
	}
	return false
}

func TestLintSpec_NoIssues(t *testing.T) {
	recentDate := time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	p := writeSpec(t, "stripe.yaml", "provider_id: stripe\nalgorithm: hmac-sha256\nnotes: this is a sufficiently long notes section\nsource_docs:\n  - url: https://example.com\n    retrieved_date: "+recentDate+"\npayload_construction: raw\n")
	issues := LintSpec(p)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %v", issues)
	}
}

func TestLintSpec_MissingSourceDocs_Error(t *testing.T) {
	p := writeSpec(t, "stripe.yaml", "provider_id: stripe\nnotes: this is a sufficiently long notes section\n")
	issues := LintSpec(p)
	if !issueExists(issues, SeverityError, "source_docs") {
		t.Fatalf("expected source_docs error, got %v", issues)
	}
}

func TestLintSpec_OutdatedRetrievedDate_Warning(t *testing.T) {
	oldDate := time.Now().AddDate(-2, 0, 0).Format("2006-01-02")
	p := writeSpec(t, "stripe.yaml", "provider_id: stripe\nnotes: this is a sufficiently long notes section\nsource_docs:\n  - url: https://example.com\n    retrieved_date: "+oldDate+"\n")
	issues := LintSpec(p)
	if !issueExists(issues, SeverityWarning, "retrieved_date") {
		t.Fatalf("expected outdated retrieved_date warning, got %v", issues)
	}
}

func TestLintSpec_ShortNotes_Warning(t *testing.T) {
	p := writeSpec(t, "stripe.yaml", "provider_id: stripe\nnotes: short\nsource_docs:\n  - url: https://example.com\n    retrieved_date: 2099-01-01\n")
	issues := LintSpec(p)
	if !issueExists(issues, SeverityWarning, "notes") {
		t.Fatalf("expected notes warning, got %v", issues)
	}
}

func TestLintSpec_ProviderIDMismatch_Error(t *testing.T) {
	p := writeSpec(t, "stripe.yaml", "provider_id: github\nnotes: this is a sufficiently long notes section\nsource_docs:\n  - url: https://example.com\n    retrieved_date: 2099-01-01\n")
	issues := LintSpec(p)
	if !issueExists(issues, SeverityError, "provider_id") {
		t.Fatalf("expected provider_id mismatch error, got %v", issues)
	}
}

func TestLintSpec_TimestampWithoutReplayWindow_Warning(t *testing.T) {
	p := writeSpec(t, "stripe.yaml", "provider_id: stripe\nnotes: this is a sufficiently long notes section\nsource_docs:\n  - url: https://example.com\n    retrieved_date: 2099-01-01\ntimestamp_header: X-Timestamp\n")
	issues := LintSpec(p)
	if !issueExists(issues, SeverityWarning, "replay_window_seconds") {
		t.Fatalf("expected replay_window_seconds warning, got %v", issues)
	}
}

func TestLintSpec_UnsupportedAlgorithm_Error(t *testing.T) {
	p := writeSpec(t, "stripe.yaml", "provider_id: stripe\nalgorithm: rsa-sha256\nnotes: this is a sufficiently long notes section\nsource_docs:\n  - url: https://example.com\n    retrieved_date: 2099-01-01\n")
	issues := LintSpec(p)
	if !issueExists(issues, SeverityError, "unsupported algorithm") {
		t.Fatalf("expected unsupported algorithm error, got %v", issues)
	}
}

func TestLintSpec_CustomPayloadWithoutTemplate_Error(t *testing.T) {
	p := writeSpec(t, "stripe.yaml", "provider_id: stripe\npayload_construction: custom\nnotes: this is a sufficiently long notes section\nsource_docs:\n  - url: https://example.com\n    retrieved_date: 2099-01-01\n")
	issues := LintSpec(p)
	if !issueExists(issues, SeverityError, "payload_template") {
		t.Fatalf("expected payload_template error, got %v", issues)
	}
}

func TestLintSpec_EmptyExtensions_Warning(t *testing.T) {
	p := writeSpec(t, "stripe.yaml", "provider_id: stripe\nnotes: this is a sufficiently long notes section\nsource_docs:\n  - url: https://example.com\n    retrieved_date: 2099-01-01\nextensions: {}\n")
	issues := LintSpec(p)
	if !issueExists(issues, SeverityWarning, "extensions") {
		t.Fatalf("expected empty extensions warning, got %v", issues)
	}
}
