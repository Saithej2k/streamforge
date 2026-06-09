# StreamForge

StreamForge is a lakehouse replay and diagnostics toolkit for checking whether Kafka events land in Iceberg exactly as expected. It reads expected replay records and actual Iceberg write manifests, compares them by record id, and emits a trace-rich report for missing, late, duplicate, mismatched, and unexpected records.

## Quick Start

Run the fixture diagnostics:

```bash
go run ./cmd/streamforge diagnose \
  --expected examples/orders/expected_iceberg.jsonl \
  --actual examples/orders/actual_iceberg.jsonl \
  --late-after 5m
```

Write a JSON report:

```bash
go run ./cmd/streamforge diagnose \
  --expected examples/orders/expected_iceberg.jsonl \
  --actual examples/orders/actual_iceberg.jsonl \
  --format json \
  --out build/diagnostics-report.json
```

## Current Capabilities

- Compare expected replay records against actual Iceberg writes.
- Flag missing, late, duplicate, hash-mismatched, and unexpected records.
- Preserve Kafka topic, partition, offset, key, run id, Iceberg snapshot id, and data file path in each finding.
- Produce text output for operators and JSON output for CI or follow-up automation.

## Input Shape

Expected records are newline-delimited JSON:

```json
{"record_id":"ord-1001","event_time":"2026-01-12T10:00:01Z","payload_hash":"sha256:4f45775c51e2c6a9","trace":{"topic":"orders","partition":0,"offset":101,"key":"ord-1001","run_id":"demo-20260112"}}
```

Actual writes include lakehouse placement metadata:

```json
{"record_id":"ord-1001","event_time":"2026-01-12T10:00:01Z","written_at":"2026-01-12T10:01:01Z","payload_hash":"sha256:4f45775c51e2c6a9","snapshot_id":"915204001","file_path":"s3://warehouse/orders/data/00001.parquet","trace":{"topic":"orders","partition":0,"offset":101,"key":"ord-1001","run_id":"demo-20260112"}}
```

## Development

```bash
go test ./...
```
