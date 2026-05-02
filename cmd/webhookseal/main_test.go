package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupWarningOnlyProvider(t *testing.T, providerID string) {
	t.Helper()
	tmp := t.TempDir()
	providersDir := filepath.Join(tmp, "providers")
	if err := os.MkdirAll(providersDir, 0o755); err != nil {
		t.Fatal(err)
	}
	spec := "provider_id: " + providerID + "\nnotes: short\nsource_docs:\n  - url: https://example.com\n    retrieved_date: 2099-01-01\n"
	if err := os.WriteFile(filepath.Join(providersDir, providerID+".yaml"), []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	originalProviders := providers
	providers = []string{providerID}
	t.Cleanup(func() {
		providers = originalProviders
		_ = os.Chdir(wd)
	})
}

func chdirToModuleRoot(t *testing.T) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Clean(filepath.Join(wd, "..", ".."))
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
}

func TestRun_GlobalHelpLong(t *testing.T) {
	var out bytes.Buffer
	code := run([]string{"--help"}, &out)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(out.String(), "commands:") {
		t.Fatalf("expected help output, got %q", out.String())
	}
}

func TestRun_GlobalHelpShort(t *testing.T) {
	var out bytes.Buffer
	code := run([]string{"-h"}, &out)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(out.String(), "webhookseal CLI") {
		t.Fatalf("expected help output, got %q", out.String())
	}
}

func TestRun_GlobalVersion(t *testing.T) {
	var out bytes.Buffer
	code := run([]string{"--version"}, &out)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if strings.TrimSpace(out.String()) != version {
		t.Fatalf("expected version %q, got %q", version, out.String())
	}
}

func TestRun_LintStrictFailsOnWarnings(t *testing.T) {
	setupWarningOnlyProvider(t, "shopify")
	var out bytes.Buffer
	code := run([]string{"lint", "--provider", "shopify", "--strict"}, &out)
	if code == 0 {
		t.Fatalf("expected non-zero exit code in strict mode when warnings exist; output: %q", out.String())
	}
	if !strings.Contains(out.String(), "0 lint errors, 1 lint warnings") {
		t.Fatalf("expected warning summary output, got %q", out.String())
	}
}

func TestRun_LintNonStrictAllowsWarnings(t *testing.T) {
	setupWarningOnlyProvider(t, "shopify")
	var out bytes.Buffer
	code := run([]string{"lint", "--provider", "shopify"}, &out)
	if code != 0 {
		t.Fatalf("expected zero exit code in non-strict mode with warnings only; output: %q", out.String())
	}
	if !strings.Contains(out.String(), "0 lint errors, 1 lint warnings") {
		t.Fatalf("expected warning summary output, got %q", out.String())
	}
}
