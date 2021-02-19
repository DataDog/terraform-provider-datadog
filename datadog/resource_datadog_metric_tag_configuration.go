package datadog

import (
	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

func resourceDatadogMetricTagConfiguration() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a Datadog metric_tag_configuration resource. This can be used to manage a custom metric's tags.",
		Create:      resourceDatadogMetricTagConfigurationCreate,
		Read:        resourceDatadogMetricTagConfigurationRead,
		Update:      resourceDatadogMetricTagConfigurationUpdate,
		Delete:      resourceDatadogMetricTagConfigurationDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogMetricTagConfigurationImport,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The metric name for this resource.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"metric_type": {
				Description:  "The metric's type. This field can't be updated after creation. Allowed enum values: gauge,count,distribution",
				Type:         schema.TypeString,
				Required:     false,
				ForceNew:     true,
				ValidateFunc: validators.ValidateEnumValue(datadogV2.NewMetricTagConfigurationMetricTypesFromValue),
			},
			"tags": {
				Description: "A list of tag keys that will be queryable for your metric.",
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    true,
				ForceNew:    true,
			},
			"include_percentiles": {
				Description: "Toggle to include/exclude percentiles for a distribution metric. Defaults to false. Can only be applied to metrics that have a metric_type of distribution.",
				Type:        schema.TypeBool,
				Required:    false,
			},
		},
	}
}

func buildDatadogMetricTagConfiguration(d *schema.ResourceData) (*datadogV2.MetricTagConfigurationCreateData, error) {
	result := datadogV2.NewMetricTagConfigurationCreateDataWithDefaults()
	result.SetId(d.Get("id").(string))

	attributes := datadogV2.NewMetricTagConfigurationCreateAttributesWithDefaults()
	tags := d.Get("tags").([]string)
	attributes.SetTags(tags)

	metric_type := d.Get("metric_type").(datadogV2.MetricTagConfigurationMetricTypes)
	attributes.SetMetricType(metric_type)

	if metric_type == datadogV2.METRICTAGCONFIGURATIONMETRICTYPES_DISTRIBUTION {
		include_percentiles := d.Get("include_percentiles").(bool)
		attributes.SetIncludePercentiles(include_percentiles)
	}

	result.SetAttributes(*attributes)

	return result, nil
}

func buildDatadogMetricTagConfigurationUpdate(d *schema.ResourceData, existing_metric_type *datadogV2.MetricTagConfigurationMetricTypes) (*datadogV2.MetricTagConfigurationUpdateData, error) {
	result := datadogV2.NewMetricTagConfigurationUpdateDataWithDefaults()
	id := d.Get("id").(string)
	result.SetId(id)

	attributes := datadogV2.NewMetricTagConfigurationUpdateAttributesWithDefaults()
	tags := d.Get("tags").([]string)
	attributes.SetTags(tags)

	if *existing_metric_type == datadogV2.METRICTAGCONFIGURATIONMETRICTYPES_DISTRIBUTION {
		include_percentiles := d.Get("include_percentiles").(bool)
		attributes.SetIncludePercentiles(include_percentiles)
	}

	result.SetAttributes(*attributes)
	return result, nil
}

func resourceDatadogMetricTagConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV2
	auth := providerConf.AuthV2

	id := d.Id()

	resultMetricTagConfigurationData, err := buildDatadogMetricTagConfiguration(d)
	if err != nil {
		return utils.TranslateClientError(err, "error building MetricTagConfiguration object")
	}
	ddObject := datadogV2.NewMetricTagConfigurationCreateRequestWithDefaults()
	ddObject.SetData(*resultMetricTagConfigurationData)

	response, _, err := datadogClient.MetricsApi.CreateTagConfiguration(auth, id).Body(*ddObject).Execute()
	if err != nil {
		return utils.TranslateClientError(err, "error creating MetricTagConfiguration")
	}
	response_id := *response.GetData().Id
	d.SetId(response_id)

	return updateMetricTagConfigurationState(d, response.Data)
}

func updateMetricTagConfigurationState(d *schema.ResourceData, metricTagConfiguration *datadogV2.MetricTagConfiguration) error {
	if attributes, ok := metricTagConfiguration.GetAttributesOk(); ok {
		if metric_type, ok := attributes.GetMetricTypeOk(); ok {
			if err := d.Set("metric_type", metric_type); err != nil {
				return err
			}
			if *metric_type == datadogV2.METRICTAGCONFIGURATIONMETRICTYPES_DISTRIBUTION {
				if err := d.Set("include_percentiles", attributes.GetIncludePercentiles()); err != nil {
					return err
				}
			}
		}
		if err := d.Set("tags", attributes.GetTags()); err != nil {
			return err
		}
	}

	// we do not care about the created_at nor modified_at fields

	return nil
}

func resourceDatadogMetricTagConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV2
	auth := providerConf.AuthV2

	id := d.Id()

	metricTagConfigurationResponse, httpresp, err := datadogClient.MetricsApi.ListTagConfigurationByName(auth, id).Execute()
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientError(err, "metric not found")
	}

	resource := metricTagConfigurationResponse.GetData()
	return updateMetricTagConfigurationState(d, &resource)
}

func resourceDatadogMetricTagConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV2
	auth := providerConf.AuthV2

	id := d.Id()

	metricTagConfigurationResponse, _, err := datadogClient.MetricsApi.ListTagConfigurationByName(auth, id).Execute()
	if err != nil {
		return utils.TranslateClientError(err, "metric not found")
	}

	existing_metric_type := metricTagConfigurationResponse.GetData().Attributes.GetMetricType()

	resultMetricTagConfigurationUpdateData, err := buildDatadogMetricTagConfigurationUpdate(d, &existing_metric_type)
	if err != nil {
		return utils.TranslateClientError(err, "error building MetricTagConfiguration object")
	}

	ddObject := datadogV2.NewMetricTagConfigurationUpdateRequestWithDefaults()
	ddObject.SetData(*resultMetricTagConfigurationUpdateData)

	response, _, err := datadogClient.MetricsApi.UpdateTagConfiguration(auth, id).Body(*ddObject).Execute()
	if err != nil {
		return utils.TranslateClientError(err, "error updating MetricTagConfiguration")
	}

	return updateMetricTagConfigurationState(d, response.Data)
}

func resourceDatadogMetricTagConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV2
	auth := providerConf.AuthV2
	var err error

	id := d.Id()

	_, err = datadogClient.MetricsApi.DeleteTagConfiguration(auth, id).Execute()

	if err != nil {
		return utils.TranslateClientError(err, "error deleting MetricTagConfiguration")
	}

	return nil
}

func resourceDatadogMetricTagConfigurationImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogMetricTagConfigurationRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
