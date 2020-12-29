package datadog

import (
	"fmt"
	"strings"

	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
			"archive_ids": {
				Description: "The archive IDs list. The order of archive IDs in this attribute defines the overall archive order for logs. If archive_ids is empty or not specified, it will import the actual archive order, and create the resource. Otherwise, it will try to update the order.",
				Type:        schema.TypeList,
				Computed:    true,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceDatadogLogsArchiveOrderCreate(d *schema.ResourceData, meta interface{}) error {
	ddArchiveList, err := buildDatadogArchiveOrderCreateReq(d)
	if err != nil {
		return err
	}

	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	if len(ddArchiveList.Data.Attributes.GetArchiveIds()) > 0 {
		return resourceDatadogLogsArchiveOrderUpdate(d, meta)
	}
	_, _, err = datadogClientV2.LogsArchivesApi.UpdateLogsArchiveOrder(authV2).Body(*ddArchiveList).Execute()
	if err != nil {
		if strings.Contains(err.Error(), "422 Unprocessable Entity") {
			fmt.Printf("cannot map archives to existing ones, will try to import it with Id `archiveOrderID`\n")
		} else {
			return translateClientError(err, "error creating logs archive order")
		}
	}
	d.SetId("archiveOrderID")
	return resourceDatadogLogsArchiveOrderRead(d, meta)
}

func resourceDatadogLogsArchiveOrderRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2
	ddList, _, err := datadogClientV2.LogsArchivesApi.GetLogsArchiveOrder(authV2).Execute()
	if err != nil {
		return translateClientError(err, "error getting logs archive order")
	}

	if err = d.Set("archive_ids", ddList.Data.Attributes.ArchiveIds); err != nil {
		return err
	}

	return nil
}

func resourceDatadogLogsArchiveOrderUpdate(d *schema.ResourceData, meta interface{}) error {
	ddArchiveList, err := buildDatadogArchiveOrderCreateReq(d)
	if err != nil {
		return err
	}

	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2
	if _, _, err := datadogClientV2.LogsArchivesApi.UpdateLogsArchiveOrder(authV2).Body(*ddArchiveList).Execute(); err != nil {
		// Cannot map archives to existing ones
		if strings.Contains(err.Error(), "422 Unprocessable Entity") {
			ddArchiveOrder, _, getErr := datadogClientV2.LogsArchivesApi.GetLogsArchiveOrder(authV2).Execute()
			if getErr != nil {
				return translateClientError(err, "error getting logs archive order")
			}
			return fmt.Errorf("cannot map archives to existing ones\n existing archives: %s\n archive to be updated: %s",
				ddArchiveOrder.Data.Attributes.ArchiveIds,
				ddArchiveList.Data.Attributes.GetArchiveIds())
		}
		return translateClientError(err, "error updating logs archive order")
	}
	d.SetId("archiveOrderID")
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

func getArchiveIds(d *schema.ResourceData) []string {
	tfList := d.Get("archive_ids").([]interface{})
	ddList := make([]string, len(tfList))
	for i, id := range tfList {
		ddList[i] = id.(string)
	}
	return ddList
}

//Map to model
func buildDatadogArchiveOrderCreateReq(d *schema.ResourceData) (*datadogV2.LogsArchiveOrder, error) {
	archiveOrderAttributes := datadogV2.NewLogsArchiveOrderAttributes(getArchiveIds(d))

	archiveOrderDefinition := datadogV2.NewLogsArchiveOrderDefinitionWithDefaults()
	archiveOrderDefinition.SetAttributes(*archiveOrderAttributes)

	archiveOrder := datadogV2.NewLogsArchiveOrderWithDefaults()
	archiveOrder.SetData(*archiveOrderDefinition)
	return archiveOrder, nil
}
