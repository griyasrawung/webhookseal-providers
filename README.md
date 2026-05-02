# WebhookSeal Providers

`webhookseal-providers` is the data and validation repository for WebhookSeal provider definitions.

This repository is open-core infrastructure for webhook signature verification:
- It stores provider verification specs in a machine-validated format.
- It stores deterministic fixtures for reproducible verification tests.
- It ships a CLI that validates specs and runs fixture-based checks.

The goal is simple, stable, reviewable provider metadata that can be used by WebhookSeal tooling and downstream integrations.

## Repository structure

- `schemas/provider-spec.schema.json`, JSON Schema for provider specification files.
- `providers/*.yaml`, provider specification documents.
- `fixtures/*.fixtures.json`, deterministic provider test fixtures.
- `cmd/webhookseal`, CLI entrypoint.
- `internal/validator`, schema and spec validation logic.
- `internal/runner`, fixture execution and result checking.
- `internal/linter`, aggregate checks behind `lint`.

## Quick start

From this directory:

```bash
go mod download
go run ./cmd/webhookseal validate-schema --all
go run ./cmd/webhookseal run-fixtures --all
go run ./cmd/webhookseal lint --all
```

Expected flow:
1. Install dependencies.
2. Validate provider specs against schema.
3. Run deterministic fixtures.
4. Run full lint checks.

## CI/local verification commands

Run from `webhookseal-providers/`:

```bash
go test ./... -v
go run ./cmd/webhookseal validate-schema --all
go run ./cmd/webhookseal run-fixtures --all
go run ./cmd/webhookseal lint --all --strict
go build -o webhookseal ./cmd/webhookseal
```

Expected passing outcome:
- All commands exit `0`.
- `go test ./... -v` ends with `PASS`.
- Validation/lint commands print success summaries without errors.

CI uses `go test ./... -v -race -coverprofile=coverage.txt`; local `-race` is optional and not required in this environment.

## Provider spec format summary

Provider specs live in `providers/*.yaml` and must conform to `schemas/provider-spec.schema.json`.

Required top-level fields:
- `spec_version`
- `provider_id`
- `display_name`
- `algorithm`
- `signature_header`
- `signature_prefix`
- `signature_encoding`
- `timestamp_header`
- `timestamp_format`
- `timestamp_location`
- `payload_construction`
- `payload_template`
- `replay_window_seconds`
- `source_docs`
- `notes`

Common optional fields:
- `signature_parse_pattern`
- `timestamp_parse_pattern`
- `multiple_signatures`
- `extra_headers`
- `extensions`

See `docs/spec-format.md` for full field semantics, constraints, and examples.

## Fixture format summary

Fixture files live in `fixtures/*.fixtures.json` and define deterministic verification cases for a single provider.

Top-level shape:
- `provider_id`, must match target provider spec.
- `cases`, array of case objects.

Each case includes:
- `name`, unique test case identifier.
- `secret`, HMAC or signing secret used for the case.
- `headers`, request headers map.
- `body`, raw request body string.
- `url`, request URL when relevant.
- `params`, auxiliary request parameters.
- `timestamp`, fixed verifier clock input.
- `expected_error`, `null` for valid or an error code for invalid.

See `docs/fixture-format.md` for full schema guidance and error taxonomy.

## CLI reference

Run from repository root:

```bash
go run ./cmd/webhookseal <command> [flags]
```

Commands:
- `validate-schema`, validate provider specs against JSON Schema.
- `run-fixtures`, execute fixture cases against provider rules.
- `lint`, run combined checks, schema plus fixture coverage and consistency.

Common flags:
- `--all`, process all provider or fixture files.
- `--provider <provider_id>`, target a single provider where command supports it.

Use `-h` on the root command or any subcommand for the current flag set:

```bash
go run ./cmd/webhookseal -h
go run ./cmd/webhookseal lint -h
```

## Open-core boundary and secret handling

`webhookseal-providers` is an open provider-registry repository. It contains provider signing specifications, validation schema, and deterministic fixtures.

What this repository includes:
- Open provider definitions and fixture data.
- Validation and lint tooling for provider rule quality.

What this repository does not include:
- Any proprietary/closed service deployment layer.
- Any secret storage implementation.
- Any runtime secret values in repository data.

Secret handling expectations:
- Fixtures and provider specs must never contain live secrets.
- Only non-sensitive sample values are allowed for test data.
- Secret lifecycle/operations are outside this repository scope.

Version baseline:
- `webhookseal-providers v0.1.0-beta.1`

## Contributing

See `CONTRIBUTING.md` for contribution workflow, validation checks, and review expectations.

## License

This repository is licensed under the terms in `LICENSE`.
