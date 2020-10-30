package datadog

import (
	"fmt"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceDatadogDashboard() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDatadogDashboardRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			// Computed values
			"title": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceDatadogDashboardRead(d *schema.ResourceData, meta interface{}) error {

	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	dashResponse, _, err := datadogClientV1.DashboardsApi.ListDashboards(authV1).Execute()

	if err != nil {
		return translateClientError(err, "error querying dashboard")
	}

	searchedName := d.Get("name")
	var foundDashes []*datadogV1.DashboardSummaryDashboards

	for _, dash := range dashResponse.GetDashboards() {
		if dash.GetTitle() == searchedName {
			foundDashes = append(foundDashes, &dash)
		}
	}

	if len(foundDashes) == 0 {
		return fmt.Errorf("Couldn't find a dashboard named %s", searchedName)
	} else if len(foundDashes) > 1 {
		return fmt.Errorf("%s returned more than one dashboard", searchedName)
	}

	d.SetId(foundDashes[0].GetId())
	d.Set("url", foundDashes[0].GetUrl())
	d.Set("title", foundDashes[0].GetTitle())

	return nil
}
