package datadog

import (
	"context"
	"strconv"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceDatadogDashboardList() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about an existing dashboard list, for use in other resources. In particular, it can be used in a dashboard to register it in the list.",
		ReadContext: dataSourceDatadogDashboardListRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Description:  "A dashboard list name to limit the search.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}

func dataSourceDatadogDashboardListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	listResponse, httpresp, err := datadogClientV1.DashboardListsApi.ListDashboardLists(authV1)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error querying dashboard lists")
	}
	if err := utils.CheckForUnparsed(listResponse); err != nil {
		return diag.FromErr(err)
	}

	searchedName := d.Get("name")
	var foundList *datadogV1.DashboardList

	for _, dashList := range listResponse.GetDashboardLists() {
		if dashList.GetName() == searchedName {
			foundList = &dashList
			break
		}
	}

	if foundList == nil {
		return diag.Errorf("Couldn't find a dashboard list named %s", searchedName)
	}

	id := foundList.GetId()
	d.SetId(strconv.Itoa(int(id)))

	return nil
}
