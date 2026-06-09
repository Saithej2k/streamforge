from __future__ import annotations

import hashlib
import json
import time
from typing import Any

CONTENT_TYPE = "application/json"
REGISTRY_NAME = "streamforge"
REGISTRY_VERSION = "0.1.0"


def build_proposals(spec: dict[str, Any], observed_at_ms: int | None = None) -> list[dict[str, Any]]:
    observed_at_ms = observed_at_ms or int(time.time() * 1000)
    run_id = spec.get("run_id", "streamforge-datahub")
    dataset = dataset_urn(spec["platform"], spec["name"], spec.get("env", "PROD"))
    actor = spec.get("actor", "urn:li:corpuser:streamforge")

    proposals = [
        mcp("dataset", dataset, "datasetProperties", dataset_properties(spec), run_id, observed_at_ms),
        mcp("dataset", dataset, "ownership", ownership(spec, actor, observed_at_ms), run_id, observed_at_ms),
        mcp("dataset", dataset, "schemaMetadata", schema_metadata(spec), run_id, observed_at_ms),
        mcp("dataset", dataset, "upstreamLineage", upstream_lineage(spec, dataset), run_id, observed_at_ms),
        mcp("dataset", dataset, "globalTags", global_tags(spec), run_id, observed_at_ms),
    ]

    for check in spec.get("quality_checks", []):
        assertion = assertion_urn(dataset, check["id"])
        proposals.append(
            mcp(
                "assertion",
                assertion,
                "assertionInfo",
                assertion_info(check, dataset, actor, observed_at_ms),
                run_id,
                observed_at_ms,
            )
        )

    return proposals


def dataset_urn(platform: str, name: str, env: str = "PROD") -> str:
    return f"urn:li:dataset:(urn:li:dataPlatform:{platform},{name},{env})"


def schema_field_urn(dataset: str, field_path: str) -> str:
    return f"urn:li:schemaField:({dataset},{field_path})"


def assertion_urn(dataset: str, check_id: str) -> str:
    digest = hashlib.sha1(f"{dataset}:{check_id}".encode("utf-8")).hexdigest()
    return f"urn:li:assertion:streamforge-{digest}"


def mcp(entity_type: str, entity_urn: str, aspect_name: str, aspect: dict[str, Any], run_id: str, observed_at_ms: int) -> dict[str, Any]:
    return {
        "entityType": entity_type,
        "entityUrn": entity_urn,
        "changeType": "UPSERT",
        "aspectName": aspect_name,
        "aspect": {
            "value": json.dumps(aspect, separators=(",", ":"), sort_keys=True),
            "contentType": CONTENT_TYPE,
        },
        "systemMetadata": {
            "lastObserved": observed_at_ms,
            "runId": run_id,
            "registryName": REGISTRY_NAME,
            "registryVersion": REGISTRY_VERSION,
            "properties": {"source": "streamforge"},
        },
    }


def dataset_properties(spec: dict[str, Any]) -> dict[str, Any]:
    custom = {
        "env": spec.get("env", "PROD"),
        "schema_version": str(spec.get("schema_version", 1)),
        "tool": "streamforge",
    }
    for index, entry in enumerate(spec.get("schema_history", []), start=1):
        custom[f"schema_history_{index}"] = entry
    custom.update(spec.get("custom_properties", {}))
    return {
        "name": spec["name"].split(".")[-1],
        "description": spec.get("description", ""),
        "customProperties": custom,
    }


def ownership(spec: dict[str, Any], actor: str, observed_at_ms: int) -> dict[str, Any]:
    return {
        "owners": [
            {
                "owner": owner["urn"],
                "type": owner.get("type", "DATAOWNER"),
            }
            for owner in spec.get("owners", [])
        ],
        "lastModified": {
            "time": observed_at_ms,
            "actor": actor,
        },
    }


def schema_metadata(spec: dict[str, Any]) -> dict[str, Any]:
    fields = spec.get("schema", [])
    raw_schema = json.dumps(fields, separators=(",", ":"), sort_keys=True)
    return {
        "schemaName": spec["name"].split(".")[-1],
        "platform": f"urn:li:dataPlatform:{spec['platform']}",
        "version": spec.get("schema_version", 1),
        "hash": hashlib.sha256(raw_schema.encode("utf-8")).hexdigest(),
        "platformSchema": {
            "com.linkedin.schema.OtherSchema": {
                "rawSchema": raw_schema,
            }
        },
        "fields": [schema_field(field) for field in fields],
    }


def schema_field(field: dict[str, Any]) -> dict[str, Any]:
    return {
        "fieldPath": field["name"],
        "type": {
            "type": {
                datahub_type(field.get("type", "string")): {},
            }
        },
        "nativeDataType": field.get("native_type", field.get("type", "string").upper()),
        "description": field.get("description", ""),
        "nullable": field.get("nullable", True),
        "recursive": False,
    }


def datahub_type(value: str) -> str:
    normalized = value.lower()
    if normalized in {"int", "integer", "long", "float", "double", "decimal", "number"}:
        return "com.linkedin.schema.NumberType"
    if normalized in {"timestamp", "time", "date", "datetime"}:
        return "com.linkedin.schema.TimeType"
    if normalized in {"bool", "boolean"}:
        return "com.linkedin.schema.BooleanType"
    return "com.linkedin.schema.StringType"


def upstream_lineage(spec: dict[str, Any], dataset: str) -> dict[str, Any]:
    upstreams = []
    fine_grained = []
    for upstream in spec.get("upstreams", []):
        upstream_dataset = dataset_urn(upstream["platform"], upstream["name"], upstream.get("env", spec.get("env", "PROD")))
        upstreams.append({"dataset": upstream_dataset, "type": upstream.get("type", "TRANSFORMED")})

        for mapping in upstream.get("field_lineage", []):
            fine_grained.append(
                {
                    "upstreamType": "FIELD_SET",
                    "upstreams": [schema_field_urn(upstream_dataset, field) for field in mapping.get("upstreams", [])],
                    "downstreamType": "FIELD",
                    "downstreams": [schema_field_urn(dataset, field) for field in mapping.get("downstreams", [])],
                    "transformOperation": mapping.get("operation", "Flink SQL projection"),
                }
            )

    return {
        "upstreams": upstreams,
        "fineGrainedLineages": fine_grained,
    }


def global_tags(spec: dict[str, Any]) -> dict[str, Any]:
    tags = spec.get("tags", ["streamforge", "lakehouse-replay"])
    return {
        "tags": [{"tag": f"urn:li:tag:{tag}"} for tag in tags],
    }


def assertion_info(check: dict[str, Any], dataset: str, actor: str, observed_at_ms: int) -> dict[str, Any]:
    audit_stamp = {"time": observed_at_ms, "actor": actor}
    custom = {
        "type": check.get("type", "CUSTOM"),
        "entity": dataset,
        "logic": check.get("logic", ""),
    }
    if check.get("field"):
        custom["field"] = check["field"]

    return {
        "type": "CUSTOM",
        "description": check.get("description", ""),
        "customAssertion": custom,
        "source": {
            "type": "EXTERNAL",
            "created": audit_stamp,
            "lastModified": audit_stamp,
        },
    }
