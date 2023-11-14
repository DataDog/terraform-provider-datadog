package datadog

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func resourceDatadogPowerpack() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog powerpack resource. This can be used to create and manage Datadog powerpacks.",
		CreateContext: resourceDatadogPowerpackCreate,
		UpdateContext: resourceDatadogPowerpackUpdate,
		ReadContext:   resourceDatadogPowerpackRead,
		DeleteContext: resourceDatadogPowerpackDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"description": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The description of the powerpack.",
				},
				"live_span": getWidgetLiveSpanSchema(),
				"name": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The name for the powerpack.",
				},
				"show_title": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Whether or not title should be displayed in the powerpack.",
				},
				"tags": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "List of tags to identify this powerpack.",
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"template_variables": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "The list of template variables for this powerpack.",
					Elem: &schema.Resource{
						Schema: getPowerpackTemplateVariableSchema(),
					},
				},
				"widget": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "The list of widgets to display in the powerpack.",
					Elem: &schema.Resource{
						Schema: getPowerpackWidgetSchema(),
					},
				},
				"layout": {
					Type:        schema.TypeList,
					MaxItems:    1,
					Optional:    true,
					Description: "The layout of the powerpack on a free-form dashboard.",
					Elem: &schema.Resource{
						Schema: getWidgetLayoutSchema(),
					},
				},
			}
		},
	}
}

func getPowerpackWidgetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"widget_layout": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Description: "The layout of the widget on a 'free' dashboard.",
			Elem: &schema.Resource{
				Schema: getWidgetLayoutSchema(),
			},
		},
		"id": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The ID of the widget.",
		},
		// A widget should implement exactly one of the following definitions
		"alert_graph_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Alert Graph widget.",
			Elem: &schema.Resource{
				Schema: getAlertGraphDefinitionSchema(),
			},
		},
		"alert_value_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Alert Value widget.",
			Elem: &schema.Resource{
				Schema: getAlertValueDefinitionSchema(),
			},
		},
		"change_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Change widget.",
			Elem: &schema.Resource{
				Schema: getChangeDefinitionSchema(),
			},
		},
		"check_status_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Check Status widget.",
			Elem: &schema.Resource{
				Schema: getCheckStatusDefinitionSchema(),
			},
		},
		"distribution_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Distribution widget.",
			Elem: &schema.Resource{
				Schema: getDistributionDefinitionSchema(),
			},
		},
		"event_stream_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Event Stream widget.",
			Elem: &schema.Resource{
				Schema: getEventStreamDefinitionSchema(),
			},
		},
		"event_timeline_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Event Timeline widget.",
			Elem: &schema.Resource{
				Schema: getEventTimelineDefinitionSchema(),
			},
		},
		"free_text_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Free Text widget.",
			Elem: &schema.Resource{
				Schema: getFreeTextDefinitionSchema(),
			},
		},
		"heatmap_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Heatmap widget.",
			Elem: &schema.Resource{
				Schema: getHeatmapDefinitionSchema(),
			},
		},
		"hostmap_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Hostmap widget.",
			Elem: &schema.Resource{
				Schema: getHostmapDefinitionSchema(),
			},
		},
		"iframe_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for an Iframe widget.",
			Elem: &schema.Resource{
				Schema: getIframeDefinitionSchema(),
			},
		},
		"image_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for an Image widget",
			Elem: &schema.Resource{
				Schema: getImageDefinitionSchema(),
			},
		},
		"list_stream_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a List Stream widget.",
			Elem: &schema.Resource{
				Schema: getListStreamDefinitionSchema(),
			},
		},
		"log_stream_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for an Log Stream widget.",
			Elem: &schema.Resource{
				Schema: getLogStreamDefinitionSchema(),
			},
		},
		"manage_status_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for an Manage Status widget.",
			Elem: &schema.Resource{
				Schema: getManageStatusDefinitionSchema(),
			},
		},
		"note_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Note widget.",
			Elem: &schema.Resource{
				Schema: getNoteDefinitionSchema(),
			},
		},
		"query_value_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Query Value widget.",
			Elem: &schema.Resource{
				Schema: getQueryValueDefinitionSchema(),
			},
		},
		// "query_table_definition": {
		// 	Type:        schema.TypeList,
		// 	Optional:    true,
		// 	MaxItems:    1,
		// 	Description: "The definition for a Query Table widget.",
		// 	Elem: &schema.Resource{
		// 		Schema: getQueryTableDefinitionSchema(),
		// 	},
		// },
		"scatterplot_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Scatterplot widget.",
			Elem: &schema.Resource{
				Schema: getScatterplotDefinitionSchema(),
			},
		},
		"servicemap_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Service Map widget.",
			Elem: &schema.Resource{
				Schema: getServiceMapDefinitionSchema(),
			},
		},
		"service_level_objective_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Service Level Objective widget.",
			Elem: &schema.Resource{
				Schema: getServiceLevelObjectiveDefinitionSchema(),
			},
		},
		// "slo_list_definition": {
		// 	Type:        schema.TypeList,
		// 	Optional:    true,
		// 	MaxItems:    1,
		// 	Description: "The definition for an SLO (Service Level Objective) List widget.",
		// 	Elem: &schema.Resource{
		// 		Schema: getSloListDefinitionSchema(),
		// 	},
		// },
		// "sunburst_definition": {
		// 	Type:        schema.TypeList,
		// 	Optional:    true,
		// 	MaxItems:    1,
		// 	Description: "The definition for a Sunburst widget.",
		// 	Elem: &schema.Resource{
		// 		Schema: getSunburstDefinitionschema(),
		// 	},
		// },
		// "timeseries_definition": {
		// 	Type:        schema.TypeList,
		// 	Optional:    true,
		// 	MaxItems:    1,
		// 	Description: "The definition for a Timeseries widget.",
		// 	Elem: &schema.Resource{
		// 		Schema: getTimeseriesDefinitionSchema(),
		// 	},
		// },
		"toplist_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Toplist widget.",
			Elem: &schema.Resource{
				Schema: getToplistDefinitionSchema(),
			},
		},
		"topology_map_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Topology Map widget.",
			Elem: &schema.Resource{
				Schema: getTopologyMapDefinitionSchema(),
			},
		},
		"trace_service_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Trace Service widget.",
			Elem: &schema.Resource{
				Schema: getTraceServiceDefinitionSchema(),
			},
		},
		// "treemap_definition": {
		// 	Type:        schema.TypeList,
		// 	Optional:    true,
		// 	MaxItems:    1,
		// 	Description: "The definition for a Treemap widget.",
		// 	Elem: &schema.Resource{
		// 		Schema: getTreemapDefinitionSchema(),
		// 	},
		// },
		// "geomap_definition": {
		// 	Type:        schema.TypeList,
		// 	Optional:    true,
		// 	MaxItems:    1,
		// 	Description: "The definition for a Geomap widget.",
		// 	Elem: &schema.Resource{
		// 		Schema: getGeomapDefinitionSchema(),
		// 	},
		// },
		"run_workflow_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Run Workflow widget.",
			Elem: &schema.Resource{
				Schema: getRunWorkflowDefinitionSchema(),
			},
		},
	}
}

func buildPowerpackTemplateVariables(terraformTemplateVariables []interface{}) *[]datadogV2.PowerpackTemplateVariable {
	ppkTemplateVariables := make([]datadogV2.PowerpackTemplateVariable, len(terraformTemplateVariables))
	for i, ttv := range terraformTemplateVariables {
		if ttv == nil {
			continue
		}
		terraformTemplateVariable := ttv.(map[string]interface{})
		var ppkTemplateVariable datadogV2.PowerpackTemplateVariable
		if v, ok := terraformTemplateVariable["name"].(string); ok && len(v) != 0 {
			ppkTemplateVariable.SetName(v)
		}
		if v, ok := terraformTemplateVariable["defaults"].([]interface{}); ok && len(v) != 0 {
			var defaults []string
			for _, s := range v {
				defaults = append(defaults, s.(string))
			}
			ppkTemplateVariable.SetDefaults(defaults)
		}
		ppkTemplateVariables[i] = ppkTemplateVariable
	}
	return &ppkTemplateVariables
}

