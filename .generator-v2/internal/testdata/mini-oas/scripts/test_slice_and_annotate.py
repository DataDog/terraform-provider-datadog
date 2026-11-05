"""Regression coverage for slicing a spec that already declares families."""
import tempfile
import unittest
from pathlib import Path

import yaml

from slice_and_annotate import main


class SliceRegistryTest(unittest.TestCase):
    def test_slice_contains_only_requested_artifact(self):
        spec = {
            "openapi": "3.0.3", "info": {"title": "Test", "version": "1"},
            "servers": [], "security": [],
            "tags": [{"name": "Things", "x-datadog-tf-generator": {
                "X-owner": "team", "existing": {"resource": {"group": {
                    "create": "CreateThing", "read": "GetThing", "delete": "DeleteThing"}}}}}],
            "paths": {"/things/{id}": {"get": {
                "operationId": "GetThing", "tags": ["Things"],
                "responses": {"200": {"description": "OK"}}}}},
        }
        with tempfile.TemporaryDirectory() as directory:
            source, output = Path(directory) / "source.yaml", Path(directory) / "slice.yaml"
            source.write_text(yaml.safe_dump(spec))
            main(["--spec", str(source), "--out", str(output),
                  "--artifact-name", "thing", "--read", "GetThing"])
            result = yaml.safe_load(output.read_text())
        registry = result["tags"][0]["x-datadog-tf-generator"]
        self.assertEqual(set(registry), {"X-owner", "thing"})
        self.assertEqual(registry["X-owner"], "team")
        self.assertEqual(registry["thing"]["datasource_singular"]["group"], {"read": "GetThing"})


if __name__ == "__main__":
    unittest.main()
