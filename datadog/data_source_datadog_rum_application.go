package datadog

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDatadogRUMApplication() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve a Datadog RUM Application.",
		ReadContext: dataSourceDatadogRUMApplicationRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the RUM application",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the RUM application",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The RUM application type. Supported values are `browser`, `ios`, `android`, `react-native`, `flutter`",
			},
			"client_token": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The client token",
			},
		},
	}
}

func dataSourceDatadogRUMApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	searchID, ok := d.GetOk("id")
	if !ok {
		return diag.Errorf("Missing ID in RUM application data source.")
	}

	resp, _, err := apiInstances.GetRumApiV2().GetRUMApplication(auth, searchID.(string))
	if err != nil {
		return diag.Errorf("Couldn't find RUM application with id %s", searchID)
	}

	rum_app := resp.Data.GetAttributes()

	d.SetId(rum_app.GetApplicationId())
	if err := d.Set("name", rum_app.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", rum_app.Type); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("client_token", rum_app.Hash); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
