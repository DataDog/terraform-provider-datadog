package test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

// TestAccDatadogSecurityFindingsDueDateRulesOrder walks a full multi-resource lifecycle: create two
// rules and order them, add a third rule, reorder, then remove the third rule, asserting the live
// order matches the configuration at every step. A final step verifies import.
func TestAccDatadogSecurityFindingsDueDateRulesOrder(t *testing.T) {
	// No t.Parallel(): The reorder endpoint is org-global, so this must not run alongside other tests that create due date rules.
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	resourceName := "datadog_security_findings_due_date_rules_order.order"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		// The order resource has no server-side delete; the rules it references are checked instead.
		CheckDestroy: testAccCheckDatadogSecurityFindingsDueDateRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				// Create two rules, order [a, b].
				Config: testAccDueDateRulesOrderConfig(uniq, []string{"a", "b"}, []string{"a", "b"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsDueDateRuleExists(providers.frameworkProvider, "datadog_security_findings_due_date_rule.a"),
					testAccCheckDatadogSecurityFindingsDueDateRuleExists(providers.frameworkProvider, "datadog_security_findings_due_date_rule.b"),
					resource.TestCheckResourceAttr(resourceName, "rule_ids.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "rule_ids.0", "datadog_security_findings_due_date_rule.a", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "rule_ids.1", "datadog_security_findings_due_date_rule.b", "id"),
				),
			},
			{
				// Add a third rule, order [a, b, c].
				Config: testAccDueDateRulesOrderConfig(uniq, []string{"a", "b", "c"}, []string{"a", "b", "c"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsDueDateRuleExists(providers.frameworkProvider, "datadog_security_findings_due_date_rule.c"),
					resource.TestCheckResourceAttr(resourceName, "rule_ids.#", "3"),
					resource.TestCheckResourceAttrPair(resourceName, "rule_ids.0", "datadog_security_findings_due_date_rule.a", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "rule_ids.1", "datadog_security_findings_due_date_rule.b", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "rule_ids.2", "datadog_security_findings_due_date_rule.c", "id"),
				),
			},
			{
				// Reorder to [c, a, b].
				Config: testAccDueDateRulesOrderConfig(uniq, []string{"a", "b", "c"}, []string{"c", "a", "b"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "rule_ids.#", "3"),
					resource.TestCheckResourceAttrPair(resourceName, "rule_ids.0", "datadog_security_findings_due_date_rule.c", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "rule_ids.1", "datadog_security_findings_due_date_rule.a", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "rule_ids.2", "datadog_security_findings_due_date_rule.b", "id"),
				),
			},
			{
				// Remove rule c (from config and the order), back to [a, b].
				Config: testAccDueDateRulesOrderConfig(uniq, []string{"a", "b"}, []string{"a", "b"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "rule_ids.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "rule_ids.0", "datadog_security_findings_due_date_rule.a", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "rule_ids.1", "datadog_security_findings_due_date_rule.b", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccDatadogSecurityFindingsDueDateRulesOrder_Drift reorders the rules directly through the API
// and asserts Terraform detects the drift (non-empty plan) and restores the configured order.
func TestAccDatadogSecurityFindingsDueDateRulesOrder_Drift(t *testing.T) {
	// No t.Parallel(): The reorder endpoint is org-global, so this must not run alongside other tests that create due date rules.
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	resourceName := "datadog_security_findings_due_date_rules_order.order"
	config := testAccDueDateRulesOrderConfig(uniq, []string{"a", "b"}, []string{"a", "b"})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogSecurityFindingsDueDateRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "rule_ids.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "rule_ids.0", "datadog_security_findings_due_date_rule.a", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "rule_ids.1", "datadog_security_findings_due_date_rule.b", "id"),
				),
			},
			{
				// Reverse the order out-of-band; the refreshed plan must be non-empty (drift detected).
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccDueDateRulesReverseOrderOutOfBand(providers.frameworkProvider),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				// Re-applying the unchanged config must restore the configured order.
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "rule_ids.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "rule_ids.0", "datadog_security_findings_due_date_rule.a", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "rule_ids.1", "datadog_security_findings_due_date_rule.b", "id"),
				),
			},
		},
	})
}

// TestAccDatadogSecurityFindingsDueDateRulesOrder_OutOfBandRuleAdded establishes the order resource,
// then creates a due date rule outside Terraform. The order resource owns the full org-global order,
// so the next apply submits a rule_ids list that omits the new rule; the reorder endpoint rejects
// any submission that does not include every rule in the org, and the apply must fail.
func TestAccDatadogSecurityFindingsDueDateRulesOrder_OutOfBandRuleAdded(t *testing.T) {
	// No t.Parallel(): The reorder endpoint is org-global, so this must not run alongside other tests that create due date rules.
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	resourceName := "datadog_security_findings_due_date_rules_order.order"
	config := testAccDueDateRulesOrderConfig(uniq, []string{"a", "b"}, []string{"a", "b"})

	var outOfBandID uuid.UUID
	// The out-of-band rule is not managed by Terraform, so neither destroy nor CheckDestroy removes
	// it. Delete it here so it does not leak into the org and break the org-global reorder elsewhere.
	t.Cleanup(func() {
		var zero uuid.UUID
		if outOfBandID == zero {
			return
		}
		api := providers.frameworkProvider.DatadogApiInstances.GetSecurityMonitoringApiV2()
		_, _ = api.DeleteSecurityFindingsAutomationDueDateRule(providers.frameworkProvider.Auth, outOfBandID)
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogSecurityFindingsDueDateRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				// Create rules [a, b] and order them, then add a third rule out-of-band in the Check.
				// The out-of-band rule is picked up by Read, so the framework's post-step refresh plan
				// is non-empty (it wants to drop the untracked rule); that drift is expected here.
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "rule_ids.#", "2"),
					testAccDueDateRuleCreateOutOfBand(providers.frameworkProvider, uniq+"-out-of-band", &outOfBandID),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				// Re-applying the unchanged config must fail: rule_ids omits the out-of-band rule, and
				// the org-global reorder endpoint rejects a submission that does not include every rule.
				Config:      config,
				ExpectError: regexp.MustCompile(`reorder must include all 3 rules, got 2`),
			},
		},
	})
}

