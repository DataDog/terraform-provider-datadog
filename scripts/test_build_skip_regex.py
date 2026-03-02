#!/usr/bin/env python3
"""Unit tests for build_skip_regex.py."""

import os
import tempfile
import unittest

from build_skip_regex import build_skip_regex


class TestBuildSkipRegex(unittest.TestCase):
    def test_file_not_found(self):
        self.assertEqual(build_skip_regex("/nonexistent/path.yaml"), "")

    def test_empty_skipped_tests_list(self):
        with tempfile.NamedTemporaryFile(mode="w", suffix=".yaml", delete=False) as f:
            f.write("skipped_tests: []\n")
            f.flush()
            self.assertEqual(build_skip_regex(f.name), "")
        os.unlink(f.name)

    def test_one_test_entry(self):
        with tempfile.NamedTemporaryFile(mode="w", suffix=".yaml", delete=False) as f:
            f.write(
                "skipped_tests:\n"
                "  - test: TestAccFoo\n"
                '    reason: "flaky"\n'
            )
            f.flush()
            self.assertEqual(build_skip_regex(f.name), "^TestAccFoo$")
        os.unlink(f.name)

    def test_multiple_entries(self):
        with tempfile.NamedTemporaryFile(mode="w", suffix=".yaml", delete=False) as f:
            f.write(
                "skipped_tests:\n"
                "  - test: TestA\n"
                "  - test: TestB\n"
            )
            f.flush()
            result = build_skip_regex(f.name)
            # Order is preserved from file
            self.assertEqual(result, "^TestA$|^TestB$")
        os.unlink(f.name)

    def test_comment_only_and_blank_lines(self):
        with tempfile.NamedTemporaryFile(mode="w", suffix=".yaml", delete=False) as f:
            f.write(
                "skipped_tests:\n"
                "# This is a comment\n"
                "\n"
                "  - test: TestOnly\n"
                "\n"
                "# Another comment\n"
            )
            f.flush()
            self.assertEqual(build_skip_regex(f.name), "^TestOnly$")
        os.unlink(f.name)

    def test_escape_for_make(self):
        with tempfile.NamedTemporaryFile(mode="w", suffix=".yaml", delete=False) as f:
            f.write(
                "skipped_tests:\n"
                "  - test: TestA\n"
                "  - test: TestB\n"
            )
            f.flush()
            result = build_skip_regex(f.name, escape_for_make=True)
            self.assertEqual(result, "^TestA$$|^TestB$$")
        os.unlink(f.name)

    def test_extra_whitespace_in_test_name(self):
        with tempfile.NamedTemporaryFile(mode="w", suffix=".yaml", delete=False) as f:
            f.write(
                "skipped_tests:\n"
                "  - test:   TestSpaced   \n"
            )
            f.flush()
            self.assertEqual(build_skip_regex(f.name), "^TestSpaced$")
        os.unlink(f.name)


if __name__ == "__main__":
    unittest.main()
