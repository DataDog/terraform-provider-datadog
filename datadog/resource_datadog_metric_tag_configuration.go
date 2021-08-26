package datadog

import (
	"context"
	"fmt"
	"regexp"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceDatadogMetricTagConfiguration() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog metric tag configuration resource. This can be used to modify tag configurations for metrics.",
		CreateContext: resourceDatadogMetricTagConfigurationCreate,
		ReadContext:   resourceDatadogMetricTagConfigurationRead,
		UpdateContext: resourceDatadogMetricTagConfigurationUpdate,
		DeleteContext: resourceDatadogMetricTagConfigurationDelete,
		CustomizeDiff: func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
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
				return fmt.Errorf("error validating diff: %w", err)
			}
			if includePercentilesOk && *metricTypeValidated != datadogV2.METRICTAGCONFIGURATIONMETRICTYPES_DISTRIBUTION {
				return fmt.Errorf("cannot use include_percentiles with a metric_type of %s, must use metric_type of 'distribution'", metricType)
			}
			return nil
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"metric_name": {
				Description:  "The metric name for this resource.",
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`^[A-Za-z][A-Za-z0-9\.\_]*$`), "metric name must be valid"), validation.StringLenBetween(1, 200)),
			},
			"metric_type": {
				Description:      "The metric's type. This field can't be updated after creation.",
				Type:             schema.TypeString,
				ForceNew:         true,
				Required:         true,
				ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewMetricTagConfigurationMetricTypesFromValue),
			},
			"tags": {
				Description: "A list of tag keys that will be queryable for your metric.",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`^[A-Za-z][A-Za-z0-9\.\-\_:\/]*[^:]$`), "tags must be valid"), validation.StringLenBetween(1, 200)),
				},
				Required: true,
			},
			"include_percentiles": {
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
	stringTags := []string{}
	for _, tag := range tags {
		stringTags = append(stringTags, tag.(string))
	}
	attributes.SetTags(stringTags)

	metricType, err := datadogV2.NewMetricTagConfigurationMetricTypesFromValue(d.Get("metric_type").(string))
	if err != nil {
		return nil, fmt.Errorf("error building MetricTagConfiguration: %w", err)
	}
	attributes.SetMetricType(*metricType)

	includePercentiles, iclFieldSet := d.GetOk("include_percentiles")

	if iclFieldSet {
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
	stringTags := []string{}
	for _, tag := range tags {
		stringTags = append(stringTags, tag.(string))
	}
	attributes.SetTags(stringTags)

	includePercentiles, iclFieldSet := d.GetOk("include_percentiles")
	if iclFieldSet {
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

func resourceDatadogMetricTagConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV2
	auth := providerConf.AuthV2

	resultMetricTagConfigurationData, err := buildDatadogMetricTagConfiguration(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error building MetricTagConfiguration object: %s", err.Error()))
	}
	ddObject := datadogV2.NewMetricTagConfigurationCreateRequestWithDefaults()
	ddObject.SetData(*resultMetricTagConfigurationData)
	metricName := d.Get("metric_name").(string)

	response, httpResponse, err := datadogClient.MetricsApi.CreateTagConfiguration(auth, metricName, *ddObject)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating MetricTagConfiguration")
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(metricName)

	return updateMetricTagConfigurationState(d, response.Data)
}

func updateMetricTagConfigurationState(d *schema.ResourceData, metricTagConfiguration *datadogV2.MetricTagConfiguration) diag.Diagnostics {
	if attributes, ok := metricTagConfiguration.GetAttributesOk(); ok {
		if metricType, ok := attributes.GetMetricTypeOk(); ok {
			if err := d.Set("metric_type", metricType); err != nil {
				return diag.FromErr(err)
			}
			if *metricType == datadogV2.METRICTAGCONFIGURATIONMETRICTYPES_DISTRIBUTION {
				if err := d.Set("include_percentiles", attributes.GetIncludePercentiles()); err != nil {
					return diag.FromErr(err)
				}
			}
		}
		tags := attributes.GetTags()
		if tags == nil {
			tags = []string{}
		}
		if err := d.Set("tags", tags); err != nil {
			return diag.FromErr(err)
		}
	}

	metricName := metricTagConfiguration.GetId()
	if err := d.Set("metric_name", metricName); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(metricName)
	// we do not care about the created_at nor modified_at fields

	return nil
}

func resourceDatadogMetricTagConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV2
	auth := providerConf.AuthV2

	metricName := d.Id()
	metricTagConfigurationResponse, httpresp, err := datadogClient.MetricsApi.ListTagConfigurationByName(auth, metricName)
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		if err != nil {
			return utils.TranslateClientErrorDiag(err, httpresp, "metric tag configuration not found")
		}
		return diag.Errorf("error fetching metric tag configuration by name")
	}
	if httpresp.StatusCode != 200 {
		return diag.Errorf("error fetching metric tag configuration by name, unexpected status code %d", httpresp.StatusCode)
	}
	if err := utils.CheckForUnparsed(metricTagConfigurationResponse); err != nil {
		return diag.FromErr(err)
	}

	resource := metricTagConfigurationResponse.GetData()
	return updateMetricTagConfigurationState(d, &resource)
}

func resourceDatadogMetricTagConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV2
	auth := providerConf.AuthV2

	metricName := d.Id()
	metricTagConfigurationResponse, httpresp, err := datadogClient.MetricsApi.ListTagConfigurationByName(auth, metricName)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "metric not found")
	}
	if httpresp == nil {
		return diag.Errorf("error determining if tag configuration for metric exists")
	}
	if httpresp != nil && httpresp.StatusCode == 404 {
		return diag.Errorf("error updating tag configuration for metric, tag configuration does not exist")
	}
	if err := utils.CheckForUnparsed(metricTagConfigurationResponse); err != nil {
		return diag.FromErr(err)
	}

	existingMetricType := metricTagConfigurationResponse.GetData().Attributes.GetMetricType()

	resultMetricTagConfigurationUpdateData, err := buildDatadogMetricTagConfigurationUpdate(d, &existingMetricType)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error building MetricTagConfiguration object")
	}

	ddObject := datadogV2.NewMetricTagConfigurationUpdateRequestWithDefaults()
	ddObject.SetData(*resultMetricTagConfigurationUpdateData)

	response, _, err := datadogClient.MetricsApi.UpdateTagConfiguration(auth, metricName, *ddObject)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error updating MetricTagConfiguration")
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		return diag.FromErr(err)
	}

	return updateMetricTagConfigurationState(d, response.Data)
}

func resourceDatadogMetricTagConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV2
	auth := providerConf.AuthV2
	var err error

	metricName := d.Id()
	httpResponse, err := datadogClient.MetricsApi.DeleteTagConfiguration(auth, metricName)

	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting MetricTagConfiguration")
	}

	return nil
}
