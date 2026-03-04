#!/usr/bin/env python3
"""Unit tests for select_pr_tests.py."""

import os
import tempfile
import textwrap
import unittest

from select_pr_tests import select_pr_tests


class TestSelectPrTests(unittest.TestCase):
    """Tests use a temporary directory tree that mimics the repo layout."""

    def setUp(self):
        self.tmpdir = tempfile.mkdtemp()
        # Create datadog/tests/ directory
        self.tests_dir = os.path.join(self.tmpdir, "datadog", "tests")
        os.makedirs(self.tests_dir)

    # -- helpers --

    def _write_file(self, relpath, content):
        full = os.path.join(self.tmpdir, relpath)
        os.makedirs(os.path.dirname(full), exist_ok=True)
        with open(full, "w") as f:
            f.write(textwrap.dedent(content))
        return full

    # -- test cases --

    def test_no_input(self):
        result = select_pr_tests([], self.tmpdir)
        self.assertEqual(result, "")

    def test_non_resource_file(self):
        result = select_pr_tests(["docs/foo.md"], self.tmpdir)
        self.assertEqual(result, "")

    def test_sdkv2_resource_file(self):
        # Create resource source (not strictly needed, but realistic)
        self._write_file(
            "datadog/resource_datadog_monitor.go",
            'package datadog\n// resource impl\n',
        )
        # Create a test file that references this resource
        self._write_file(
            "datadog/tests/resource_datadog_monitor_test.go",
            '''\
            package test

            func TestAccDatadogMonitor_Basic(t *testing.T) {
                resource "datadog_monitor" "foo" {}
            }

            func TestAccDatadogMonitor_Updated(t *testing.T) {
                resource "datadog_monitor" "bar" {}
            }
            ''',
        )
        result = select_pr_tests(
            ["datadog/resource_datadog_monitor.go"], self.tmpdir
        )
        parts = result.split("|")
        self.assertIn("^TestAccDatadogMonitor_Basic$", parts)
        self.assertIn("^TestAccDatadogMonitor_Updated$", parts)

    def test_framework_resource_file(self):
        self._write_file(
            "datadog/tests/resource_datadog_synthetics_concurrency_cap_test.go",
            '''\
            package test

            func TestAccSyntheticsConcurrencyCap(t *testing.T) {
                resource "datadog_synthetics_concurrency_cap" "test" {}
            }
            ''',
        )
        result = select_pr_tests(
            ["datadog/fwprovider/resource_datadog_synthetics_concurrency_cap.go"],
            self.tmpdir,
        )
        self.assertEqual(result, "^TestAccSyntheticsConcurrencyCap$")

    def test_data_source_file(self):
        self._write_file(
            "datadog/tests/data_source_datadog_monitor_test.go",
            '''\
            package test

            func TestAccDataSourceMonitor(t *testing.T) {
                data "datadog_monitor" "existing" {}
            }
            ''',
        )
        result = select_pr_tests(
            ["datadog/data_source_datadog_monitor.go"], self.tmpdir
        )
        self.assertEqual(result, "^TestAccDataSourceMonitor$")

    def test_direct_test_file_change(self):
        self._write_file(
            "datadog/tests/resource_datadog_dashboard_test.go",
            '''\
            package test

            func TestAccDashboard_Basic(t *testing.T) {}
            func TestAccDashboard_Updated(t *testing.T) {}
            func helperNotATest() {}
            ''',
        )
        result = select_pr_tests(
            ["datadog/tests/resource_datadog_dashboard_test.go"], self.tmpdir
        )
        parts = result.split("|")
        self.assertIn("^TestAccDashboard_Basic$", parts)
        self.assertIn("^TestAccDashboard_Updated$", parts)
        # helper should NOT appear
        for p in parts:
            self.assertNotIn("helper", p)

    def test_multiple_changed_files_deduplicates(self):
        self._write_file(
            "datadog/tests/resource_datadog_monitor_test.go",
            '''\
            package test

            func TestAccMonitor_One(t *testing.T) {
                resource "datadog_monitor" "a" {}
            }
            ''',
        )
        # Both the resource file AND the test file changed
        result = select_pr_tests(
            [
                "datadog/resource_datadog_monitor.go",
                "datadog/tests/resource_datadog_monitor_test.go",
            ],
            self.tmpdir,
        )
        # Should not duplicate
        self.assertEqual(result, "^TestAccMonitor_One$")

    def test_output_format(self):
        self._write_file(
            "datadog/tests/resource_datadog_monitor_test.go",
            '''\
            package test
            func TestA(t *testing.T) { resource "datadog_monitor" "x" {} }
            func TestB(t *testing.T) { resource "datadog_monitor" "y" {} }
            ''',
        )
        result = select_pr_tests(
            ["datadog/resource_datadog_monitor.go"], self.tmpdir
        )
        parts = result.split("|")
        self.assertEqual(len(parts), 2)
        for p in parts:
            self.assertTrue(p.startswith("^"), f"{p} should start with ^")
            self.assertTrue(p.endswith("$"), f"{p} should end with $")

    def test_escape_for_make(self):
        self._write_file(
            "datadog/tests/resource_datadog_monitor_test.go",
            '''\
            package test
            func TestA(t *testing.T) { resource "datadog_monitor" "x" {} }
            ''',
        )
        result = select_pr_tests(
            ["datadog/resource_datadog_monitor.go"],
            self.tmpdir,
            escape_for_make=True,
        )
        self.assertEqual(result, "^TestA$$")

    def test_no_matching_tests(self):
        # Resource changed but no test file references it
        self._write_file(
            "datadog/tests/resource_datadog_other_test.go",
            '''\
            package test
            func TestOther(t *testing.T) { resource "datadog_other" "x" {} }
            ''',
        )
        result = select_pr_tests(
            ["datadog/resource_datadog_monitor.go"], self.tmpdir
        )
        self.assertEqual(result, "")


if __name__ == "__main__":
    unittest.main()
