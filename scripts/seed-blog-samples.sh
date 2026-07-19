#!/usr/bin/env bash
# Seed ~48 sample blog articles into the local SQLite dev database.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT/backend"
export DB_DSN="${DB_DSN:-file:./data/inkless.db?cache=shared&mode=rwc}"
mkdir -p data
go run ./cmd/inkless/ seed blog-samples --dsn "$DB_DSN"
