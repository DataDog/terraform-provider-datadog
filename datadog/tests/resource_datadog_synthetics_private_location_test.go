package test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccDatadogSyntheticsPrivateLocation_importBasic(t *testing.T) {
	t.Parallel()
	if !isReplaying() {
		log.Println("Skipping private locations tests in non replaying mode")
		return
	}
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	privateLocationName := uniqueEntityName(ctx, t)
	frameworkProvider := providers.frameworkProvider

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testSyntheticsPrivateLocationIsDestroyed(frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: createSyntheticsPrivateLocationConfig(privateLocationName),
			},
			{
				ResourceName:            "datadog_synthetics_private_location.foo",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"config", "api_key"},
			},
		},
	})
}

func TestAccDatadogSyntheticsPrivateLocation_Basic(t *testing.T) {
	t.Parallel()
	if !isReplaying() {
		log.Println("Skipping private locations tests in non replaying mode")
		return
	}
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	frameworkProvider := providers.frameworkProvider

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testSyntheticsPrivateLocationIsDestroyed(frameworkProvider),
		Steps: []resource.TestStep{
			createSyntheticsPrivateLocationStep(ctx, frameworkProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsPrivateLocation_Updated(t *testing.T) {
	t.Parallel()
	if !isReplaying() {
		log.Println("Skipping private locations tests in non replaying mode")
		return
	}
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	frameworkProvider := providers.frameworkProvider

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testSyntheticsPrivateLocationIsDestroyed(frameworkProvider),
		Steps: []resource.TestStep{
			createSyntheticsPrivateLocationStep(ctx, frameworkProvider, t),
			updateSyntheticsPrivateLocationStep(ctx, frameworkProvider, t),
		},
	})
}

func createSyntheticsPrivateLocationStep(ctx context.Context, accProvider *fwprovider.FrameworkProvider, t *testing.T) resource.TestStep {
	privateLocationName := uniqueEntityName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsPrivateLocationConfig(privateLocationName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsPrivateLocationExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_private_location.foo", "name", privateLocationName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_private_location.foo", "description", "a private location"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_private_location.foo", "tags.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_private_location.foo", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_private_location.foo", "tags.1", "baz"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_private_location.foo", "api_key", "1234567890"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_private_location.foo", "config"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_private_location.foo", "id"),
			checkRessourceAttributeRegex("datadog_synthetics_private_location.foo", "restriction_policy_resource_id", "synthetics-private-location:pl:.*"),
		),
	}
}

func createSyntheticsPrivateLocationConfig(uniqPrivateLocation string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_private_location" "foo" {
	name = "%s"
	description = "a private location"
	tags = ["foo:bar", "baz"]
	api_key = "1234567890"
}`, uniqPrivateLocation)
}

func updateSyntheticsPrivateLocationStep(ctx context.Context, accProvider *fwprovider.FrameworkProvider, t *testing.T) resource.TestStep {
	privateLocationName := uniqueEntityName(ctx, t) + "_updated"
	return resource.TestStep{
		Config: updateSyntheticsPrivateLocationConfig(privateLocationName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsPrivateLocationExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_private_location.foo", "name", privateLocationName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_private_location.foo", "description", "an updated private location"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_private_location.foo", "tags.#", "3"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_private_location.foo", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_private_location.foo", "tags.1", "baz"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_private_location.foo", "tags.2", "env:test"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_private_location.foo", "api_key", "0987654321"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_private_location.foo", "config"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_private_location.foo", "id"),
			checkRessourceAttributeRegex("datadog_synthetics_private_location.foo", "restriction_policy_resource_id", "synthetics-private-location:pl:.*"),
		),
	}
}

func updateSyntheticsPrivateLocationConfig(uniqPrivateLocation string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_private_location" "foo" {
	name = "%s"
	description = "an updated private location"
	tags = ["foo:bar", "baz", "env:test"]
	api_key = "0987654321"
}`, uniqPrivateLocation)
}

func testSyntheticsPrivateLocationExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type == "datadog_role" {
				continue
			}

			if _, _, err := apiInstances.GetSyntheticsApiV1().GetPrivateLocation(auth, r.Primary.ID); err != nil {
				return fmt.Errorf("received an error retrieving synthetics private location %s", err)
			}
		}
		return nil
	}
}

func testSyntheticsPrivateLocationIsDestroyed(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := SyntheticsPrivateLocationDestroyerHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func SyntheticsPrivateLocationDestroyerHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_synthetics_private_location" {
			continue
		}
		err := utils.Retry(2, 10, func() error {
			_, httpresp, err := apiInstances.GetSyntheticsApiV1().GetPrivateLocation(auth, r.Primary.ID)
			if err != nil {
				if httpresp != nil && httpresp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("error retrieving synthetics private location: %s", err)}
			}
			return &utils.RetryableError{Prob: "synthetics private location still exists"}
		})

		if err != nil {
			return err
		}
	}
	return nil
}
