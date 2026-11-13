#!/usr/bin/env python3
"""Unit tests for detect_affected_resources.py."""

import json
import os
import shutil
import subprocess
import sys
import tempfile
import textwrap
import unittest

from detect_affected_resources import (
    BOT_LOGIN as BOT,
    MAX_LABEL_LEN,
    affected_section,
    detect_labels,
    label_for,
    load_known_names,
    plan_label_changes,
)

CODEOWNERS = """\
*                                         @DataDog/api-reliability
/docs/                                    @DataDog/api-reliability @DataDog/documentation
datadog/*datadog_monitor*                 @DataDog/monitor-app
datadog/**/*datadog_monitor*              @DataDog/api-reliability @DataDog/monitor-app
datadog/**/*datadog_downtime*             @DataDog/api-reliability @DataDog/monitor-app
datadog/**/*datadog_team*                 @DataDog/api-reliability @DataDog/teams-and-ownership
"""

RESOURCES = [
    "monitor",
    "monitor_json",
    "monitor_config_policy",
    "monitor_notification_rule",
    "downtime",
    "downtime_schedule",
    "team",
    "user",
    "api_key",
    "dashboard",
    "dashboard_list",
    "dashboard_v2",
    "powerpack_v2",
    "synthetics_test",
    "security_findings_severity_modification_rule",
    "security_findings_ticket_creation_rule",
    "security_findings_ticket_creation_rules_order",
]
DATASOURCES = [
    "monitor",
    "monitors",
    "team",
    "user",
    # "datasource/datadog_" + this is exactly MAX_LABEL_LEN characters, and the next one cuts to it.
    "integration_aws_iam_permissions",
    "integration_aws_iam_permissions_standard",
]


def issue(title, body="", labels=()):
    return {"title": title, "body": body, "labels": [{"name": n} for n in labels]}


def pr(title, labels=(), author_association="NONE"):
    return dict(issue(title, labels=labels), pull_request={"url": "x"}, author_association=author_association)


def ev(event, actor, label):
    return {"event": event, "actor": actor, "label": label}


# Dedented before the answers are substituted, so multiline answers can't keep
# the headings indented.
_FORM_TEMPLATE = textwrap.dedent(
    """\
    ### Datadog Terraform Provider Version

    v4.29.0

    ### What resources or data sources are affected?

    {affected}

    ### Terraform Configuration Files

    ```hcl
    {config}
    ```
    """
)


def form_body(affected, config=""):
    return _FORM_TEMPLATE.format(affected=affected, config=config)


