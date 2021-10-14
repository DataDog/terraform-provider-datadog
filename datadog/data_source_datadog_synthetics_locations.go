package datadog

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func dataSourceDatadogSyntheticsLocations() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve Datadog's Synthetics Locations (to be used in Synthetics tests).",
		ReadContext: dataSourceDatadogSyntheticsLocationsRead,

		// Locations are a map of IDs to names
		Schema: map[string]*schema.Schema{
			"locations": {
				Description: "A map of available Synthetics location IDs to names for Synthetics tests.",
				Type:        schema.TypeMap,
				Computed:    true,
			},
		},
	}
}

func dataSourceDatadogSyntheticsLocationsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	syntheticsLocations, _, err := datadogClientV1.SyntheticsApi.ListLocations(authV1)

	if err != nil {
		return diag.FromErr(err)
	}
	if err := utils.CheckForUnparsed(syntheticsLocations); err != nil {
		return diag.FromErr(err)
	}

	locationsMap := make(map[string]string)

	for _, location := range syntheticsLocations.GetLocations() {
		locationsMap[location.GetId()] = location.GetName()
	}

	if len(locationsMap) > 0 {
		d.SetId("datadog-synthetics-location")
	}
	d.Set("locations", locationsMap)

	return nil
}
