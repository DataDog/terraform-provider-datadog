"""Exercise the reporting script without credentials or GitHub API writes."""
import json
import os
from pathlib import Path
import subprocess
import tempfile
import textwrap
import unittest


SCRIPT = Path(__file__).with_name("report_api_spec_integration.sh")


class ReportIntegrationTests(unittest.TestCase):
    def run_report(self, conclusion="success", branch="datadog-api-spec/generated/421", fail_lookup=False):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            calls = root / "calls.jsonl"
            gh = root / "gh"
            gh.write_text(
                "#!/usr/bin/env python3\n"
                "import json, os, sys\n"
                "with open(os.environ['CALLS'], 'a') as f:\n"
                "    f.write(json.dumps(sys.argv[1:]) + '\\n')\n"
                "if '/pulls/' in sys.argv[2]:\n"
                "    if os.environ['FAIL_LOOKUP'] == '1': sys.exit(1)\n"
                "    print('spec-head-sha')\n"
            )
            gh.chmod(0o755)
            env = {
                **os.environ,
                "PATH": f"{root}:{os.environ['PATH']}",
                "CALLS": str(calls),
                "FAIL_LOOKUP": "1" if fail_lookup else "0",
                "HEAD_BRANCH": branch,
                "INTEGRATION_CONCLUSION": conclusion,
                "GH_TOKEN": "test-token",
                "GITHUB_REPOSITORY": "DataDog/terraform-provider-datadog",
                "GITHUB_WORKFLOW": "PR Integration Tests",
                "GITHUB_SERVER_URL": "https://github.com",
                "GITHUB_RUN_ID": "123",
            }
            result = subprocess.run(["bash", str(SCRIPT)], env=env, capture_output=True, text=True)
            requests = [json.loads(line) for line in calls.read_text().splitlines()] if calls.exists() else []
            return result, requests

    def test_sdk_status_contract(self):
        for conclusion, state in [("pending", "pending"), ("success", "success"), ("failure", "failure"), ("cancelled", "pending")]:
            with self.subTest(conclusion=conclusion):
                result, calls = self.run_report(conclusion)
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertEqual(calls, [
                    ["api", "repos/ddoghq/datadog-api-spec/pulls/421", "--jq", ".head.sha"],
                    ["api", "--method", "POST", "repos/ddoghq/datadog-api-spec/statuses/spec-head-sha",
                     "-f", f"state={state}",
                     "-f", "context=terraform-provider-datadog/integration",
                     "-f", "description=PR Integration Tests",
                     "-f", "target_url=https://github.com/DataDog/terraform-provider-datadog/actions/runs/123"],
                ])

    def test_unrelated_and_malformed_branches_do_not_write(self):
        for branch in ["feature", "datadog-api-spec/test/421", "datadog-api-spec/generated/nope", "datadog-api-spec/generated/421/extra"]:
            with self.subTest(branch=branch):
                result, calls = self.run_report(branch=branch)
                self.assertNotEqual(result.returncode, 0)
                self.assertEqual(calls, [])

    def test_unknown_conclusion_does_not_write(self):
        result, calls = self.run_report("skipped")
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(calls, [])

    def test_failed_pr_lookup_does_not_write(self):
        result, calls = self.run_report(fail_lookup=True)
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(len(calls), 1)


class CompleteIntegrationCheckTests(unittest.TestCase):
    def test_conclusion_survives_check_publication_failure(self):
        workflow = (SCRIPT.parent.parent / ".github/workflows/test_pr_integration.yml").read_text()
        # Exercise the actual shell step, including GitHub's fail-fast behavior.
        step = workflow.split("      - name: Complete check run\n", 1)[1]
        command = textwrap.dedent(step.split("        run: |\n", 1)[1].split("\n  report-api-spec-finish:", 1)[0])
        for check_id in ["123", ""]:
            for tests_result in ["success", "failure", "cancelled"]:
                with self.subTest(check_id=check_id, tests_result=tests_result), tempfile.TemporaryDirectory() as directory:
                    root = Path(directory)
                    output = root / "output"
                    calls = root / "calls"
                    gh = root / "gh"
                    gh.write_text('#!/bin/sh\necho called >> "$CALLS"\nexit 1\n')
                    gh.chmod(0o755)
                    env = {
                        **os.environ,
                        "PATH": f"{root}:{os.environ['PATH']}",
                        "GITHUB_OUTPUT": str(output),
                        "GITHUB_REPOSITORY": "DataDog/terraform-provider-datadog",
                        "CALLS": str(calls),
                        "CHECK_RUN_ID": check_id,
                        "SELECTION_CONCLUSION": "success",
                        "SELECTION_RESULT": "success",
                        "SHOULD_RUN": "true",
                        "TESTS_RESULT": tests_result,
                        "TESTS_TITLE": "",
                        "TESTS_SUMMARY": "",
                    }
                    result = subprocess.run(["bash", "-e", "-c", command], env=env, capture_output=True, text=True)
                    self.assertNotEqual(result.returncode, 0)
                    self.assertEqual(output.read_text(), f"conclusion={tests_result}\n")
                    self.assertEqual(calls.exists(), bool(check_id))


if __name__ == "__main__":
    unittest.main()
