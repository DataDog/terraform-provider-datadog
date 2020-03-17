package datadog

import (
	"fmt"
	"strings"

	"github.com/zorkian/go-datadog-api"

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
		Type:           datadog.String(d.Get("type").(string)),
		Description:    datadog.String(d.Get("description").(string)),
		ShortName:      datadog.String(d.Get("short_name").(string)),
		Unit:           datadog.String(d.Get("unit").(string)),
		PerUnit:        datadog.String(d.Get("per_unit").(string)),
		StatsdInterval: datadog.Int(d.Get("statsd_interval").(int)),
	}
}

func resourceDatadogMetricMetadataExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	// Exists - This is called to verify a resource still exists. It is called prior to Read,
	// and lowers the burden of Read to be able to assume the resource exists.
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	id, _ := buildMetricMetadataStruct(d)

	if _, err := client.ViewMetricMetadata(id); err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return false, nil
		}
		return false, translateClientError(err, "error checking metric metadata exists")
	}

	return true, nil
}

func resourceDatadogMetricMetadataCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	id, m := buildMetricMetadataStruct(d)
	_, err := client.EditMetricMetadata(id, m)
	if err != nil {
		return translateClientError(err, "error creating metric metadata")
	}

	d.SetId(id)

	return resourceDatadogMetricMetadataRead(d, meta)
}

func resourceDatadogMetricMetadataRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	id, _ := buildMetricMetadataStruct(d)

	m, err := client.ViewMetricMetadata(id)
	if err != nil {
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
	client := providerConf.CommunityClient

	m := &datadog.MetricMetadata{}
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
		m.SetStatsdInterval(attr.(int))
	}

	if _, err := client.EditMetricMetadata(id, m); err != nil {
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
