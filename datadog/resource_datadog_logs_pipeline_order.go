package datadog

import (
	"context"
	"slices"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogLogsPipelineOrder() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages the order in which Datadog processes log pipelines. Use at most one pipeline order resource per organization. Each update replaces the entire order, so this resource can overwrite changes made at the same time in the Datadog UI, by automation, or by another Terraform run, even when `ignore_unlisted_pipelines` is `true`.",
		CreateContext: resourceDatadogLogsPipelineOrderCreate,
		UpdateContext: resourceDatadogLogsPipelineOrderUpdate,
		ReadContext:   resourceDatadogLogsPipelineOrderRead,
		DeleteContext: resourceDatadogLogsPipelineOrderDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"name": {
					Description: "The name attribute in the resource `datadog_logs_pipeline_order` needs to be unique. It's recommended to use the same value as the resource name. No related field is available in [Logs Pipeline API](https://docs.datadoghq.com/api/v1/logs-pipelines/#get-pipeline-order).",
					Type:        schema.TypeString,
					Required:    true,
				},
				"pipelines": {
					Description: "The pipeline IDs in execution order. Include every pipeline in the organization unless `ignore_unlisted_pipelines` is `true`.",
					Type:        schema.TypeList,
					Required:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"ignore_unlisted_pipelines": {
					Description: "When `true`, the resource manages only the relative order of the pipelines listed in `pipelines`. Each listed ID must be unique and match an existing pipeline. Unlisted pipelines keep their current positions, and pipelines added outside Terraform do not cause drift. For example, if the current order is `[A, X, B]` and you configure `[B, A]`, the result is `[B, X, A]`: pipeline `X` stays second but runs before `A` instead of after it. Leave this setting `false` to manage the full order, including which pipelines run first and last.",
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
				},
			}
		},
	}
}

func resourceDatadogLogsPipelineOrderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDatadogLogsPipelineOrderUpdate(ctx, d, meta)
}

func listedPipelineSlots(listed, current []string) []int {
	ids := make(map[string]bool, len(listed))
	for _, id := range listed {
		ids[id] = true
	}
	var slots []int
	for i, id := range current {
		if ids[id] {
			slots = append(slots, i)
			delete(ids, id)
		}
	}
	return slots
}

func updateLogsPipelineOrderState(d *schema.ResourceData, order *datadogV1.LogsPipelinesOrder) diag.Diagnostics {
	ids := order.PipelineIds
	if d.Get("ignore_unlisted_pipelines").(bool) {
		slots := listedPipelineSlots(utils.GetStringSlice(d, "pipelines"), ids)
		listed := make([]string, len(slots))
		for i, slot := range slots {
			listed[i] = ids[slot]
		}
		ids = listed
	}
	if err := d.Set("pipelines", ids); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func getLogsPipelineOrder(conf *ProviderConfiguration) (*datadogV1.LogsPipelinesOrder, diag.Diagnostics) {
	order, httpResponse, err := conf.DatadogApiInstances.GetLogsPipelinesApiV1().GetLogsPipelineOrder(conf.Auth)
	if err != nil {
		return nil, utils.TranslateClientErrorDiag(err, httpResponse, "error getting logs pipeline order")
	}
	if err := utils.CheckForUnparsed(order); err != nil {
		return nil, diag.FromErr(err)
	}
	return &order, nil
}

func resourceDatadogLogsPipelineOrderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	order, diags := getLogsPipelineOrder(meta.(*ProviderConfiguration))
	if diags.HasError() {
		return diags
	}
	return updateLogsPipelineOrderState(d, order)
}

func resourceDatadogLogsPipelineOrderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	relative := d.Get("ignore_unlisted_pipelines").(bool)
	order, diags := applyLogsPipelineOrder(ctx, meta.(*ProviderConfiguration), utils.GetStringSlice(d, "pipelines"), relative)
	if diags.HasError() {
		if relative {
			d.Partial(true)
		}
		return diags
	}
	d.SetId(d.Get("name").(string))
	return updateLogsPipelineOrderState(d, order)
}

func applyLogsPipelineOrder(ctx context.Context, providerConf *ProviderConfiguration, pipelineIDs []string, relative bool) (*datadogV1.LogsPipelinesOrder, diag.Diagnostics) {
	if relative {
		logCustomPipelineMutex.Lock()
		defer logCustomPipelineMutex.Unlock()
		current, diags := getLogsPipelineOrder(providerConf)
		if diags.HasError() {
			return nil, diags
		}
		slots := listedPipelineSlots(pipelineIDs, current.PipelineIds)
		if len(slots) != len(pipelineIDs) {
			return nil, diag.Errorf("each ID in pipelines must be unique and match an existing pipeline: matched %d of %d configured IDs", len(slots), len(pipelineIDs))
		}
		merged := slices.Clone(current.PipelineIds)
		for i, slot := range slots {
			merged[slot] = pipelineIDs[i]
		}
		if slices.Equal(merged, current.PipelineIds) {
			return current, nil
		}
		if err := ctx.Err(); err != nil {
			return nil, diag.FromErr(err)
		}
		pipelineIDs = merged
	}
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	updatedOrder, httpResponse, err := apiInstances.GetLogsPipelinesApiV1().
		UpdateLogsPipelineOrder(auth, datadogV1.LogsPipelinesOrder{PipelineIds: pipelineIDs})
	if err != nil {
		// Cannot map pipelines to existing ones
		if strings.Contains(err.Error(), "422 Unprocessable Entity") {
			ddPipelineOrder, diags := getLogsPipelineOrder(providerConf)
			if diags.HasError() {
				return nil, diags
			}
			return nil, diag.Errorf("cannot map pipelines to existing ones\n existing pipelines: %s\n pipeline to be updated: %s",
				ddPipelineOrder.PipelineIds,
				pipelineIDs)
		}
		return nil, utils.TranslateClientErrorDiag(err, httpResponse, "error updating logs pipeline order")
	}
	if err := utils.CheckForUnparsed(updatedOrder); err != nil {
		return nil, diag.FromErr(err)
	}
	return &updatedOrder, nil
}

// The deletion of pipeline order is not supported from config API.
// This function simply delete the pipeline order resource from terraform state.
func resourceDatadogLogsPipelineOrderDelete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {

	return nil
}
