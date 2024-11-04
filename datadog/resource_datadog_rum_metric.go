package datadog

import (
	"context"
	"fmt"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogRumMetric() *schema.Resource {
	return &schema.Resource{
		Description:   "Resource for interacting with the rum_metrics API",
		CreateContext: resourceDatadogRumMetricCreate,
		ReadContext:   resourceDatadogRumMetricRead,
		UpdateContext: resourceDatadogRumMetricUpdate,
		DeleteContext: resourceDatadogRumMetricDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"compute": {
					Type:        schema.TypeList,
					Required:    true,
					ForceNew:    true,
					Description: "The compute rule to compute the log-based metric. This field can't be updated after creation.",
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{

							"aggregation_type": {
								Type:             schema.TypeString,
								Required:         true,
								ForceNew:         true,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewRumMetricComputeAggregationTypeFromValue),
								Description:      "The type of aggregation to use. This field can't be updated after creation.",
							},

							"path": {
								Type:        schema.TypeString,
								Optional:    true,
								ForceNew:    true,
								Description: "The path to the value the log-based metric will aggregate on (only used if the aggregation type is a \"distribution\"). This field can't be updated after creation.",
							},

							"include_percentiles": {
								Description: "Toggle to include/exclude percentiles for a distribution metric. Defaults to false. Can only be applied to metrics that have an `aggregation_type` of distribution.",
								Type:        schema.TypeBool,
								Optional:    true,
							},
						},
					},
				},

				"filter": {
					Type:        schema.TypeList,
					Required:    true,
					Description: "The log-based metric filter. RUM events matching this filter will be aggregated in this metric.",
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{

							"query": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The search query - following the log search syntax.",
							},
						},
					},
				},

				"group_by": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "The rules for the group by.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{

							"path": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The path to the value the log-based metric will be aggregated over.",
							},

							"tag_name": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Name of the tag that gets created.",
							},
						},
					},
				},

				"name": {
					Type:        schema.TypeString,
					Required:    true,
					ForceNew:    true,
					Description: "The name of the log-based metric. This field can't be updated after creation.",
					StateFunc: func(val any) string {
						return utils.NormMetricNameParse(val.(string))
					},
				},
			}
		},
	}
}

func buildDatadogRumMetric(d *schema.ResourceData) (*datadogV2.RumMetricCreateData, error) {
	result := datadogV2.NewRumMetricCreateDataWithDefaults()
	result.SetId(d.Get("name").(string))

	attributes := datadogV2.NewRumMetricCreateAttributesWithDefaults()

	compute, err := getRumMetricCompute(d)
	if err != nil {
		return nil, err
	}
	attributes.SetCompute(*compute)

	filter, err := getRumMetricFilter(d)
	if err != nil {
		return nil, err
	}
	attributes.SetFilter(*filter)

	groupBy, err := getRumMetricGroupBy(d)
	if err != nil {
		return nil, err
	}
	attributes.SetGroupBy(groupBy)

	result.SetAttributes(*attributes)
	return result, nil
}

func getRumMetricCompute(d *schema.ResourceData) (*datadogV2.RumMetricCompute, error) {
	resourceCompute := d.Get("compute").([]interface{})[0].(map[string]interface{})
	compute := datadogV2.NewRumMetricComputeWithDefaults()

	if aggregationType, ok := resourceCompute["aggregation_type"]; ok {
		aggregation_type := datadogV2.RumMetricComputeAggregationType(aggregationType.(string))
		compute.SetAggregationType(aggregation_type)
		if aggregation_type == datadogV2.RUMMETRICCOMPUTEAGGREGATIONTYPE_DISTRIBUTION {
			if includePercentiles, ok := resourceCompute["include_percentiles"]; ok {
				compute.SetIncludePercentiles(includePercentiles.(bool))
			}
		}
	}

	if path, ok := resourceCompute["path"]; ok && path != "" {
		compute.SetPath(path.(string))
	}

	return compute, nil
}

func getRumMetricUpdateCompute(d *schema.ResourceData) (*datadogV2.RumMetricUpdateCompute, error) {
	resourceCompute := d.Get("compute").([]interface{})[0].(map[string]interface{})
	updateCompute := datadogV2.NewRumMetricUpdateComputeWithDefaults()

	if aggregationType, ok := resourceCompute["aggregation_type"]; ok {
		aggregation_type := datadogV2.RumMetricComputeAggregationType(aggregationType.(string))
		if aggregation_type == datadogV2.RUMMETRICCOMPUTEAGGREGATIONTYPE_DISTRIBUTION {
			if includePercentiles, ok := resourceCompute["include_percentiles"]; ok {
				updateCompute.SetIncludePercentiles(includePercentiles.(bool))
			}
		}
	}

	return updateCompute, nil
}

func getRumMetricFilter(d *schema.ResourceData) (*datadogV2.RumMetricFilter, error) {
	resourceFilter := d.Get("filter").([]interface{})[0].(map[string]interface{})
	filter := datadogV2.NewRumMetricFilterWithDefaults()

	if query, ok := resourceFilter["query"]; ok {
		filter.SetQuery(query.(string))
	}

	return filter, nil
}

