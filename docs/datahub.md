# DataHub Publishing

StreamForge emits DataHub Metadata Change Proposals for the replayed Iceberg table. The publisher covers:

- dataset properties and schema version history
- ownership
- schema metadata
- upstream Kafka lineage and field lineage
- quality checks as external assertion entities
- operational tags

Preview the proposals:

```bash
make datahub-dry-run
```

The generated JSONL file is written to `build/datahub-mcps.jsonl`. Each line is a Metadata Change Proposal with an `aspect.value` payload serialized as JSON.

Publish to DataHub GMS:

```bash
DATAHUB_GMS_URL=http://localhost:8080 \
DATAHUB_TOKEN=optional-token \
PYTHONPATH=python python3 -m streamforge_datahub publish \
  --spec metadata/datahub/orders_dataset.json
```

The publisher posts each proposal to `/aspects?action=ingestProposal` with the Rest.li protocol header. Use `--dry-run` before publishing when changing schema or assertion definitions.
