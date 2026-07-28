#!/usr/bin/env python3
"""Unit tests for summarize_testacc.py."""

import json
import unittest

from summarize_testacc import parse_results, summarize


def _stream(*events) -> str:
    return "\n".join(json.dumps(e) for e in events)


def _event(test, action, pkg="datadog/tests"):
    return {"Action": action, "Test": test, "Package": pkg}


class TestParseResults(unittest.TestCase):
    def test_empty_stream(self):
        self.assertEqual(parse_results(""), {})

    def test_ignores_non_terminal_actions(self):
        stream = _stream(_event("TestA", "run"), _event("TestA", "output"))
        self.assertEqual(parse_results(stream), {})

    def test_ignores_package_level_events(self):
        stream = _stream({"Action": "fail", "Package": "datadog/tests"})
        self.assertEqual(parse_results(stream), {})

    def test_folds_subtests_into_parent(self):
        stream = _stream(_event("TestA/sub", "fail"), _event("TestA", "fail"))
        self.assertEqual(parse_results(stream), {"TestA": ["fail"]})

    def test_records_actions_in_order(self):
        stream = _stream(_event("TestA", "fail"), _event("TestA", "pass"))
        self.assertEqual(parse_results(stream), {"TestA": ["fail", "pass"]})

    def test_tolerates_malformed_lines(self):
        stream = "not json\n" + _stream(_event("TestA", "pass"))
        self.assertEqual(parse_results(stream), {"TestA": ["pass"]})


class TestSummarize(unittest.TestCase):
    def test_no_results(self):
        result = summarize("")
        self.assertEqual(result["title"], "No tests ran")

    def test_all_passed(self):
        stream = _stream(_event("TestA", "pass"), _event("TestB", "pass"))
        result = summarize(stream)
        self.assertEqual(result["title"], "2 passed")
        self.assertIn("`TestA` | :white_check_mark: passed", result["summary"])

    def test_failure_reported(self):
        stream = _stream(_event("TestA", "fail"), _event("TestB", "pass"))
        result = summarize(stream)
        self.assertEqual(result["title"], "1 failed, 1 passed")
        self.assertIn("`TestA` | :x: failed", result["summary"])

    def test_fail_then_pass_is_flaky_not_failed(self):
        stream = _stream(_event("TestA", "fail"), _event("TestA", "pass"))
        result = summarize(stream)
        self.assertEqual(result["title"], "1 flaky")
        self.assertIn(":warning: flaky", result["summary"])

    def test_pass_then_fail_is_failed(self):
        stream = _stream(_event("TestA", "pass"), _event("TestA", "fail"))
        result = summarize(stream)
        self.assertEqual(result["title"], "1 failed")

    def test_selected_but_absent_is_not_run(self):
        stream = _stream(_event("TestA", "pass"))
        result = summarize(stream, selected_regex="^TestA$|^TestB$")
        self.assertEqual(result["title"], "1 passed, 1 not run")
        self.assertIn("`TestB` | :fast_forward: not run", result["summary"])
        self.assertIn("2 test(s) selected", result["summary"])

    def test_skipped_test(self):
        stream = _stream(_event("TestA", "skip"))
        result = summarize(stream)
        self.assertEqual(result["title"], "1 skipped")

    def test_details_url_appended(self):
        stream = _stream(_event("TestA", "pass"))
        result = summarize(stream, details_url="https://example.test/run/1")
        self.assertIn("[Full logs](https://example.test/run/1)", result["summary"])

    def test_selected_regex_ignores_unanchored_alternatives(self):
        stream = _stream(_event("TestA", "pass"))
        result = summarize(stream, selected_regex="^TestA$|TestB")
        self.assertEqual(result["title"], "1 passed")


if __name__ == "__main__":
    unittest.main()
