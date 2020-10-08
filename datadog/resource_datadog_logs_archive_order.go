package datadog

import (
	"fmt"
	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"strings"
)

func resourceDatadogLogsArchiveOrder() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogLogsArchiveOrderCreate,
		Update: resourceDatadogLogsArchiveOrderUpdate,
		Read:   resourceDatadogLogsArchiveOrderRead,
		Delete: resourceDatadogLogsArchiveOrderDelete,
		Exists: resourceDatadogLogsArchiveOrderExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {Type: schema.TypeString, Required: true},
			"archives": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceDatadogLogsArchiveOrderCreate(d *schema.ResourceData, meta interface{}) error {
	return resourceDatadogLogsArchiveOrderUpdate(d, meta)
}

func resourceDatadogLogsArchiveOrderRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2
	ddList, _, err := datadogClientV2.LogsArchivesApi.GetLogsArchiveOrder(authV2).Execute()
	if err != nil {
		return translateClientError(err, "error getting logs archive order")
	}

	if err = d.Set("archives", ddList.Data.Attributes.ArchiveIds); err != nil {
		return err
	}

	return nil
}

func resourceDatadogLogsArchiveOrderUpdate(d *schema.ResourceData, meta interface{}) error {
	var ddArchiveList datadogV2.LogsArchiveOrder
	tfList := d.Get("archives").([]interface{})
	ddList := make([]string, len(tfList))
	for i, id := range tfList {
		ddList[i] = id.(string)
	}
	ddArchiveList.Data.Attributes.ArchiveIds = ddList
	var tfId string
	if name, exists := d.GetOk("name"); exists {
		tfId = name.(string)
	}
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2
	if _, _, err := datadogClientV2.LogsArchivesApi.UpdateLogsArchiveOrder(authV2).Body(ddArchiveList).Execute(); err != nil {
		// Cannot map archives to existing ones
		if strings.Contains(err.Error(), "422 Unprocessable Entity") {
			ddArchiveOrder, _, getErr := datadogClientV2.LogsArchivesApi.GetLogsArchiveOrder(authV2).Execute()
			if getErr != nil {
				return translateClientError(err, "error getting logs archive order")
			}
			return fmt.Errorf("cannot map archives to existing ones\n existing archives: %s\n archive to be updated: %s",
				ddArchiveOrder.Data.Attributes.ArchiveIds,
				ddList)
		}
		return translateClientError(err, "error updating logs archive order")
	}
	d.SetId(tfId)
	return resourceDatadogLogsArchiveOrderRead(d, meta)
}

// The deletion of archive order is not supported from config API.
// This function simply delete the archive order resource from terraform state.
func resourceDatadogLogsArchiveOrderDelete(_ *schema.ResourceData, _ interface{}) error {
	return nil
}

func resourceDatadogLogsArchiveOrderExists(_ *schema.ResourceData, _ interface{}) (bool, error) {
	return true, nil
}
