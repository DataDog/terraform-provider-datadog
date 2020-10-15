package datadog

import (
	"fmt"
	"strconv"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceDatadogDashboardList() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDatadogDashboardListRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}

func dataSourceDatadogDashboardListRead(d *schema.ResourceData, meta interface{}) error {

	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	listResponse, _, err := datadogClientV1.DashboardListsApi.ListDashboardLists(authV1).Execute()

	if err != nil {
		return translateClientError(err, "error querying dashboard lists")
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
		return fmt.Errorf("Couldn't find a dashboard list named %s", searchedName)
	}

	id := foundList.GetId()
	d.SetId(strconv.Itoa(int(id)))

	return nil
}
