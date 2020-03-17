package datadog

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	datadog "github.com/zorkian/go-datadog-api"
)

func resourceDatadogDashboardList() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogDashboardListCreate,
		Update: resourceDatadogDashboardListUpdate,
		Read:   resourceDatadogDashboardListRead,
		Delete: resourceDatadogDashboardListDelete,
		Exists: resourceDatadogDashboardListExists,
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
				Description: "A set of dashbaord items that belong to this list",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The type of this dashboard",
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
	client := providerConf.CommunityClient

	dashboardList, err := buildDatadogDashboardList(d)
	if err != nil {
		fmt.Printf("Error building the dashboard list %s", err.Error())
	}
	dashboardList, err = client.CreateDashboardList(dashboardList)
	if err != nil {
		return translateClientError(err, "error creating dashboard list")
	}
	id := dashboardList.GetId()
	d.SetId(strconv.Itoa(id))

	// Add all the dash list items into the List
	if len(d.Get("dash_item").(*schema.Set).List()) > 0 {
		dashboardListV2Items, _ := buildDatadogDashboardListItemsV2(d)
		_, err := client.UpdateDashboardListItemsV2(id, dashboardListV2Items)
		if err != nil {
			return translateClientError(err, "error updating dashboard list item")
		}
	}

	return resourceDatadogDashboardListRead(d, meta)
}

func resourceDatadogDashboardListUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	id, err := strconv.Atoi(d.Id())

	// Make any necessary updates to the Overall Dashboard List Object
	dashList, err := buildDatadogDashboardList(d)
	dashList.SetId(id)
	dashList.SetName(d.Get("name").(string))
	err = client.UpdateDashboardList(dashList)
	if err != nil {
		return translateClientError(err, "error updating dashboard list")
	}

	// Delete all elements from the dash list and add back only the ones in the config
	completeDashListV2, err := client.GetDashboardListItemsV2(id)
	if err != nil {
		return translateClientError(err, "error getting dashboard list item")
	}
	completeDashListV2, err = client.DeleteDashboardListItemsV2(id, completeDashListV2)
	if err != nil {
		return translateClientError(err, "error deleting dashboard list item")
	}
	if len(d.Get("dash_item").(*schema.Set).List()) > 0 {
		dashboardListV2Items, _ := buildDatadogDashboardListItemsV2(d)
		_, err := client.UpdateDashboardListItemsV2(id, dashboardListV2Items)
		if err != nil {
			return translateClientError(err, "error updating dashboard list item")
		}
	}

	return resourceDatadogDashboardListRead(d, meta)
}

func resourceDatadogDashboardListRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	id, err := strconv.Atoi(d.Id())

	//Read the overall Dashboard List object
	dashList, err := client.GetDashboardList(id)
	if err != nil {
		return translateClientError(err, "error getting dashboard list")
	}
	d.SetId(strconv.Itoa(id))
	d.Set("name", dashList.GetName())

	// Read and set all the dashboard list elements
	completeItemListV2, err := client.GetDashboardListItemsV2(id)
	if err != nil {
		return translateClientError(err, "error getting dashboard list item")
	}
	dashItemListV2, err := buildTerraformDashboardListItemsV2(d, completeItemListV2)
	d.Set("dash_item", dashItemListV2)
	return err
}

func resourceDatadogDashboardListDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	id, _ := strconv.Atoi(d.Id())
	// Deleting the overall List will also take care of deleting its sub elements
	// Deletion of individual dash items happens in the Update method
	// Note this doesn't delete the actual dashboards, just removes them from the deleted list
	err := client.DeleteDashboardList(id)
	if err != nil {
		return translateClientError(err, "error deleting dashboard list")
	}
	return nil
}

func resourceDatadogDashboardListExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	id, _ := strconv.Atoi(d.Id())
	// Only check existence of the overall Dash List, not its sub items
	if _, err := client.GetDashboardList(id); err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return false, nil
		}
		return false, translateClientError(err, "error checking dashboard list exists")
	}
	return true, nil
}

func resourceDatadogDashboardListImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogDashboardListRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func buildDatadogDashboardList(d *schema.ResourceData) (*datadog.DashboardList, error) {
	var dashboardList datadog.DashboardList
	dashboardList.SetName(d.Get("name").(string))
	return &dashboardList, nil
}

func buildDatadogDashboardListItemsV2(d *schema.ResourceData) ([]datadog.DashboardListItemV2, error) {
	var dashboardListV2Items []datadog.DashboardListItemV2
	for _, dashItem := range d.Get("dash_item").(*schema.Set).List() {
		dashItemRaw := dashItem.(map[string]interface{})
		var dashItem datadog.DashboardListItemV2
		dashItem.SetID(dashItemRaw["dash_id"].(string))
		dashItem.SetType(dashItemRaw["type"].(string))
		dashboardListV2Items = append(dashboardListV2Items, dashItem)
	}
	return dashboardListV2Items, nil
}

func buildTerraformDashboardListItemsV2(d *schema.ResourceData, completeItemListV2 []datadog.DashboardListItemV2) ([]map[string]interface{}, error) {
	dashItemListV2 := make([]map[string]interface{}, 0, 1)
	for _, item := range completeItemListV2 {
		dashItem := make(map[string]interface{})
		dashItem["type"] = item.GetType()
		dashItem["dash_id"] = item.GetID()
		dashItemListV2 = append(dashItemListV2, dashItem)
	}
	return dashItemListV2, nil
}