func buildPowerpackTerraformTemplateVariables(powerpackTemplateVariables []datadogV2.PowerpackTemplateVariable) *[]map[string]interface{} {
	terraformTemplateVariables := make([]map[string]interface{}, len(powerpackTemplateVariables))
	for i, templateVariable := range powerpackTemplateVariables {
		terraformTemplateVariable := map[string]interface{}{}
		if v, ok := templateVariable.GetNameOk(); ok {
			terraformTemplateVariable["name"] = *v
		}
		if v, ok := templateVariable.GetDefaultsOk(); ok && len(*v) > 0 {
			var tags []string
			tags = append(tags, *v...)
			terraformTemplateVariable["defaults"] = tags
		}
		terraformTemplateVariables[i] = terraformTemplateVariable
	}
	return &terraformTemplateVariables
}
func getPowerpackTemplateVariableSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the powerpack template variable.",
		},
		"defaults": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			Description: "One or many default values for powerpack template variables on load. If more than one default is specified, they will be unioned together with `OR`.",
		},
	}
}

func resourceDatadogPowerpackCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	powerpackPayload, diags := buildDatadogPowerpack(ctx, d)
	if diags.HasError() {
		return diags
	}
	powerpack, httpresp, err := apiInstances.GetPowerpackApiV2().CreatePowerpack(auth, *powerpackPayload)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error creating powerpack")
	}
	if err := utils.CheckForUnparsed(powerpack); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(*powerpack.Data.Id)

	var getPowerpackResponse datadogV2.PowerpackResponse
	var httpResponse *http.Response

	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		getPowerpackResponse, httpResponse, err = apiInstances.GetPowerpackApiV2().GetPowerpack(auth, *powerpack.Data.Id)

		if err != nil {
			if httpResponse != nil {
				return retry.RetryableError(fmt.Errorf("powerpack not created yet"))
			}
			return retry.NonRetryableError(err)
		}

		if err := utils.CheckForUnparsed(getPowerpackResponse); err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return updatePowerpackState(d, &getPowerpackResponse)
}

func resourceDatadogPowerpackUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	id := d.Id()
	powerpack, diags := buildDatadogPowerpack(ctx, d)
	if diags.HasError() {
		return diags
	}

	updatedPowerpackResponse, httpResponse, err := apiInstances.GetPowerpackApiV2().UpdatePowerpack(auth, id, *powerpack)
	if err != nil {
		if httpResponse != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("error updating powerpack: %s", err),
			})
			return diags
		}
	}
	if err := utils.CheckForUnparsed(updatedPowerpackResponse); err != nil {
		return diag.FromErr(err)
	}
	return updatePowerpackState(d, &updatedPowerpackResponse)
}

func resourceDatadogPowerpackRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	id := d.Id()
	powerpack, httpResponse, err := apiInstances.GetPowerpackApiV2().GetPowerpack(auth, id)
	if err != nil {
		if httpResponse != nil {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpResponse, "error getting powerpack")
	}
	if err := utils.CheckForUnparsed(powerpack); err != nil {
		return diag.FromErr(err)
	}

	return updatePowerpackState(d, &powerpack)
}

func validatePowerpackGroupWidgetLayout(layout map[string]interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	height := int64(layout["height"].(int))
	width := int64(layout["width"].(int))
	x := int64(layout["x"].(int))
	y := int64(layout["y"].(int))

	layoutDict := map[string]interface{}{
		"height": height,
		"width":  width,
		"x":      x,
		"y":      y,
	}

	for _, v := range []string{"height", "width"} {
		if layoutDict[v].(int64) < 1 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("powerpack layout contains an invalid value. %s must be greater than 0", v),
			})
		}
	}

	for _, v := range []string{"x", "y"} {
		if layoutDict[v].(int64) < 0 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("powerpack layout contains an invalid value. %s must be 0 or greater", v),
			})
		}
	}

	if width+x > 12 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("powerpack layout contains an invalid value. sum of x and width is greater than the maximum of 12."),
		})
	}

	return diags
}

