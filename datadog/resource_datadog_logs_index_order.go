package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogLogsIndexOrder() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog Logs Index API resource. This can be used to manage the order of Datadog logs indexes.",
		CreateContext: resourceDatadogLogsIndexOrderCreate,
		UpdateContext: resourceDatadogLogsIndexOrderUpdate,
		ReadContext:   resourceDatadogLogsIndexOrderRead,
		DeleteContext: resourceDatadogLogsIndexOrderDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"name": {
					Description: "The unique name of the index order resource.",
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
				},
				"indexes": {
					Description: "The index resource list. Logs are tested against the query filter of each index one by one following the order of the list.",
					Type:        schema.TypeList,
					Required:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
			}
		},
	}
}

func resourceDatadogLogsIndexOrderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDatadogLogsIndexOrderUpdate(ctx, d, meta)
}

func resourceDatadogLogsIndexOrderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var ddIndexList datadogV1.LogsIndexesOrder
	tfList := d.Get("indexes").([]interface{})
	ddList := make([]string, len(tfList))
	for i, tfName := range tfList {
		ddList[i] = tfName.(string)
	}
	ddIndexList.IndexNames = ddList
	var tfID = "logs_index_order"
	if name, exists := d.GetOk("name"); exists {
		tfID = name.(string)
	}
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	updatedOrder, httpResponse, err := apiInstances.GetLogsIndexesApiV1().UpdateLogsIndexOrder(auth, ddIndexList)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating logs index list")
	}
	if err := utils.CheckForUnparsed(updatedOrder); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(tfID)
	return updateLogsIndexOrderState(d, &updatedOrder)
}

func updateLogsIndexOrderState(d *schema.ResourceData, order *datadogV1.LogsIndexesOrder) diag.Diagnostics {
	if err := d.Set("indexes", order.GetIndexNames()); err != nil {
		return diag.FromErr(err)
	}
	if _, ok := d.GetOk("name"); !ok {
		d.Set("name", d.Id())
	}
	return nil
}

func resourceDatadogLogsIndexOrderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	ddIndexList, httpResponse, err := apiInstances.GetLogsIndexesApiV1().GetLogsIndexOrder(auth)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error getting logs index list")
	}
	if err := utils.CheckForUnparsed(ddIndexList); err != nil {
		return diag.FromErr(err)
	}
	return updateLogsIndexOrderState(d, &ddIndexList)
}

func resourceDatadogLogsIndexOrderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
