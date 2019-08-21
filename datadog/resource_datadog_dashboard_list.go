package datadog

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
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
				Description: "The name of the Dashboard ListV2",
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
							Description: "The type of these dashboards",
						},
						"dash_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The list of dashboard IDs to add",
							// Elem: &schema.Schema{
							// 	Type: schema.TypeString,
							// },
						},
					},
				},
			},
		},
	}
}

func resourceDatadogDashboardListCreate(d *schema.ResourceData, meta interface{}) error {
	dashboardList, err := buildDatadogDashboardList(d)
	if err != nil {
		fmt.Printf("Error building the dashboard list %s", err.Error())
	}
	dashboardList, err = meta.(*datadog.Client).CreateDashboardList(dashboardList)
	if err != nil {
		return fmt.Errorf("Failed to create dashboard list using Datadog API: %s", err.Error())
	}
	id := dashboardList.GetId()
	d.SetId(strconv.Itoa(id))

	// Add all the dash list items into the List
	if len(d.Get("dash_item").(*schema.Set).List()) > 0 {
		dashboardListV2Items, _ := buildDatadogDashboardListItemsV2(d)
		_, err := meta.(*datadog.Client).UpdateDashboardListItemsV2(id, dashboardListV2Items)
		if err != nil {
			return err
		}
	}

	return resourceDatadogDashboardListRead(d, meta)
}

func resourceDatadogDashboardListUpdate(d *schema.ResourceData, meta interface{}) error {
	id, err := strconv.Atoi(d.Id())

	// Make any necessary updates to the Overall Dashboard ListV2 Object
	dashList, err := buildDatadogDashboardList(d)
	dashList.SetId(id)
	dashList.SetName(d.Get("name").(string))
	err = meta.(*datadog.Client).UpdateDashboardList(dashList)
	if err != nil {
		return err
	}

	// Delete all elements from the dash list and add back only the ones in the config
	completeDashListV2, err := meta.(*datadog.Client).GetDashboardListItemsV2(id)
	if err != nil {
		return err
	}
	completeDashListV2, err = meta.(*datadog.Client).DeleteDashboardListItemsV2(id, completeDashListV2)
	if err != nil {
		return err
	}
	if len(d.Get("dash_item").(*schema.Set).List()) > 0 {
		dashboardListV2Items, _ := buildDatadogDashboardListItemsV2(d)
		_, err := meta.(*datadog.Client).UpdateDashboardListItemsV2(id, dashboardListV2Items)
		if err != nil {
			return err
		}
	}

	return resourceDatadogDashboardListRead(d, meta)
}

func resourceDatadogDashboardListRead(d *schema.ResourceData, meta interface{}) error {
	id, err := strconv.Atoi(d.Id())

	//Read the overall Dashboard List object
	dashListV2, err := meta.(*datadog.Client).GetDashboardList(id)
	if err != nil {
		return err
	}
	d.SetId(strconv.Itoa(id))
	d.Set("name", dashListV2.GetName())

	// Read and set all the dashboard list elements
	completeItemListV2, err := meta.(*datadog.Client).GetDashboardListItemsV2(id)
	if err != nil {
		return err
	}
	dashItemListV2, err := buildTerraformDashboardListItemsV2(d, completeItemListV2)
	d.Set("dash_item", dashItemListV2)
	return err
}

func resourceDatadogDashboardListDelete(d *schema.ResourceData, meta interface{}) error {
	id, _ := strconv.Atoi(d.Id())
	// Deleting the overall ListV2 will also take care of deleting its sub elements
	// Deletion of individual dash items happens in the Update method
	err := meta.(*datadog.Client).DeleteDashboardList(id)
	if err != nil {
		return err
	}
	return nil
}

func resourceDatadogDashboardListExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	id, _ := strconv.Atoi(d.Id())
	// Only check existence of the overall Dash ListV2, not its sub items
	if _, err := meta.(*datadog.Client).GetDashboardList(id); err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return false, nil
		}
		return false, err
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
	var dashboardListV2 datadog.DashboardList
	dashboardListV2.SetName(d.Get("name").(string))
	return &dashboardListV2, nil
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
