package datadog

import (
	"context"
	"strconv"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogDashboardList() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog dashboard_list resource. This can be used to create and manage Datadog Dashboard Lists and the individual dashboards within them.",
		CreateContext: resourceDatadogDashboardListCreate,
		UpdateContext: resourceDatadogDashboardListUpdate,
		ReadContext:   resourceDatadogDashboardListRead,
		DeleteContext: resourceDatadogDashboardListDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewDashboardTypeFromValue),
							Description:      "The type of this dashboard.",
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

func resourceDatadogDashboardListCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	dashboardListPayload, err := buildDatadogDashboardList(d)
	if err != nil {
		return diag.Errorf("failed to parse resource configuration: %s", err.Error())
	}
	dashboardList, httpresp, err := apiInstances.GetDashboardListsApiV1().CreateDashboardList(auth, *dashboardListPayload)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error creating dashboard list")
	}
	if err := utils.CheckForUnparsed(dashboardList); err != nil {
		return diag.FromErr(err)
	}
	id := dashboardList.GetId()
	d.SetId(strconv.Itoa(int(id)))

	// Add all the dash list items into the List
	if len(d.Get("dash_item").(*schema.Set).List()) > 0 {
		dashboardListV2Items, err := buildDatadogDashboardListUpdateItemsV2(d)
		if err != nil {
			return diag.Errorf("failed to parse resource configuration: %s", err.Error())
		}
		_, _, err = apiInstances.GetDashboardListsApiV2().UpdateDashboardListItems(auth, id, *dashboardListV2Items)
		if err != nil {
			return utils.TranslateClientErrorDiag(err, httpresp, "error updating dashboard list item")
		}
	}

	return resourceDatadogDashboardListRead(ctx, d, meta)
}

func resourceDatadogDashboardListUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.Errorf("failed to parse resource id: %s", err.Error())
	}

	// Make any necessary updates to the Overall Dashboard List Object
	dashList, err := buildDatadogDashboardList(d)
	if err != nil {
		return diag.Errorf("failed to parse resource configuration: %s", err.Error())
	}
	dashList.SetName(d.Get("name").(string))
	_, httpresp, err := apiInstances.GetDashboardListsApiV1().UpdateDashboardList(auth, id, *dashList)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error updating dashboard list")
	}

	// Delete all elements from the dash list and add back only the ones in the config
	completeDashListV2, httpresp, err := apiInstances.GetDashboardListsApiV2().GetDashboardListItems(auth, id)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting dashboard list item")
	}
	if err := utils.CheckForUnparsed(completeDashListV2); err != nil {
		return diag.FromErr(err)
	}
	completeDashListDeleteV2, err := buildDatadogDashboardListDeleteItemsV2(completeDashListV2)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error creating dashboard list delete item")
	}
	_, httpresp, err = apiInstances.GetDashboardListsApiV2().DeleteDashboardListItems(auth, id, *completeDashListDeleteV2)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error deleting dashboard list item")
	}
	if len(d.Get("dash_item").(*schema.Set).List()) > 0 {
		dashboardListV2Items, err := buildDatadogDashboardListUpdateItemsV2(d)
		if err != nil {
			return diag.Errorf("failed to parse resource configuration: %s", err.Error())
		}
		_, httpresp, err = apiInstances.GetDashboardListsApiV2().UpdateDashboardListItems(auth, id, *dashboardListV2Items)
		if err != nil {
			return utils.TranslateClientErrorDiag(err, httpresp, "error updating dashboard list item")
		}
	}

	return resourceDatadogDashboardListRead(ctx, d, meta)
}

func resourceDatadogDashboardListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	id, _ := strconv.ParseInt(d.Id(), 10, 64)

	//Read the overall Dashboard List object
	dashList, httpresp, err := apiInstances.GetDashboardListsApiV1().GetDashboardList(auth, id)
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting dashboard list")
	}
	d.Set("name", dashList.GetName())

	// Read and set all the dashboard list elements
	completeItemListV2, _, err := apiInstances.GetDashboardListsApiV2().GetDashboardListItems(auth, id)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting dashboard list item")
	}
	if err := utils.CheckForUnparsed(completeItemListV2); err != nil {
		return diag.FromErr(err)
	}
	dashItemListV2, err := buildTerraformDashboardListItemsV2(completeItemListV2)
	if err != nil {
		return diag.Errorf("failed to parse resource configuration: %s", err.Error())
	}
	d.Set("dash_item", dashItemListV2)
	return diag.FromErr(err)
}

func resourceDatadogDashboardListDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	// Deleting the overall List will also take care of deleting its sub elements
	// Deletion of individual dash items happens in the Update method
	// Note this doesn't delete the actual dashboards, just removes them from the deleted list
	_, httpresp, err := apiInstances.GetDashboardListsApiV1().DeleteDashboardList(auth, id)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error deleting dashboard list")
	}
	return nil
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
		dashID := dashItem.GetId()
		dashItem := datadogV2.NewDashboardListItemRequest(dashID, dashType)
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
