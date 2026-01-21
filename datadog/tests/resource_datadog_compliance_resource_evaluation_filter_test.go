package test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

func TestAccResourceEvaluationFilter(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	accountId := "123456789"
	resourceName := "datadog_compliance_resource_evaluation_filter.filter_test"
	simpleTags := []string{"tag1:val1", "tag2:val2", "tag3:val3"}
	reorderedTags := []string{"tag3:val3", "tag2:val2", "tag1:val1"}
	provider := "aws"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckResourceEvaluationFilterDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "datadog_compliance_resource_evaluation_filter" "filter_test" {
					tags = ["tag1:val1", "tag2:val2", "tag3:val3"]
					cloud_provider = "%s"
					resource_id = "%s"
				}
				`, provider, accountId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceEvaluationFilterExists(providers.frameworkProvider, resourceName),
					checkResourceEvaluationFilterContent(
						resourceName,
						accountId,
						provider,
						simpleTags,
					),
				),
			},
			// Check if reordering tags caused update
			{
				// Same tags as step 1 but reordered
				Config: fmt.Sprintf(`
				resource "datadog_compliance_resource_evaluation_filter" "filter_test" {
					tags = ["tag3:val3", "tag1:val1", "tag2:val2"]
					cloud_provider = "%s"
					resource_id = "%s"
				}
				`, provider, accountId),
				// This should trigger a diff or update because tags are now a list
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceEvaluationFilterExists(providers.frameworkProvider, resourceName),
					checkResourceEvaluationFilterContent(
						resourceName,
						accountId,
						provider,
						reorderedTags,
					),
				),
			},
			{
				// Changing the cloud provider, but keeping the rest should force deletion
				Config: fmt.Sprintf(`
				resource "datadog_compliance_resource_evaluation_filter" "filter_test" {
					tags = ["tag3:val3", "tag1:val1", "tag2:val2"]
					cloud_provider = "azure"
					resource_id = "%s"
				}
				`, accountId),
				// This should trigger a diff or update because tags are now a list
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				// Changing the resource_id, but keeping the rest should force deletion
				Config: fmt.Sprintf(`
				resource "datadog_compliance_resource_evaluation_filter" "filter_test" {
					tags = ["tag3:val3", "tag1:val1", "tag2:val2"]
					cloud_provider = "%s"
					resource_id = "123"
				}
				`, provider),
				// This should trigger a diff or update because tags are now a list
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccResourceEvaluationFilterImport(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	accountId := "223456789"
	resourceName := "datadog_compliance_resource_evaluation_filter.filter_test"
	simpleTags := []string{"tag1:val1", "tag2:val2", "tag3:val3"}
	provider := "aws"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "datadog_compliance_resource_evaluation_filter" "filter_test" {
					tags = ["tag1:val1", "tag2:val2", "tag3:val3"]
					cloud_provider = "%s"
					resource_id = "%s"
				}
				`, provider, accountId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceEvaluationFilterExists(providers.frameworkProvider, resourceName),
					checkResourceEvaluationFilterContent(
						resourceName,
						accountId,
						provider,
						simpleTags,
					),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s:%s", provider, accountId),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceEvaluationFilterInvalid(t *testing.T) {
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	accountId := "323456789"
	provider := "aws"
	invalidProvider := "invalid"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "datadog_compliance_resource_evaluation_filter" "filter_test" {
					tags = ["tag1:val1", "invalidTag:asdasf:InvalidTag", "tag3:val3"]
					cloud_provider = "%s"
					resource_id = "%s"
				}
				`, provider, accountId),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value Match`),
				PlanOnly:    true,
			},
			{
				Config: fmt.Sprintf(`
				resource "datadog_compliance_resource_evaluation_filter" "filter_test" {
					tags = ["tag1:val1", "tag3:val3"]
					cloud_provider = "%s"
					resource_id = "%s"
				}
				`, invalidProvider, accountId),
				ExpectError: regexp.MustCompile(`Invalid cloud provider invalid`),
			},
		},
	})
}

