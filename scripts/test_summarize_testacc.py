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

    def test_selected_but_absent_and_skip_listed_is_not_run(self):
        stream = _stream(_event("TestA", "pass"))
        result = summarize(
            stream, selected_regex="^TestA$|^TestB$", skipped_regex="^TestB$"
        )
        self.assertEqual(result["title"], "1 passed, 1 not run")
        self.assertIn("`TestB` | :fast_forward: not run", result["summary"])
        self.assertIn("2 test(s) selected", result["summary"])

    def test_selected_but_absent_without_skip_list_has_no_result(self):
        # A compile failure produces no terminal events; those tests were not
        # skipped and must not be reported as if they were.
        stream = _stream(_event("TestA", "pass"))
        result = summarize(
            stream, selected_regex="^TestA$|^TestB$", skipped_regex="^TestOther$"
        )
        self.assertEqual(result["title"], "1 without a result, 1 passed")
        self.assertIn("`TestB` | :grey_question: no result reported", result["summary"])
        self.assertNotIn("flaky_tests.yaml", result["summary"])

    def test_absent_with_no_skip_regex_has_no_result(self):
        result = summarize("", selected_regex="^TestA$")
        self.assertEqual(result["title"], "1 without a result")
        self.assertIn(":grey_question: no result reported", result["summary"])

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


class TestSummaryLimit(unittest.TestCase):
    """The check-run output.summary field is capped at 65535 characters.

    Sized against the real suite: ~1050 acceptance tests averaging 44 characters
    of name, which renders to roughly 84 KB untruncated.
    """

    LIMIT = 65535
    SUITE_SIZE = 1050
    # Padded to the measured mean name length so the row width is realistic.
    NAME = "TestAccDatadog{:04d}".format(0).ljust(44, "x")

    def _many(self, count, action="pass"):
        return _stream(
            *(_event(f"TestAccDatadog{i:04d}".ljust(44, "x"), action)
              for i in range(count))
        )

    def test_full_suite_overflows_without_truncation(self):
        # Guards the premise. If rows ever shrink enough to all fit, truncation
        # stops being exercised and the tests below would pass vacuously.
        rows = self.SUITE_SIZE * (len(f"| `{self.NAME}` | :white_check_mark: passed |") + 1)
        self.assertGreater(rows, self.LIMIT)

    def test_truncates_and_says_so(self):
        result = summarize(
            self._many(self.SUITE_SIZE), details_url="https://example.test/r/1"
        )
        self.assertLessEqual(len(result["summary"]), self.LIMIT)
        self.assertIn(f"of {self.SUITE_SIZE} tests", result["summary"])
        # The link out must survive truncation: it is the only way to reach the
        # results that were dropped.
        self.assertIn("[Full logs](https://example.test/r/1)", result["summary"])

    def test_failures_survive_truncation(self):
        # Rows are ordered worst-first, so failures must be what survives.
        stream = "\n".join(
            [_stream(_event("TestAccFailingOne", "fail")), self._many(self.SUITE_SIZE)]
        )
        result = summarize(stream)
        self.assertLessEqual(len(result["summary"]), self.LIMIT)
        self.assertIn("`TestAccFailingOne` | :x: failed", result["summary"])

    def test_title_still_counts_every_test(self):
        result = summarize(self._many(self.SUITE_SIZE, "fail"))
        self.assertEqual(result["title"], f"{self.SUITE_SIZE} failed")

    def test_small_result_is_not_truncated(self):
        result = summarize(self._many(3))
        self.assertNotIn("Showing", result["summary"])


if __name__ == "__main__":
    unittest.main()
