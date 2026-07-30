import importlib.util
from pathlib import Path
import sys
import unittest


MODULE_PATH = Path(__file__).with_name("coverage_sweep.py")
SPEC = importlib.util.spec_from_file_location("coverage_sweep", MODULE_PATH)
assert SPEC and SPEC.loader
coverage = importlib.util.module_from_spec(SPEC)
sys.modules[SPEC.name] = coverage
SPEC.loader.exec_module(coverage)


def response(data):
    return {
        "responses": {
            "200": {
                "content": {
                    "application/json": {
                        "schema": {
                            "type": "object",
                            "properties": {"data": data},
                        }
                    }
                }
            }
        }
    }


class FakeSDK:
    version = "test"

    def contains(self, operation_id):
        return True


class CoverageSweepTests(unittest.TestCase):
    def test_oas_cardinality_and_set_regions(self):
        item = {
            "type": "object",
            "properties": {
                "id": {"type": "string"},
                "attributes": {
                    "type": "object",
                    "properties": {"name": {"type": "string"}},
                },
            },
        }
        spec = {
            "paths": {
                "/api/v2/widgets": {
                    "get": {
                        "operationId": "ListWidgets",
                        "tags": ["Widgets"],
                        **response({"type": "array", "items": item}),
                    }
                },
                "/api/v2/widgets/{widget_id}": {
                    "get": {
                        "operationId": "GetWidget",
                        "tags": ["Widgets"],
                        **response(item),
                    }
                },
                "/api/v2/settings": {
                    "get": {
                        "operationId": "GetSettings",
                        "tags": ["Settings"],
                        **response(item),
                    }
                },
                # Array shape wins over the trailing path parameter.
                "/api/v2/scores/{aggregation}": {
                    "get": {
                        "operationId": "ListScores",
                        "tags": ["Scores"],
                        **response({"type": "array", "items": item}),
                    }
                },
            }
        }
        resources, excluded, counts = coverage.inventory(spec, FakeSDK())
        self.assertEqual(excluded, [])
        self.assertEqual(counts["collection"], 2)
        self.assertEqual(counts["by_id"], 1)
        self.assertEqual(counts["singleton"], 1)
        regions = {resource.key: resource.endpoint_set for resource in resources}
        self.assertEqual(regions["/api/v2/widgets"], "S∩P")
        self.assertEqual(regions["/api/v2/settings"], "S\\P")
        self.assertEqual(regions["/api/v2/scores/{aggregation}"], "P\\S")

    def test_no_data_is_excluded_and_unstable_remains_visible(self):
        spec = {
            "paths": {
                "/api/v2/raw": {
                    "get": {
                        "operationId": "GetRaw",
                        "tags": ["Raw"],
                        "responses": {
                            "200": {
                                "content": {
                                    "application/json": {
                                        "schema": {
                                            "type": "object",
                                            "properties": {"value": {"type": "string"}},
                                        }
                                    }
                                }
                            }
                        },
                    }
                },
                "/api/unstable/fleet": {
                    "get": {
                        "operationId": "GetFleet",
                        "tags": ["Fleet"],
                        **response(
                            {
                                "type": "object",
                                "properties": {"attributes": {"type": "object"}},
                            }
                        ),
                    }
                },
            }
        }
        resources, excluded, _ = coverage.inventory(spec, FakeSDK())
        self.assertEqual(len(resources), 1)
        self.assertEqual(len(excluded), 1)
        reasons = {operation.operation_id: operation.excluded_reason for operation in excluded}
        self.assertIn("no JSON:API", reasons["GetRaw"])
        self.assertIn("unstable", resources[0].singular.scope_reason)

    def test_gap_taxonomy_keeps_unclassified_loud(self):
        self.assertEqual(
            coverage.classify_gap("map not yet supported", "emit")[0],
            "maps-under-attributes",
        )
        self.assertEqual(
            coverage.classify_gap("cannot use id uuid.UUID as string", "build")[0],
            "uuid-id",
        )
        self.assertEqual(
            coverage.classify_gap(
                "cannot use state.ID.ValueString() as int64 value in argument", "build"
            )[0],
            "sdk-arg-binding",
        )
        self.assertEqual(
            coverage.classify_gap(
                "providerData.DatadogApiInstances.GetScorecardsApiV2 undefined", "build"
            )[0],
            "api-accessor-resolution",
        )
        self.assertEqual(
            coverage.classify_gap("other declaration of CreatorModel", "build")[0],
            "id-collision",
        )
        self.assertEqual(
            coverage.classify_gap(
                'model: map value kind "unsupported" is not representable', "emit"
            )[0],
            "unsupported-schema-kind",
        )
        self.assertEqual(
            coverage.classify_gap(
                "parser: circular $ref at #/components/schemas/Component", "emit"
            )[0],
            "recursive-schema",
        )
        self.assertEqual(
            coverage.classify_gap("brand new compiler failure", "build")[0],
            "unclassified",
        )

    def test_build_error_attribution_uses_generated_filename(self):
        resource = coverage.Resource(
            key="/api/v2/widgets", display="widgets", service="Widgets"
        )
        op = coverage.Operation(
            "ListWidgets", "/api/v2/widgets", "Widgets", "collection", {}, False
        )
        candidate = coverage.Candidate(
            "widgets#plural",
            "tfgen_coverage_list_widgets_plural",
            resource,
            "plural",
            op,
            None,
            "",
            50,
            False,
            True,
        )
        output = (
            "datadog/fwprovider/"
            "data_source_datadog_tfgen_coverage_list_widgets_plural.go:10:2: boom"
        )
        errors = coverage.attribute_build_errors(
            output, [candidate], Path("/definitely/not/a/provider/repo")
        )
        self.assertEqual(
            errors["tfgen_coverage_list_widgets_plural"], "boom"
        )


if __name__ == "__main__":
    unittest.main()
