package datadog

import (
	"context"
	"fmt"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogLogsArchiveOrder() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog [Logs Archive API](https://docs.datadoghq.com/api/v2/logs-archives/) resource, which is used to manage Datadog log archives order.",
		CreateContext: resourceDatadogLogsArchiveOrderCreate,
		UpdateContext: resourceDatadogLogsArchiveOrderUpdate,
		ReadContext:   resourceDatadogLogsArchiveOrderRead,
		DeleteContext: resourceDatadogLogsArchiveOrderDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"archive_ids": {
				Description: "The archive IDs list. The order of archive IDs in this attribute defines the overall archive order for logs. If `archive_ids` is empty or not specified, it will import the actual archive order, and create the resource. Otherwise, it will try to update the order.",
				Type:        schema.TypeList,
				Computed:    true,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceDatadogLogsArchiveOrderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ddArchiveList, err := buildDatadogArchiveOrderCreateReq(d)
	if err != nil {
		return diag.FromErr(err)
	}

	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	if len(ddArchiveList.Data.Attributes.GetArchiveIds()) > 0 {
		return resourceDatadogLogsArchiveOrderUpdate(ctx, d, meta)
	}
	order, httpResponse, err := datadogClientV2.LogsArchivesApi.UpdateLogsArchiveOrder(authV2, *ddArchiveList)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 422 {
			fmt.Printf("cannot map archives to existing ones, will try to import it with Id `archiveOrderID`\n")
			d.SetId("archiveOrderID")
			return resourceDatadogLogsArchiveOrderRead(ctx, d, meta)
		}
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating logs archive order")
	}
	if err := utils.CheckForUnparsed(order); err != nil {
		return diag.FromErr(err)
	}
	d.SetId("archiveOrderID")
	return updateLogsArchiveOrderState(d, &order)
}

func resourceDatadogLogsArchiveOrderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2
	order, httpResponse, err := datadogClientV2.LogsArchivesApi.GetLogsArchiveOrder(authV2)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error getting logs archive order")
	}
	if err := utils.CheckForUnparsed(order); err != nil {
		return diag.FromErr(err)
	}

	return updateLogsArchiveOrderState(d, &order)
}

func updateLogsArchiveOrderState(d *schema.ResourceData, order *datadogV2.LogsArchiveOrder) diag.Diagnostics {
	if err := d.Set("archive_ids", order.Data.Attributes.ArchiveIds); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceDatadogLogsArchiveOrderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ddArchiveList, err := buildDatadogArchiveOrderCreateReq(d)
	if err != nil {
		return diag.FromErr(err)
	}

	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2
	updatedOrder, httpResponse, err := datadogClientV2.LogsArchivesApi.UpdateLogsArchiveOrder(authV2, *ddArchiveList)
	if err != nil {
		// Cannot map archives to existing ones
		if httpResponse != nil && httpResponse.StatusCode == 422 {
			ddArchiveOrder, _, getErr := datadogClientV2.LogsArchivesApi.GetLogsArchiveOrder(authV2)
			if getErr != nil {
				return utils.TranslateClientErrorDiag(err, httpResponse, "error getting logs archive order")
			}
			return diag.Errorf("cannot map archives to existing ones\n existing archives: %s\n archive to be updated: %s",
				ddArchiveOrder.Data.Attributes.ArchiveIds,
				ddArchiveList.Data.Attributes.GetArchiveIds())
		}
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating logs archive order")
	}
	if err := utils.CheckForUnparsed(updatedOrder); err != nil {
		return diag.FromErr(err)
	}
	d.SetId("archiveOrderID")
	return updateLogsArchiveOrderState(d, &updatedOrder)
}

// The deletion of archive order is not supported from config API.
// This function simply delete the archive order resource from terraform state.
func resourceDatadogLogsArchiveOrderDelete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
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
