package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

// usersListedTimeout bounds the wait for the eventually consistent users and
// team members search indexes after a create.
const usersListedTimeout = 2 * time.Minute

// Concurrent provider construction can return shared provider instances.
// Replay and recording run serially to keep recorders isolated between tests.
// Passthrough mode retains parallel execution.
func parallelIfLive(t *testing.T) {
	if !isRecording() && !isReplaying() {
		t.Parallel()
	}
}

type apiClientFunc func() (context.Context, *utils.ApiInstances)

func sdkAPIClient(accProvider func() (*schema.Provider, error)) apiClientFunc {
	return func() (context.Context, *utils.ApiInstances) {
		provider, _ := accProvider()
		conf := provider.Meta().(*datadog.ProviderConfiguration)
		return conf.Auth, conf.DatadogApiInstances
	}
}

func frameworkAPIClient(provider *fwprovider.FrameworkProvider) apiClientFunc {
	return func() (context.Context, *utils.ApiInstances) {
		return provider.Auth, provider.DatadogApiInstances
	}
}

// testAccCheckUsersListed waits until ListUsers returns at least minCount users
// for filter and optional status. Data source steps that run before this can
// see a partial result.
func testAccCheckUsersListed(client apiClientFunc, filter, status string, minCount int) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		auth, api := client()
		params := datadogV2.ListUsersOptionalParameters{Filter: &filter}
		if status != "" {
			params.FilterStatus = &status
		}
		return retryUntilUsersListed(auth, usersListedTimeout, func() (int, error) {
			resp, _, err := api.GetUsersApiV2().ListUsers(auth, params)
			if err != nil {
				return 0, fmt.Errorf("error listing users: %w", err)
			}
			if len(resp.GetData()) < minCount {
				return 0, nil
			}
			return len(resp.GetData()), nil
		}, fmt.Errorf("users list returned fewer than %d results for filter %q", minCount, filter))
	}
}

// testAccCheckTeamMembershipsListed waits until the team members search for
// keyword returns at least minCount memberships.
func testAccCheckTeamMembershipsListed(client apiClientFunc, teamResource, keyword string, minCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[teamResource]
		if !ok {
			return fmt.Errorf("resource %s not found in state", teamResource)
		}
		auth, api := client()
		return retryUntilUsersListed(auth, usersListedTimeout, func() (int, error) {
			resp, _, err := api.GetTeamsApiV2().GetTeamMemberships(auth, rs.Primary.ID, *datadogV2.NewGetTeamMembershipsOptionalParameters().WithFilterKeyword(keyword))
			if err != nil {
				return 0, fmt.Errorf("error listing team memberships: %w", err)
			}
			if len(resp.GetData()) < minCount {
				return 0, nil
			}
			return len(resp.GetData()), nil
		}, fmt.Errorf("team memberships returned fewer than %d results for keyword %q", minCount, keyword))
	}
}