func buildDatadogPowerpack(ctx context.Context, d *schema.ResourceData) (*datadogV2.Powerpack, diag.Diagnostics) {
	var diags diag.Diagnostics
	attributes := datadogV2.NewPowerpackAttributesWithDefaults()

	// Set Description
	if v, ok := d.GetOk("description"); ok {
		attributes.SetDescription(v.(string))
	}

	// Set Name
	if v, ok := d.GetOk("name"); ok {
		attributes.SetName(v.(string))
	}

	// Set Tags
	if v, ok := d.GetOk("tags"); ok {
		tags := make([]string, v.(*schema.Set).Len())
		for i, tag := range v.(*schema.Set).List() {
			tags[i] = tag.(string)
		}
		attributes.SetTags(tags)
	} else {
		attributes.SetTags([]string{})
	}

	// Set TemplateVariables
	if v, ok := d.GetOk("template_variables"); ok {
		templateVariables := *buildPowerpackTemplateVariables(v.([]interface{}))
		attributes.SetTemplateVariables(templateVariables)
	} else {
		attributes.SetTemplateVariables(*buildPowerpackTemplateVariables([]interface{}{}))
	}

	// Create group widget object
	var groupWidget datadogV2.PowerpackGroupWidget

	var definition datadogV2.PowerpackGroupWidgetDefinition

	// Group Widget type and layout type should always be set to the following values
	definition.SetLayoutType("ordered")
	definition.SetType("group")

	// User configurable properties defined in the group widget
	if v, ok := d.GetOk("show_title"); ok {
		definition.SetShowTitle(v.(bool))
	}

	// Note: The Powerpack name is the group title.
	if v, ok := d.GetOk("name"); ok {
		definition.SetTitle(v.(string))
	}

	// Fetch widgets in the request form
	requestWidgets := d.Get("widget").([]interface{})
	// Convert and validate them using the Dashboard widget type
	datadogWidgets, err := buildDatadogWidgets(&requestWidgets)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("error constructing widgets: %s", err),
		})
	}
	// Convert to TF widget type for easier parsing
	terraformWidgets, err := buildTerraformWidgets(datadogWidgets, d)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("error constructing widgets: %s", err),
		})
	}

	var columnWidth int64
	if v, ok := d.GetOk("layout"); ok {
		unparsedLayout := v.([]interface{})[0].(map[string]interface{})
		diags := validatePowerpackGroupWidgetLayout(unparsedLayout)
		if diags.HasError() {
			return nil, diags
		}

		columnWidth = int64(unparsedLayout["width"].(int))
		layout := datadogV2.NewPowerpackGroupWidgetLayout(
			int64(unparsedLayout["height"].(int)),
			columnWidth,
			int64(unparsedLayout["x"].(int)),
			int64(unparsedLayout["y"].(int)))
		groupWidget.SetLayout(*layout)
	} else {
		// Temporary fix: set a reasonable default layout value for the layout property
		columnWidth = 12
		groupWidget.Layout = datadogV2.NewPowerpackGroupWidgetLayout(1, 12, 0, 0)
	}

	// Finally, build JSON Powerpack API compatible widgets
	powerpackWidgets, diags := dashboardWidgetsToPpkWidgets(terraformWidgets, columnWidth)

	if diags != nil {
		return nil, diags
	}

	// Set Widget
	definition.SetWidgets(powerpackWidgets)

	groupWidget.Definition = definition

	// Set Live span for all powerpack widgets.
	if v, ok := d.GetOk("live_span"); ok {
		liveSpan, err := datadogV2.NewWidgetLiveSpanFromValue(v.(string))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("live_span is invalid: %s", v.(string)),
			})
			return nil, diags
		}
		groupWidget.LiveSpan = liveSpan
	}

	attributes.GroupWidget = groupWidget

	req := datadogV2.NewPowerpackWithDefaults()
	req.Data = datadogV2.NewPowerpackDataWithDefaults()
	// Set type to powerpack, which is the only acceptable value for a powerpack request
	req.Data.SetType("powerpack")

	req.Data.SetAttributes(*attributes)

	return req, diags

}

