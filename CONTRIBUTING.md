# Contributing to WebhookSeal Providers

Thanks for contributing. This repository is data-first, so quality comes from deterministic specs and fixtures.

## Contribution workflow

1. Fork the repository and create a focused branch.
2. Update or add a provider spec in `providers/*.yaml`.
3. Update or add matching fixtures in `fixtures/*.fixtures.json`.
4. Run required checks locally.
5. Open a pull request with context, source docs, and rationale.

## What to include in a provider update

- Schema-valid spec fields.
- Source documentation links with retrieval date.
- Notes that capture parsing caveats and edge conditions.
- Deterministic valid and invalid fixture cases.

## Required local checks

Run from `webhookseal-providers`:

```bash
& "C:\Program Files\Go\bin\go.exe" test ./...
& "C:\Program Files\Go\bin\go.exe" run ./cmd/webhookseal validate-schema --all
& "C:\Program Files\Go\bin\go.exe" run ./cmd/webhookseal run-fixtures --all
& "C:\Program Files\Go\bin\go.exe" run ./cmd/webhookseal lint --all
```

All commands should pass before requesting review.

## Fixture quality expectations

- Keep every case deterministic.
- Use fixed secrets, timestamps, and bodies.
- Avoid dynamic time or generated identifiers.
- Cover both success paths and expected error paths.
- Keep `expected_error` stable and explicit.

## Pull request checklist

- [ ] Provider spec validates against schema.
- [ ] Fixture file maps to correct `provider_id`.
- [ ] Cases cover required parser and signature behavior.
- [ ] Lint and tests pass locally.
- [ ] Documentation is updated when format or behavior changes.

## Commit guidance

Use small, reviewable commits:
- one logical provider change per commit when possible.
- include fixtures in the same commit as spec changes.
- keep unrelated formatting or refactors out of provider changes.

## Reporting issues

When filing an issue, include:
- provider id and relevant spec snippet.
- failing fixture case and expected behavior.
- command output from `validate-schema`, `run-fixtures`, or `lint`.
