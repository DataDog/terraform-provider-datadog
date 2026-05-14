package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

func TestAccTeamSync(t *testing.T) {
	ctx := context.WithValue(context.Background(), "http_retry_enable", true)
	_, providers, accProviders := testAccFrameworkMuxProviders(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTeamSyncDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			// Step 1: Create with all fields
			{
				PreConfig: pauseExistingTeamSync(t, providers.frameworkProvider),
				Config: `
resource "datadog_team_sync" "foo" {
	source          = "github"
	type            = "provision"
	frequency       = "once"
	sync_membership = true

	selection_state {
		external_id {
			type  = "organization"
			value = "12345"
		}
	}
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_team_sync.foo", "id", "github"),
					resource.TestCheckResourceAttr("datadog_team_sync.foo", "source", "github"),
					resource.TestCheckResourceAttr("datadog_team_sync.foo", "type", "provision"),
					resource.TestCheckResourceAttr("datadog_team_sync.foo", "frequency", "once"),
					resource.TestCheckResourceAttr("datadog_team_sync.foo", "sync_membership", "true"),
					resource.TestCheckResourceAttr("datadog_team_sync.foo", "selection_state.#", "1"),
					resource.TestCheckResourceAttr("datadog_team_sync.foo", "selection_state.0.operation", "include"),
					resource.TestCheckResourceAttr("datadog_team_sync.foo", "selection_state.0.scope", "subtree"),
					resource.TestCheckResourceAttr("datadog_team_sync.foo", "selection_state.0.external_id.type", "organization"),
					resource.TestCheckResourceAttr("datadog_team_sync.foo", "selection_state.0.external_id.value", "12345"),
				),
			},
		},
	})
}

func pauseExistingTeamSync(t *testing.T, accProvider *fwprovider.FrameworkProvider) func() {
	return func() {
		t.Helper()
		api := accProvider.DatadogApiInstances.GetTeamsApiV2()
		auth := accProvider.Auth

		// Check if sync already exists and is paused to avoid unnecessary POST
		resp, _, err := api.GetTeamSync(auth, datadogV2.TEAMSYNCATTRIBUTESSOURCE_GITHUB)
		if err == nil {
			data := resp.GetData()
			for _, d := range data {
				attrs := d.GetAttributes()
				if freq, ok := attrs.GetFrequencyOk(); ok && *freq == datadogV2.TEAMSYNCATTRIBUTESFREQUENCY_PAUSED {
					return // already paused, no POST needed
				}
			}
		}

		attrs := datadogV2.NewTeamSyncAttributes(datadogV2.TEAMSYNCATTRIBUTESSOURCE_GITHUB, datadogV2.TEAMSYNCATTRIBUTESTYPE_LINK)
		attrs.SetFrequency(datadogV2.TEAMSYNCATTRIBUTESFREQUENCY_PAUSED)
		syncData := datadogV2.NewTeamSyncData(*attrs, datadogV2.TEAMSYNCBULKTYPE_TEAM_SYNC_BULK)
		body := datadogV2.NewTeamSyncRequest(*syncData)

		_, err = api.SyncTeams(auth, *body)
		if err != nil {
			t.Fatalf("failed to pause existing team sync: %s", err)
		}
	}
}

func testAccCheckDatadogTeamSyncDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		resp, httpResp, err := accProvider.DatadogApiInstances.GetTeamsApiV2().GetTeamSync(accProvider.Auth, datadogV2.TEAMSYNCATTRIBUTESSOURCE_GITHUB)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				return nil
			}
			return fmt.Errorf("error retrieving TeamSync: %s", err)
		}

		for _, d := range resp.GetData() {
			attrs := d.GetAttributes()
			if freq, ok := attrs.GetFrequencyOk(); ok && *freq != datadogV2.TEAMSYNCATTRIBUTESFREQUENCY_PAUSED {
				return fmt.Errorf("TeamSync still active")
			}
		}
		return nil
	}
}