func normalizeWidgetDefRequests(widgetDefRequests []map[string]interface{}, widgetType string) []map[string]interface{} {
	normalizedWidgetDefRequests := widgetDefRequests
	for i, widgetDefRequest := range normalizedWidgetDefRequests {
		for _, v := range []string{"style", "apm_stats_query", "x", "y"} {
			// Properties listed above are always a single value, not a list
			if widgetDefRequest[v] != nil {
				widgetDefRequest[v] = widgetDefRequest[v].([]map[string]interface{})[0]
			}
		}
		for _, v := range []string{"formula"} {
			// Properties listed above are defined as single, but API Spec expects a plural name
			if widgetDefRequest[v] != nil {
				widgetDefRequest[v+"s"] = widgetDefRequest[v]
				delete(widgetDefRequest, v)
			}
		}
		if widgetType != "topology_map" && widgetDefRequest["query"] != nil {
			widgetDefRequest["query"] = widgetDefRequest["query"].([]map[string]interface{})[0]
		}
		normalizedWidgetDefRequests[i] = widgetDefRequest
	}
	return normalizedWidgetDefRequests
}

func normalizeDashboardWidgetDef(widgetDef map[string]interface{}, columnWidth int64) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	// SLO widgets are spelled out in dashboard widgets, but API expects "slo" as the widget type
	if widgetDef["type"] == "service_level_objective" {
		widgetDef["type"] = "slo"
	}
	// Dashboard widgets set live span at the widget level, we don't allow that for powerpack widgets
	// where live span is set at the resource level.
	if widgetDef["live_span"] != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("live_span must be set for all powerpack resources (to be applied to all widgets within the powerpack)"),
		})
		return nil, diags
	}

	if widgetDef["request"] != nil {
		if widgetDef["type"] == "scatterplot" || widgetDef["type"] == "hostmap" {
			// Scatterplot and hostmap widgets expect a single requests object instead of a list
			widgetDefRequest := widgetDef["request"].([]map[string]interface{})[0]
			if len(widgetDefRequest) == 0 {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("request should be defined for widget: %s", widgetDef["type"]),
				})
				return nil, diags
			}
			for _, v := range []string{"fill", "size", "x", "y"} {
				// Properties listed above are always a single value, not a list
				if widgetDefRequest[v] != nil {
					widgetDefRequest[v] = widgetDefRequest[v].([]map[string]interface{})[0]
				}
			}
			widgetDef["requests"] = widgetDefRequest
		} else {
			castWidgetDefReq := *widgetDef["request"].(*[]map[string]interface{})
			if len(castWidgetDefReq) == 0 {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("at least one request should be defined for widget: %s", widgetDef["type"]),
				})
				return nil, diags
			}
			// Distribution/change/heatmap widgets have a "requests" field, while API Spec has a "request" field
			// Here we set the "requests" field and remove "request"
			widgetDef["requests"] = normalizeWidgetDefRequests(castWidgetDefReq, widgetDef["type"].(string))
		}
		delete(widgetDef, "request")
	}
	for _, v := range []string{"style", "xaxis", "yaxis"} {
		// Properties listed above are always a single value, not a list
		if widgetDef[v] != nil {
			widgetDef[v] = widgetDef[v].([]map[string]interface{})[0]
		}
	}
	for _, v := range []string{"event", "custom_link", "input"} {
		// Properties listed above are defined as single, but API Spec expects a plural name
		if widgetDef[v] != nil {
			widgetDef[v+"s"] = widgetDef[v]
			delete(widgetDef, v)
		}
	}
	if widgetDef["type"] == "log_stream" && widgetDef["sort"] != nil {
		// For log stream only, sort is not a string value but an object
		// that is a single value
		widgetDef["sort"] = widgetDef["sort"].([]map[string]interface{})[0]
	}
	return widgetDef, diags
}

