package test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

const actionExecutionPolicyResourceName = "datadog_action_execution_policy.test"

func TestAccDatadogActionExecutionPolicy_Kubernetes(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogActionExecutionPolicyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatadogActionExecutionPolicyKubernetesConfig(name, "allow", `["default"]`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogActionExecutionPolicyExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(actionExecutionPolicyResourceName, "name", name),
					resource.TestCheckResourceAttr(actionExecutionPolicyResourceName, "effect", "allow"),
					resource.TestCheckResourceAttr(actionExecutionPolicyResourceName, "action_pattern.integration", "INTEGRATION_KUBERNETES"),
					resource.TestCheckResourceAttr(actionExecutionPolicyResourceName, "action_pattern.action_fqns.0", "*"),
					resource.TestCheckResourceAttr(actionExecutionPolicyResourceName, "scope.kubernetes.rule.0.target_namespaces.0", "default"),
					resource.TestCheckTypeSetElemNestedAttrs(actionExecutionPolicyResourceName, "target.*", map[string]string{
						"name":         "test-agents",
						"agent_tags.#": "1",
					}),
					resource.TestCheckResourceAttrSet(actionExecutionPolicyResourceName, "id"),
					resource.TestCheckResourceAttrSet(actionExecutionPolicyResourceName, "version"),
				),
			},
			{
				Config: testAccDatadogActionExecutionPolicyKubernetesConfig(name+"-updated", "deny", `["default", "kube-system"]`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogActionExecutionPolicyExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(actionExecutionPolicyResourceName, "name", name+"-updated"),
					resource.TestCheckResourceAttr(actionExecutionPolicyResourceName, "effect", "deny"),
					resource.TestCheckResourceAttr(actionExecutionPolicyResourceName, "scope.kubernetes.rule.0.target_namespaces.#", "2"),
					resource.TestCheckResourceAttr(actionExecutionPolicyResourceName, "scope.kubernetes.rule.0.target_namespaces.1", "kube-system"),
				),
			},
			{
				ResourceName:      actionExecutionPolicyResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDatadogActionExecutionPolicy_TargetCanonicalization(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogActionExecutionPolicyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatadogActionExecutionPolicyTargetCanonicalizationConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogActionExecutionPolicyExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(actionExecutionPolicyResourceName, "target.#", "2"),
					testAccCheckDatadogActionExecutionPolicyTargetTags("z-target", "region:a", "env:prod"),
					testAccCheckDatadogActionExecutionPolicyTargetTags("a-target", "service:web"),
				),
			},
		},
	})
}

func TestAccDatadogActionExecutionPolicy_Scripts(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogActionExecutionPolicyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatadogActionExecutionPolicyScriptsConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogActionExecutionPolicyExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(actionExecutionPolicyResourceName, "effect", "deny"),
					resource.TestCheckResourceAttr(actionExecutionPolicyResourceName, "action_pattern.integration", "INTEGRATION_SCRIPT"),
					resource.TestCheckResourceAttr(actionExecutionPolicyResourceName, "scope.scripts.rule.0.target_script_names.0", "maintenance.sh"),
				),
			},
		},
	})
}

func TestAccDatadogActionExecutionPolicy_RemoteActionRshell(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogActionExecutionPolicyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatadogActionExecutionPolicyRemoteActionRshellConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogActionExecutionPolicyExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(actionExecutionPolicyResourceName, "effect", "allow"),
					resource.TestCheckResourceAttr(actionExecutionPolicyResourceName, "action_pattern.integration", "INTEGRATION_REMOTE_ACTION"),
					resource.TestCheckResourceAttr(actionExecutionPolicyResourceName, "scope.remote_action_rshell.rule.0.target_paths.0", "/tmp"),
					resource.TestCheckResourceAttr(actionExecutionPolicyResourceName, "scope.remote_action_rshell.rule.0.access", "read_only"),
				),
			},
		},
	})
}

func TestAccDatadogActionExecutionPolicy_EmptyScope(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)
	config := testAccDatadogActionExecutionPolicyScopeLessConfig(name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogActionExecutionPolicyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogActionExecutionPolicyExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(actionExecutionPolicyResourceName, "name", name),
				),
			},
			{
				ResourceName:      actionExecutionPolicyResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDatadogActionExecutionPolicy_EmptyScopeRejected(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDatadogActionExecutionPolicyEmptyScopeConfig(name),
				ExpectError: regexp.MustCompile("Empty scope block"),
			},
		},
	})
}

func TestAccDatadogActionExecutionPolicy_EmptyScopeVariant(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	for _, variant := range []struct {
		name        string
		integration string
	}{
		{name: "kubernetes", integration: "INTEGRATION_KUBERNETES"},
		{name: "scripts", integration: "INTEGRATION_SCRIPT"},
		{name: "remote_action_rshell", integration: "INTEGRATION_REMOTE_ACTION"},
	} {
		t.Run(variant.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: accProviders,
				Steps: []resource.TestStep{
					{
						Config:      testAccDatadogActionExecutionPolicyEmptyScopeVariantConfig(variant.integration, variant.name),
						ExpectError: regexp.MustCompile("must contain at least one `rule`\\s+block"),
					},
				},
			})
		})
	}
}

