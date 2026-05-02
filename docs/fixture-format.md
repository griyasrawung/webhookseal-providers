# Fixture Format

Fixture files define deterministic verification cases per provider.

Location pattern:
- `fixtures/<provider_id>.fixtures.json`

## Top-level schema

```json
{
  "provider_id": "stripe",
  "cases": [
    {
      "name": "valid_basic",
      "secret": "whsec_test_alpha",
      "headers": {"Stripe-Signature": "t=1714600000,v1=<hex>"},
      "body": "{\"id\":\"evt_001\",\"object\":\"event\"}",
      "url": "",
      "params": {},
      "timestamp": 1714600000,
      "expected_error": null
    }
  ]
}
```

## Case fields

### `name` (string)
Unique identifier for the case within the file.

### `secret` (string)
Signing secret used to generate or validate signature.

### `headers` (object)
Request headers for the case. Include exact signature and timestamp headers expected by provider spec.

### `body` (string)
Raw request body as verifier input. Keep exact bytes stable.

### `url` (string)
Request URL if provider signature depends on URL components. Use empty string when not needed.

### `params` (object)
Auxiliary request parameters when provider algorithm references them. Use empty object when unused.

### `timestamp` (integer)
Deterministic verifier clock value used to evaluate replay window logic.

### `expected_error` (string or null)
- `null` for valid cases.
- explicit error code for invalid cases.

## Deterministic guidance

- Use fixed timestamps only.
- Use static, readable test secrets.
- Keep payload text unchanged between generations and checks.
- Avoid runtime-generated identifiers unless fixed in fixture body.
- Do not depend on wall clock or random data.

Determinism keeps fixture results reproducible in CI and local runs.

## Coverage guidance

Each provider fixture set should include:
- at least one valid signature case.
- malformed signature format case.
- missing signature header case.
- bad signature mismatch case.
- replay or timestamp expiry case when applicable.
- provider-specific parsing edge cases.

## Error taxonomy

Common expected error codes used by fixture cases:
- `ERR_MISSING_SIGNATURE`, required signature header missing.
- `ERR_BAD_FORMAT`, signature or timestamp field parse failure.
- `ERR_BAD_SIGNATURE`, signature value does not verify.
- `ERR_TIMESTAMP_EXPIRED`, timestamp outside replay tolerance.

Use stable canonical error names. Keep naming consistent across providers and tests.

## Naming conventions

Suggested case name style:
- `valid_basic`
- `valid_multi_signature`
- `missing_header`
- `bad_format_no_timestamp`
- `expired_timestamp`
- `bad_signature`

Clear names make fixture failures easy to diagnose.