func checkResourceEvaluationFilterContent(resourceName string, resource_id string, provider string, expectedTags []string) resource.TestCheckFunc {
	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(resourceName, "cloud_provider", provider),
		resource.TestCheckResourceAttr(resourceName, "resource_id", resource_id),
	}

	checks = append(checks, resource.TestCheckFunc(func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}

		actualCountStr := rs.Primary.Attributes["tags.#"]
		actualCount, err := strconv.Atoi(actualCountStr)
		if err != nil {
			return fmt.Errorf("failed to parse tag count: %v", err)
		}

		if actualCount != len(expectedTags) {
			return fmt.Errorf("expected %d tags, but found %d", len(expectedTags), actualCount)
		}

		// Build actual tag set
		actualTagSet := make(map[string]struct{}, actualCount)
		for i := 0; i < actualCount; i++ {
			tag := rs.Primary.Attributes[fmt.Sprintf("tags.%d", i)]
			actualTagSet[tag] = struct{}{}
		}

		// Check all expected tags are present
		for _, expectedTag := range expectedTags {
			if _, ok := actualTagSet[expectedTag]; !ok {
				return fmt.Errorf("expected tag %q not found in actual set", expectedTag)
			}
		}

		// Optional: fail if extra tags are present (strict match)
		if len(actualTagSet) != len(expectedTags) {
			return fmt.Errorf("tag set contains unexpected tags")
		}

		return nil
	}))

	return resource.ComposeTestCheckFunc(checks...)
}

func testAccCheckResourceEvaluationFilterExists(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource '%s' not found in the state %s", resourceName, s.RootModule().Resources)
		}

		if r.Type != "datadog_compliance_resource_evaluation_filter" {
			return fmt.Errorf("resource %s is not of type datadog_compliance_resource_evaluation_filter, found %s instead", resourceName, r.Type)
		}

		auth := accProvider.Auth
		apiInstances := accProvider.DatadogApiInstances
		provider := r.Primary.Attributes["cloud_provider"]
		resource_id := r.Primary.Attributes["resource_id"]
		skipCache := true

		params := datadogV2.GetResourceEvaluationFiltersOptionalParameters{
			CloudProvider: &provider,
			AccountId:     &resource_id,
			SkipCache:     &skipCache,
		}
		_, _, err := apiInstances.GetSecurityMonitoringApiV2().GetResourceEvaluationFilters(auth, params)
		if err != nil {
			return fmt.Errorf("received an error retrieving agent rule: %s", err)
		}

		return nil
	}
}

func testAccCheckResourceEvaluationFilterDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		auth := accProvider.Auth
		apiInstances := accProvider.DatadogApiInstances

		for _, r := range s.RootModule().Resources {
			if r.Type == "datadog_compliance_resource_evaluation_filter" {
				provider := r.Primary.Attributes["cloud_provider"]
				resource_id := r.Primary.Attributes["resource_id"]
				skipCache := true

				params := datadogV2.GetResourceEvaluationFiltersOptionalParameters{
					CloudProvider: &provider,
					AccountId:     &resource_id,
					SkipCache:     &skipCache,
				}
				response, httpResponse, err := apiInstances.GetSecurityMonitoringApiV2().GetResourceEvaluationFilters(auth, params)

				if err != nil {
					return errors.New("Error retrieving resource evaluation filter")
				}

				if len(response.Data.Attributes.CloudProvider[r.Primary.Attributes["cloud_provider"]][resource_id]) != 0 {
					bytes, _ := json.MarshalIndent(response.Data.Attributes, "", "  ")
					fmt.Println(string(bytes))
					return errors.New("filters were not destroyed")
				}

				if httpResponse == nil {
					return fmt.Errorf("received an error while getting the resource evaluation filter: %s", err)
				}
			}
		}

		return nil
	}
}
