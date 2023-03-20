package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogSDSGroupOrder() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog Sensitive Data Scanner Group Order API resource. This can be used to manage the order of Datadog Sensitive Data Scanner Groups.",
		CreateContext: resourceDatadogSDSGroupOrderCreate,
		UpdateContext: resourceDatadogSDSGroupOrderUpdate,
		ReadContext:   resourceDatadogSDSGroupOrderRead,
		DeleteContext: resourceDatadogSDSGroupOrderDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"groups": {
				Description: "The list of Sensitive Data Scanner group IDs, in order. Logs are tested against the query filter of each index one by one following the order of the list.",
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceDatadogSDSGroupOrderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDatadogSDSGroupOrderUpdate(ctx, d, meta)
}

func resourceDatadogSDSGroupOrderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	tfList := d.Get("groups").([]interface{})
	ddList := make([]datadogV2.SensitiveDataScannerGroupItem, len(tfList))
	for i, tfName := range tfList {
		ddList[i] = *datadogV2.NewSensitiveDataScannerGroupItemWithDefaults()
		ddList[i].SetId(tfName.(string))
	}
	ddSDSGroupsList, httpResponse, err := apiInstances.GetSensitiveDataScannerApiV2().ListScanningGroups(auth)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error getting Sensitive Data Scanner groups list")
	}

	SDSGroupOrderRequest := datadogV2.NewSensitiveDataScannerConfigRequestWithDefaults()
	SDSGroupOrderRequestConfig := datadogV2.NewSensitiveDataScannerReorderConfigWithDefaults()
	SDSGroupOrderRequestRelationships := datadogV2.NewSensitiveDataScannerConfigurationRelationshipsWithDefaults()
	SDSGroupOrderRequestGroups := datadogV2.NewSensitiveDataScannerGroupListWithDefaults()
	SDSGroupOrderRequestGroups.SetData(ddList)
	SDSGroupOrderRequestRelationships.SetGroups(*SDSGroupOrderRequestGroups)
	SDSGroupOrderRequestConfig.SetRelationships(*SDSGroupOrderRequestRelationships)
	SDSGroupOrderRequestConfig.SetId(ddSDSGroupsList.Data.GetId())
	SDSGroupOrderRequest.SetData(*SDSGroupOrderRequestConfig)

	updatedOrder, httpResponse, err := apiInstances.GetSensitiveDataScannerApiV2().ReorderScanningGroups(auth, *SDSGroupOrderRequest)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating Sensitive Data Scanner group list")
	}
	if err := utils.CheckForUnparsed(updatedOrder); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(ddSDSGroupsList.Data.GetId())
	return nil
}

func updateSDSGroupOrderState(d *schema.ResourceData, groups datadogV2.SensitiveDataScannerGroupList) diag.Diagnostics {
	groupItems := groups.Data
	tfList := make([]string, len(groupItems))
	for i, ddGroup := range groupItems {
		tfList[i] = ddGroup.GetId()
	}
	if err := d.Set("groups", tfList); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceDatadogSDSGroupOrderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	ddSDSGroupsList, httpResponse, err := apiInstances.GetSensitiveDataScannerApiV2().ListScanningGroups(auth)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error getting Sensitive Data Scanner groups list")
	}
	if err := utils.CheckForUnparsed(ddSDSGroupsList); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(ddSDSGroupsList.Data.GetId())
	return updateSDSGroupOrderState(d, *ddSDSGroupsList.Data.Relationships.Groups)
}

func resourceDatadogSDSGroupOrderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
