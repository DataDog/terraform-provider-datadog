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
    "dashboard_v2",
    "powerpack_v2",
    "synthetics_test",
    "security_findings_severity_modification_rule",
    "security_findings_ticket_creation_rule",
    "security_findings_ticket_creation_rules_order",
]
DATASOURCES = ["monitor", "monitors", "team", "user"]


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

    def test_issue_module_addresses(self):
        for affected, expected in (
            ("module.alerts.datadog_monitor.example", ["resource/datadog_monitor"]),
            ("module.alerts.data.datadog_monitor.example", ["datasource/datadog_monitor"]),
            ('module.a.module.b.datadog_team.this["x"]', ["resource/datadog_team"]),
            ("# module.alerts.datadog_monitor.cpu will be updated in-place", ["resource/datadog_monitor"]),
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
            "- Datasource: datadog_monitor\n- datadog_team",
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

    def test_issue_kind_heading_for_missing_kind_is_honored(self):
        # datadog_downtime has no data source yet; the heading asks for one.
        self.assertEqual(
            self.labels(issue("Feature request", form_body("Data sources:\n- datadog_downtime"))),
            ["datasource/datadog_downtime"],
        )

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

    def test_cli_requires_changed_files_for_prs(self):
        tmp = tempfile.mkdtemp()
        self.addCleanup(shutil.rmtree, tmp)
        item_path, files_path = os.path.join(tmp, "item.json"), os.path.join(tmp, "files.txt")
        with open(item_path, "w") as fh:
            json.dump(pr("Fix"), fh)
        with open(files_path, "w") as fh:
            fh.write("datadog/resource_datadog_monitor.go\n")
        script = os.path.join(os.path.dirname(os.path.abspath(__file__)), "detect_affected_resources.py")
        cmd = [sys.executable, script, "--item", item_path, "--files", files_path]
        missing = subprocess.run(cmd, capture_output=True, text=True)
        self.assertNotEqual(missing.returncode, 0)
        self.assertEqual(missing.stdout, "")
        self.assertIn("--changed-files", missing.stderr)
        short = subprocess.run(cmd + ["--changed-files", "2"], check=True, capture_output=True, text=True)
        self.assertEqual(short.stdout, "")
        full = subprocess.run(cmd + ["--changed-files", "1"], check=True, capture_output=True, text=True)
        self.assertEqual(full.stdout, "add\tresource/datadog_monitor\tResource datadog_monitor\n")

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
        with open(item_path, "w") as fh:
            json.dump(issue("[datadog_monitor] Bug", labels=["resource/datadog_downtime"]), fh)
        with open(events_path, "w") as fh:
            fh.write(json.dumps(ev("labeled", BOT, "resource/datadog_downtime")) + "\n")
        script = os.path.join(os.path.dirname(os.path.abspath(__file__)), "detect_affected_resources.py")
        out = subprocess.run(
            [sys.executable, script, "--item", item_path, "--events", events_path],
            check=True, capture_output=True, text=True,
        ).stdout
        self.assertEqual(
            out.splitlines(),
            ["add\tresource/datadog_monitor\tResource datadog_monitor", "remove\tresource/datadog_downtime"],
        )

    def test_short_label_is_unchanged(self):
        self.assertEqual(label_for("resource", "datadog_monitor"), "resource/datadog_monitor")
        name = "datadog_" + "a" * (MAX_LABEL_LEN - len("datasource/datadog_"))
        self.assertEqual(label_for("datasource", name), "datasource/" + name)
        self.assertEqual(len("datasource/" + name), MAX_LABEL_LEN)

    def test_long_label_is_shortened_with_hash(self):
        result = detect_labels(
            issue("[datadog_security_findings_severity_modification_rule] Bug"), [], self.root
        )
        self.assertEqual(len(result), 1)
        label, description = result[0]
        self.assertEqual(len(label), MAX_LABEL_LEN)
        self.assertRegex(label, r"^resource/datadog_security_findings_severity-[0-9a-f]{6}$")
        self.assertEqual(description, "Resource datadog_security_findings_severity_modification_rule")

    def test_long_labels_sharing_a_prefix_stay_distinct(self):
        body = form_body(
            "- datadog_security_findings_ticket_creation_rule\n"
            "- datadog_security_findings_ticket_creation_rules_order"
        )
        result = detect_labels(issue("Bug", body), [], self.root)
        self.assertEqual(len(result), 2)
        self.assertNotEqual(result[0][0], result[1][0])

    def test_real_repo_labels_are_unique(self):
        repo_root = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..")
        known = load_known_names(repo_root)
        # A stated kind may pair any documented name with either kind.
        names = set().union(*known.values())
        for kind in known:
            with self.subTest(kind=kind):
                labels = [label_for(kind, name) for name in names]
                self.assertTrue(all(len(lbl) <= MAX_LABEL_LEN for lbl in labels))
                self.assertEqual(len(labels), len(set(labels)))

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
