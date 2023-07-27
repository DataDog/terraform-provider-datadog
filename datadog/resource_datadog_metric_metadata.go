package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogMetricMetadata() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog metric_metadata resource. This can be used to manage a metric's metadata.",
		CreateContext: resourceDatadogMetricMetadataCreate,
		ReadContext:   resourceDatadogMetricMetadataRead,
		UpdateContext: resourceDatadogMetricMetadataUpdate,
		DeleteContext: resourceDatadogMetricMetadataDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"metric": {
					Description: "The name of the metric.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"type": {
					Description: "Metric type such as `gauge` or `rate`.",
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
			}
		},
	}
}

func buildMetricMetadataStruct(d *schema.ResourceData) (string, *datadogV1.MetricMetadata) {
	return d.Get("metric").(string), &datadogV1.MetricMetadata{
		Type:           datadog.PtrString(d.Get("type").(string)),
		Description:    datadog.PtrString(d.Get("description").(string)),
		ShortName:      datadog.PtrString(d.Get("short_name").(string)),
		Unit:           datadog.PtrString(d.Get("unit").(string)),
		PerUnit:        datadog.PtrString(d.Get("per_unit").(string)),
		StatsdInterval: datadog.PtrInt64(int64(d.Get("statsd_interval").(int))),
	}
}

func resourceDatadogMetricMetadataCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	id, m := buildMetricMetadataStruct(d)
	createdMetadata, httpResponse, err := apiInstances.GetMetricsApiV1().UpdateMetricMetadata(auth, id, *m)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating metric metadata")
	}
	if err := utils.CheckForUnparsed(createdMetadata); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)

	return updateMetricMetadataState(d, &createdMetadata)
}

func updateMetricMetadataState(d *schema.ResourceData, metadata *datadogV1.MetricMetadata) diag.Diagnostics {
	if err := d.Set("type", metadata.GetType()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", metadata.GetDescription()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("short_name", metadata.GetShortName()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("unit", metadata.GetUnit()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("per_unit", metadata.GetPerUnit()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("statsd_interval", metadata.GetStatsdInterval()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDatadogMetricMetadataRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	id := d.Id()

	m, httpresp, err := apiInstances.GetMetricsApiV1().GetMetricMetadata(auth, id)
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting metric metadata")
	}
	if err := utils.CheckForUnparsed(m); err != nil {
		return diag.FromErr(err)
	}
	return updateMetricMetadataState(d, &m)
}

func resourceDatadogMetricMetadataUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

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

	updatedMetadata, httpResponse, err := apiInstances.GetMetricsApiV1().UpdateMetricMetadata(auth, id, *m)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating metric metadata")
	}
	if err := utils.CheckForUnparsed(updatedMetadata); err != nil {
		return diag.FromErr(err)
	}

	return updateMetricMetadataState(d, &updatedMetadata)
}

func resourceDatadogMetricMetadataDelete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}
