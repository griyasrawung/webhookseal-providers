# Provider Spec Format

Provider specs are YAML files under `providers/` and must satisfy `schemas/provider-spec.schema.json`.

## Required fields

### `spec_version` (string)
Semantic version for the spec document, format `MAJOR.MINOR.PATCH`.

Example:
```yaml
spec_version: "1.0.0"
```

### `provider_id` (string)
Stable provider key, lowercase alphanumeric with hyphens, pattern `^[a-z][a-z0-9-]*$`.

Example:
```yaml
provider_id: "stripe"
```

### `display_name` (string)
Human-readable provider name.

### `algorithm` (enum)
Allowed values:
- `hmac-sha256`
- `hmac-sha1`
- `hmac-sha512`

### `signature_header` (string)
Request header that carries signature data.

### `signature_prefix` (string or null)
Prefix used before signature value, for example `v1=`. Use `null` if no prefix is expected.

### `signature_encoding` (enum)
Allowed values:
- `hex`
- `base64`

### `timestamp_header` (string or null)
Header that carries timestamp when separate from signature header.

### `timestamp_format` (enum or null)
Allowed values:
- `epoch_seconds`
- `epoch_ms`
- `iso8601`
- `embedded`
- `null`

### `timestamp_location` (enum or null)
Allowed values:
- `separate_header`
- `embedded_in_signature`
- `null`

### `payload_construction` (enum)
How signed payload is built:
- `raw_body`
- `json_canonical`
- `custom`

### `payload_template` (string or null)
Template used for `custom` payload construction.

Example:
```yaml
payload_template: "{timestamp}.{body}"
```

### `replay_window_seconds` (integer or null)
Non-negative replay tolerance window. Use `null` when not applicable.

### `source_docs` (array)
At least one source document object. Each object requires:
- `url`, valid URI
- `retrieved_date`, date in `YYYY-MM-DD`

### `notes` (string)
Implementation caveats, parsing details, and behavior constraints.

## Optional fields

### `signature_parse_pattern` (string)
Regex used to extract signature value from composite header.

### `timestamp_parse_pattern` (string)
Regex used to extract timestamp value from composite header.

### `multiple_signatures` (boolean)
Whether multiple signatures may appear in one header. Default is `false`.

### `extra_headers` (array of strings)
Additional headers required for payload construction or verification context.

### `extensions` (object)
Provider-specific metadata for future tooling.

## Example, Stripe style embedded timestamp

```yaml
spec_version: "1.0.0"
provider_id: "stripe"
display_name: "Stripe"
algorithm: "hmac-sha256"
signature_header: "Stripe-Signature"
signature_prefix: "v1="
signature_encoding: "hex"
timestamp_header: null
timestamp_format: "epoch_seconds"
timestamp_location: "embedded_in_signature"
payload_construction: "custom"
payload_template: "{timestamp}.{body}"
replay_window_seconds: 300
signature_parse_pattern: "v1=([a-f0-9]+)"
timestamp_parse_pattern: "t=(\\d+)"
multiple_signatures: true
source_docs:
  - url: "https://docs.stripe.com/webhooks/signatures"
    retrieved_date: "2026-05-02"
notes: "Accept if any v1 value matches during key rotation."
```

## Validation notes

- Additional unknown top-level fields are rejected.
- Required fields must exist even when value is `null` for nullable properties.
- Regex patterns should be anchored or specific enough to avoid accidental matches.
