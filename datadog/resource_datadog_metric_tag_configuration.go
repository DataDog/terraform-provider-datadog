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
		CustomizeDiff: func(diff *schema.ResourceDiff, meta interface{}) error {
			_, includePercentilesOk := diff.GetOkExists("include_percentiles")
			if !includePercentilesOk {
				// if there was no change to include_percentiles we don't need special handling
				return nil
			}

			metricType, ok := diff.GetOkExists("metric_type")
			if !ok {
				// no change to metric_type so no special handling
				return nil
			}
			metricTypeValidated, err := datadogV2.NewMetricTagConfigurationMetricTypesFromValue(metricType.(string))
			if err != nil {
				return utils.TranslateClientError(err, "error validating diff")
			}
			if includePercentilesOk && *metricTypeValidated != datadogV2.METRICTAGCONFIGURATIONMETRICTYPES_DISTRIBUTION {
				return fmt.Errorf("Cannot use include_percentiles with a metric_type of %s, must use metric_type of 'distribution'.", metricType)
			}
			return nil
		},
		Importer: &schema.ResourceImporter{
			State: resourceDatadogMetricTagConfigurationImport,
		},

		Schema: map[string]*schema.Schema{
			"metric_name": {
				Description: "The metric name for this resource.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				ValidateFunc: func(val interface{}, k string) (warns []string, errs []error) {
					v := val.(string)
					re := regexp.MustCompile(`^[A-Za-z][A-Za-z0-9\\.\\_]*$`)
					if len(v) < 1 || len(v) > 200 {
						errs = append(errs, fmt.Errorf("expected metric name length of %s to be in the range (%d - %d), got %s", k, 1, 200, v))
					}
					if !re.MatchString(v) {
						errs = append(errs, fmt.Errorf("metric name not allowed"))
					}
					// todo[efraese] ensure metric name is not a standard metric or should this be done in backend only?
					return
				},
			},
			"metric_type": {
				Description:  "The metric's type. This field can't be updated after creation. Allowed enum values: gauge,count,distribution",
				Type:         schema.TypeString,
				ForceNew:     true,
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
						re := regexp.MustCompile(`^[A-Za-z][A-Za-z0-9\\.\\-\\_:\\/]*$`)
						if len(v) < 1 || len(v) > 200 {
							errs = append(errs, fmt.Errorf("expected tag length of %s to be in the range (%d - %d), got %s", k, 1, 200, v))
						}
						if strings.HasSuffix(v, ":") {
							errs = append(errs, fmt.Errorf("tag ends in : which is not allowed"))
						}
						if !re.MatchString(v) {
							errs = append(errs, fmt.Errorf("tag not allowed"))
						}
						return
					},
				},
				Required: true,
			},
			"include_percentiles": {
				// TODO[efraese] fix schema to only allow this field when the metric type is a distribution (done via build funcs but still can show in a plan)
				Description: "Toggle to include/exclude percentiles for a distribution metric. Defaults to false. Can only be applied to metrics that have a metric_type of distribution.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
		},
	}
}

func buildDatadogMetricTagConfiguration(d *schema.ResourceData) (*datadogV2.MetricTagConfigurationCreateData, error) {
	result := datadogV2.NewMetricTagConfigurationCreateDataWithDefaults()
	result.SetId(d.Get("metric_name").(string))

	attributes := datadogV2.NewMetricTagConfigurationCreateAttributesWithDefaults()
	tags := d.Get("tags").(*schema.Set).List()
	var stringTags []string
	for _, tag := range tags {
		stringTags = append(stringTags, tag.(string))
	}
	attributes.SetTags(stringTags)

	metricType, err := datadogV2.NewMetricTagConfigurationMetricTypesFromValue(d.Get("metric_type").(string))
	if err != nil {
		return nil, utils.TranslateClientError(err, "error building MetricTagConfiguration")
	}
	attributes.SetMetricType(*metricType)

	includePercentiles, icFieldSet := d.GetOk("include_percentiles")

	if icFieldSet {
		if *metricType != datadogV2.METRICTAGCONFIGURATIONMETRICTYPES_DISTRIBUTION {
			return nil, fmt.Errorf("include_percentiles field not allowed with metric_type: %s, only with metric_type distribution", *metricType)
		}
		attributes.SetIncludePercentiles(includePercentiles.(bool))
	} else {
		// if the include_percentiles field is not set and the metric is not a distribution, we need to remove the include_percentiles field from the payload
		if *metricType != datadogV2.METRICTAGCONFIGURATIONMETRICTYPES_DISTRIBUTION {
			attributes.IncludePercentiles = nil
		}
	}
	result.SetAttributes(*attributes)

	return result, nil
}

