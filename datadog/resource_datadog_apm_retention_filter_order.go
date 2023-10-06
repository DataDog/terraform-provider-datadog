package datadog

import (
	"context"
	"fmt"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogApmRetentionFilterOrder() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog [APM Retention Filters API](https://docs.datadoghq.com/api/v2/apm-retention-filters/) resource, which is used to manage Datadog APM retention filters order.",
		CreateContext: resourceDatadogApmRetentionFilterOrderCreate,
		UpdateContext: resourceDatadogApmRetentionFilterOrderUpdate,
		ReadContext:   resourceDatadogApmRetentionFilterOrderRead,
		DeleteContext: resourceDatadogApmRetentionFilterOrderDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"filter_ids": {
					Description: "The filter IDs list. The order of filters IDs in this attribute defines the overall APM retention filters order.. If `filter_ids` is empty or not specified, it will import the actual order, and create the resource. Otherwise, it will try to update the order.",
					Type:        schema.TypeList,
					Optional:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
			}
		},
	}
}

func resourceDatadogApmRetentionFilterOrderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rfList, err := buildDatadogApmRetentionFiltersCreateReq(d)
	if err != nil {
		return diag.FromErr(err)
	}

	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	// if len(rfList.Data) > 0 {
	// 	return resourceDatadogApmRetentionFilterOrderUpdate(ctx, d, meta)
	// }
	httpResponse, err := apiInstances.GetApmRetentionFiltersApiV2().ReorderApmRetentionFilters(auth, *rfList)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 400 {
			fmt.Printf("cannot map retention filters to existing ones, will try to import it with Id `filterOrderId`\n")
			d.SetId("filterOrderId")
			return resourceDatadogApmRetentionFilterOrderRead(ctx, d, meta)
		}
		return utils.TranslateClientErrorDiag(err, httpResponse, "error re-ordering APM retention filters")
	}
	d.SetId("filterOrderId")
	listData, httpResponse, err := apiInstances.GetApmRetentionFiltersApiV2().ListApmRetentionFilters(auth)
	return updateApmRetentionFilterOrderState(d, &listData)
}

func resourceDatadogApmRetentionFilterOrderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	listData, httpResponse, err := apiInstances.GetApmRetentionFiltersApiV2().ListApmRetentionFilters(auth)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error getting apm retention filters order")
	}
	if err := utils.CheckForUnparsed(listData); err != nil {
		return diag.FromErr(err)
	}
	return updateApmRetentionFilterOrderState(d, &listData)
}

func updateApmRetentionFilterOrderState(d *schema.ResourceData, listData *datadogV2.RetentionFiltersResponse) diag.Diagnostics {
	filterIds := GetApmFilterIds(*listData)
	if err := d.Set("filter_ids", filterIds); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceDatadogApmRetentionFilterOrderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ddOrderList, err := buildDatadogApmRetentionFiltersCreateReq(d)
	if err != nil {
		return diag.FromErr(err)
	}

	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	httpResponse, err := apiInstances.GetApmRetentionFiltersApiV2().ReorderApmRetentionFilters(auth, *ddOrderList)
	currentOrder, _, listErr := apiInstances.GetApmRetentionFiltersApiV2().ListApmRetentionFilters(auth)
	if err != nil {
		// Cannot map filters to existing ones
		if httpResponse != nil && httpResponse.StatusCode == 400 {
			if listErr != nil {
				return utils.TranslateClientErrorDiag(err, httpResponse, "error getting APM retention filters order")
			}
			return diag.Errorf("cannot map filters to existing ones\n existing filters: %s\n filters to be updated: %s",
				currentOrder.Data,
				ddOrderList.Data)
		}
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating APM retention filters order")
	}
	d.SetId("filterOrderID")
	return updateApmRetentionFilterOrderState(d, &currentOrder)
}

func resourceDatadogApmRetentionFilterOrderDelete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}

func getFilterIdList(d *schema.ResourceData) []datadogV2.RetentionFilterWithoutAttributes {
	tfList := d.Get("filter_ids").([]interface{})
	ddList := make([]datadogV2.RetentionFilterWithoutAttributes, len(tfList))
	for i, id := range tfList {
		ddList[i] = datadogV2.RetentionFilterWithoutAttributes{
			Id:   id.(string),
			Type: "apm_retention_filter",
		}
	}
	return ddList
}

// Map to model
func buildDatadogApmRetentionFiltersCreateReq(d *schema.ResourceData) (*datadogV2.ReorderRetentionFiltersRequest, error) {
	filtersOrder := datadogV2.NewReorderRetentionFiltersRequestWithDefaults()
	filtersOrder.SetData(getFilterIdList(d))
	return filtersOrder, nil
}
