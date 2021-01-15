package datadog

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
)

func TestAccDatadogSloCorrection_Basic(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	uniqueSloCorrection := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogSloCorrectionDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSloCorrectionConfig_Basic(uniqueSloCorrection),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSloCorrectionExists(accProvider, "datadog_slo_correction.testing_slo_correction"),
					// FIXME: add steps to verify attributes were set
				),
			},
		},
	})
}

func testAccCheckDatadogSloCorrectionConfig_Basic(uniq string) string {
	// FIXME: implement usage of the `uniq` argument as a title/name/description of the created entity.
	// This ensures uniqueness in case of parallel-running test cases and easier trackability when
	// cleanup fails.
	return fmt.Sprintf(`
        # uniq: %s
        resource "datadog_slo_correction" "testing_slo_correction" {
			category = "example category"
			description = "example description"
			end = 1600000000
			slo_id = "sloId"
			start = 1600000000
			timezone = "UTC"
        }
    `, uniq)
}

func testAccCheckDatadogSloCorrectionExists(accProvider *schema.Provider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := accProvider.Meta()
		resourceId := s.RootModule().Resources[resourceName].Primary.ID
		providerConf := meta.(*ProviderConfiguration)
		datadogClient := providerConf.DatadogClientV1
		auth := providerConf.AuthV1
		var err error

		id := resourceId

		_, _, err = datadogClient.ServiceLevelObjectiveCorrectionsApi.GetSLOCorrection(auth, id).Execute()

		if err != nil {
			return translateClientError(err, "error checking slo_correction existence")
		}

		return nil
	}
}

func testAccCheckDatadogSloCorrectionDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		meta := accProvider.Meta()
		providerConf := meta.(*ProviderConfiguration)
		datadogClient := providerConf.DatadogClientV1
		auth := providerConf.AuthV1
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_slo_correction" {
				continue
			}

			var err error

			id := r.Primary.ID

			_, resp, err := datadogClient.ServiceLevelObjectiveCorrectionsApi.GetSLOCorrection(auth, id).Execute()

			if err != nil {
				if resp.StatusCode == 404 {
					continue // resource not found => all ok
				} else {
					return fmt.Errorf("received an error retrieving slo_correction: %s", err.Error())
				}
			} else {
				return fmt.Errorf("slo_correction %s still exists", r.Primary.ID)
			}
		}

		return nil
	}
}
