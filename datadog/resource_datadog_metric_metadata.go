package datadog

import (
	"fmt"
	"strings"

	"github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceDatadogMetricMetadata() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogMetricMetadataCreate,
		Read:   resourceDatadogMetricMetadataRead,
		Update: resourceDatadogMetricMetadataUpdate,
		Delete: resourceDatadogMetricMetadataDelete,
		Exists: resourceDatadogMetricMetadataExists,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogMetricMetadataImport,
		},

		Schema: map[string]*schema.Schema{
			"metric": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"short_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"unit": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"per_unit": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"statsd_interval": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

func buildMetricMetadataStruct(d *schema.ResourceData) (string, *datadog.MetricMetadata) {
	return d.Get("metric").(string), &datadog.MetricMetadata{
		Type:           d.Get("type").(string),
		Description:    datadog.PtrString(d.Get("description").(string)),
		ShortName:      datadog.PtrString(d.Get("short_name").(string)),
		Unit:           datadog.PtrString(d.Get("unit").(string)),
		PerUnit:        datadog.PtrString(d.Get("per_unit").(string)),
		StatsdInterval: datadog.PtrInt64(int64(d.Get("statsd_interval").(int))),
	}
}

func resourceDatadogMetricMetadataExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	// Exists - This is called to verify a resource still exists. It is called prior to Read,
	// and lowers the burden of Read to be able to assume the resource exists.
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	metricName, _ := buildMetricMetadataStruct(d)

	if _, _, err := client.MetricsApi.GetMetricMetadata(auth, metricName).Execute(); err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func resourceDatadogMetricMetadataCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	metricName, m := buildMetricMetadataStruct(d)
	_, _, err := client.MetricsApi.EditMetricMetadata(auth, metricName).Body(*m).Execute()
	if err != nil {
		fmt.Println("this si the err:", err)
		return translateClientError(err, "error updating MetricMetadata")
	}

	d.SetId(metricName)

	return resourceDatadogMetricMetadataRead(d, meta)
}

func resourceDatadogMetricMetadataRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	metricName, _ := buildMetricMetadataStruct(d)

	m, _, err := client.MetricsApi.GetMetricMetadata(auth, metricName).Execute()
	if err != nil {
		return translateClientError(err, "error getting MetricMetadata")
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
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	m := &datadog.MetricMetadata{}
	metricName := d.Get("metric").(string)

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

	if _, _, err := client.MetricsApi.EditMetricMetadata(auth, metricName).Body(*m).Execute(); err != nil {
		return translateClientError(err, "error updating MetricMetadata")
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
