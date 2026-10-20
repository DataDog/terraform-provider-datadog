import unittest

from registry_services import artifact_services


class RegistryServicesTest(unittest.TestCase):
    def test_metadata_cannot_claim_an_artifact_or_hide_later_families(self):
        spec = {"tags": [
            {"name": "Metadata", "x-datadog-tf-generator": {
                "X-owner": "pets", "x-null": None,
                "x-fake": {"datasource_list": {"group": {"read": "ListFake"}}}}},
            {"name": "Real Service", "x-datadog-tf-generator": {
                "x-owner": "pets", "X-list": [], "x-enabled": True,
                "policy": {"datasource_list": {"group": {"read": "ListPolicies"}}},
                "ignored": {"skip": True},
                "thing": {"datasource_singular": {"skip": True},
                          "datasource_list": {"artifact_name": "custom", "group": {"read": "ListThings"}}}}},
        ]}
        self.assertEqual(list(artifact_services(spec)), [
            ("policies", "real-service"), ("custom", "real-service")])


if __name__ == "__main__":
    unittest.main()
