package datadog

import (
	"fmt"
	"regexp"
	"strings"

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
			"metric_name": {
				Description: "The metric name for this resource.",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: func(val interface{}, k string) (warns []string, errs []error) {
					v := val.(string)
					// pulled from dogweb/dogweb/lib/validation/schemas/metric.py
					re := regexp.MustCompile(`^[A-Za-z][A-Za-z0-9\\.\\_]*$`)
					if len(v) < 1 || len(v) > 200 {
						errs = append(errs, fmt.Errorf("expected metric name length of %s to be in the range (%d - %d), got %s", k, 1, 200, v))
					}
					if !re.MatchString(v) {
						errs = append(errs, fmt.Errorf("metric name not allowed."))
					}
					// todo[efraese] ensure metric name is not a standard metric
					return
				},
			},
			"metric_type": {
				Description:  "The metric's type. This field can't be updated after creation. Allowed enum values: gauge,count,distribution",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validators.ValidateEnumValue(datadogV2.NewMetricTagConfigurationMetricTypesFromValue),
			},
			"tags": {
				Description: "A list of tag keys that will be queryable for your metric.",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: func(val interface{}, k string) (warns []string, errs []error) {
						v := val.(string)
						// pulled from dogweb/dogweb/lib/validation/schemas/metric.py
						re := regexp.MustCompile(`^[A-Za-z][A-Za-z0-9\\.\\-\\_:\\/]*$`)
						if len(v) < 1 || len(v) > 200 {
							errs = append(errs, fmt.Errorf("expected tag length of %s to be in the range (%d - %d), got %s", k, 1, 200, v))
						}
						if strings.HasSuffix(v, ":") {
							errs = append(errs, fmt.Errorf("tag ends in : which is not allowed."))
						}
						if !re.MatchString(v) {
							errs = append(errs, fmt.Errorf("tag not allowed."))
						}
						return
					},
				},
				Required: true,
			},
			"include_percentiles": {
				// TODO[efraese] fix schema to only allow this field when the metric type is a distribution (I think this is done via teh build funcs?)
				Description: "Toggle to include/exclude percentiles for a distribution metric. Defaults to false. Can only be applied to metrics that have a metric_type of distribution.",
				Type:        schema.TypeBool,
				Required:    false,
				Optional:    true,
			},
		},
	}
}

func buildDatadogMetricTagConfiguration(d *schema.ResourceData) (*datadogV2.MetricTagConfigurationCreateData, error) {
	print("HERE2")
	result := datadogV2.NewMetricTagConfigurationCreateDataWithDefaults()
	result.SetId(d.Get("metric_name").(string))

	attributes := datadogV2.NewMetricTagConfigurationCreateAttributesWithDefaults()
	tags := d.Get("tags").(*schema.Set).List()
	var stringTags []string
	for _, tag := range tags {
		stringTags = append(stringTags, tag.(string))
	}
	attributes.SetTags(stringTags)

	metric_type, err := datadogV2.NewMetricTagConfigurationMetricTypesFromValue(d.Get("metric_type").(string))
	if err != nil {
		return nil, utils.TranslateClientError(err, "error building MetricTagConfiguration")
	}
	attributes.SetMetricType(*metric_type)

	include_percentiles, ic_field_set := d.GetOk("include_percentiles")

	if ic_field_set {
		if *metric_type != datadogV2.METRICTAGCONFIGURATIONMETRICTYPES_DISTRIBUTION {
			return nil, fmt.Errorf("include_percentiles field not allowed with metric_type: %s, only with metric_type distribution", *metric_type)
		}
		attributes.SetIncludePercentiles(include_percentiles.(bool))
	}

	result.SetAttributes(*attributes)

	return result, nil
}

func buildDatadogMetricTagConfigurationUpdate(d *schema.ResourceData, existing_metric_type *datadogV2.MetricTagConfigurationMetricTypes) (*datadogV2.MetricTagConfigurationUpdateData, error) {
	print("HERE3")
	result := datadogV2.NewMetricTagConfigurationUpdateDataWithDefaults()
	id := d.Get("metric_name").(string)
	result.SetId(id)

	attributes := datadogV2.NewMetricTagConfigurationUpdateAttributesWithDefaults()
	tags := d.Get("tags").(*schema.Set).List()
	var stringTags []string
	for _, tag := range tags {
		stringTags = append(stringTags, tag.(string))
	}
	attributes.SetTags(stringTags)

	include_percentiles, ic_field_set := d.GetOk("include_percentiles")
	if ic_field_set {
		if *existing_metric_type != datadogV2.METRICTAGCONFIGURATIONMETRICTYPES_DISTRIBUTION {
			return nil, fmt.Errorf("include_percentiles field not allowed with metric_type: %s, only with metric_type distribution", *existing_metric_type)
		}
		attributes.SetIncludePercentiles(include_percentiles.(bool))
	}

	result.SetAttributes(*attributes)
	return result, nil
}

func resourceDatadogMetricTagConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	print("HERE4")
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV2
	auth := providerConf.AuthV2

	resultMetricTagConfigurationData, err := buildDatadogMetricTagConfiguration(d)
	if err != nil {
		return utils.TranslateClientError(err, "error building MetricTagConfiguration object")
	}
	ddObject := datadogV2.NewMetricTagConfigurationCreateRequestWithDefaults()
	ddObject.SetData(*resultMetricTagConfigurationData)
	metric_name := d.Get("metric_name").(string)
	response, _, err := datadogClient.MetricsApi.CreateTagConfiguration(auth, metric_name).Body(*ddObject).Execute()
	if err != nil {
		return utils.TranslateClientError(err, "error creating MetricTagConfiguration")
	}
	d.SetId(metric_name)

	return updateMetricTagConfigurationState(d, response.Data)
}

func updateMetricTagConfigurationState(d *schema.ResourceData, metricTagConfiguration *datadogV2.MetricTagConfiguration) error {
	print("HERE5")
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
	print("HERE6")
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV2
	auth := providerConf.AuthV2

	metric_name := d.Get("metric_name").(string)
	metricTagConfigurationResponse, httpresp, err := datadogClient.MetricsApi.ListTagConfigurationByName(auth, metric_name).Execute()
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
	print("HERE7")
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV2
	auth := providerConf.AuthV2

	metric_name := d.Get("metric_name").(string)
	metricTagConfigurationResponse, _, err := datadogClient.MetricsApi.ListTagConfigurationByName(auth, metric_name).Execute()
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

	response, _, err := datadogClient.MetricsApi.UpdateTagConfiguration(auth, metric_name).Body(*ddObject).Execute()
	if err != nil {
		return utils.TranslateClientError(err, "error updating MetricTagConfiguration")
	}

	return updateMetricTagConfigurationState(d, response.Data)
}

func resourceDatadogMetricTagConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	print("HERE8")
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV2
	auth := providerConf.AuthV2
	var err error

	metric_name := d.Get("metric_name").(string)
	_, err = datadogClient.MetricsApi.DeleteTagConfiguration(auth, metric_name).Execute()

	if err != nil {
		return utils.TranslateClientError(err, "error deleting MetricTagConfiguration")
	}

	return nil
}

func resourceDatadogMetricTagConfigurationImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	print("HERE9")
	if err := resourceDatadogMetricTagConfigurationRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
