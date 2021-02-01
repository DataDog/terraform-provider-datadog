package datadog

import (
	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogLogsMetric() *schema.Resource {
	return &schema.Resource{
		Description: "Resource for interacting with the logs_metric API",
		Create:      resourceDatadogLogsMetricCreate,
		Read:        resourceDatadogLogsMetricRead,
		Update:      resourceDatadogLogsMetricUpdate,
		Delete:      resourceDatadogLogsMetricDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogLogsMetricImport,
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
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validateEnumValue(datadogV2.NewLogsMetricComputeAggregationTypeFromValue),
							Description:  "The type of aggregation to use. This field can't be updated after creation.",
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

func resourceDatadogLogsMetricCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV2
	auth := providerConf.AuthV2

	resultLogsMetricCreateData, err := buildDatadogLogsMetric(d)
	if err != nil {
		return translateClientError(err, "error building LogsMetric object")
	}

	ddObject := datadogV2.NewLogsMetricCreateRequestWithDefaults()
	ddObject.SetData(*resultLogsMetricCreateData)

	response, _, err := datadogClient.LogsMetricsApi.CreateLogsMetric(auth).Body(*ddObject).Execute()
	if err != nil {
		return translateClientError(err, "error creating LogsMetric")
	}
	id := *response.GetData().Id
	d.SetId(id)

	return resourceDatadogLogsMetricRead(d, meta)
}

func resourceDatadogLogsMetricRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV2
	auth := providerConf.AuthV2
	var err error

	id := d.Id()

	resourceLogsMetricResponse, httpResp, err := datadogClient.LogsMetricsApi.GetLogsMetric(auth, id).Execute()

	if err != nil {
		if httpResp.StatusCode == 404 {
			// this condition takes on the job of the deprecated Exists handlers
			d.SetId("")
			return nil
		}
		return translateClientError(err, "error reading LogsMetric")
	}

	resource := resourceLogsMetricResponse.GetData()

	if ddAttributes, ok := resource.GetAttributesOk(); ok {
		if computeDDModel, ok := ddAttributes.GetComputeOk(); ok {
			computeMap := map[string]interface{}{}
			if v, ok := computeDDModel.GetAggregationTypeOk(); ok {
				computeMap["aggregation_type"] = *v
			}
			if v, ok := computeDDModel.GetPathOk(); ok {
				computeMap["path"] = *v
			}
			d.Set("compute", []map[string]interface{}{computeMap})
		}
		if filterDDModel, ok := ddAttributes.GetFilterOk(); ok {
			filterMap := map[string]interface{}{}
			if v, ok := filterDDModel.GetQueryOk(); ok {
				filterMap["query"] = *v
			}
			d.Set("filter", []map[string]interface{}{filterMap})
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
			d.Set("group_by", mapAttributesArray)
		}
	}

	if v, ok := resource.GetIdOk(); ok {
		d.Set("name", *v)
	}

	return nil
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

func resourceDatadogLogsMetricUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV2
	auth := providerConf.AuthV2

	resultLogsMetricUpdateData, err := buildDatadogLogsMetricUpdate(d)
	if err != nil {
		return translateClientError(err, "error building LogsMetric object")
	}

	ddObject := datadogV2.NewLogsMetricUpdateRequestWithDefaults()
	ddObject.SetData(*resultLogsMetricUpdateData)
	id := d.Id()

	_, _, err = datadogClient.LogsMetricsApi.UpdateLogsMetric(auth, id).Body(*ddObject).Execute()
	if err != nil {
		return translateClientError(err, "error updating LogsMetric")
	}

	return resourceDatadogLogsMetricRead(d, meta)
}

func resourceDatadogLogsMetricDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV2
	auth := providerConf.AuthV2
	var err error

	id := d.Id()

	_, err = datadogClient.LogsMetricsApi.DeleteLogsMetric(auth, id).Execute()

	if err != nil {
		return translateClientError(err, "error deleting LogsMetric")
	}

	return nil
}

func resourceDatadogLogsMetricImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogLogsMetricRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