func getRumMetricGroupBy(d *schema.ResourceData) ([]datadogV2.RumMetricGroupBy, error) {
	resourceGroupBys := d.Get("group_by").(*schema.Set).List()
	groupBys := make([]datadogV2.RumMetricGroupBy, len(resourceGroupBys))
	for i, v := range resourceGroupBys {
		if v == nil {
			continue
		}
		resourceGroupBy := v.(map[string]interface{})
		groupBy := datadogV2.NewRumMetricGroupByWithDefaults()
		if path, ok := resourceGroupBy["path"]; ok {
			groupBy.SetPath(path.(string))
		}
		if path, ok := resourceGroupBy["tag_name"]; ok {
			groupBy.SetTagName(path.(string))
		}
		groupBys[i] = *groupBy
	}

	return groupBys, nil
}

func resourceDatadogRumMetricCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	resultRumMetricCreateData, err := buildDatadogRumMetric(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error building RumMetric object: %w", err))
	}

	ddObject := datadogV2.NewRumMetricCreateRequestWithDefaults()
	ddObject.SetData(*resultRumMetricCreateData)

	response, httpResponse, err := apiInstances.GetRumMetricsApiV2().CreateRumMetric(auth, *ddObject)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating RumMetric")
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		return diag.FromErr(err)
	}
	id := *response.GetData().Id
	d.SetId(id)

	return updateRumMetricState(d, response.Data)
}

func updateRumMetricState(d *schema.ResourceData, resource *datadogV2.RumMetricResponseData) diag.Diagnostics {
	if ddAttributes, ok := resource.GetAttributesOk(); ok {
		if computeDDModel, ok := ddAttributes.GetComputeOk(); ok {
			computeMap := map[string]interface{}{}

			if v, ok := computeDDModel.GetAggregationTypeOk(); ok {
				computeMap["aggregation_type"] = *v
				if *v == datadogV2.RUMMETRICCOMPUTEAGGREGATIONTYPE_DISTRIBUTION {
					if w, ok := computeDDModel.GetIncludePercentilesOk(); ok {
						computeMap["include_percentiles"] = *w
					}
				}
			}
			if v, ok := computeDDModel.GetPathOk(); ok {
				computeMap["path"] = *v
			}
			if err := d.Set("compute", []map[string]interface{}{computeMap}); err != nil {
				return diag.FromErr(err)
			}
		}
		if filterDDModel, ok := ddAttributes.GetFilterOk(); ok {
			filterMap := map[string]interface{}{}
			if v, ok := filterDDModel.GetQueryOk(); ok {
				filterMap["query"] = *v
			}
			if err := d.Set("filter", []map[string]interface{}{filterMap}); err != nil {
				return diag.FromErr(err)
			}
		}
		if groupByArray, ok := ddAttributes.GetGroupByOk(); ok {
			mapAttributesArray := make([]map[string]interface{}, 0)

			for _, groupByArrayItem := range *groupByArray {
				mapAttributesArrayIntf := map[string]interface{}{}
				if v, ok := groupByArrayItem.GetPathOk(); ok {
					mapAttributesArrayIntf["path"] = *v
				}
				if v, ok := groupByArrayItem.GetTagNameOk(); ok {
					mapAttributesArrayIntf["tag_name"] = *v
				}

				mapAttributesArray = append(mapAttributesArray, mapAttributesArrayIntf)
			}
			if err := d.Set("group_by", mapAttributesArray); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if v, ok := resource.GetIdOk(); ok {
		if err := d.Set("name", *v); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceDatadogRumMetricRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	var err error

	id := d.Id()

	resourceRumMetricResponse, httpResp, err := apiInstances.GetRumMetricsApiV2().GetRumMetric(auth, id)

	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			// this condition takes on the job of the deprecated Exists handlers
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpResp, "error reading RumMetric")
	}
	if err := utils.CheckForUnparsed(resourceRumMetricResponse); err != nil {
		return diag.FromErr(err)
	}

	resource := resourceRumMetricResponse.GetData()
	return updateRumMetricState(d, &resource)
}

func buildDatadogRumMetricUpdate(d *schema.ResourceData) (*datadogV2.RumMetricUpdateData, error) {
	result := datadogV2.NewRumMetricUpdateDataWithDefaults()
	attributes := datadogV2.NewRumMetricUpdateAttributesWithDefaults()

	updateCompute, err := getRumMetricUpdateCompute(d)
	if err != nil {
		return nil, err
	}
	attributes.SetCompute(*updateCompute)

	filter, err := getRumMetricFilter(d)
	if err != nil {
		return nil, err
	}
	attributes.SetFilter(*filter)

	groupBy, err := getRumMetricGroupBy(d)
	if err != nil {
		return nil, err
	}
	attributes.SetGroupBy(groupBy)

	result.SetAttributes(*attributes)

	return result, nil
}

func resourceDatadogRumMetricUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	resultRumMetricUpdateData, err := buildDatadogRumMetricUpdate(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error building RumMetric object: %w", err))
	}

	ddObject := datadogV2.NewRumMetricUpdateRequestWithDefaults()
	ddObject.SetData(*resultRumMetricUpdateData)
	id := d.Id()

	response, httpResponse, err := apiInstances.GetRumMetricsApiV2().UpdateRumMetric(auth, id, *ddObject)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating RumMetric")
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		return diag.FromErr(err)
	}

	return updateRumMetricState(d, response.Data)
}

func resourceDatadogRumMetricDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	var err error

	id := d.Id()

	httpResponse, err := apiInstances.GetRumMetricsApiV2().DeleteRumMetric(auth, id)

	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting RumMetric")
	}

	return nil
}
