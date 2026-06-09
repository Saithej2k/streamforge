# Local Workflow

This workflow runs the replay fixture, starts the local streaming stack, submits the Flink SQL job, and compares the expected records with Iceberg write evidence.

## Host Demo

Use this path when Docker is not available:

```bash
scripts/run_local_demo.sh
```

The script runs the Go test suite, validates the replay fixture in dry-run mode, and writes `build/diagnostics-report.json` from the example Iceberg write manifest.

Build DataHub metadata proposals:

```bash
make datahub-dry-run
```

## Docker Stack

Start the services:

```bash
docker compose -f deploy/docker-compose.yml up -d redpanda minio minio-init iceberg-rest flink-jobmanager flink-taskmanager
```

Replay the order fixture into Redpanda:

```bash
go run ./cmd/streamforge replay \
  --events examples/orders/source_events.jsonl \
  --topic orders \
  --brokers localhost:19092
```

Submit the Flink SQL job:

```bash
docker compose -f deploy/docker-compose.yml exec flink-jobmanager \
  /opt/flink/bin/sql-client.sh -f /opt/streamforge/flink/sql/orders_replay_to_iceberg.sql
```

Run diagnostics against the captured write manifest:

```bash
go run ./cmd/streamforge diagnose \
  --expected examples/orders/expected_iceberg.jsonl \
  --actual examples/orders/actual_iceberg.jsonl \
  --late-after 5m \
  --format json \
  --out build/diagnostics-report.json
```

Stop the stack:

```bash
docker compose -f deploy/docker-compose.yml down -v
```

See `docs/troubleshooting.md` for common replay, Flink, Iceberg, and DataHub issues.
