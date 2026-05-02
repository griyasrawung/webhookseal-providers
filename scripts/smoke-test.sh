#!/usr/bin/env bash
set -euo pipefail
go run ./cmd/webhookseal validate-schema --all
go run ./cmd/webhookseal run-fixtures --all
go run ./cmd/webhookseal lint --all