func normalizeTerraformWidgetDef(widgetDef map[string]interface{}) map[string]interface{} {
	// Dashboard TF widgets have a "requests" field, while API Spec has a "request" field
	// Here we set the "request" field and remove "requests"
	if widgetDef["requests"] != nil {
		if widgetDef["type"] == "scatterplot" || widgetDef["type"] == "hostmap" {
			widgetDefRequest := widgetDef["requests"].(map[string]interface{})
			for _, v := range []string{"fill", "size", "x", "y"} {
				// Properties listed above need to be converted from single values in the API to plural values for TF
				if widgetDefRequest[v] != nil {
					widgetDefRequest[v] = []interface{}{widgetDefRequest[v].(interface{})}
				}
			}
			widgetDef["request"] = []interface{}{widgetDefRequest}
		} else {
			widgetDefRequests := widgetDef["requests"].([]interface{})
			for i, widgetDefRequest := range widgetDefRequests {
				widgetDefRequestNormalized := widgetDefRequest.(map[string]interface{})
				for _, v := range []string{"style", "apm_stats_query"} {
					// Properties listed above need to be converted from single values in the API to plural values for TF
					if widgetDefRequestNormalized[v] != nil {
						widgetDefRequestNormalized[v] = []interface{}{widgetDefRequestNormalized[v].(interface{})}
					}
				}
				if widgetDefRequestNormalized["limit"] != nil {
					// Dashboard widget can't typecast the float64 limit value so we need to convert it first
					widgetDefRequestNormalized["limit"] = int(widgetDefRequestNormalized["limit"].(float64))
				}
				if widgetDef["type"] != "topology_map" && widgetDefRequestNormalized["query"] != nil {
					widgetDefRequestNormalized["query"] = []interface{}{widgetDefRequestNormalized["query"].(interface{})}
				}
				widgetDefRequests[i] = widgetDefRequestNormalized
			}
			widgetDef["request"] = widgetDefRequests
		}
		delete(widgetDef, "requests")
	}
	for _, v := range []string{"style", "xaxis", "yaxis", "y"} {
		// Properties listed above need to be converted from single values in the API to plural values for TF
		if widgetDef[v] != nil {
			widgetDef[v] = []interface{}{widgetDef[v].(map[string]interface{})}
		}
	}
	for _, v := range []string{"custom_links", "inputs", "events"} {
		// Properties listed above are defined as plural for the API Spec and need to be converted back to single values for TF
		if widgetDef[v] != nil {
			widgetDef[v[:len(v)-1]] = widgetDef[v]
			delete(widgetDef, v)
		}
	}
	if widgetDef["type"] == "log_stream" && widgetDef["sort"] != nil {
		// For log stream only, sort is not a string value and we need to cast it from a
		// single object to a list of objects for TF
		widgetDef["sort"] = []interface{}{widgetDef["sort"].(map[string]interface{})}
	}
	// TF -> json conversion processes precision as float64, it needs to be converted to
	// an int value to be saved successfully
	if widgetDef["precision"] != nil {
		widgetDef["precision"] = int(widgetDef["precision"].(float64))
	}
	return widgetDef
}

func dashboardWidgetsToPpkWidgets(terraformWidgets *[]map[string]interface{}, columnWidth int64) ([]datadogV2.PowerpackInnerWidgets, diag.Diagnostics) {
	var diags diag.Diagnostics

	widgets := make([]datadogV2.PowerpackInnerWidgets, len(*terraformWidgets))
	for i, terraformWidget := range *terraformWidgets {
		if terraformWidget == nil {
			continue
		}
		widgetDef := make(map[string]interface{})
		var widgetLayout *datadogV2.PowerpackInnerWidgetLayout

		for widgetType, terraformDefinition := range terraformWidget {
			// Each terraform definition contains an ID field which is unused,
			// and a widget definition which we need to process
			if widgetType == "id" {
				continue
			}
			if widgetType == "widget_layout" {
				dimensions := terraformDefinition.([]map[string]interface{})[0]
				height := dimensions["height"].(int64)
				width := dimensions["width"].(int64)
				x := dimensions["x"].(int64)
				y := dimensions["y"].(int64)

				if x+width > columnWidth {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  fmt.Sprintf("sum of x [%d] and width [%d] is greater than the maximum of %d", x, width, columnWidth),
					})
					return nil, diags
				}
				widgetLayout = datadogV2.NewPowerpackInnerWidgetLayout(height, width, x, y)
			} else if strings.HasSuffix(widgetType, "_definition") {
				widgetDef = terraformDefinition.([]map[string]interface{})[0]
				// The type in the dictionary is in the format <widget_type>_definition, where <widget_type> can contain
				// a type with multiple underscores. To parse a valid type name, we take a substring up until the last
				// underscore. Ex: free_text_definition -> free_text, hostmap_definition -> hostmap
				widgetDef["type"] = widgetType[:strings.LastIndex(widgetType, "_")]
				widgetDef, diags = normalizeDashboardWidgetDef(widgetDef, columnWidth)
				if diags.HasError() {
					return nil, diags
				}
			}
		}
		widgetsDDItem := datadogV2.NewPowerpackInnerWidgets(widgetDef)

		if widgetLayout != nil {
			widgetsDDItem.SetLayout(*widgetLayout)
		}

		widgets[i] = *widgetsDDItem
	}
	return widgets, diags
}