func buildDatadogMetricTagConfigurationUpdate(d *schema.ResourceData, existingMetricType *datadogV2.MetricTagConfigurationMetricTypes) (*datadogV2.MetricTagConfigurationUpdateData, error) {
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

	includePercentiles, icFieldSet := d.GetOk("include_percentiles")
	if icFieldSet {
		if *existingMetricType != datadogV2.METRICTAGCONFIGURATIONMETRICTYPES_DISTRIBUTION {
			return nil, fmt.Errorf("include_percentiles field not allowed with metric_type: %s, only with metric_type distribution", *existingMetricType)
		}
		attributes.SetIncludePercentiles(includePercentiles.(bool))
	} else {
		// if the include_percentiles field is not set and the metric is not a distribution, we need to remove the include_percentiles field from the payload
		if *existingMetricType != datadogV2.METRICTAGCONFIGURATIONMETRICTYPES_DISTRIBUTION {
			attributes.IncludePercentiles = nil
		}
	}

	result.SetAttributes(*attributes)
	return result, nil
}

func resourceDatadogMetricTagConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV2
	auth := providerConf.AuthV2

	resultMetricTagConfigurationData, err := buildDatadogMetricTagConfiguration(d)
	if err != nil {
		return utils.TranslateClientError(err, "error building MetricTagConfiguration object")
	}
	ddObject := datadogV2.NewMetricTagConfigurationCreateRequestWithDefaults()
	ddObject.SetData(*resultMetricTagConfigurationData)
	metricName := d.Get("metric_name").(string)

	// check if the tag configuration already exists, if so return an error
	_, httpresp, err := datadogClient.MetricsApi.ListTagConfigurationByName(auth, metricName).Execute()
	if err != nil || httpresp == nil {
		if httpresp != nil && httpresp.StatusCode != 404 {
			return utils.TranslateClientError(err, "could not determine if metric already exists")
		}
		if httpresp == nil {
			return fmt.Errorf("error creating MetricTagConfiguration: could not determine if metric already exists")
		}
		// if neither of these cases hit is it ok because the api will return 404 when we can create a tag-configuration
	}
	if httpresp != nil && httpresp.StatusCode == 200 {
		return fmt.Errorf("error creating MetricTagConfiguration: a tag configuration already exists for metric, import it first")
	}

	response, _, err := datadogClient.MetricsApi.CreateTagConfiguration(auth, metricName).Body(*ddObject).Execute()
	if err != nil {
		return utils.TranslateClientError(err, "error creating MetricTagConfiguration")
	}
	d.SetId(metricName)

	return updateMetricTagConfigurationState(d, response.Data)
}

func updateMetricTagConfigurationState(d *schema.ResourceData, metricTagConfiguration *datadogV2.MetricTagConfiguration) error {
	if attributes, ok := metricTagConfiguration.GetAttributesOk(); ok {
		if metricType, ok := attributes.GetMetricTypeOk(); ok {
			if err := d.Set("metric_type", metricType); err != nil {
				return err
			}
			if *metricType == datadogV2.METRICTAGCONFIGURATIONMETRICTYPES_DISTRIBUTION {
				if err := d.Set("include_percentiles", attributes.GetIncludePercentiles()); err != nil {
					return err
				}
			}
		}
		if err := d.Set("tags", attributes.GetTags()); err != nil {
			return err
		}
	}

	metricName := metricTagConfiguration.GetId()
	if err := d.Set("metric_name", metricName); err != nil {
		return err
	}

	d.SetId(metricName)
	// we do not care about the created_at nor modified_at fields

	return nil
}

func resourceDatadogMetricTagConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV2
	auth := providerConf.AuthV2

	metricName := d.Id()
	metricTagConfigurationResponse, httpresp, err := datadogClient.MetricsApi.ListTagConfigurationByName(auth, metricName).Execute()
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

	metricName := d.Id()
	metricTagConfigurationResponse, _, err := datadogClient.MetricsApi.ListTagConfigurationByName(auth, metricName).Execute()
	if err != nil {
		return utils.TranslateClientError(err, "metric not found")
	}

	existingMetricType := metricTagConfigurationResponse.GetData().Attributes.GetMetricType()

	resultMetricTagConfigurationUpdateData, err := buildDatadogMetricTagConfigurationUpdate(d, &existingMetricType)
	if err != nil {
		return utils.TranslateClientError(err, "error building MetricTagConfiguration object")
	}

	ddObject := datadogV2.NewMetricTagConfigurationUpdateRequestWithDefaults()
	ddObject.SetData(*resultMetricTagConfigurationUpdateData)

	response, _, err := datadogClient.MetricsApi.UpdateTagConfiguration(auth, metricName).Body(*ddObject).Execute()
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

	metricName := d.Id()
	_, err = datadogClient.MetricsApi.DeleteTagConfiguration(auth, metricName).Execute()

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
