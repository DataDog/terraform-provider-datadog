package datadog

import (
	"context"
	"fmt"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogLogsMetric() *schema.Resource {
	return &schema.Resource{
		Description:   "Resource for interacting with the logs_metric API",
		CreateContext: resourceDatadogLogsMetricCreate,
		ReadContext:   resourceDatadogLogsMetricRead,
		UpdateContext: resourceDatadogLogsMetricUpdate,
		DeleteContext: resourceDatadogLogsMetricDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{

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
							ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewLogsMetricComputeAggregationTypeFromValue),
							Description:      "The type of aggregation to use. This field can't be updated after creation.",
						},

						"path": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "The path to the value the log-based metric will aggregate on (only used if the aggregation type is a \"distribution\"). This field can't be updated after creation.",
						},
					},
				},
			},

			"filter": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "The log-based metric filter. Logs matching this filter will be aggregated in this metric.",
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
				Type:        schema.TypeList,
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
			},
		},
	}
}

func buildDatadogLogsMetric(d *schema.ResourceData) (*datadogV2.LogsMetricCreateData, error) {
	result := datadogV2.NewLogsMetricCreateDataWithDefaults()
	result.SetId(d.Get("name").(string))

	attributes := datadogV2.NewLogsMetricCreateAttributesWithDefaults()

	compute, err := getCompute(d)
	if err != nil {
		return nil, err
	}
	attributes.SetCompute(*compute)

	filter, err := getFilter(d)
	if err != nil {
		return nil, err
	}
	attributes.SetFilter(*filter)

	groupBys, err := getGroupBys(d)
	if err != nil {
		return nil, err
	}
	attributes.SetGroupBy(groupBys)

	result.SetAttributes(*attributes)
	return result, nil
}

func getCompute(d *schema.ResourceData) (*datadogV2.LogsMetricCompute, error) {
	resourceCompute := d.Get("compute").([]interface{})[0].(map[string]interface{})
	compute := datadogV2.NewLogsMetricComputeWithDefaults()

	if aggregationType, ok := resourceCompute["aggregation_type"]; ok {
		compute.SetAggregationType(datadogV2.LogsMetricComputeAggregationType(aggregationType.(string)))
	}

	path, ok := resourceCompute["path"]
	if ok && path != "" {
		compute.SetPath(path.(string))
	}

	return compute, nil
}

func getFilter(d *schema.ResourceData) (*datadogV2.LogsMetricFilter, error) {
	resourceFilter := d.Get("filter").([]interface{})[0].(map[string]interface{})
	filter := datadogV2.NewLogsMetricFilterWithDefaults()

	if query, ok := resourceFilter["query"]; ok {
		filter.SetQuery(query.(string))
	}

	return filter, nil
}

func getGroupBys(d *schema.ResourceData) ([]datadogV2.LogsMetricGroupBy, error) {
	resourceGroupBys := d.Get("group_by").([]interface{})
	groupBys := make([]datadogV2.LogsMetricGroupBy, len(resourceGroupBys))

	for i, v := range resourceGroupBys {
		resourceGroupBy := v.(map[string]interface{})
		groupBy := datadogV2.NewLogsMetricGroupByWithDefaults()
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

func resourceDatadogLogsMetricCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV2
	auth := providerConf.AuthV2

	resultLogsMetricCreateData, err := buildDatadogLogsMetric(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error building LogsMetric object: %w", err))
	}

	ddObject := datadogV2.NewLogsMetricCreateRequestWithDefaults()
	ddObject.SetData(*resultLogsMetricCreateData)

	response, httpResponse, err := datadogClient.LogsMetricsApi.CreateLogsMetric(auth, *ddObject)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating LogsMetric")
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		return diag.FromErr(err)
	}
	id := *response.GetData().Id
	d.SetId(id)

	return updateLogsMetricState(d, response.Data)
}

func updateLogsMetricState(d *schema.ResourceData, resource *datadogV2.LogsMetricResponseData) diag.Diagnostics {
	if ddAttributes, ok := resource.GetAttributesOk(); ok {
		if computeDDModel, ok := ddAttributes.GetComputeOk(); ok {
			computeMap := map[string]interface{}{}
			if v, ok := computeDDModel.GetAggregationTypeOk(); ok {
				computeMap["aggregation_type"] = *v
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

func resourceDatadogLogsMetricRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV2
	auth := providerConf.AuthV2
	var err error

	id := d.Id()

	resourceLogsMetricResponse, httpResp, err := datadogClient.LogsMetricsApi.GetLogsMetric(auth, id)

	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			// this condition takes on the job of the deprecated Exists handlers
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpResp, "error reading LogsMetric")
	}
	if err := utils.CheckForUnparsed(resourceLogsMetricResponse); err != nil {
		return diag.FromErr(err)
	}

	resource := resourceLogsMetricResponse.GetData()
	return updateLogsMetricState(d, &resource)
}

func buildDatadogLogsMetricUpdate(d *schema.ResourceData) (*datadogV2.LogsMetricUpdateData, error) {
	result := datadogV2.NewLogsMetricUpdateDataWithDefaults()
	attributes := datadogV2.NewLogsMetricUpdateAttributesWithDefaults()

	filter, err := getFilter(d)
	if err != nil {
		return nil, err
	}
	attributes.SetFilter(*filter)

	groupBys, err := getGroupBys(d)
	if err != nil {
		return nil, err
	}
	attributes.SetGroupBy(groupBys)

	result.SetAttributes(*attributes)

	return result, nil
}

func resourceDatadogLogsMetricUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV2
	auth := providerConf.AuthV2

	resultLogsMetricUpdateData, err := buildDatadogLogsMetricUpdate(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error building LogsMetric object: %w", err))
	}

	ddObject := datadogV2.NewLogsMetricUpdateRequestWithDefaults()
	ddObject.SetData(*resultLogsMetricUpdateData)
	id := d.Id()

	response, httpResponse, err := datadogClient.LogsMetricsApi.UpdateLogsMetric(auth, id, *ddObject)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating LogsMetric")
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		return diag.FromErr(err)
	}

	return updateLogsMetricState(d, response.Data)
}

func resourceDatadogLogsMetricDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV2
	auth := providerConf.AuthV2
	var err error

	id := d.Id()

	httpResponse, err := datadogClient.LogsMetricsApi.DeleteLogsMetric(auth, id)

	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting LogsMetric")
	}

	return nil
}
