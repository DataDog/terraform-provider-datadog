package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceDatadogDashboard() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about an existing dashboard, for use in other resources. In particular, it can be used in a monitor message to link to a specific dashboard.",
		ReadContext: dataSourceDatadogDashboardRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Description:  "The dashboard name to search for. Must only match one dashboard.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			// Computed values
			"title": {
				Description: "The name of the dashboard.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"url": {
				Description: "The URL to a specific dashboard.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceDatadogDashboardRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	dashResponse, _, err := datadogClientV1.DashboardsApi.ListDashboards(authV1).Execute()

	if err != nil {
		return utils.TranslateClientErrorDiag(err, "error querying dashboard")
	}

	searchedName := d.Get("name")
	var foundDashes []datadogV1.DashboardSummaryDefinition

	for _, dash := range dashResponse.GetDashboards() {
		if dash.GetTitle() == searchedName {
			foundDashes = append(foundDashes, dash)
		}
	}

	if len(foundDashes) == 0 {
		return diag.Errorf("Couldn't find a dashboard named %s", searchedName)
	} else if len(foundDashes) > 1 {
		return diag.Errorf("%s returned more than one dashboard", searchedName)
	}

	d.SetId(foundDashes[0].GetId())
	d.Set("url", foundDashes[0].GetUrl())
	d.Set("title", foundDashes[0].GetTitle())

	return nil
}
