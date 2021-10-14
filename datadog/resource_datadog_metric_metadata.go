package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
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

func resourceDatadogMetricMetadataCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	id, m := buildMetricMetadataStruct(d)
	createdMetadata, httpResponse, err := datadogClientV1.MetricsApi.UpdateMetricMetadata(authV1, id, *m)
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
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	id := d.Id()

	m, httpresp, err := datadogClientV1.MetricsApi.GetMetricMetadata(authV1, id)
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

	updatedMetadata, httpResponse, err := datadogClientV1.MetricsApi.UpdateMetricMetadata(authV1, id, *m)
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