class TestDetectAffectedResources(unittest.TestCase):
    def setUp(self):
        self.root = tempfile.mkdtemp()
        self.addCleanup(shutil.rmtree, self.root)
        for subdir, names in (("resources", RESOURCES), ("data-sources", DATASOURCES)):
            os.makedirs(os.path.join(self.root, "docs", subdir))
            for name in names:
                open(os.path.join(self.root, "docs", subdir, name + ".md"), "w").close()
        os.makedirs(os.path.join(self.root, ".github"))
        with open(os.path.join(self.root, ".github", "CODEOWNERS"), "w") as fh:
            fh.write(CODEOWNERS)

    def labels(self, item, files=()):
        return [label for label, _ in detect_labels(item, list(files), self.root)]

    # -- issues --

    def test_issue_form_section(self):
        body = form_body("- datadog_monitor\n- datadog_downtime")
        self.assertEqual(
            self.labels(issue("Plan diff on every run", body)),
            ["resource/datadog_downtime", "resource/datadog_monitor"],
        )

    def test_form_body_fixture_has_form_section(self):
        section = affected_section(form_body("- datadog_monitor\n- datadog_team", "line 1\nline 2"))
        self.assertEqual(section.strip(), "- datadog_monitor\n- datadog_team")

    def test_issue_form_section_keeps_level_three_kind_headings(self):
        for affected in (
            "### Data sources\n- datadog_monitor",
            "### Data sources:\r\n- datadog_monitor",
            "### Resources\n- datadog_team\n\n### Data sources\n- datadog_monitor",
        ):
            with self.subTest(affected=affected):
                labels = self.labels(issue("Bug", form_body(affected)))
                self.assertIn("datasource/datadog_monitor", labels)
                self.assertNotIn("resource/datadog_monitor", labels)

    def test_issue_form_section_still_ends_at_next_field(self):
        body = form_body("### Data sources\n- datadog_monitor") + "\n### Expected Behavior\n\ndatadog_team works\n"
        self.assertEqual(self.labels(issue("Bug", body)), ["datasource/datadog_monitor"])

    def test_issue_title_tag(self):
        self.assertEqual(
            self.labels(issue("[datadog_team] Members are reordered")),
            ["resource/datadog_team"],
        )

    def test_issue_ignores_pasted_config(self):
        config = 'resource "datadog_user" "a" {}\nresource "datadog_team" "b" {}'
        body = form_body("datadog_monitor", config)
        self.assertEqual(self.labels(issue("Bug", body)), ["resource/datadog_monitor"])

    def test_issue_without_form_skips_code_blocks(self):
        body = 'The datadog_user resource fails.\n\n```hcl\nresource "datadog_team" "x" {}\n```'
        self.assertEqual(self.labels(issue("Bug", body)), ["resource/datadog_user"])

    def test_issue_without_form_skips_code_block_variants(self):
        config = 'resource "datadog_team" "x" {}'
        for body in (
            f"datadog_user fails\n\n~~~hcl\n{config}\n~~~\n",
            f"datadog_user fails\n\n~~~~\n```\n{config}\n~~~~\n",
            f"datadog_user fails\r\n```hcl\r\n{config}\r\n```\r\n",
            f"1. Apply:\n\n    ```hcl\n    {config}\n    ```\n2. datadog_user fails\n",
            f"Config: ```{config}``` and datadog_user fails\n",
        ):
            with self.subTest(body=body):
                self.assertEqual(self.labels(issue("Bug", body)), ["resource/datadog_user"])

    def test_issue_without_form_keeps_names_after_code_block(self):
        for fence in ("```", "~~~"):
            with self.subTest(fence=fence):
                body = f"{fence}\r\nresource \"datadog_team\" \"x\" {{}}\r\n{fence}\r\nThe datadog_user resource fails\r\n"
                self.assertEqual(self.labels(issue("Bug", body)), ["resource/datadog_user"])

    def test_issue_without_form_ignores_provider_wide_wording_in_code(self):
        body = "datadog_user fails\n\n~~~\n# applies to all resources\n~~~\n"
        self.assertEqual(self.labels(issue("Bug", body)), ["resource/datadog_user"])

    def test_issue_strikethrough_is_not_a_code_block(self):
        body = "~~datadog_team~~ datadog_user fails"
        self.assertEqual(self.labels(issue("Bug", body)), ["resource/datadog_team", "resource/datadog_user"])

    def test_issue_data_source_prefixes(self):
        for affected in ("data.datadog_monitor", 'data "datadog_monitor" "x"', "data source datadog_monitor",
                         "data_source_datadog_monitor"):
            with self.subTest(affected=affected):
                self.assertEqual(
                    self.labels(issue("Bug", form_body(affected))), ["datasource/datadog_monitor"]
                )

    def test_issue_stated_kind_overrides_title_guess(self):
        item = issue("[datadog_monitor] Data source returns an error", form_body("data.datadog_monitor"))
        self.assertEqual(self.labels(item), ["datasource/datadog_monitor"])

    def test_issue_both_kinds_stated(self):
        body = form_body('data.datadog_monitor and resource "datadog_monitor"')
        self.assertEqual(
            self.labels(issue("[datadog_monitor] Bug", body)),
            ["datasource/datadog_monitor", "resource/datadog_monitor"],
        )

    def test_issue_plain_text_resource_and_data_source(self):
        for affected in (
            "resource datadog_monitor and data source datadog_monitor",
            "Resource: datadog_monitor\nData source: datadog_monitor",
            "the `datadog_monitor` data source and the resource `datadog_monitor`",
        ):
            with self.subTest(affected=affected):
                self.assertEqual(
                    self.labels(issue("[datadog_monitor] Bug", form_body(affected))),
                    ["datasource/datadog_monitor", "resource/datadog_monitor"],
                )

    def test_issue_both_kinds_before_name(self):
        for affected in (
            "resource and data source datadog_monitor",
            "Data source or resource: `datadog_monitor`",
            "data sources and resource datadog_monitor",
        ):
            with self.subTest(affected=affected):
                self.assertEqual(
                    self.labels(issue("Bug", form_body(affected))),
                    ["datasource/datadog_monitor", "resource/datadog_monitor"],
                )

    def test_issue_joined_kinds_before_name_need_singular_resource(self):
        # Plural "resources" next to the name stays a guess, even after a stated kind.
        affected = "data source datadog_team; affected resources: datadog_monitor"
        self.assertEqual(
            self.labels(issue("Bug", form_body(affected))), ["datasource/datadog_team", "resource/datadog_monitor"]
        )

    def test_issue_longer_identifiers_are_not_documented_names(self):
        for affected in (
            "datadog_monitor-alerts",
            "my-datadog_monitor",
            "terraform-provider-datadog_monitor",
            "datadog_monitor_${var.suffix}",
            "datadog_monitor${var.suffix}",
            "datadog_monitor_*",
            "datadog_monitor_",
            "datadog_monitor\u00e9",
            "\u00e9datadog_monitor",
        ):
            with self.subTest(affected=affected):
                self.assertEqual(self.labels(issue("Bug", form_body(affected))), [])

    def test_issue_name_boundaries_that_still_count(self):
        for affected in (
            "datadog_monitor's thresholds",
            "**datadog_monitor**, `datadog_monitor`, (datadog_monitor)",
            "datadog_monitor.cpu.id",
            "datadog_monitor: fails",
            "datadog_monitor - fails on apply",
            "[datadog_monitor](https://registry.terraform.io)",
        ):
            with self.subTest(affected=affected):
                self.assertEqual(self.labels(issue("Bug", form_body(affected))), ["resource/datadog_monitor"])

    def test_issue_module_addresses(self):
        for affected, expected in (
            ("module.alerts.datadog_monitor.example", ["resource/datadog_monitor"]),
            ("module.alerts.data.datadog_monitor.example", ["datasource/datadog_monitor"]),
            ('module.a.module.b.datadog_team.this["x"]', ["resource/datadog_team"]),
            ("# module.alerts.datadog_monitor.cpu will be updated in-place", ["resource/datadog_monitor"]),
        ):
            with self.subTest(affected=affected):
                self.assertEqual(self.labels(issue("Bug", form_body(affected))), expected)

    def test_issue_names_after_a_dot_need_a_module_address(self):
        for affected, expected in (
            ('module.a["x"].datadog_team.this', ["resource/datadog_team"]),
            ("module.a[0].module.b.datadog_team.this", ["resource/datadog_team"]),
            ("module.my-alerts.datadog_team.this", ["resource/datadog_team"]),
            ("data.datadog_monitor.x.id", ["datasource/datadog_monitor"]),
            ("datadog_monitor.cpu.id", ["resource/datadog_monitor"]),
            ("It fails on datadog_monitor. Retrying helps.", ["resource/datadog_monitor"]),
            ("api_key = var.datadog_api_key", []),
            ("ids = local.datadog_team", []),
            ("each.datadog_team, count.datadog_team, self.datadog_team", []),
            ("datadog_team.datadog_monitor", ["resource/datadog_team"]),
            ("module.datadog_monitor", []),
            ("resource.datadog_monitor", []),
            ("foo.module.datadog_team", []),
            ("notmodule.x.datadog_team.y", []),
        ):
            with self.subTest(affected=affected):
                self.assertEqual(self.labels(issue("Bug", form_body(affected))), expected)

    def test_issue_terraform_names_that_are_not_types(self):
        for affected in (
            'variable "datadog_api_key" {}',
            'output "datadog_monitor" {}',
            'module "datadog_monitor" {',
            'provider "datadog_team" {',
        ):
            with self.subTest(affected=affected):
                self.assertEqual(self.labels(issue("Bug", form_body(affected))), [])
        self.assertEqual(
            self.labels(issue("Bug", form_body('resource "datadog_monitor" "x" {}'))), ["resource/datadog_monitor"]
        )

    def test_issue_hcl_instance_names_and_index_keys_are_not_types(self):
        for affected, expected in (
            ('resource "datadog_monitor_json" "datadog_team" {}', ["resource/datadog_monitor_json"]),
            ('resource  "datadog_monitor_json"   "datadog_team" {', ["resource/datadog_monitor_json"]),
            ('data "datadog_monitor" "datadog_team" {}', ["datasource/datadog_monitor"]),
            ('datadog_monitor.this["datadog_team"]', ["resource/datadog_monitor"]),
            ("datadog_monitor.this['datadog_team']", ["resource/datadog_monitor"]),
            ('x[ "datadog_team" ]', []),
            ('module.alerts["datadog_team"].datadog_monitor.cpu', ["resource/datadog_monitor"]),
            ('module.alerts["a]b"].datadog_monitor.cpu', ["resource/datadog_monitor"]),
            ("module.alerts[each.key].datadog_monitor.cpu", ["resource/datadog_monitor"]),
            # Still names: types, list literals, links, later mentions.
            ('resource "datadog_monitor" {}', ["resource/datadog_monitor"]),
            ('ignore_changes = ["datadog_monitor"]', ["resource/datadog_monitor"]),
            ('types: ["datadog_monitor", "datadog_team"]', ["resource/datadog_monitor", "resource/datadog_team"]),
            ('foo([ "datadog_monitor" ])', ["resource/datadog_monitor"]),
            ("[datadog_monitor](https://registry.terraform.io)", ["resource/datadog_monitor"]),
            (
                'resource "datadog_monitor_json" "datadog_team" {} breaks the datadog_team resource',
                ["resource/datadog_monitor_json", "resource/datadog_team"],
            ),
        ):
            with self.subTest(affected=affected):
                self.assertEqual(self.labels(issue("Bug", form_body(affected))), expected)

    def test_issue_module_names_are_not_resources(self):
        affected = "module.datadog_monitor.datadog_team.example"
        self.assertEqual(self.labels(issue("Bug", form_body(affected))), ["resource/datadog_team"])

    def test_issue_both_kinds_after_name(self):
        for affected in (
            "`datadog_monitor` (data source and resource)",
            "datadog_monitor resource/data source",
        ):
            with self.subTest(affected=affected):
                self.assertEqual(
                    self.labels(issue("Bug", form_body(affected))),
                    ["datasource/datadog_monitor", "resource/datadog_monitor"],
                )

    def test_issue_joined_names_share_kind_words(self):
        for affected in (
            "datadog_monitor and datadog_team data sources",
            "data sources datadog_monitor, datadog_team",
            "`datadog_monitor` or `datadog_team` data source",
        ):
            with self.subTest(affected=affected):
                self.assertEqual(
                    self.labels(issue("Bug", form_body(affected))),
                    ["datasource/datadog_monitor", "datasource/datadog_team"],
                )

    def test_issue_kind_words_dont_cross_unrelated_text(self):
        affected = "datadog_team fails, while datadog_monitor data source is fine"
        self.assertEqual(
            self.labels(issue("Bug", form_body(affected))),
            ["datasource/datadog_monitor", "resource/datadog_team"],
        )

    def test_issue_name_prefix_beats_group_words(self):
        affected = "resource_datadog_monitor and datadog_team data sources"
        self.assertEqual(
            self.labels(issue("Bug", form_body(affected))),
            ["datasource/datadog_team", "resource/datadog_monitor"],
        )

    def test_issue_kind_word_after_name_in_title(self):
        item = issue("[datadog_monitor] Data source returns an error", form_body("datadog_monitor"))
        self.assertEqual(self.labels(item), ["datasource/datadog_monitor"])

    def test_issue_generic_resources_wording_is_a_guess(self):
        item = issue("data.datadog_monitor fails", form_body("Affected resources: datadog_monitor"))
        self.assertEqual(self.labels(item), ["datasource/datadog_monitor"])

    def test_issue_stated_kind_only_affects_that_name(self):
        item = issue("[datadog_monitor] Bug", form_body("data.datadog_team"))
        self.assertEqual(self.labels(item), ["datasource/datadog_team", "resource/datadog_monitor"])

    def test_issue_requested_kind_that_does_not_exist_yet_is_honored(self):
        # datadog_downtime has no data source yet; the request is for one.
        for title, body in (
            ("Add data source datadog_downtime", ""),
            ("[datadog_downtime] Feature request", form_body("data.datadog_downtime")),
            ("Feature request", form_body('data "datadog_downtime" "x" {}')),
        ):
            with self.subTest(title=title, body=body):
                self.assertEqual(self.labels(issue(title, body)), ["datasource/datadog_downtime"])

    def test_issue_requested_kind_for_undocumented_name_is_ignored(self):
        self.assertEqual(self.labels(issue("Add data source datadog_made_up_thing")), [])

    def test_issue_kind_heading_applies_to_list(self):
        for affected in (
            "Data sources:\n- datadog_monitor\n- datadog_team",
            "**Data sources**\n\n* datadog_monitor\n* datadog_team",
            "Data sources: datadog_monitor, datadog_team",
            "- Datasource: datadog_monitor, datadog_team",
            "## Data sources\n- datadog_monitor\n- datadog_team",
            "#### **Data sources:**\n\n- datadog_monitor\n- datadog_team",
            "## Data sources: datadog_monitor, datadog_team",
        ):
            with self.subTest(affected=affected):
                self.assertEqual(
                    self.labels(issue("Bug", form_body(affected))),
                    ["datasource/datadog_monitor", "datasource/datadog_team"],
                )

    def test_issue_kind_headings_switch(self):
        affected = "Resources:\n- datadog_team\n\nData sources:\n- datadog_monitor"
        self.assertEqual(
            self.labels(issue("Bug", form_body(affected))),
            ["datasource/datadog_monitor", "resource/datadog_team"],
        )

    def test_issue_markdown_kind_headings_switch(self):
        affected = "## Resources\n- datadog_team\n\n## Data sources\n- datadog_monitor"
        self.assertEqual(
            self.labels(issue("Bug", form_body(affected))),
            ["datasource/datadog_monitor", "resource/datadog_team"],
        )

    def test_issue_unrelated_markdown_heading_ends_markdown_kind_heading(self):
        affected = "## Data sources\n- datadog_monitor\n\n## Notes\n- datadog_team"
        self.assertEqual(
            self.labels(issue("Bug", form_body(affected))),
            ["datasource/datadog_monitor", "resource/datadog_team"],
        )

    def test_issue_markdown_heading_with_more_words_is_not_a_kind_heading(self):
        affected = "## Data sources look fine\n- datadog_monitor"
        self.assertEqual(self.labels(issue("Bug", form_body(affected))), ["resource/datadog_monitor"])

    def test_issue_kind_heading_overrides_title_guess(self):
        item = issue("[datadog_monitor] Bug", form_body("Data sources:\n- datadog_monitor"))
        self.assertEqual(self.labels(item), ["datasource/datadog_monitor"])

    def test_issue_name_prefix_beats_kind_heading(self):
        affected = "Data sources:\n- resource_datadog_monitor\n- datadog_team"
        self.assertEqual(
            self.labels(issue("Bug", form_body(affected))),
            ["datasource/datadog_team", "resource/datadog_monitor"],
        )

    def test_issue_kind_heading_never_makes_up_a_kind(self):
        # datadog_downtime has no data source and datadog_monitors no resource.
        for affected in ("Data sources:\n- datadog_downtime", "Resources:\n- datadog_monitors"):
            with self.subTest(affected=affected):
                self.assertEqual(self.labels(issue("Feature request", form_body(affected))), [])

    def test_issue_inline_request_for_missing_kind_is_honored(self):
        for affected in ("Data source: datadog_downtime", "Add data source datadog_downtime"):
            with self.subTest(affected=affected):
                self.assertEqual(
                    self.labels(issue("Feature request", form_body(affected))), ["datasource/datadog_downtime"]
                )

    def test_issue_kind_heading_scope(self):
        for affected, expected in (
            # A kind label covers only its own line.
            ("- Resource: datadog_team\n- datadog_monitors", ["datasource/datadog_monitors", "resource/datadog_team"]),
            ("- Data source: datadog_monitor\n- datadog_downtime", ["datasource/datadog_monitor", "resource/datadog_downtime"]),
            # A heading covers its block, including loose and numbered lists.
            ("## Data sources\n\n- datadog_monitor\n\n- datadog_team", ["datasource/datadog_monitor", "datasource/datadog_team"]),
            ("Data sources:\n1. datadog_monitor\n2. datadog_team", ["datasource/datadog_monitor", "datasource/datadog_team"]),
            ("Data sources:\n\n\n- datadog_monitor", ["datasource/datadog_monitor"]),
            # ...but not the text after it.
            ("**Data sources**\n- datadog_user\n\nThis breaks my datadog_team", ["datasource/datadog_user", "resource/datadog_team"]),
            ("Data sources:\r\n- datadog_user\r\n\r\nLater datadog_team broke", ["datasource/datadog_user", "resource/datadog_team"]),
            ("Data sources:\nThe datadog_user one.\n\n- datadog_team", ["datasource/datadog_user", "resource/datadog_team"]),
        ):
            with self.subTest(affected=affected):
                self.assertEqual(self.labels(issue("Bug", form_body(affected))), expected)

    def test_issue_kind_words_belong_to_one_name(self):
        for affected, expected in (
            ("datadog_team resource, datadog_user data source", ["datasource/datadog_user", "resource/datadog_team"]),
            ("datadog_team (resource), datadog_user (data source)", ["datasource/datadog_user", "resource/datadog_team"]),
            ("datadog_team resource / datadog_user data source", ["datasource/datadog_user", "resource/datadog_team"]),
            ("datadog_team data source; resource datadog_user", ["datasource/datadog_team", "resource/datadog_user"]),
            ("data source datadog_team, resource datadog_user", ["datasource/datadog_team", "resource/datadog_user"]),
            # Joined names still share them.
            ("datadog_team and datadog_user data sources", ["datasource/datadog_team", "datasource/datadog_user"]),
        ):
            with self.subTest(affected=affected):
                self.assertEqual(self.labels(issue("Bug", form_body(affected))), expected)

    def test_issue_markdown_emphasis_and_escapes(self):
        for affected, expected in (
            ("_datadog_monitor_", ["resource/datadog_monitor"]),
            ("__datadog_monitor__", ["resource/datadog_monitor"]),
            ("datadog\\_monitor", ["resource/datadog_monitor"]),
            ("data source \\_datadog\\_monitor\\_", ["datasource/datadog_monitor"]),
            ("my_datadog_monitor_", []),
            ("_datadog_monitor", []),
            ("datadog_monitor_", []),
        ):
            with self.subTest(affected=affected):
                self.assertEqual(self.labels(issue("Bug", form_body(affected))), expected)

    def test_issue_kind_words_in_a_sentence_are_not_a_heading(self):
        affected = "The data sources are fine.\nBut datadog_monitor fails."
        self.assertEqual(self.labels(issue("Bug", form_body(affected))), ["resource/datadog_monitor"])

    def test_issue_markdown_heading_ends_kind_heading(self):
        body = "Data sources:\n- datadog_monitor\n\n## Notes\n\nAlso seen on datadog_team."
        self.assertEqual(
            self.labels(issue("Bug", body)), ["datasource/datadog_monitor", "resource/datadog_team"]
        )

    def test_issue_resource_prefix_placeholder_style(self):
        self.assertEqual(
            self.labels(issue("Bug", form_body("resource_datadog_monitor"))), ["resource/datadog_monitor"]
        )

    def test_issue_data_source_only_name(self):
        self.assertEqual(
            self.labels(issue("Bug", form_body("datadog_monitors"))), ["datasource/datadog_monitors"]
        )

    def test_issue_unknown_and_env_var_names_ignored(self):
        body = form_body("datadog_monitr, DATADOG_API_KEY, my_datadog_user, datadog_monitor_xyz")
        self.assertEqual(self.labels(issue("Bug", body)), [])

    def test_issue_provider_wide(self):
        body = form_body("All resources")
        self.assertEqual(self.labels(issue("Provider crashes on init", body)), [])

    def test_issue_provider_wide_with_example_name(self):
        for affected in (
            "All resources, for example datadog_monitor",
            "every resource (seen on datadog_monitor)",
            "Any data source, e.g. data.datadog_monitor",
            "datadog_monitor, but it looks provider-wide",
            "Probably Terraform core, seen with datadog_monitor",
            "The provider itself; datadog_monitor is just where I saw it",
        ):
            with self.subTest(affected=affected):
                self.assertEqual(self.labels(issue("Bug", form_body(affected))), [])

    def test_issue_negated_provider_wide_wording_keeps_labels(self):
        for affected in (
            "Only datadog_monitor; this is not provider-wide",
            "datadog_monitor. It doesn't affect all resources",
            "datadog_monitor — isn’t a provider-wide issue",
            "datadog_monitor, not every resource",
            "datadog_monitor; this cannot be provider-wide",
        ):
            with self.subTest(affected=affected):
                self.assertEqual(
                    self.labels(issue("Bug", form_body(affected))), ["resource/datadog_monitor"]
                )

    def test_issue_provider_wide_wording_not_negated_by_distant_words(self):
        for affected in (
            "Not just datadog_monitor but all resources",
            "Not sure, maybe all resources; seen on datadog_monitor",
            "No idea. All resources, e.g. datadog_monitor",
            "Not provider-wide for reads, but all resources fail on apply: datadog_monitor",
        ):
            with self.subTest(affected=affected):
                self.assertEqual(self.labels(issue("Bug", form_body(affected))), [])

    def test_issue_instance_wording_is_not_provider_wide(self):
        for affected in (
            "datadog_monitor (all resources of this type are recreated after upgrade)",
            "datadog_monitor - any resource with a browser step",
            "datadog_monitor: all resources that are currently managed drift",
            "datadog_monitor; all resources in my state are recreated",
            "datadog_monitor, every resource of the same kind",
            "datadog_monitor: all resources managed by this module",
        ):
            with self.subTest(affected=affected):
                self.assertEqual(self.labels(issue("Bug", form_body(affected))), ["resource/datadog_monitor"])

    def test_issue_unqualified_scope_wording_is_still_provider_wide(self):
        for affected in (
            "All resources\nwith Terraform 1.5, e.g. datadog_monitor",
            "all resources without tags, e.g. datadog_monitor",
            "all resources withheld from state, e.g. datadog_monitor",
            "All data sources fail with 500, e.g. data.datadog_monitor",
            "ALL RESOURCES. With datadog_monitor it is obvious",
        ):
            with self.subTest(affected=affected):
                self.assertEqual(self.labels(issue("Bug", form_body(affected))), [])

    def test_issue_provider_wide_title(self):
        for title in ("Provider crashes for all resources", "[provider] Retries ignore the timeout"):
            with self.subTest(title=title):
                self.assertEqual(self.labels(issue(title, form_body("datadog_monitor"))), [])

    def test_issue_multiple_instances_are_not_provider_wide(self):
        for title in (
            "[datadog_monitor] Creating multiple resources with for_each fails",
            "Creating multiple resources with for_each fails",
            "Several resources of the same type time out",
            "Many data sources in one plan are slow",
        ):
            with self.subTest(title=title):
                self.assertEqual(
                    self.labels(issue(title, form_body("datadog_monitor"))), ["resource/datadog_monitor"]
                )

    def test_issue_tagged_title_wording_is_about_instances(self):
        for title in (
            "[datadog_monitor] Every resource is recreated on apply",
            "[ACTP][datadog_monitor] All resources drift after import",
        ):
            with self.subTest(title=title):
                self.assertEqual(
                    self.labels(issue(title, form_body("datadog_monitor"))), ["resource/datadog_monitor"]
                )

    def test_tagged_title_explicit_provider_markers_still_count(self):
        files = ["datadog/fwprovider/resource_datadog_monitor.go"]
        for title in (
            "[datadog_monitor] Provider-wide crash on init",
            "[datadog_monitor] Terraform core crash on init",
            "[datadog_monitor] Crash in the provider itself",
            "[datadog_monitor][provider] Retries ignore the timeout",
        ):
            with self.subTest(title=title):
                self.assertEqual(self.labels(issue(title, form_body("datadog_monitor"))), [])
                self.assertEqual(self.labels(pr(title), files), [])

    def test_tagged_title_negated_provider_marker_keeps_labels(self):
        title = "[datadog_monitor] Crash on init, not provider-wide"
        self.assertEqual(self.labels(issue(title, form_body("datadog_monitor"))), ["resource/datadog_monitor"])
        files = ["datadog/fwprovider/resource_datadog_monitor.go"]
        self.assertEqual(self.labels(pr(title), files), ["resource/datadog_monitor"])

    def test_issue_affected_section_still_wins_over_tagged_title(self):
        item = issue("[datadog_monitor] Crash", form_body("All resources, for example datadog_monitor"))
        self.assertEqual(self.labels(item), [])

    def test_issue_multiple_related_types_are_labeled(self):
        affected = "Multiple resources: datadog_monitor, datadog_downtime"
        self.assertEqual(
            self.labels(issue("Bug", form_body(affected))),
            ["resource/datadog_downtime", "resource/datadog_monitor"],
        )

    def test_bot_label_kept_for_multiple_instances(self):
        item = issue(
            "[datadog_monitor] Creating multiple resources with for_each fails",
            form_body("datadog_monitor"),
            labels=["resource/datadog_monitor"],
        )
        events = [ev("labeled", BOT, "resource/datadog_monitor")]
        self.assertEqual(self.plan(item, events=events), ([], []))

    def test_issue_provider_wide_without_form(self):
        body = "Every resource fails to plan, including datadog_monitor."
        self.assertEqual(self.labels(issue("Bug", body)), [])

    def test_issue_scope_wording_needs_whole_words(self):
        for affected in (
            "datadog_monitor resources fail",
            "datadog_monitor overall resources look fine",
            "datadog_monitor allows resources to be imported",
            "datadog_monitor should keep all resource metric types enabled",
        ):
            with self.subTest(affected=affected):
                self.assertEqual(
                    self.labels(issue("Bug", form_body(affected))), ["resource/datadog_monitor"]
                )

    def test_issue_scope_wording_in_pasted_config_ignored(self):
        body = form_body("datadog_monitor", "# applies to all resources\nresource \"datadog_monitor\" \"m\" {}")
        self.assertEqual(self.labels(issue("Bug", body)), ["resource/datadog_monitor"])

    def test_pr_provider_wide_title(self):
        files = ["datadog/fwprovider/resource_datadog_monitor.go"]
        self.assertEqual(self.labels(pr("[provider] Update Go SDK"), files), [])

    def test_issue_many_unrelated_names(self):
        body = form_body("datadog_monitor, datadog_team, datadog_user, datadog_api_key")
        self.assertEqual(self.labels(issue("Bug", body)), [])

    def test_issue_many_names_same_team(self):
        body = form_body("datadog_monitor, datadog_monitor_json, datadog_downtime, datadog_downtime_schedule")
        self.assertEqual(len(self.labels(issue("Bug", body))), 4)

    # -- label changes --

    def plan(self, item, files=(), events=(), changed_files=None):
        to_add, to_remove = plan_label_changes(item, list(files), list(events), self.root, changed_files)
        return [label for label, _ in to_add], to_remove

    def test_existing_labels_not_repeated(self):
        item = issue("[datadog_team] Bug", labels=["resource/datadog_team", "bug"])
        self.assertEqual(self.plan(item), ([], []))

    def test_bot_label_that_no_longer_applies_is_removed(self):
        item = issue("[datadog_team] Bug", labels=["resource/datadog_monitor", "resource/datadog_team"])
        events = [ev("labeled", BOT, "resource/datadog_monitor"), ev("labeled", BOT, "resource/datadog_team")]
        self.assertEqual(self.plan(item, events=events), ([], ["resource/datadog_monitor"]))

    def test_bot_label_kept_for_negated_scope_and_kind_subheading(self):
        events = [ev("labeled", BOT, "resource/datadog_monitor"), ev("labeled", BOT, "datasource/datadog_monitor")]
        for affected, label in (
            ("Only datadog_monitor; this is not provider-wide", "resource/datadog_monitor"),
            ("### Data sources\n- datadog_monitor", "datasource/datadog_monitor"),
        ):
            with self.subTest(affected=affected):
                item = issue("Bug", form_body(affected), labels=[label])
                self.assertEqual(self.plan(item, events=events), ([], []))

    def test_bot_labels_removed_when_item_becomes_unclear(self):
        events = [ev("labeled", BOT, "resource/datadog_monitor")]
        for title, body in (
            ("Bug", form_body("All resources, for example datadog_monitor")),
            ("Bug", form_body("datadog_monitor, datadog_team, datadog_user, datadog_api_key")),
            ("Bug", form_body("No idea")),
        ):
            with self.subTest(body=body):
                item = issue(title, body, labels=["resource/datadog_monitor"])
                self.assertEqual(self.plan(item, events=events), ([], ["resource/datadog_monitor"]))

    def test_label_a_person_added_is_kept(self):
        item = issue("[datadog_team] Bug", labels=["resource/datadog_monitor"])
        for events in ([ev("labeled", "octocat", "resource/datadog_monitor")], []):
            with self.subTest(events=events):
                self.assertEqual(self.plan(item, events=events), (["resource/datadog_team"], []))

    def test_label_a_person_readded_after_the_bot_removed_it_is_kept(self):
        item = issue("[datadog_team] Bug", labels=["resource/datadog_team", "resource/datadog_monitor"])
        events = [
            ev("labeled", BOT, "resource/datadog_monitor"),
            ev("unlabeled", BOT, "resource/datadog_monitor"),
            ev("labeled", "octocat", "resource/datadog_monitor"),
        ]
        self.assertEqual(self.plan(item, events=events), ([], []))

    def test_label_a_person_removed_is_not_added_back(self):
        item = issue("[datadog_team] Bug")
        events = [ev("labeled", BOT, "resource/datadog_team"), ev("unlabeled", "octocat", "resource/datadog_team")]
        self.assertEqual(self.plan(item, events=events), ([], []))

    def test_label_the_bot_removed_can_be_added_again(self):
        item = issue("[datadog_team] Bug")
        events = [ev("labeled", BOT, "resource/datadog_team"), ev("unlabeled", BOT, "resource/datadog_team")]
        self.assertEqual(self.plan(item, events=events), (["resource/datadog_team"], []))

    def test_other_labels_are_never_removed(self):
        item = issue("[datadog_team] Bug", labels=["resource/datadog_team", "bug", "changelog/bugfix"])
        events = [ev("labeled", BOT, "bug"), ev("labeled", BOT, "changelog/bugfix")]
        self.assertEqual(self.plan(item, events=events), ([], []))

    def test_pr_bot_label_removed_when_files_change(self):
        item = pr("Fix", labels=["resource/datadog_monitor"])
        events = [ev("labeled", BOT, "resource/datadog_monitor")]
        files = ["datadog/fwprovider/resource_datadog_team.go"]
        self.assertEqual(self.plan(item, files, events), (["resource/datadog_team"], ["resource/datadog_monitor"]))

    def test_pr_rename_keeps_the_old_resource(self):
        files = ["datadog/fwprovider/resource_datadog_monitor_renamed.go"]
        previous = ["datadog/fwprovider/resource_datadog_monitor.go"]
        events = [ev("labeled", BOT, "resource/datadog_monitor")]
        kept = pr("Rename", labels=["resource/datadog_monitor"])
        self.assertEqual(
            plan_label_changes(kept, files, events, self.root, changed_files=1, previous_files=previous), ([], [])
        )
        to_add, _ = plan_label_changes(pr("Rename"), files, [], self.root, changed_files=1, previous_files=previous)
        self.assertEqual([lbl for lbl, _ in to_add], ["resource/datadog_monitor"])
        # Without the old path the label would look stale.
        self.assertEqual(
            plan_label_changes(kept, files, events, self.root, changed_files=1), ([], ["resource/datadog_monitor"])
        )

    def test_pr_rename_count_and_duplicates_use_new_paths_only(self):
        files = ["docs/resources/team.md", "docs/resources/user.md"]
        swapped = ["docs/resources/user.md", "docs/resources/team.md"]
        to_add, _ = plan_label_changes(pr("Swap"), files, [], self.root, changed_files=2, previous_files=swapped)
        self.assertEqual([lbl for lbl, _ in to_add], ["resource/datadog_team", "resource/datadog_user"])

    def test_pr_rename_from_fork_labels_only_documented_old_names(self):
        files = ["docs/resources/anything_i_want.md"]
        previous = ["docs/resources/monitor.md"]
        to_add, _ = plan_label_changes(pr("Rename"), files, [], self.root, changed_files=1, previous_files=previous)
        self.assertEqual([lbl for lbl, _ in to_add], ["resource/datadog_monitor"])

    def test_pr_incomplete_file_list_changes_nothing(self):
        # The files endpoint stops at 3,000 files; the team file is past the cutoff.
        labels = ["resource/datadog_monitor", "resource/datadog_team"]
        item = pr("Fix", labels=labels)
        events = [ev("labeled", BOT, label) for label in labels]
        files = [f"datadog/tests/cassettes/TestAccDatadogMonitor_{i}.yaml" for i in range(2999)]
        files += ["datadog/resource_datadog_monitor.go", "datadog/fwprovider/resource_datadog_team.go"]
        self.assertEqual(self.plan(item, files, events, changed_files=len(files)), ([], []))
        self.assertEqual(self.plan(item, files[:3000], events, changed_files=len(files)), ([], []))
        # Without the count the partial list would drop the team label
        self.assertEqual(self.plan(item, files[:3000], events), ([], ["resource/datadog_team"]))

    def test_pr_incomplete_file_list_adds_nothing(self):
        files = ["datadog/resource_datadog_monitor.go"]
        self.assertEqual(self.plan(pr("Fix"), files, changed_files=2), ([], []))
        self.assertEqual(self.plan(pr("Fix"), files, changed_files=1), (["resource/datadog_monitor"], []))

    def test_pr_file_list_longer_than_metadata_changes_nothing(self):
        # A push between metadata and file-list reads can grow the diff.
        # An inconsistent snapshot must neither add nor remove labels.
        item = pr("Fix", labels=["resource/datadog_user"])
        events = [ev("labeled", BOT, "resource/datadog_user")]
        files = [
            "datadog/resource_datadog_monitor.go",
            "datadog/fwprovider/resource_datadog_team.go",
        ]
        self.assertEqual(self.plan(item, files, events, changed_files=1), ([], []))

    def test_pr_duplicate_paths_cannot_hide_a_missing_file(self):
        item = pr("Fix", labels=["resource/datadog_team"])
        events = [ev("labeled", BOT, "resource/datadog_team")]
        files = ["datadog/resource_datadog_monitor.go"] * 2
        self.assertEqual(self.plan(item, files, events, changed_files=2), ([], []))

    def run_cli(self, item, files, pull=None, repo_labels=()):
        tmp = tempfile.mkdtemp()
        self.addCleanup(shutil.rmtree, tmp)
        paths = {name: os.path.join(tmp, name) for name in ("item.json", "files.jsonl", "pull.json", "labels.jsonl")}
        with open(paths["labels.jsonl"], "w") as fh:
            fh.write("".join(json.dumps({"name": n, "description": d}) + "\n" for n, d in repo_labels))
        with open(paths["item.json"], "w") as fh:
            json.dump(item, fh)
        with open(paths["files.jsonl"], "w") as fh:
            for f in files:
                filename, previous = (f, None) if isinstance(f, str) else f
                fh.write(json.dumps({"filename": filename, "previous_filename": previous}) + "\n")
        script = os.path.join(os.path.dirname(os.path.abspath(__file__)), "detect_affected_resources.py")
        cmd = [sys.executable, script, "--item", paths["item.json"], "--files", paths["files.jsonl"]]
        cmd += ["--labels", paths["labels.jsonl"]]
        if pull is not None:
            with open(paths["pull.json"], "w") as fh:
                json.dump(pull, fh)
            cmd += ["--pull", paths["pull.json"]]
        return subprocess.run(cmd, capture_output=True, text=True)

    @staticmethod
    def pull_json(changed_files, head_repo="DataDog/terraform-provider-datadog"):
        base = {"repo": {"full_name": "DataDog/terraform-provider-datadog"}}
        head = {"repo": {"full_name": head_repo} if head_repo else None}
        return {"changed_files": changed_files, "head": head, "base": base}

    def test_cli_requires_pull_for_prs(self):
        files = ["datadog/resource_datadog_monitor.go"]
        missing = self.run_cli(pr("Fix"), files)
        self.assertNotEqual(missing.returncode, 0)
        self.assertEqual(missing.stdout, "")
        self.assertIn("--pull", missing.stderr)
        short = self.run_cli(pr("Fix"), files, self.pull_json(2))
        self.assertEqual((short.returncode, short.stdout), (0, ""))
        full = self.run_cli(pr("Fix"), files, self.pull_json(1))
        self.assertEqual((full.returncode, full.stdout), (0, "add\tresource/datadog_monitor\tResource datadog_monitor\t\n"))

    def test_cli_reads_previous_filenames(self):
        files = [("datadog/fwprovider/resource_datadog_monitor_renamed.go", "datadog/fwprovider/resource_datadog_monitor.go")]
        out = self.run_cli(pr("Rename"), files, self.pull_json(1))
        self.assertEqual((out.returncode, out.stdout), (0, "add\tresource/datadog_monitor\tResource datadog_monitor\t\n"))

    def test_cli_prints_hashed_fallback_for_50_character_labels(self):
        # Runs against the real repo's docs.
        name = "datadog_metric_active_tags_and_aggregations"
        cut = ("datasource/" + name)[:MAX_LABEL_LEN]
        out = self.run_cli(issue(f"[data.{name}] Bug"), [])
        action, label, description, fallback = out.stdout.rstrip("\n").split("\t")
        self.assertEqual((action, label, description), ("add", cut, "Data source " + name))
        self.assertRegex(fallback, r"^datasource/datadog_metric_active_tags_and_a-[0-9a-f]{6}$")
        # A label that is already hashed has no fallback.
        out = self.run_cli(issue(f"[data.{name}] Bug"), [], repo_labels=[(cut, "Data source datadog_other")])
        _, label, _, fallback = out.stdout.rstrip("\n").split("\t")
        self.assertRegex(label, r"-[0-9a-f]{6}$")
        self.assertEqual(fallback, "")

    def test_cli_trusts_branches_in_this_repo(self):
        # Generation bots open PRs from branches here but show as CONTRIBUTOR.
        files = ["datadog/fwprovider/resource_datadog_zz_brand_new.go", "docs/resources/zz_brand_new.md"]
        item = pr("[datadog_zz_brand_new] Add resource", author_association="CONTRIBUTOR")
        for head_repo, expected in (
            ("DataDog/terraform-provider-datadog", "add\tresource/datadog_zz_brand_new\tResource datadog_zz_brand_new\t\n"),
            ("someone/terraform-provider-datadog", ""),
            ("datadog/terraform-provider-datadog", ""),  # not the same full_name
            (None, ""),  # deleted fork
        ):
            with self.subTest(head_repo=head_repo):
                out = self.run_cli(item, files, self.pull_json(len(files), head_repo))
                self.assertEqual((out.returncode, out.stdout), (0, expected))

    def test_missing_docs_changes_nothing(self):
        shutil.rmtree(os.path.join(self.root, "docs"))
        item = issue("Bug", labels=["resource/datadog_monitor"])
        with self.assertRaises(SystemExit):
            plan_label_changes(item, [], [ev("labeled", BOT, "resource/datadog_monitor")], self.root)

    def test_cli_output(self):
        # Runs against the real repo's docs.
        tmp = tempfile.mkdtemp()
        self.addCleanup(shutil.rmtree, tmp)
        item_path, events_path = os.path.join(tmp, "item.json"), os.path.join(tmp, "events.jsonl")
        labels_path = os.path.join(tmp, "labels.jsonl")
        with open(item_path, "w") as fh:
            json.dump(issue("[datadog_monitor] Bug", labels=["resource/datadog_downtime"]), fh)
        with open(events_path, "w") as fh:
            fh.write(json.dumps(ev("labeled", BOT, "resource/datadog_downtime")) + "\n")
        with open(labels_path, "w") as fh:
            fh.write(json.dumps({"name": "bug", "description": None}) + "\n")
        script = os.path.join(os.path.dirname(os.path.abspath(__file__)), "detect_affected_resources.py")
        out = subprocess.run(
            [sys.executable, script, "--item", item_path, "--events", events_path, "--labels", labels_path],
            check=True, capture_output=True, text=True,
        ).stdout
        self.assertEqual(
            out.splitlines(),
            ["add\tresource/datadog_monitor\tResource datadog_monitor\t", "remove\tresource/datadog_downtime"],
        )

    # -- label names --

    LONG = "datadog_security_findings_severity_modification_rule"  # cut: resource/datadog_security_findings_severity_modifi
    RULE = "datadog_security_findings_ticket_creation_rule"
    ORDER = "datadog_security_findings_ticket_creation_rules_order"

    def name_of(self, kind, name, documented=None, repo_labels=None):
        documented = documented if documented is not None else set().union(*load_known_names(self.root).values())
        return label_for(kind, name, documented, repo_labels or {})

    def test_short_label_is_unchanged(self):
        self.assertEqual(self.name_of("resource", "datadog_monitor"), "resource/datadog_monitor")
        name = "datadog_" + "a" * (MAX_LABEL_LEN - len("datasource/datadog_"))
        self.assertEqual(self.name_of("datasource", name, {name}), "datasource/" + name)

    def test_long_unique_label_is_cut(self):
        label = self.name_of("resource", self.LONG)
        self.assertEqual(label, ("resource/" + self.LONG)[:MAX_LABEL_LEN])
        self.assertEqual(len(label), MAX_LABEL_LEN)

    def test_long_label_reuses_cut_label_it_owns(self):
        cut = ("resource/" + self.LONG)[:MAX_LABEL_LEN]
        for name_case in (cut, cut.upper()):
            with self.subTest(name_case=name_case):
                repo_labels = {name_case.lower(): "Resource " + self.LONG}
                self.assertEqual(self.name_of("resource", self.LONG, repo_labels=repo_labels), cut)

    def test_long_label_never_takes_a_cut_label_it_does_not_own(self):
        cut = ("resource/" + self.LONG)[:MAX_LABEL_LEN]
        for description in ("", "Resource datadog_something_else", "resource " + self.LONG, "Data source " + self.LONG):
            with self.subTest(description=description):
                label = self.name_of("resource", self.LONG, repo_labels={cut.lower(): description})
                self.assertNotEqual(label, cut)
                self.assertRegex(label, r"^resource/datadog_security_findings_severity-[0-9a-f]{6}$")

    def test_names_that_cut_the_same_get_distinct_hashed_labels(self):
        rule, order = self.name_of("resource", self.RULE), self.name_of("resource", self.ORDER)
        self.assertNotEqual(rule, order)
        for label in (rule, order):
            self.assertRegex(label, r"-[0-9a-f]{6}$")
            self.assertEqual(len(label), MAX_LABEL_LEN)

    def test_names_that_cut_the_same_across_kinds_get_hashed(self):
        # Both are resources only, but a stated kind can ask for either as a data source.
        self.assertNotEqual(self.name_of("datasource", self.RULE), self.name_of("datasource", self.ORDER))

    def test_exact_length_name_keeps_its_label_and_longer_names_get_hashed(self):
        exact = "datadog_" + "a" * (MAX_LABEL_LEN - len("datasource/datadog_"))
        longer = exact + "_standard"
        documented = {exact, longer}
        self.assertEqual(self.name_of("datasource", exact, documented), "datasource/" + exact)
        self.assertRegex(self.name_of("datasource", longer, documented), r"-[0-9a-f]{6}$")
        # Even when the exact-length label already exists without a description.
        repo_labels = {"datasource/" + exact: ""}
        self.assertRegex(self.name_of("datasource", longer, documented, repo_labels), r"-[0-9a-f]{6}$")

    def test_owner_keeps_cut_label_when_a_clashing_name_is_added_later(self):
        cut = ("resource/" + self.RULE)[:MAX_LABEL_LEN]
        documented = {self.RULE, self.ORDER}
        repo_labels = {cut: "Resource " + self.RULE}
        self.assertEqual(self.name_of("resource", self.RULE, documented, repo_labels), cut)
        self.assertRegex(self.name_of("resource", self.ORDER, documented, repo_labels), r"-[0-9a-f]{6}$")

    def test_detect_labels_uses_repo_labels(self):
        item = issue(f"[{self.LONG}] Bug")
        cut = ("resource/" + self.LONG)[:MAX_LABEL_LEN]
        self.assertEqual(detect_labels(item, [], self.root), [(cut, "Resource " + self.LONG)])
        owned_elsewhere = {cut: "Resource datadog_other"}
        (label, description), = detect_labels(item, [], self.root, repo_labels=owned_elsewhere)
        self.assertRegex(label, r"-[0-9a-f]{6}$")
        self.assertEqual(description, "Resource " + self.LONG)

    def test_bot_hashed_label_moves_to_owned_cut_label(self):
        cut = ("resource/" + self.LONG)[:MAX_LABEL_LEN]
        hashed = self.name_of("resource", self.LONG, repo_labels={cut: ""})
        item = issue(f"[{self.LONG}] Bug", labels=[hashed])
        to_add, to_remove = plan_label_changes(
            item, [], [ev("labeled", BOT, hashed)], self.root, repo_labels={cut: "Resource " + self.LONG}
        )
        self.assertEqual(([lbl for lbl, _ in to_add], to_remove), ([cut], [hashed]))

    def long_label_names(self):
        cut = ("resource/" + self.LONG)[:MAX_LABEL_LEN]
        return cut, self.name_of("resource", self.LONG, repo_labels={cut: ""}), {cut: "Resource " + self.LONG}

    def test_person_removing_one_name_of_a_long_label_blocks_the_other(self):
        cut, hashed, owned = self.long_label_names()
        item = issue(f"[{self.LONG}] Bug")
        for removed, repo_labels in ((hashed, owned), (cut, {cut: ""}), (hashed.upper(), owned)):
            with self.subTest(removed=removed):
                events = [ev("labeled", BOT, removed), ev("unlabeled", "octocat", removed)]
                self.assertEqual(
                    plan_label_changes(item, [], events, self.root, repo_labels=repo_labels), ([], [])
                )

    def test_person_added_name_of_a_long_label_is_not_duplicated(self):
        # E.g. the hand-made cut labels before they have a description.
        cut, hashed, _ = self.long_label_names()
        item = issue(f"[{self.LONG}] Bug", labels=[cut])
        for events in ([], [ev("labeled", "octocat", cut)]):
            with self.subTest(events=events):
                self.assertEqual(
                    plan_label_changes(item, [], events, self.root, repo_labels={cut: ""}), ([], [])
                )

    def test_person_label_for_another_name_does_not_block(self):
        cut, _, owned = self.long_label_names()
        other = "resource/datadog_security_findings_severity_other_rule"
        item = issue(f"[{self.LONG}] Bug", labels=[other])
        to_add, _ = plan_label_changes(item, [], [ev("unlabeled", "octocat", other)], self.root, repo_labels=owned)
        self.assertEqual([lbl for lbl, _ in to_add], [cut])

    def test_person_switching_kinds_is_kept(self):
        item = issue("[datadog_monitor] Bug", labels=["datasource/datadog_monitor"])
        events = [
            ev("labeled", BOT, "resource/datadog_monitor"),
            ev("unlabeled", "octocat", "resource/datadog_monitor"),
            ev("labeled", "octocat", "datasource/datadog_monitor"),
        ]
        self.assertEqual(self.plan(item, events=events), ([], []))

    EXACT = "datadog_integration_aws_iam_permissions"
    STANDARD = "datadog_integration_aws_iam_permissions_standard"

    def exact_plan(self, name, labels, events, repo_labels):
        to_add, to_remove = plan_label_changes(
            issue(f"[data.{name}] Bug", labels=labels), [], events, self.root, repo_labels=repo_labels
        )
        return [lbl for lbl, _ in to_add], to_remove

    def test_exact_length_label_belongs_to_its_name(self):
        exact = "datasource/" + self.EXACT
        self.assertEqual(len(exact), MAX_LABEL_LEN)
        for repo_labels in ({}, {exact: ""}, {exact: "Data source " + self.EXACT}):
            with self.subTest(repo_labels=repo_labels):
                self.assertEqual(self.name_of("datasource", self.EXACT, repo_labels=repo_labels), exact)
                self.assertRegex(self.name_of("datasource", self.STANDARD, repo_labels=repo_labels), r"-[0-9a-f]{6}$")

    def test_exact_length_label_is_not_an_alias_of_a_longer_name(self):
        # As on GitHub today: the exact-length label exists without a description.
        exact, today = "datasource/" + self.EXACT, {"datasource/" + self.EXACT: ""}
        standard = self.name_of("datasource", self.STANDARD, repo_labels=today)
        for labels, events in (
            ([exact], [ev("labeled", "octocat", exact)]),
            ([], [ev("labeled", BOT, exact), ev("unlabeled", "octocat", exact)]),
        ):
            with self.subTest(events=events):
                self.assertEqual(self.exact_plan(self.STANDARD, labels, events, today), ([standard], []))
        removed = [ev("labeled", BOT, standard), ev("unlabeled", "octocat", standard)]
        self.assertEqual(self.exact_plan(self.STANDARD, [], removed, today), ([], []))
        removed = [ev("labeled", BOT, exact), ev("unlabeled", "octocat", exact)]
        self.assertEqual(self.exact_plan(self.EXACT, [], removed, today), ([], []))

    def test_exact_length_name_documented_after_a_longer_name_took_its_text(self):
        exact = "datasource/" + self.EXACT
        taken = {exact: "Data source " + self.STANDARD}
        hashed = self.name_of("datasource", self.EXACT, repo_labels=taken)
        self.assertRegex(hashed, r"^datasource/datadog_integration_aws_iam_perm-[0-9a-f]{6}$")
        self.assertEqual(self.name_of("datasource", self.STANDARD, repo_labels=taken), exact)
        removed = [ev("labeled", BOT, exact), ev("unlabeled", "octocat", exact)]
        self.assertEqual(self.exact_plan(self.STANDARD, [], removed, taken), ([], []))
        self.assertEqual(self.exact_plan(self.EXACT, [], removed, taken), ([hashed], []))
        self.assertEqual(self.exact_plan(self.EXACT, [exact], [ev("labeled", "octocat", exact)], taken), ([hashed], []))

    def test_real_repo_labels_are_unique(self):
        repo_root = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..")
        known = load_known_names(repo_root)
        # A stated kind may pair any documented name with either kind.
        names = set().union(*known.values())
        for kind in known:
            with self.subTest(kind=kind):
                labels = [label_for(kind, name, names, {}) for name in names]
                self.assertTrue(all(len(lbl) <= MAX_LABEL_LEN for lbl in labels))
                self.assertEqual(len({lbl.lower() for lbl in labels}), len(labels))

    # -- pull requests --

    def test_pr_files(self):
        files = [
            "datadog/fwprovider/resource_datadog_team.go",
            "datadog/tests/resource_datadog_team_test.go",
            "datadog/tests/cassettes/TestAccDatadogTeam_Basic.yaml",
            "docs/resources/team.md",
            "examples/resources/datadog_team/resource.tf",
        ]
        self.assertEqual(self.labels(pr("[datadog_team] Fix"), files), ["resource/datadog_team"])

    def test_pr_data_source_files(self):
        files = ["datadog/fwprovider/data_source_datadog_team.go", "docs/data-sources/team.md"]
        self.assertEqual(self.labels(pr("[datadog_team] Fix"), files), ["datasource/datadog_team"])

    def test_pr_title_only(self):
        self.assertEqual(
            self.labels(pr("[datadog_monitor] Fix"), ["datadog/internal/utils/utils.go"]),
            ["resource/datadog_monitor"],
        )

    def test_pr_test_file_maps_to_longest_known_name(self):
        for f in ("datadog/resource_datadog_monitor_json_unit_test.go",
                  "datadog/tests/resource_datadog_monitor_json_heatmap_test.go"):
            with self.subTest(file=f):
                self.assertEqual(self.labels(pr("Fix"), [f]), ["resource/datadog_monitor_json"])

    def test_pr_list_stream_widget_tests_are_dashboard(self):
        for f in (
            "datadog/tests/resource_datadog_dashboard_list_stream_test.go",
            "datadog/tests/resource_datadog_dashboard_list_stream_storage_test.go",
            "datadog/resource_datadog_dashboard_list_stream_unit_test.go",
        ):
            with self.subTest(f=f):
                self.assertEqual(self.labels(pr("Fix"), [f]), ["resource/datadog_dashboard"])
        # Only whole name segments: dashboard_list tests are still dashboard_list.
        for f in (
            "datadog/tests/resource_datadog_dashboard_list_test.go",
            "datadog/tests/resource_datadog_dashboard_list_streamer_test.go",
        ):
            with self.subTest(f=f):
                self.assertEqual(self.labels(pr("Fix"), [f]), ["resource/datadog_dashboard_list"])

    def test_pr_non_test_source_needs_exact_name(self):
        files = ["datadog/resource_datadog_monitor_json_helpers.go"]
        self.assertEqual(self.labels(pr("Fix"), files), [])

    def test_pr_source_file_with_trailing_underscore(self):
        files = ["datadog/resource_datadog_synthetics_test_.go"]
        self.assertEqual(self.labels(pr("Fix"), files), ["resource/datadog_synthetics_test"])

    def test_pr_draft_new_resource_without_docs(self):
        files = [
            "datadog/fwprovider/resource_datadog_monitor_new_feature.go",
            "datadog/tests/resource_datadog_monitor_new_feature_test.go",
        ]
        for association in ("MEMBER", "CONTRIBUTOR"):
            with self.subTest(author_association=association):
                item = pr("[datadog_monitor_new_feature] Add resource", author_association=association)
                self.assertEqual(self.labels(item, files), [])

    def test_pr_draft_new_resource_keeps_real_existing_changes(self):
        files = [
            "datadog/fwprovider/resource_datadog_monitor_new_feature.go",
            "datadog/tests/resource_datadog_monitor_new_feature_test.go",
            "datadog/tests/resource_datadog_monitor_test.go",
        ]
        self.assertEqual(self.labels(pr("Add resource"), files), ["resource/datadog_monitor"])

    def test_pr_special_directories(self):
        self.assertEqual(
            self.labels(pr("Refactor"), ["datadog/dashboardmapping/engine.go"]),
            ["resource/datadog_dashboard_v2", "resource/datadog_powerpack_v2"],
        )

    def test_pr_new_resource_from_member(self):
        files = [
            "datadog/fwprovider/resource_datadog_flaky_tests.go",
            "docs/resources/flaky_tests.md",
        ]
        item = pr("[datadog_flaky_tests] Add resource", author_association="MEMBER")
        self.assertEqual(self.labels(item, files), ["resource/datadog_flaky_tests"])

    def test_pr_new_resource_from_same_repo_branch(self):
        files = ["datadog/fwprovider/resource_datadog_flaky_tests.go", "docs/resources/flaky_tests.md"]
        for association in ("CONTRIBUTOR", "NONE"):
            with self.subTest(association=association):
                item = pr("[datadog_flaky_tests] Add resource", author_association=association)
                self.assertEqual(
                    [l for l, _ in detect_labels(item, files, self.root, same_repo=True)], ["resource/datadog_flaky_tests"]
                )
                self.assertEqual(self.labels(item, files), [])

    def test_pr_new_resource_from_fork_ignored(self):
        files = ["docs/resources/anything_i_want.md"]
        item = pr("[datadog_anything_i_want] Add resource", author_association="CONTRIBUTOR")
        self.assertEqual(self.labels(item, files), [])

    NEW_RESOURCE_FILES = [
        "datadog/fwprovider/resource_datadog_monitor_new_feature.go",
        "datadog/tests/resource_datadog_monitor_new_feature_test.go",
        "docs/resources/monitor_new_feature.md",
        "examples/resources/datadog_monitor_new_feature/resource.tf",
    ]

    def test_pr_unknown_docs_and_examples_need_exact_names(self):
        for f in ("docs/resources/monitor_new_feature.md", "examples/resources/datadog_monitor_new_feature/resource.tf",
                  "templates/resources/monitor_new_feature.md.tmpl"):
            with self.subTest(file=f):
                self.assertEqual(self.labels(pr("Add docs"), [f]), [])

    def test_pr_new_resource_from_fork_not_labeled_as_shorter_name(self):
        item = pr("[datadog_monitor_new_feature] Add resource", author_association="CONTRIBUTOR")
        self.assertEqual(self.labels(item, self.NEW_RESOURCE_FILES), [])

    def test_pr_new_resource_from_fork_keeps_real_existing_changes(self):
        files = self.NEW_RESOURCE_FILES + ["datadog/resource_datadog_monitor.go"]
        item = pr("[datadog_monitor_new_feature] Add resource", author_association="CONTRIBUTOR")
        self.assertEqual(self.labels(item, files), ["resource/datadog_monitor"])

    def test_pr_new_resource_from_member_gets_only_its_own_label(self):
        item = pr("[datadog_monitor_new_feature] Add resource", author_association="MEMBER")
        self.assertEqual(self.labels(item, self.NEW_RESOURCE_FILES), ["resource/datadog_monitor_new_feature"])

    NEW_DOWNTIME_DATA_SOURCE = [
        "datadog/fwprovider/data_source_datadog_downtime.go",
        "datadog/tests/data_source_datadog_downtime_test.go",
        "docs/data-sources/downtime.md",
        "examples/data-sources/datadog_downtime/data-source.tf",
    ]

    def test_pr_new_data_source_for_documented_resource(self):
        # datadog_downtime is documented only as a resource.
        for title in ("[datadog_downtime] Add data source", "[datadog_downtime] Add lookup", "Add lookup"):
            for author in ("NONE", "MEMBER"):
                for files in (self.NEW_DOWNTIME_DATA_SOURCE, self.NEW_DOWNTIME_DATA_SOURCE[:2]):
                    with self.subTest(title=title, author=author, files=files):
                        self.assertEqual(
                            self.labels(pr(title, author_association=author), files),
                            ["datasource/datadog_downtime"],
                        )

    def test_pr_new_data_source_test_file_alone(self):
        files = ["datadog/tests/data_source_datadog_monitor_json_test.go"]
        self.assertEqual(self.labels(pr("Fix"), files), ["datasource/datadog_monitor_json"])

    def test_pr_existing_file_needs_its_documented_kind(self):
        # Existing files can define a differently named data source, e.g.
        # data_source_datadog_csm_threats_policy.go is datadog_csm_threats_policies.
        files = [
            "datadog/fwprovider/data_source_datadog_downtime.go",
            "datadog/tests/data_source_datadog_downtime_test.go",
        ]
        for f in files:
            os.makedirs(os.path.join(self.root, os.path.dirname(f)), exist_ok=True)
            open(os.path.join(self.root, f), "w").close()
        self.assertEqual(self.labels(pr("Fix"), files), [])

    def test_pr_ci_only(self):
        files = [".github/workflows/test.yml", "scripts/select_pr_tests.py", "go.mod"]
        self.assertEqual(self.labels(pr("[CI] Bump actions"), files), [])

    def test_pr_release(self):
        self.assertEqual(self.labels(pr("Release 4.29.0"), ["CHANGELOG.md", "version/version.go"]), [])

    def test_pr_cross_team_refactor(self):
        files = [
            "datadog/fwprovider/resource_datadog_monitor.go",
            "datadog/fwprovider/resource_datadog_team.go",
            "datadog/fwprovider/resource_datadog_user.go",
            "datadog/fwprovider/resource_datadog_api_key.go",
        ]
        self.assertEqual(self.labels(pr("[generated] Regenerate"), files), [])

    def test_pr_same_team_family(self):
        files = [
            "datadog/resource_datadog_monitor.go",
            "datadog/resource_datadog_monitor_json.go",
            "datadog/fwprovider/resource_datadog_monitor_config_policy.go",
            "datadog/fwprovider/resource_datadog_monitor_notification_rule.go",
            "datadog/fwprovider/resource_datadog_downtime.go",
        ]
        self.assertEqual(len(self.labels(pr("[datadog_monitor] Shared fix"), files)), 5)


if __name__ == "__main__":
    unittest.main()