func ppkWidgetsToDashboardWidgets(ppkWidgets []datadogV2.PowerpackInnerWidgets) (*[]datadogV1.Widget, diag.Diagnostics) {
	var diags diag.Diagnostics
	var datadogWidgets []datadogV1.Widget
	for _, terraformWidget := range ppkWidgets {
		var definition datadogV1.WidgetDefinition
		widgetDefinition := terraformWidget.Definition
		if widgetDefinition == nil {
			continue
		}
		widgetDefinition = normalizeTerraformWidgetDef(widgetDefinition)
		// Add new powerpack-supported widgets here
		// We save Powerpack widgets as Dashboard widgets so we need to convert them to the appropriate widget definition object.
		widgetType := widgetDefinition["type"]
		switch widgetType {
		case "alert_graph":
			definition = datadogV1.AlertGraphWidgetDefinitionAsWidgetDefinition(buildDatadogAlertGraphDefinition(widgetDefinition))
		case "alert_value":
			definition = datadogV1.AlertValueWidgetDefinitionAsWidgetDefinition(buildDatadogAlertValueDefinition(widgetDefinition))
		case "change":
			definition = datadogV1.ChangeWidgetDefinitionAsWidgetDefinition(buildDatadogChangeDefinition(widgetDefinition))
		case "check_status":
			definition = datadogV1.CheckStatusWidgetDefinitionAsWidgetDefinition(buildDatadogCheckStatusDefinition(widgetDefinition))
		case "distribution":
			definition = datadogV1.DistributionWidgetDefinitionAsWidgetDefinition(buildDatadogDistributionDefinition(widgetDefinition))
		case "event_stream":
			definition = datadogV1.EventStreamWidgetDefinitionAsWidgetDefinition(buildDatadogEventStreamDefinition(widgetDefinition))
		case "event_timeline":
			definition = datadogV1.EventTimelineWidgetDefinitionAsWidgetDefinition(buildDatadogEventTimelineDefinition(widgetDefinition))
		case "free_text":
			definition = datadogV1.FreeTextWidgetDefinitionAsWidgetDefinition(buildDatadogFreeTextDefinition(widgetDefinition))
		case "heatmap":
			definition = datadogV1.HeatMapWidgetDefinitionAsWidgetDefinition(buildDatadogHeatmapDefinition(widgetDefinition))
		case "hostmap":
			definition = datadogV1.HostMapWidgetDefinitionAsWidgetDefinition(buildDatadogHostmapDefinition(widgetDefinition))
		case "iframe":
			definition = datadogV1.IFrameWidgetDefinitionAsWidgetDefinition(buildDatadogIframeDefinition(widgetDefinition))
		case "image":
			definition = datadogV1.ImageWidgetDefinitionAsWidgetDefinition(buildDatadogImageDefinition(widgetDefinition))
		case "list_stream":
			definition = datadogV1.ListStreamWidgetDefinitionAsWidgetDefinition(buildDatadogListStreamDefinition(widgetDefinition))
		case "log_stream":
			definition = datadogV1.LogStreamWidgetDefinitionAsWidgetDefinition(buildDatadogLogStreamDefinition(widgetDefinition))
		case "manage_status":
			definition = datadogV1.MonitorSummaryWidgetDefinitionAsWidgetDefinition(buildDatadogManageStatusDefinition(widgetDefinition))
		case "note":
			definition = datadogV1.NoteWidgetDefinitionAsWidgetDefinition(buildDatadogNoteDefinition(widgetDefinition))
		case "query_value":
			definition = datadogV1.QueryValueWidgetDefinitionAsWidgetDefinition(buildDatadogQueryValueDefinition(widgetDefinition))
		case "run_workflow":
			definition = datadogV1.RunWorkflowWidgetDefinitionAsWidgetDefinition(buildDatadogRunWorkflowDefinition(widgetDefinition))
		case "scatterplot":
			definition = datadogV1.ScatterPlotWidgetDefinitionAsWidgetDefinition(buildDatadogScatterplotDefinition(widgetDefinition))
		case "servicemap":
			definition = datadogV1.ServiceMapWidgetDefinitionAsWidgetDefinition(buildDatadogServiceMapDefinition(widgetDefinition))
		case "slo":
			definition = datadogV1.SLOWidgetDefinitionAsWidgetDefinition(buildDatadogServiceLevelObjectiveDefinition(widgetDefinition))
		case "toplist":
			definition = datadogV1.ToplistWidgetDefinitionAsWidgetDefinition(buildDatadogToplistDefinition(widgetDefinition))
		case "topology_map":
			definition = datadogV1.TopologyMapWidgetDefinitionAsWidgetDefinition(buildDatadogTopologyMapDefinition(widgetDefinition))
		case "trace_service":
			definition = datadogV1.ServiceSummaryWidgetDefinitionAsWidgetDefinition(buildDatadogTraceServiceDefinition(widgetDefinition))
		case "group":
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("powerpacks cannot contain group widgets"),
			})
			return nil, diags
		case "powerpack":
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("powerpacks cannot contain powerpack widgets"),
			})
			return nil, diags
		default:
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("widget type is not supported: %s", terraformWidget.Definition["type"]),
			})
			continue
		}

		datadogWidget := datadogV1.NewWidget(definition)

		if terraformWidget.Layout != nil {
			layout := map[string]interface{}{
				"x":      terraformWidget.Layout.X,
				"y":      terraformWidget.Layout.Y,
				"width":  terraformWidget.Layout.Width,
				"height": terraformWidget.Layout.Height,
			}
			datadogWidget.SetLayout(*buildPowerpackWidgetLayout(layout))
		}

		datadogWidgets = append(datadogWidgets, *datadogWidget)
	}
	return &datadogWidgets, diags
}
func buildPowerpackWidgetLayout(terraformLayout map[string]interface{}) *datadogV1.WidgetLayout {
	datadogLayout := datadogV1.NewWidgetLayoutWithDefaults()
	datadogLayout.SetX(terraformLayout["x"].(int64))
	datadogLayout.SetY(terraformLayout["y"].(int64))
	datadogLayout.SetHeight(terraformLayout["height"].(int64))
	datadogLayout.SetWidth(terraformLayout["width"].(int64))
	return datadogLayout
}

