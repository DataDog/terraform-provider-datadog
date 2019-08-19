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
		dashboardListItems, _ := buildDatadogDashboardListItems(d)
		_, err := meta.(*datadog.Client).UpdateDashboardListItems(id, dashboardListItems)
		if err != nil {
			return err
		}
	}

	return resourceDatadogDashboardListRead(d, meta)
}

func resourceDatadogDashboardListUpdate(d *schema.ResourceData, meta interface{}) error {
	id, err := strconv.Atoi(d.Id())

	// Make any necessary updates to the Overall Dashboard List Object
	dashList, err := buildDatadogDashboardList(d)
	dashList.SetId(id)
	dashList.SetName(d.Get("name").(string))
	err = meta.(*datadog.Client).UpdateDashboardList(dashList)
	if err != nil {
		return err
	}

	// Delete all elements from the dash list and add back only the ones in the config
	completeDashList, err := meta.(*datadog.Client).GetDashboardListItems(id)
	if err != nil {
		return err
	}
	completeDashList, err = meta.(*datadog.Client).DeleteDashboardListItems(id, completeDashList)
	if err != nil {
		return err
	}
	if len(d.Get("dash_item").(*schema.Set).List()) > 0 {
		dashboardListItems, _ := buildDatadogDashboardListItems(d)
		_, err := meta.(*datadog.Client).UpdateDashboardListItems(id, dashboardListItems)
		if err != nil {
			return err
		}
	}

	return resourceDatadogDashboardListRead(d, meta)
}

func resourceDatadogDashboardListRead(d *schema.ResourceData, meta interface{}) error {
	id, err := strconv.Atoi(d.Id())

	//Read the overall Dashboard List object
	dashList, err := meta.(*datadog.Client).GetDashboardList(id)
	if err != nil {
		return err
	}
	d.SetId(strconv.Itoa(id))
	d.Set("name", dashList.GetName())

	// Read all the dashboard list elements
	completeItemList, err := meta.(*datadog.Client).GetDashboardListItems(id)
	if err != nil {
		return err
	}
	dashItemList := make([]map[string]interface{}, 0, 1)
	for _, item := range completeItemList {
		dashItem := make(map[string]interface{})
		dashItem["type"] = item.GetType()
		dashItem["dash_id"] = item.GetId()
		dashItemList = append(dashItemList, dashItem)
	}
	d.Set("dash_item", dashItemList)
	return err
}

func resourceDatadogDashboardListDelete(d *schema.ResourceData, meta interface{}) error {
	id, _ := strconv.Atoi(d.Id())

	// Deleting the overall List will also take care of deleting its sub elements
	// Deletion of individual dash items happens in the Update method
	err := meta.(*datadog.Client).DeleteDashboardList(id)
	if err != nil {
		return err
	}
	return nil
}

func resourceDatadogDashboardListExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	id, _ := strconv.Atoi(d.Id())

	// Only check existence of the overall Dash List, not its sub items
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
	var dashboardList datadog.DashboardList
	dashboardList.SetName(d.Get("name").(string))
	return &dashboardList, nil
}

func buildDatadogDashboardListItems(d *schema.ResourceData) ([]datadog.DashboardListItem, error) {
	var dashboardListItems []datadog.DashboardListItem
	for _, dashItem := range d.Get("dash_item").(*schema.Set).List() {
		dashItemRaw := dashItem.(map[string]interface{})
		var dashItem datadog.DashboardListItem
		dashItem.SetId(dashItemRaw["dash_id"].(string))
		dashItem.SetType(dashItemRaw["type"].(string))
		dashboardListItems = append(dashboardListItems, dashItem)
	}
	return dashboardListItems, nil
}
