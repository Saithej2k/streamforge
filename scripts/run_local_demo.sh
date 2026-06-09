#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

mkdir -p build

go test ./...

go run ./cmd/streamforge replay \
  --events examples/orders/source_events.jsonl \
  --topic orders \
  --dry-run

go run ./cmd/streamforge diagnose \
  --expected examples/orders/expected_iceberg.jsonl \
  --actual examples/orders/actual_iceberg.jsonl \
  --late-after 5m \
  --format json \
  --out build/diagnostics-report.json

echo "Wrote build/diagnostics-report.json"
