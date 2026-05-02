package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunFixtureFile_StripeRealValidation(t *testing.T) {
	dir := t.TempDir()
	fixturesDir := filepath.Join(dir, "fixtures")
	providersDir := filepath.Join(dir, "providers")
	if err := os.MkdirAll(fixturesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(providersDir, 0o755); err != nil {
		t.Fatal(err)
	}

	spec := `provider_id: "stripe"
algorithm: "hmac-sha256"
signature_header: "Stripe-Signature"
signature_prefix: "v1="
signature_encoding: "hex"
timestamp_header: null
payload_construction: "custom"
payload_template: "{timestamp}.{body}"
replay_window_seconds: 300
signature_parse_pattern: "v1=([a-f0-9]+)"
timestamp_parse_pattern: "t=(\\d+)"
multiple_signatures: true
`
	if err := os.WriteFile(filepath.Join(providersDir, "stripe.yaml"), []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}

	fixture := `{
  "provider_id": "stripe",
  "cases": [
    {
      "name": "valid with multiple v1 signatures accepts any",
      "secret": "whsec_test",
      "headers": {
        "Stripe-Signature": "t=1715000000,v1=deadbeef,v1=1b461b1198dbfa7fa7ca8c9fbc1e3fc0b34c44b05e1722cceb250c252f9634dd"
      },
      "body": "{\"id\":\"evt_1\"}",
      "url": "",
      "params": {},
      "timestamp": 1715000000,
      "expected_error": null
    },
    {
      "name": "expired timestamp",
      "secret": "whsec_test",
      "headers": {
        "Stripe-Signature": "t=1715000000,v1=1b461b1198dbfa7fa7ca8c9fbc1e3fc0b34c44b05e1722cceb250c252f9634dd"
      },
      "body": "{\"id\":\"evt_1\"}",
      "url": "",
      "params": {},
      "timestamp": 1715000401,
      "expected_error": "ERR_TIMESTAMP_EXPIRED"
    }
  ]
}`
	fixturePath := filepath.Join(fixturesDir, "stripe.fixtures.json")
	if err := os.WriteFile(fixturePath, []byte(fixture), 0o644); err != nil {
		t.Fatal(err)
	}

	pass, fail, err := RunFixtureFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	if pass != 2 || fail != 0 {
		t.Fatalf("got pass=%d fail=%d", pass, fail)
	}
}

func TestVerifyFixtureCaseMissingErrors(t *testing.T) {
	spec := &ProviderSpec{
		ProviderID:          "github",
		Algorithm:           "hmac-sha256",
		SignatureHeader:     "X-Hub-Signature-256",
		SignatureEncoding:   "hex",
		PayloadConstruction: "raw_body",
	}

	if err := verifyFixtureCase(spec, FixtureCase{Headers: map[string]string{}}); err == nil || err.Error() != "ERR_MISSING_SIGNATURE" {
		t.Fatalf("expected ERR_MISSING_SIGNATURE, got %v", err)
	}
}
