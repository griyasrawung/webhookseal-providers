package linter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

type Issue struct {
	Severity Severity
	Message  string
}

func (i Issue) String() string {
	return fmt.Sprintf("[%s] %s", i.Severity, i.Message)
}

func LintSpec(path string) []Issue {
	data, err := os.ReadFile(path)
	if err != nil {
		return []Issue{{Severity: SeverityError, Message: fmt.Sprintf("read error: %v", err)}}
	}
	var m map[string]any
	if err := yaml.Unmarshal(data, &m); err != nil {
		return []Issue{{Severity: SeverityError, Message: fmt.Sprintf("yaml parse error: %v", err)}}
	}

	var issues []Issue

	docs, docsOK := m["source_docs"].([]any)
	if !docsOK || len(docs) == 0 {
		issues = append(issues, Issue{Severity: SeverityError, Message: "source_docs must contain at least one item"})
	} else {
		cutoff := time.Now().AddDate(-1, 0, 0)
		for _, rawDoc := range docs {
			docMap, ok := rawDoc.(map[string]any)
			if !ok {
				continue
			}
			rawDateValue, exists := docMap["retrieved_date"]
			if !exists || rawDateValue == nil {
				continue
			}
			var retrievedDate time.Time
			var rawDate string
			switch v := rawDateValue.(type) {
			case string:
				rawDate = strings.TrimSpace(v)
				if rawDate == "" {
					continue
				}
				parsed, err := time.Parse("2006-01-02", rawDate)
				if err != nil {
					continue
				}
				retrievedDate = parsed
			case time.Time:
				retrievedDate = v
				rawDate = v.Format("2006-01-02")
			default:
				continue
			}
			if retrievedDate.Before(cutoff) {
				issues = append(issues, Issue{Severity: SeverityWarning, Message: fmt.Sprintf("source_docs retrieved_date is older than 1 year: %s", rawDate)})
			}
		}
	}

	notes, notesOK := m["notes"].(string)
	trimmedNotes := strings.TrimSpace(notes)
	if !notesOK || len(trimmedNotes) < 20 {
		issues = append(issues, Issue{Severity: SeverityWarning, Message: "notes should be at least 20 non-whitespace characters"})
	}

	providerID, providerIDOK := m["provider_id"].(string)
	filenameID := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	if !providerIDOK || strings.TrimSpace(providerID) == "" || providerID != filenameID {
		issues = append(issues, Issue{Severity: SeverityError, Message: fmt.Sprintf("provider_id must match filename: expected %q", filenameID)})
	}

	hasTimestamp := false
	if header, ok := m["timestamp_header"].(string); ok && strings.TrimSpace(header) != "" {
		hasTimestamp = true
	}
	if format, ok := m["timestamp_format"].(string); ok && strings.TrimSpace(format) != "" {
		hasTimestamp = true
	}
	if hasTimestamp {
		if v, exists := m["replay_window_seconds"]; !exists || v == nil {
			issues = append(issues, Issue{Severity: SeverityWarning, Message: "replay_window_seconds should be set when timestamp fields are configured"})
		}
	}

	if alg, ok := m["algorithm"].(string); ok {
		switch alg {
		case "hmac-sha256", "hmac-sha1", "hmac-sha512":
		default:
			issues = append(issues, Issue{Severity: SeverityError, Message: fmt.Sprintf("unsupported algorithm: %s", alg)})
		}
	}

	if payloadConstruction, ok := m["payload_construction"].(string); ok && payloadConstruction == "custom" {
		template, templateOK := m["payload_template"].(string)
		if !templateOK || strings.TrimSpace(template) == "" {
			issues = append(issues, Issue{Severity: SeverityError, Message: "payload_template is required when payload_construction=custom"})
		}
	}

	if ext, exists := m["extensions"]; exists {
		if extMap, ok := ext.(map[string]any); ok && len(extMap) == 0 {
			issues = append(issues, Issue{Severity: SeverityWarning, Message: "extensions object is present but empty"})
		}
	}

	return issues
}
