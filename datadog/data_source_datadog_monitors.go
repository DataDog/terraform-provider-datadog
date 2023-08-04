package datadog

import (
	"context"
	"fmt"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDatadogMonitors() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to list several existing monitors for use in other resources.",
		ReadContext: dataSourceDatadogMonitorsRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"name_filter": {
					Description: "A monitor name to limit the search.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"tags_filter": {
					Description: "A list of tags to limit the search. This filters on the monitor scope.",
					Type:        schema.TypeList,
					Optional:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"monitor_tags_filter": {
					Description: "A list of monitor tags to limit the search. This filters on the tags set on the monitor itself.",
					Type:        schema.TypeList,
					Optional:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},

				// Computed values
				"monitors": {
					Description: "List of monitors",
					Type:        schema.TypeList,
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Description: "ID of the monitor",
								Type:        schema.TypeInt,
								Computed:    true,
							},
							"name": {
								Description: "Name of the monitor",
								Type:        schema.TypeString,
								Computed:    true,
							},
							"type": {
								Description: "Type of the monitor.",
								Type:        schema.TypeString,
								Computed:    true,
							},
						},
					},
				},
			}
		},
	}
}

func dataSourceDatadogMonitorsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	optionalParams := datadogV1.NewListMonitorsOptionalParameters()
	if v, ok := d.GetOk("name_filter"); ok {
		optionalParams = optionalParams.WithName(v.(string))
	}
	if v, ok := d.GetOk("tags_filter"); ok {
		optionalParams = optionalParams.WithTags(strings.Join(expandStringList(v.([]interface{})), ","))
	}
	if v, ok := d.GetOk("monitor_tags_filter"); ok {
		optionalParams = optionalParams.WithMonitorTags(strings.Join(expandStringList(v.([]interface{})), ","))
	}

	monitors, httpresp, err := apiInstances.GetMonitorsApiV1().ListMonitors(auth, *optionalParams)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error querying monitors")
	}
	if len(monitors) == 0 {
		return diag.Errorf("your query returned no result, please try a less specific search criteria")
	}

	d.SetId(computeMonitorsDatasourceID(d))

	diags := diag.Diagnostics{}
	tfMonitors := make([]map[string]interface{}, len(monitors))
	for i, m := range monitors {
		if err := utils.CheckForUnparsed(m); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("skipping monitor with id: %v", m.GetId()),
				Detail:   fmt.Sprintf("monitor contains unparsed object: %v", err),
			})
			continue
		}

		tfMonitors[i] = map[string]interface{}{
			"id":   m.GetId(),
			"name": m.GetName(),
			"type": m.GetType(),
		}
	}
	if err := d.Set("monitors", tfMonitors); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func computeMonitorsDatasourceID(d *schema.ResourceData) string {
	var dsID strings.Builder
	if v, ok := d.GetOk("name_filter"); ok {
		dsID.WriteString(v.(string))
	}
	dsID.WriteRune('|')
	if v, ok := d.GetOk("tags_filter"); ok {
		dsID.WriteString(strings.Join(expandStringList(v.([]interface{})), ","))
	}
	dsID.WriteRune('|')
	if v, ok := d.GetOk("monitor_tags_filter"); ok {
		dsID.WriteString(strings.Join(expandStringList(v.([]interface{})), ","))
	}
	return dsID.String()
}
