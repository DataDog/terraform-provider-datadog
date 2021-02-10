package datadog

import (
	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func resourceDatadogLogsIndexOrder() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a Datadog Logs Index API resource. This can be used to manage the order of Datadog logs indexes.",
		Create:      resourceDatadogLogsIndexOrderCreate,
		Update:      resourceDatadogLogsIndexOrderUpdate,
		Read:        resourceDatadogLogsIndexOrderRead,
		Delete:      resourceDatadogLogsIndexOrderDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The unique name of the index order resource.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"indexes": {
				Description: "The index resource list. Logs are tested against the query filter of each index one by one following the order of the list.",
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceDatadogLogsIndexOrderCreate(d *schema.ResourceData, meta interface{}) error {
	return resourceDatadogLogsIndexOrderUpdate(d, meta)
}

func resourceDatadogLogsIndexOrderUpdate(d *schema.ResourceData, meta interface{}) error {
	var ddIndexList datadogV1.LogsIndexesOrder
	tfList := d.Get("indexes").([]interface{})
	ddList := make([]string, len(tfList))
	for i, tfName := range tfList {
		ddList[i] = tfName.(string)
	}
	ddIndexList.IndexNames = ddList
	var tfId string
	if name, exists := d.GetOk("name"); exists {
		tfId = name.(string)
	}
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	if _, _, err := datadogClientV1.LogsIndexesApi.UpdateLogsIndexOrder(authV1).Body(ddIndexList).Execute(); err != nil {
		return utils.TranslateClientError(err, "error updating logs index list")
	}
	d.SetId(tfId)
	return resourceDatadogLogsIndexOrderRead(d, meta)
}

func resourceDatadogLogsIndexOrderRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient
	ddIndexList, err := client.GetLogsIndexList()
	if err != nil {
		return utils.TranslateClientError(err, "error getting logs index list")
	}
	if err := d.Set("indexes", ddIndexList.IndexNames); err != nil {
		return err
	}
	return nil
}

func resourceDatadogLogsIndexOrderDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