func updatePowerpackState(d *schema.ResourceData, powerpack *datadogV2.PowerpackResponse) diag.Diagnostics {
	if powerpack.Data == nil {
		return diag.Errorf("error updating powerpack")
	}
	// Set description
	if err := d.Set("description", powerpack.Data.Attributes.GetDescription()); err != nil {
		return diag.FromErr(err)
	}

	// Set name
	if err := d.Set("name", powerpack.Data.Attributes.GetName()); err != nil {
		return diag.FromErr(err)
	}

	// Set tags
	if err := d.Set("tags", powerpack.Data.Attributes.GetTags()); err != nil {
		return diag.FromErr(err)
	}

	// Set template variables
	templateVariables := buildPowerpackTerraformTemplateVariables(powerpack.Data.Attributes.GetTemplateVariables())
	if err := d.Set("template_variables", templateVariables); err != nil {
		return diag.FromErr(err)
	}

	// Set widgets
	dashWidgets, diags := ppkWidgetsToDashboardWidgets(powerpack.Data.Attributes.GetGroupWidget().Definition.Widgets)
	if diags.HasError() {
		return diags
	}
	terraformWidgets, err := buildTerraformWidgets(dashWidgets, d)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("widget", terraformWidgets); err != nil {
		return diag.FromErr(fmt.Errorf("trouble setting widget"))
	}

	return nil
}

func resourceDatadogPowerpackDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	id := d.Id()
	if httpresp, err := apiInstances.GetPowerpackApiV2().DeletePowerpack(auth, id); err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error deleting powerpack")
	}
	return nil
}
