package test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

var (
	notificationRuleResourceType = "datadog_security_notification_rule"

	signalResourceName = "signal"
	signalResourcePath = fmt.Sprintf("%s.%s", notificationRuleResourceType, signalResourceName)

	vulnerabilityResourceName = "vulnerability"
	vulnerabilityResourcePath = fmt.Sprintf("%s.%s", notificationRuleResourceType, vulnerabilityResourceName)
)

func simpleSignalResourceConfig(name string) string {
	return resourceConfig(signalResourceName, name, "security_signals", []string{"attack_path", "misconfiguration"}, []string{"critical"}, "simple:query", 0, false, []string{"email@datad0g.com"})
}

func updatedSignalResourceConfig(name string) string {
	return resourceConfig(signalResourceName, name, "security_signals", []string{"misconfiguration", "attack_path"}, []string{}, "updated:query", 0, true, []string{"updated_email@datad0g.com"})
}

func simpleVulnerabilityResourceConfig(name string) string {
	return resourceConfig(vulnerabilityResourceName, name, "security_findings", []string{"misconfiguration"}, []string{"critical"}, "simple:query", 3600, false, []string{"email@datad0g.com"})
}

func updatedVulnerabilityResourceConfig(name string) string {
	return resourceConfig(vulnerabilityResourceName, name, "security_findings", []string{"misconfiguration", "attack_path"}, []string{}, "updated:query", 0, true, []string{"updated_email@datad0g.com"})
}

func resourceConfig(resourceName, name, triggerSource string, ruleTypes, severities []string, query string, timeAggregation int64, enabled bool, targets []string) string {
	return fmt.Sprintf(`
		resource "%s" "%s" {
			name    = "%s"
			selectors {
				trigger_source = "%s"
				rule_types = [%s]
				severities = [%s]
				query      = "%s"
			}
			time_aggregation = %d
			enabled = %t
			targets = [%s]
		}
	`, notificationRuleResourceType, resourceName, name, triggerSource, formatSlice(ruleTypes), formatSlice(severities), query, timeAggregation, enabled, formatSlice(targets))
}

func TestAccDatadogSecurityNotificationRuleSignalRuleSimple(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccNotificationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: simpleSignalResourceConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNotificationRuleExists(providers.frameworkProvider, signalResourcePath),
					checkSimpleSignalNotificationRuleContent(name),
				),
			},
		},
	})
}

func TestAccDatadogSecurityNotificationRuleVulnerabilityRuleSimple(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccNotificationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: simpleVulnerabilityResourceConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNotificationRuleExists(providers.frameworkProvider, vulnerabilityResourcePath),
					checkSimpleVulnerabilityNotificationRuleContent(name),
				),
			},
		},
	})
}

func TestAccDatadogSecurityNotificationRuleFull(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	signalRuleName := uniqueEntityName(ctx, t) + "signal"
	vulnerabilityRuleName := uniqueEntityName(ctx, t) + "vulnerability"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccNotificationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: simpleSignalResourceConfig(signalRuleName) + "\n\n" + simpleVulnerabilityResourceConfig(vulnerabilityRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccNotificationRuleExists(providers.frameworkProvider, signalResourcePath),
					checkSimpleSignalNotificationRuleContent(signalRuleName),
					testAccNotificationRuleExists(providers.frameworkProvider, vulnerabilityResourcePath),
					checkSimpleVulnerabilityNotificationRuleContent(vulnerabilityRuleName),
				),
			},
			{
				// Update various notification rule attributes
				Config: updatedSignalResourceConfig(signalRuleName) + "\n\n" + updatedVulnerabilityResourceConfig(vulnerabilityRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccNotificationRuleExists(providers.frameworkProvider, signalResourcePath),
					checkUpdatedSignalNotificationRuleContent(signalRuleName),
					testAccNotificationRuleExists(providers.frameworkProvider, vulnerabilityResourcePath),
					checkUpdatedVulnerabilityNotificationRuleContent(vulnerabilityRuleName),
				),
			},
		},
	})
}

func testAccNotificationRuleExists(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource '%s' not found in the state %s", resourceName, s.RootModule().Resources)
		}

		if r.Type != notificationRuleResourceType {
			return fmt.Errorf("resource %s is not of type %s, found %s instead", resourceName, notificationRuleResourceType, r.Type)
		}

		auth := accProvider.Auth
		apiInstances := accProvider.DatadogApiInstances

		var err error
		if resourceName == signalResourcePath {
			_, _, err = apiInstances.GetSecurityMonitoringApiV2().GetSignalNotificationRule(auth, r.Primary.ID)
		} else {
			_, _, err = apiInstances.GetSecurityMonitoringApiV2().GetVulnerabilityNotificationRule(auth, r.Primary.ID)
		}
		if err != nil {
			return fmt.Errorf("received an error retrieving the notification rule: %s", err)
		}

		return nil
	}
}

func testAccNotificationRuleDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		auth := accProvider.Auth
		apiInstances := accProvider.DatadogApiInstances

		for _, r := range s.RootModule().Resources {
			if r.Type != notificationRuleResourceType {
				continue
			}

			_, httpResponse, err := apiInstances.GetSecurityMonitoringApiV2().GetSignalNotificationRule(auth, r.Primary.ID)
			if err == nil {
				return errors.New("notification rule still exists")
			}
			if httpResponse == nil || httpResponse.StatusCode != 404 {
				return fmt.Errorf("received an error while getting the notification rule: %s", err)
			}

			_, httpResponse, err = apiInstances.GetSecurityMonitoringApiV2().GetVulnerabilityNotificationRule(auth, r.Primary.ID)
			if err == nil {
				return errors.New("notification rule still exists")
			}
			if httpResponse == nil || httpResponse.StatusCode != 404 {
				return fmt.Errorf("received an error while getting the notification rule: %s", err)
			}
		}

		return nil
	}
}

