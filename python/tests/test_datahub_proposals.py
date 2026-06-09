import json
import unittest

from streamforge_datahub.proposals import build_proposals, dataset_urn


class DataHubProposalTest(unittest.TestCase):
    def test_builds_dataset_and_assertion_proposals(self):
        spec = {
            "platform": "iceberg",
            "name": "warehouse.default.orders",
            "env": "PROD",
            "schema_version": 3,
            "owners": [{"urn": "urn:li:corpuser:data-platform", "type": "DATAOWNER"}],
            "schema": [
                {"name": "order_id", "type": "string", "native_type": "STRING", "nullable": False},
                {"name": "amount", "type": "decimal", "native_type": "DECIMAL(12,2)", "nullable": False},
            ],
            "upstreams": [
                {
                    "platform": "kafka",
                    "name": "orders",
                    "field_lineage": [{"upstreams": ["order_id"], "downstreams": ["order_id"]}],
                }
            ],
            "quality_checks": [
                {
                    "id": "orders.order_id.not_null",
                    "type": "FIELD_NOT_NULL",
                    "field": "order_id",
                    "logic": "order_id is not null",
                }
            ],
        }

        proposals = build_proposals(spec, observed_at_ms=1700000000000)
        aspect_names = [proposal["aspectName"] for proposal in proposals]

        self.assertIn("datasetProperties", aspect_names)
        self.assertIn("ownership", aspect_names)
        self.assertIn("schemaMetadata", aspect_names)
        self.assertIn("upstreamLineage", aspect_names)
        self.assertIn("assertionInfo", aspect_names)

        schema = next(proposal for proposal in proposals if proposal["aspectName"] == "schemaMetadata")
        schema_value = json.loads(schema["aspect"]["value"])
        self.assertEqual(schema_value["version"], 3)
        self.assertEqual(schema_value["fields"][0]["fieldPath"], "order_id")

    def test_dataset_urn(self):
        self.assertEqual(
            dataset_urn("iceberg", "warehouse.default.orders", "PROD"),
            "urn:li:dataset:(urn:li:dataPlatform:iceberg,warehouse.default.orders,PROD)",
        )


if __name__ == "__main__":
    unittest.main()
