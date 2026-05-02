package validator

import (
	"path/filepath"
	"testing"

	"github.com/xeipuuv/gojsonschema"
)

func schemaPath(t *testing.T) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join("..", "..", "schemas", "provider-spec.schema.json"))
	if err != nil {
		t.Fatalf("resolve schema path: %v", err)
	}
	return path
}

func TestSchemaAcceptsValid(t *testing.T) {
	schemaLoader := gojsonschema.NewReferenceLoader("file:///" + filepath.ToSlash(schemaPath(t)))
	docLoader := gojsonschema.NewStringLoader(`{
		"spec_version":"1.0.0",
		"provider_id":"stripe",
		"display_name":"Stripe",
		"algorithm":"hmac-sha256",
		"signature_header":"Stripe-Signature",
		"signature_prefix":"v1=",
		"signature_encoding":"hex",
		"timestamp_header":null,
		"timestamp_format":"embedded",
		"timestamp_location":"embedded_in_signature",
		"payload_construction":"custom",
		"payload_template":"{timestamp}.{body}",
		"replay_window_seconds":300,
		"source_docs":[{"url":"https://example.com/docs","retrieved_date":"2026-05-02"}],
		"notes":"Deterministic test spec"
	}`)

	result, err := gojsonschema.Validate(schemaLoader, docLoader)
	if err != nil {
		t.Fatalf("schema validation error: %v", err)
	}
	if !result.Valid() {
		t.Fatalf("expected valid spec, got errors: %v", result.Errors())
	}
}

func TestSchemaRejectsInvalid(t *testing.T) {
	schemaLoader := gojsonschema.NewReferenceLoader("file:///" + filepath.ToSlash(schemaPath(t)))
	docLoader := gojsonschema.NewStringLoader(`{
		"spec_version":"1.0.0",
		"display_name":"Missing Provider ID",
		"algorithm":"hmac-sha256",
		"signature_header":"X-Signature",
		"signature_prefix":null,
		"signature_encoding":"hex",
		"timestamp_header":null,
		"timestamp_format":null,
		"timestamp_location":null,
		"payload_construction":"raw_body",
		"payload_template":null,
		"replay_window_seconds":null,
		"source_docs":[{"url":"https://example.com/docs","retrieved_date":"2026-05-02"}],
		"notes":"invalid",
		"unknown_field":"should fail"
	}`)

	result, err := gojsonschema.Validate(schemaLoader, docLoader)
	if err != nil {
		t.Fatalf("schema validation error: %v", err)
	}
	if result.Valid() {
		t.Fatalf("expected invalid spec to be rejected")
	}
}

