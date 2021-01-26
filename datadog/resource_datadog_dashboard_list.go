package datadog

import (
	"fmt"
	"strconv"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceDatadogDashboardList() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a Datadog dashboard_list resource. This can be used to create and manage Datadog Dashboard Lists and the individual dashboards within them.",
		Create:      resourceDatadogDashboardListCreate,
		Update:      resourceDatadogDashboardListUpdate,
		Read:        resourceDatadogDashboardListRead,
		Delete:      resourceDatadogDashboardListDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogDashboardListImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Dashboard List",
			},
			"dash_item": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A set of dashboard items that belong to this list",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateEnumValue(datadogV2.NewDashboardTypeFromValue),
							Description:  "The type of this dashboard. Available options are: `custom_timeboard`, `custom_screenboard`, `integration_screenboard`, `integration_timeboard`, and `host_timeboard`",
						},
						"dash_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The ID of the dashboard to add",
						},
					},
				},
			},
		},
	}
}

func resourceDatadogDashboardListCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	datadogClientV2 := providerConf.DatadogClientV2
	authV1 := providerConf.AuthV1
	authV2 := providerConf.AuthV2

	dashboardListPayload, err := buildDatadogDashboardList(d)
	if err != nil {
		return fmt.Errorf("failed to parse resource configuration: %s", err.Error())
	}
	dashboardList, _, err := datadogClientV1.DashboardListsApi.CreateDashboardList(authV1).Body(*dashboardListPayload).Execute()
	if err != nil {
		return translateClientError(err, "error creating dashboard list")
	}
	id := dashboardList.GetId()
	d.SetId(strconv.Itoa(int(id)))

	// Add all the dash list items into the List
	if len(d.Get("dash_item").(*schema.Set).List()) > 0 {
		dashboardListV2Items, err := buildDatadogDashboardListUpdateItemsV2(d)
		if err != nil {
			return fmt.Errorf("failed to parse resource configuration: %s", err.Error())
		}
		_, _, err = datadogClientV2.DashboardListsApi.UpdateDashboardListItems(authV2, id).Body(*dashboardListV2Items).Execute()
		if err != nil {
			return translateClientError(err, "error updating dashboard list item")
		}
	}

	return resourceDatadogDashboardListRead(d, meta)
}

func resourceDatadogDashboardListUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	datadogClientV2 := providerConf.DatadogClientV2
	authV1 := providerConf.AuthV1
	authV2 := providerConf.AuthV2

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse resource id: %s", err.Error())
	}

	// Make any necessary updates to the Overall Dashboard List Object
	dashList, err := buildDatadogDashboardList(d)
	if err != nil {
		return fmt.Errorf("failed to parse resource configuration: %s", err.Error())
	}
	dashList.SetName(d.Get("name").(string))
	_, _, err = datadogClientV1.DashboardListsApi.UpdateDashboardList(authV1, id).Body(*dashList).Execute()
	if err != nil {
		return translateClientError(err, "error updating dashboard list")
	}

	// Delete all elements from the dash list and add back only the ones in the config
	completeDashListV2, _, err := datadogClientV2.DashboardListsApi.GetDashboardListItems(authV2, id).Execute()
	if err != nil {
		return translateClientError(err, "error getting dashboard list item")
	}
	completeDashListDeleteV2, err := buildDatadogDashboardListDeleteItemsV2(completeDashListV2)
	if err != nil {
		return translateClientError(err, "error creating dashboard list delete item")
	}
	_, _, err = datadogClientV2.DashboardListsApi.DeleteDashboardListItems(authV2, id).Body(*completeDashListDeleteV2).Execute()
	if err != nil {
		return translateClientError(err, "error deleting dashboard list item")
	}
	if len(d.Get("dash_item").(*schema.Set).List()) > 0 {
		dashboardListV2Items, err := buildDatadogDashboardListUpdateItemsV2(d)
		if err != nil {
			return fmt.Errorf("failed to parse resource configuration: %s", err.Error())
		}
		_, _, err = datadogClientV2.DashboardListsApi.UpdateDashboardListItems(authV2, id).Body(*dashboardListV2Items).Execute()
		if err != nil {
			return translateClientError(err, "error updating dashboard list item")
		}
	}

	return resourceDatadogDashboardListRead(d, meta)
}

func resourceDatadogDashboardListRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	datadogClientV2 := providerConf.DatadogClientV2
	authV1 := providerConf.AuthV1
	authV2 := providerConf.AuthV2

	id, err := strconv.ParseInt(d.Id(), 10, 64)

	//Read the overall Dashboard List object
	dashList, httpresp, err := datadogClientV1.DashboardListsApi.GetDashboardList(authV1, id).Execute()
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return translateClientError(err, "error getting dashboard list")
	}
	d.Set("name", dashList.GetName())

	// Read and set all the dashboard list elements
	completeItemListV2, _, err := datadogClientV2.DashboardListsApi.GetDashboardListItems(authV2, id).Execute()
	if err != nil {
		return translateClientError(err, "error getting dashboard list item")
	}
	dashItemListV2, err := buildTerraformDashboardListItemsV2(completeItemListV2)
	if err != nil {
		return fmt.Errorf("failed to parse resource configuration: %s", err.Error())
	}
	d.Set("dash_item", dashItemListV2)
	return err
}

func resourceDatadogDashboardListDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	// Deleting the overall List will also take care of deleting its sub elements
	// Deletion of individual dash items happens in the Update method
	// Note this doesn't delete the actual dashboards, just removes them from the deleted list
	_, _, err := datadogClientV1.DashboardListsApi.DeleteDashboardList(authV1, id).Execute()
	if err != nil {
		return translateClientError(err, "error deleting dashboard list")
	}
	return nil
}

func resourceDatadogDashboardListImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogDashboardListRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func buildDatadogDashboardList(d *schema.ResourceData) (*datadogV1.DashboardList, error) {
	var dashboardList datadogV1.DashboardList
	dashboardList.SetName(d.Get("name").(string))
	return &dashboardList, nil
}

func buildDatadogDashboardListDeleteItemsV2(dashboardListItems datadogV2.DashboardListItems) (*datadogV2.DashboardListDeleteItemsRequest, error) {
	dashboardListV2ItemsArr := make([]datadogV2.DashboardListItemRequest, 0)
	for _, dashItem := range dashboardListItems.GetDashboards() {
		dashType := dashItem.GetType()
		dashId := dashItem.GetId()
		dashItem := datadogV2.NewDashboardListItemRequest(dashId, dashType)
		dashboardListV2ItemsArr = append(dashboardListV2ItemsArr, *dashItem)
	}
	dashboardListV2Items := datadogV2.NewDashboardListDeleteItemsRequest()
	dashboardListV2Items.SetDashboards(dashboardListV2ItemsArr)
	return dashboardListV2Items, nil
}

func buildDatadogDashboardListUpdateItemsV2(d *schema.ResourceData) (*datadogV2.DashboardListUpdateItemsRequest, error) {
	dashboardListV2ItemsArr := make([]datadogV2.DashboardListItemRequest, 0)
	for _, dashItem := range d.Get("dash_item").(*schema.Set).List() {
		dashItemRaw := dashItem.(map[string]interface{})
		dashType := datadogV2.DashboardType(dashItemRaw["type"].(string))
		dashItem := datadogV2.NewDashboardListItemRequest(dashItemRaw["dash_id"].(string), dashType)
		dashboardListV2ItemsArr = append(dashboardListV2ItemsArr, *dashItem)
	}
	dashboardListV2Items := datadogV2.NewDashboardListUpdateItemsRequest()
	dashboardListV2Items.SetDashboards(dashboardListV2ItemsArr)
	return dashboardListV2Items, nil
}

func buildTerraformDashboardListItemsV2(completeItemListV2 datadogV2.DashboardListItems) ([]map[string]interface{}, error) {
	dashItemListV2 := make([]map[string]interface{}, 0, 1)
	for _, item := range completeItemListV2.GetDashboards() {
		dashItem := make(map[string]interface{})
		dashItem["type"] = item.GetType()
		dashItem["dash_id"] = item.GetId()
		dashItemListV2 = append(dashItemListV2, dashItem)
	}
	return dashItemListV2, nil
}
