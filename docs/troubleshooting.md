# Troubleshooting

## Replay Cannot Connect to Kafka

Check that the broker address matches the network you are running from:

- From the host, use `localhost:19092`.
- From Docker services, use `redpanda:9092`.

Run a dry-run first to validate the fixture:

```bash
go run ./cmd/streamforge replay --events examples/orders/source_events.jsonl --topic orders --dry-run
```

## Flink SQL Cannot Find Kafka or Iceberg Connectors

The compose stack builds `deploy/flink/Dockerfile`, which places the Kafka, Iceberg, Hadoop S3, and AWS SDK jars under `/opt/flink/lib`. Rebuild the Flink image after changing connector versions:

```bash
docker compose -f deploy/docker-compose.yml build flink-jobmanager flink-taskmanager
```

## Iceberg REST Cannot Write to MinIO

Confirm the MinIO bucket exists:

```bash
docker compose -f deploy/docker-compose.yml exec minio-init \
  mc ls local/warehouse
```

The local stack uses path-style S3 access and the `minio` / `minio123` credentials configured in `deploy/docker-compose.yml`.

## Diagnostics Report Shows Missing Records

Missing records usually mean one of three things:

- the replay fixture was not published to the topic
- Flink started from a later Kafka offset
- the actual write manifest does not cover the replay run id

Check the `run_id`, Kafka offset, and Iceberg snapshot fields in the finding contexts.

## Diagnostics Report Shows Duplicate Records

Duplicates indicate more than one actual write for the same `record_id`. Check whether the Flink job was restarted without a savepoint, whether the Iceberg table is configured for append instead of upsert, or whether the replay fixture was published multiple times.

## DataHub Publish Fails

Preview proposals before publishing:

```bash
make datahub-dry-run
```

For a live publish, set the GMS URL and optional token:

```bash
DATAHUB_GMS_URL=http://localhost:8080 \
DATAHUB_TOKEN=optional-token \
PYTHONPATH=python python3 -m streamforge_datahub publish \
  --spec metadata/datahub/orders_dataset.json
```

If GMS rejects an aspect, the publisher prints the aspect name and entity URN returned by the server error.
