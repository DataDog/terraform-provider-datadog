package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccMonitor_Fwprovider_Create(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorDestroyFwprovider(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitor(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.r", "name", uniq),
					resource.TestCheckResourceAttr(
						"datadog_monitor.r", "type", "metric alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.r", "message", "Monitor triggered. Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.r", "query", "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 4"),
				),
			},
		},
	})
}

func TestAccMonitor_Fwprovider_Update(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorDestroyFwprovider(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitor(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.r", "name", uniq),
					resource.TestCheckResourceAttr(
						"datadog_monitor.r", "type", "metric alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.r", "message", "Monitor triggered. Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.r", "query", "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 4"),
				),
			},
			{
				Config: testAccCheckDatadogMonitor_update(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.r", "name", uniq),
					resource.TestCheckResourceAttr(
						"datadog_monitor.r", "type", "metric alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.r", "message", "updated message"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.r", "query", "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 4"),
				),
			},
		},
	})
}

func testAccCheckDatadogMonitor(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`resource "datadog_monitor" "r" {
	name               = "%s"
    type               = "metric alert"
    message            = "Monitor triggered. Notify: @hipchat-channel"
    query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 4"
}`, uniq)
}

func testAccCheckDatadogMonitor_update(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`resource "datadog_monitor" "r" {
	name               = "%s"
    type               = "metric alert"
    message            = "updated message"
    query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 4"
}`, uniq)
}

func testAccCheckDatadogMonitorDestroyFwprovider(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := MonitorDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func MonitorDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_monitor" {
				continue
			}
			id, _ := strconv.ParseInt(r.Primary.ID, 10, 64)
			_, httpResp, err := apiInstances.GetMonitorsApiV1().GetMonitor(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving Monitor %s", err)}
			}
			return &utils.RetryableError{Prob: "Monitor still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogMonitorExistsFwprovider(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := monitorExistsHelperFwprovider(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func monitorExistsHelperFwprovider(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_monitor" {
			continue
		}
		id, _ := strconv.ParseInt(r.Primary.ID, 10, 64)

		_, httpResp, err := apiInstances.GetMonitorsApiV1().GetMonitor(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving Monitor")
		}
	}
	return nil
}