// TestAccDatadogSecurityFindingsDueDateRulesOrder_InvalidID asserts that a malformed rule ID is
// rejected before any API call.
func TestAccDatadogSecurityFindingsDueDateRulesOrder_InvalidID(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "datadog_security_findings_due_date_rules_order" "order" {
  name     = "%s-order"
  rule_ids = ["not-a-uuid"]
}
`, uniq),
				ExpectError: regexp.MustCompile("invalid rule ID"),
			},
		},
	})
}

// testAccDueDateRulesOrderConfig declares one due date rule per entry in rules (using the entry as
// the Terraform resource label), followed by an order resource listing the rules named in order.
func testAccDueDateRulesOrderConfig(uniq string, rules, order []string) string {
	var b strings.Builder
	for _, name := range rules {
		fmt.Fprintf(&b, `
resource "datadog_security_findings_due_date_rule" %[2]q {
  name = "%[1]s-%[2]s"
  rule {
    finding_types = ["misconfiguration"]
    query         = "env:dev"
  }
  action {
    due_from = "first_seen"
    due_days_per_severity {
      severity    = "critical"
      due_in_days = 7
    }
  }
}
`, uniq, name)
	}

	fmt.Fprintf(&b, "\nresource \"datadog_security_findings_due_date_rules_order\" \"order\" {\n  name = %q\n  rule_ids = [\n", uniq+"-order")
	for _, name := range order {
		fmt.Fprintf(&b, "    datadog_security_findings_due_date_rule.%s.id,\n", name)
	}
	b.WriteString("  ]\n}\n")
	return b.String()
}

// testAccDueDateRulesReverseOrderOutOfBand reverses the full live due date rule order through the
// API. Submitting the complete list keeps the reorder endpoint happy while flipping the relative
// order of the Terraform-managed rules, which the order resource should then report as drift.
func testAccDueDateRulesReverseOrderOutOfBand(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		api := accProvider.DatadogApiInstances.GetSecurityMonitoringApiV2()
		auth := accProvider.Auth

		resp, httpResp, err := api.ListSecurityFindingsAutomationDueDateRules(auth)
		if err != nil {
			return fmt.Errorf("error listing due date rules for out-of-band reorder (%v): %s", httpResp, err)
		}
		data := resp.GetData()
		items := make([]datadogV2.DueDateRuleReorderItem, 0, len(data))
		for i := len(data) - 1; i >= 0; i-- {
			items = append(items, *datadogV2.NewDueDateRuleReorderItem(data[i].GetId(), datadogV2.DUEDATERULETYPE_DUE_DATE_RULES))
		}
		req := datadogV2.NewDueDateRuleReorderRequest(items)
		if _, httpResp, err := api.ReorderSecurityFindingsAutomationDueDateRules(auth, *req); err != nil {
			return fmt.Errorf("error reordering due date rules out-of-band (%v): %s", httpResp, err)
		}
		return nil
	}
}

// testAccDueDateRuleCreateOutOfBand creates a due date rule directly through the API (outside
// Terraform), storing its ID in idOut for later cleanup. It simulates a rule added to the org by
// another actor, which the full-ownership order resource does not track in its rule_ids.
func testAccDueDateRuleCreateOutOfBand(accProvider *fwprovider.FrameworkProvider, name string, idOut *uuid.UUID) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		api := accProvider.DatadogApiInstances.GetSecurityMonitoringApiV2()
		auth := accProvider.Auth

		scope := datadogV2.NewAutomationRuleScopeWithDefaults()
		scope.SetFindingTypes([]datadogV2.SecurityFindingType{datadogV2.SecurityFindingType("misconfiguration")})

		action := datadogV2.NewDueDateRuleActionWithDefaults()
		action.SetDueFrom(datadogV2.DueDateFrom("first_seen"))
		action.SetDueDaysPerSeverity([]datadogV2.DueDatePerSeverityItem{
			*datadogV2.NewDueDatePerSeverityItem(7, datadogV2.DueDateSeverity("critical")),
		})

		attrs := datadogV2.NewDueDateRuleAttributesCreateWithDefaults()
		attrs.SetName(name)
		attrs.SetEnabled(true)
		attrs.SetRule(*scope)
		attrs.SetAction(*action)

		data := datadogV2.NewDueDateRuleDataCreateWithDefaults()
		data.SetType(datadogV2.DUEDATERULETYPE_DUE_DATE_RULES)
		data.SetAttributes(*attrs)

		body := datadogV2.NewDueDateRuleCreateRequestWithDefaults()
		body.SetData(*data)

		resp, httpResp, err := api.CreateSecurityFindingsAutomationDueDateRule(auth, *body)
		if err != nil {
			return fmt.Errorf("error creating due date rule out-of-band (%v): %s", httpResp, err)
		}
		respData := resp.GetData()
		*idOut = respData.GetId()
		return nil
	}
}
