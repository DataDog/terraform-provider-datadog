package datadog

import (
	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceDatadogMetricMetadata() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a Datadog metric_metadata resource. This can be used to manage a metric's metadata.",
		Create:      resourceDatadogMetricMetadataCreate,
		Read:        resourceDatadogMetricMetadataRead,
		Update:      resourceDatadogMetricMetadataUpdate,
		Delete:      resourceDatadogMetricMetadataDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogMetricMetadataImport,
		},

		Schema: map[string]*schema.Schema{
			"metric": {
				Description: "The name of the metric.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"type": {
				Description: "Type of the metric.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"description": {
				Description: "A description of the metric.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"short_name": {
				Description: "A short name of the metric.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"unit": {
				Description: "Primary unit of the metric such as `byte` or `operation`.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"per_unit": {
				Description: "Per unit of the metric such as `second` in `bytes per second`.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"statsd_interval": {
				Description: "If applicable, statsd flush interval in seconds for the metric.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
		},
	}
}

func buildMetricMetadataStruct(d *schema.ResourceData) (string, *datadogV1.MetricMetadata) {
	return d.Get("metric").(string), &datadogV1.MetricMetadata{
		Type:           datadogV1.PtrString(d.Get("type").(string)),
		Description:    datadogV1.PtrString(d.Get("description").(string)),
		ShortName:      datadogV1.PtrString(d.Get("short_name").(string)),
		Unit:           datadogV1.PtrString(d.Get("unit").(string)),
		PerUnit:        datadogV1.PtrString(d.Get("per_unit").(string)),
		StatsdInterval: datadogV1.PtrInt64(int64(d.Get("statsd_interval").(int))),
	}
}

func resourceDatadogMetricMetadataCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	id, m := buildMetricMetadataStruct(d)
	_, _, err := datadogClientV1.MetricsApi.UpdateMetricMetadata(authV1, id).Body(*m).Execute()
	if err != nil {
		return translateClientError(err, "error creating metric metadata")
	}

	d.SetId(id)

	return resourceDatadogMetricMetadataRead(d, meta)
}

func resourceDatadogMetricMetadataRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	id := d.Id()

	m, httpresp, err := datadogClientV1.MetricsApi.GetMetricMetadata(authV1, id).Execute()
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return translateClientError(err, "error getting metric metadata")
	}

	d.Set("type", m.GetType())
	d.Set("description", m.GetDescription())
	d.Set("short_name", m.GetShortName())
	d.Set("unit", m.GetUnit())
	d.Set("per_unit", m.GetPerUnit())
	d.Set("statsd_interval", m.GetStatsdInterval())

	return nil
}

func resourceDatadogMetricMetadataUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	m := &datadogV1.MetricMetadata{}
	id := d.Get("metric").(string)

	if attr, ok := d.GetOk("type"); ok {
		m.SetType(attr.(string))
	}
	if attr, ok := d.GetOk("description"); ok {
		m.SetDescription(attr.(string))
	}
	if attr, ok := d.GetOk("short_name"); ok {
		m.SetShortName(attr.(string))
	}
	if attr, ok := d.GetOk("unit"); ok {
		m.SetUnit(attr.(string))
	}
	if attr, ok := d.GetOk("per_unit"); ok {
		m.SetPerUnit(attr.(string))
	}
	if attr, ok := d.GetOk("statsd_interval"); ok {
		m.SetStatsdInterval(int64(attr.(int)))
	}

	if _, _, err := datadogClientV1.MetricsApi.UpdateMetricMetadata(authV1, id).Body(*m).Execute(); err != nil {
		return translateClientError(err, "error updating metric metadata")
	}

	return resourceDatadogMetricMetadataRead(d, meta)
}

func resourceDatadogMetricMetadataDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceDatadogMetricMetadataImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogMetricMetadataRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