func checkSimpleSignalNotificationRuleContent(name string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(signalResourcePath, "name", name),
		resource.TestCheckResourceAttr(signalResourcePath, "enabled", "false"),
		resource.TestCheckResourceAttr(signalResourcePath, "selectors.trigger_source", "security_signals"),
		resource.TestCheckResourceAttr(signalResourcePath, "selectors.rule_types.#", "2"),
		resource.TestCheckResourceAttr(signalResourcePath, "selectors.rule_types.0", "attack_path"),
		resource.TestCheckResourceAttr(signalResourcePath, "selectors.rule_types.1", "misconfiguration"),
		resource.TestCheckResourceAttr(signalResourcePath, "selectors.severities.#", "1"),
		resource.TestCheckResourceAttr(signalResourcePath, "selectors.severities.0", "critical"),
		resource.TestCheckResourceAttr(signalResourcePath, "selectors.query", "simple:query"),
		resource.TestCheckResourceAttr(signalResourcePath, "targets.#", "1"),
		resource.TestCheckResourceAttr(signalResourcePath, "targets.0", "email@datad0g.com"),
		resource.TestCheckResourceAttr(signalResourcePath, "version", "1"),
	)
}

func checkUpdatedSignalNotificationRuleContent(name string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(signalResourcePath, "name", name),
		resource.TestCheckResourceAttr(signalResourcePath, "enabled", "true"),
		resource.TestCheckResourceAttr(signalResourcePath, "selectors.trigger_source", "security_signals"),
		resource.TestCheckResourceAttr(signalResourcePath, "selectors.rule_types.#", "2"),
		resource.TestCheckResourceAttr(signalResourcePath, "selectors.rule_types.0", "attack_path"),
		resource.TestCheckResourceAttr(signalResourcePath, "selectors.rule_types.1", "misconfiguration"),
		resource.TestCheckResourceAttr(signalResourcePath, "selectors.severities.#", "0"),
		resource.TestCheckResourceAttr(signalResourcePath, "selectors.query", "updated:query"),
		resource.TestCheckResourceAttr(signalResourcePath, "targets.#", "1"),
		resource.TestCheckResourceAttr(signalResourcePath, "targets.0", "updated_email@datad0g.com"),
		resource.TestCheckResourceAttr(signalResourcePath, "version", "2"),
	)
}

func checkSimpleVulnerabilityNotificationRuleContent(name string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(vulnerabilityResourcePath, "name", name),
		resource.TestCheckResourceAttr(vulnerabilityResourcePath, "enabled", "false"),
		resource.TestCheckResourceAttr(vulnerabilityResourcePath, "selectors.trigger_source", "security_findings"),
		resource.TestCheckResourceAttr(vulnerabilityResourcePath, "selectors.rule_types.#", "1"),
		resource.TestCheckResourceAttr(vulnerabilityResourcePath, "selectors.rule_types.0", "misconfiguration"),
		resource.TestCheckResourceAttr(vulnerabilityResourcePath, "selectors.severities.#", "1"),
		resource.TestCheckResourceAttr(vulnerabilityResourcePath, "selectors.severities.0", "critical"),
		resource.TestCheckResourceAttr(vulnerabilityResourcePath, "selectors.query", "simple:query"),
		resource.TestCheckResourceAttr(vulnerabilityResourcePath, "time_aggregation", "3600"),
		resource.TestCheckResourceAttr(vulnerabilityResourcePath, "targets.#", "1"),
		resource.TestCheckResourceAttr(vulnerabilityResourcePath, "targets.0", "email@datad0g.com"),
		resource.TestCheckResourceAttr(vulnerabilityResourcePath, "version", "1"),
	)
}

func checkUpdatedVulnerabilityNotificationRuleContent(name string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(vulnerabilityResourcePath, "name", name),
		resource.TestCheckResourceAttr(vulnerabilityResourcePath, "enabled", "true"),
		resource.TestCheckResourceAttr(vulnerabilityResourcePath, "selectors.trigger_source", "security_findings"),
		resource.TestCheckResourceAttr(vulnerabilityResourcePath, "selectors.rule_types.#", "2"),
		resource.TestCheckResourceAttr(vulnerabilityResourcePath, "selectors.rule_types.0", "attack_path"),
		resource.TestCheckResourceAttr(vulnerabilityResourcePath, "selectors.rule_types.1", "misconfiguration"),
		resource.TestCheckResourceAttr(vulnerabilityResourcePath, "selectors.severities.#", "0"),
		resource.TestCheckResourceAttr(vulnerabilityResourcePath, "selectors.query", "updated:query"),
		resource.TestCheckResourceAttr(vulnerabilityResourcePath, "time_aggregation", "0"),
		resource.TestCheckResourceAttr(vulnerabilityResourcePath, "targets.#", "1"),
		resource.TestCheckResourceAttr(vulnerabilityResourcePath, "targets.0", "updated_email@datad0g.com"),
		resource.TestCheckResourceAttr(vulnerabilityResourcePath, "version", "2"),
	)
}

func formatSlice(slice []string) string {
	if len(slice) == 0 {
		return ""
	}
	return `"` + strings.Join(slice, `", "`) + `"`
}