func testAccDatadogActionExecutionPolicyKubernetesConfig(name, effect, namespaces string) string {
	return fmt.Sprintf(`
resource "datadog_action_execution_policy" "test" {
  name   = %q
  effect = %q

  action_pattern {
    integration = "INTEGRATION_KUBERNETES"
    action_fqns = ["*"]
  }

  scope {
    kubernetes {
      rule {
        target_namespaces = %s
      }
    }
  }

  target {
    name       = "test-agents"
    agent_tags = ["env:test"]
  }
}`, name, effect, namespaces)
}

func testAccDatadogActionExecutionPolicyTargetCanonicalizationConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_action_execution_policy" "test" {
  name   = %q
  effect = "allow"

  action_pattern {
    integration = "INTEGRATION_SCRIPT"
    action_fqns = ["*"]
  }

  target {
    name       = "z-target"
    agent_tags = ["region:a", "env:prod", "env:prod"]
  }

  target {
    name       = "a-target"
    agent_tags = ["service:web"]
  }
}`, name)
}

func testAccDatadogActionExecutionPolicyScriptsConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_action_execution_policy" "test" {
  name   = %q
  effect = "deny"

  action_pattern {
    integration = "INTEGRATION_SCRIPT"
    action_fqns = ["*"]
  }

  scope {
    scripts {
      rule {
        target_script_names = ["maintenance.sh"]
      }
    }
  }
}`, name)
}

func testAccDatadogActionExecutionPolicyRemoteActionRshellConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_action_execution_policy" "test" {
  name   = %q
  effect = "allow"

  action_pattern {
    integration = "INTEGRATION_REMOTE_ACTION"
    action_fqns = ["*"]
  }

  scope {
    remote_action_rshell {
      rule {
        target_paths = ["/tmp"]
        access       = "read_only"
      }
    }
  }
}`, name)
}

func testAccDatadogActionExecutionPolicyEmptyScopeConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_action_execution_policy" "test" {
  name   = %q
  effect = "allow"

  action_pattern {
    integration = "INTEGRATION_SCRIPT"
    action_fqns = ["*"]
  }

  scope {}
}`, name)
}

func testAccDatadogActionExecutionPolicyScopeLessConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_action_execution_policy" "test" {
  name   = %q
  effect = "allow"

  action_pattern {
    integration = "INTEGRATION_SCRIPT"
    action_fqns = ["*"]
  }
}`, name)
}

func testAccDatadogActionExecutionPolicyEmptyScopeVariantConfig(integration, variant string) string {
	return fmt.Sprintf(`
resource "datadog_action_execution_policy" "test" {
  name   = "invalid-empty-scope-variant"
  effect = "allow"

  action_pattern {
    integration = %q
    action_fqns = ["*"]
  }

  scope {
    %s {}
  }
}`, integration, variant)
}

func testAccCheckDatadogActionExecutionPolicyExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[actionExecutionPolicyResourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", actionExecutionPolicyResourceName)
		}

		_, _, err := accProvider.DatadogApiInstances.GetExecutionPolicyApiV2().GetExecutionPolicy(accProvider.Auth, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("received an error retrieving execution policy: %w", err)
		}
		return nil
	}
}

func testAccCheckDatadogActionExecutionPolicyTargetTags(name string, expectedTags ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[actionExecutionPolicyResourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", actionExecutionPolicyResourceName)
		}

		for key, value := range rs.Primary.Attributes {
			if !strings.HasSuffix(key, ".name") || value != name {
				continue
			}

			prefix := strings.TrimSuffix(key, "name") + "agent_tags."
			if got := rs.Primary.Attributes[prefix+"#"]; got != fmt.Sprint(len(expectedTags)) {
				return fmt.Errorf("target %q has %s tags, want %d", name, got, len(expectedTags))
			}
			for _, expectedTag := range expectedTags {
				found := false
				for tagKey, tagValue := range rs.Primary.Attributes {
					if strings.HasPrefix(tagKey, prefix) && tagValue == expectedTag {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("target %q is missing tag %q", name, expectedTag)
				}
			}
			return nil
		}

		return fmt.Errorf("target %q not found", name)
	}
}

func testAccCheckDatadogActionExecutionPolicyDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "datadog_action_execution_policy" {
				continue
			}

			_, httpResponse, err := accProvider.DatadogApiInstances.GetExecutionPolicyApiV2().GetExecutionPolicy(accProvider.Auth, rs.Primary.ID)
			if err != nil {
				if httpResponse != nil && httpResponse.StatusCode == 404 {
					continue
				}
				return fmt.Errorf("received an error retrieving execution policy: %w", err)
			}
			return fmt.Errorf("execution policy %s still exists", rs.Primary.ID)
		}
		return nil
	}
}
