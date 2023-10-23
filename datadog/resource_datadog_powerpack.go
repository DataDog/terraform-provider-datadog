package datadog

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"

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
				"widget": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "The list of widgets to display in the powerpack.",
					Elem: &schema.Resource{
						Schema: getWidgetSchema(),
					},
				},
			}
		},
	}
}

func resourceDatadogPowerpackCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	powerpackPayload, diags := buildDatadogPowerpack(ctx, d)
	if diags != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("failed to parse resource configuration"),
		})
		return diags
	}
	powerpack, httpresp, err := apiInstances.GetPowerpackApiV2().CreatePowerpack(auth, *powerpackPayload)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error creating powerpack error A")
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
			if httpResponse != nil && httpResponse.StatusCode == 404 {
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
	if diags != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("failed to parse resource configuration"),
		})
		return diags
	}

	updatedPowerpackResponse, httpResponse, err := apiInstances.GetPowerpackApiV2().UpdatePowerpack(auth, id, *powerpack)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("failure: %s", err),
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
		if httpResponse != nil && httpResponse.StatusCode == 404 {
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

func buildDatadogPowerpack(ctx context.Context, d *schema.ResourceData) (*datadogV2.Powerpack, diag.Diagnostics) {
	var diags diag.Diagnostics
	attributes := datadogV2.NewPowerpackAttributesWithDefaults()

	if v, ok := d.GetOk("description"); ok {
		attributes.SetDescription(v.(string))
	}

	// Creating group widget object
	var groupWidget datadogV2.PowerpackGroupWidget

	var definition datadogV2.PowerpackGroupWidgetDefinition

	definition.SetLayoutType("ordered")
	definition.SetType("group")

	// Should be settable by user - figure out later
	definition.SetShowTitle(true)
	definition.SetTitle("Powerpack Test")

	terraformWidgets := d.Get("widget").([]interface{})
	datadogWidgets, _ := buildDatadogWidgets(&terraformWidgets)
	terraformWidgets2, _ := buildTerraformWidgets(datadogWidgets, d)

	powerpackWidgets, diags := dashboardWidgetsToPpkWidgets(terraformWidgets2)
	if diags != nil {
		return nil, diags
	}

	definition.SetWidgets(powerpackWidgets)

	groupWidget.Definition = definition

	attributes.GroupWidget = groupWidget

	req := datadogV2.NewPowerpackWithDefaults()
	req.Data = datadogV2.NewPowerpackDataWithDefaults()
	req.Data.SetType("powerpack")
	req.Data.SetAttributes(*attributes)

	return req, diags

}

func dashboardWidgetsToPpkWidgets(terraformWidgets *[]map[string]interface{}) ([]datadogV2.PowerpackInnerWidgets, diag.Diagnostics) {
	var diags diag.Diagnostics

	widgets := make([]datadogV2.PowerpackInnerWidgets, len(*terraformWidgets))
	for i, terraformWidget := range *terraformWidgets {
		if terraformWidget == nil {
			continue
		}
		widgetDef := make(map[string]interface{})

		for widgetType, terraformDefinition := range terraformWidget {
			if widgetType == "id" {
				continue
			}
			widgetDef = terraformDefinition.([]map[string]interface{})[0]
			widgetDef["type"] = widgetType[:strings.LastIndex(widgetType, "_")]
			if widgetDef["request"] != nil {
				// Distribution/change/heatmap widgets have a "requests" field, while API Spec has a "request" field
				// Here we set the "request" field and remove "requests"
				if widgetDef["type"] == "scatterplot" || widgetDef["type"] == "hostmap" {
					// Because of course JUST one widget type expects requests to be a single value instead of a list
					widgetDefRequest := widgetDef["request"].([]map[string]interface{})[0]
					if widgetDefRequest["y"] != nil {
						widgetDefRequest["y"] = widgetDefRequest["y"].([]map[string]interface{})[0]
					}
					if widgetDefRequest["x"] != nil {
						widgetDefRequest["x"] = widgetDefRequest["x"].([]map[string]interface{})[0]
					}
					if widgetDefRequest["fill"] != nil {
						widgetDefRequest["fill"] = widgetDefRequest["fill"].([]map[string]interface{})[0]
					}
					if widgetDefRequest["size"] != nil {
						widgetDefRequest["size"] = widgetDefRequest["size"].([]map[string]interface{})[0]
					}
					widgetDef["requests"] = widgetDefRequest
					delete(widgetDef, "request")
				} else {
					widgetDefRequests := *widgetDef["request"].(*[]map[string]interface{})
					for i, widgetDefRequest := range widgetDefRequests {
						if widgetDefRequest["style"] != nil {
							// TF generates a style list, whereas API expects a single element
							widgetDefRequest["style"] = widgetDefRequest["style"].([]map[string]interface{})[0]
						}
						if widgetDefRequest["x"] != nil {
							// TF generates a style list, whereas API expects a single element
							widgetDefRequest["x"] = widgetDefRequest["x"].([]map[string]interface{})[0]
						}
						if widgetDefRequest["y"] != nil {
							// TF generates a style list, whereas API expects a single element
							widgetDefRequest["y"] = widgetDefRequest["y"].([]map[string]interface{})[0]
							//diags = append(diags, diag.Diagnostic{
							//	Severity: diag.Error,
							//	Summary:  fmt.Sprintf("this is bad: %s", widgetDef["request"]),
							//})
							//return nil, diags
						}
						widgetDefRequests[i] = widgetDefRequest
					}
					widgetDef["requests"] = widgetDefRequests
					delete(widgetDef, "request")
				}
				if widgetDef["yaxis"] != nil {
					widgetDef["yaxis"] = widgetDef["yaxis"].([]map[string]interface{})[0]
				}
				if widgetDef["xaxis"] != nil {
					widgetDef["xaxis"] = widgetDef["xaxis"].([]map[string]interface{})[0]
				}
				if widgetDef["style"] != nil {
					widgetDef["style"] = widgetDef["style"].([]map[string]interface{})[0]
				}
			}
		}
		widgetsDDItem := datadogV2.NewPowerpackInnerWidgets(widgetDef)

		widgets[i] = *widgetsDDItem
	}
	return widgets, diags
}

func ppkWidgetsToDashboardWidgets(ppkWidgets []datadogV2.PowerpackInnerWidgets) (*[]datadogV1.Widget, diag.Diagnostics) {
	var diags diag.Diagnostics
	var datadogWidgets []datadogV1.Widget
	for _, terraformWidget := range ppkWidgets {
		var definition datadogV1.WidgetDefinition
		if terraformWidget.Definition == nil {
			continue
		}
		if terraformWidget.Definition["requests"] != nil {
			terraformWidget.Definition["request"] = terraformWidget.Definition["requests"]
			delete(terraformWidget.Definition, "requests")
		}
		if terraformWidget.Definition["type"] == "note" {
			definition = datadogV1.NoteWidgetDefinitionAsWidgetDefinition(buildDatadogNoteDefinition(terraformWidget.Definition))
		} else if terraformWidget.Definition["type"] == "free_text" {
			definition = datadogV1.FreeTextWidgetDefinitionAsWidgetDefinition(buildDatadogFreeTextDefinition(terraformWidget.Definition))
		} else if terraformWidget.Definition["type"] == "iframe" {
			definition = datadogV1.IFrameWidgetDefinitionAsWidgetDefinition(buildDatadogIframeDefinition(terraformWidget.Definition))
		} else if terraformWidget.Definition["type"] == "alert_value" {
			definition = datadogV1.AlertValueWidgetDefinitionAsWidgetDefinition(buildDatadogAlertValueDefinition(terraformWidget.Definition))
		} else if terraformWidget.Definition["type"] == "alert_graph" {
			definition = datadogV1.AlertGraphWidgetDefinitionAsWidgetDefinition(buildDatadogAlertGraphDefinition(terraformWidget.Definition))
		} else if terraformWidget.Definition["type"] == "check_status" {
			definition = datadogV1.CheckStatusWidgetDefinitionAsWidgetDefinition(buildDatadogCheckStatusDefinition(terraformWidget.Definition))
		} else if terraformWidget.Definition["type"] == "servicemap" {
			definition = datadogV1.ServiceMapWidgetDefinitionAsWidgetDefinition(buildDatadogServiceMapDefinition(terraformWidget.Definition))
		} else if terraformWidget.Definition["type"] == "event_stream" {
			definition = datadogV1.EventStreamWidgetDefinitionAsWidgetDefinition(buildDatadogEventStreamDefinition(terraformWidget.Definition))
		} else if terraformWidget.Definition["type"] == "trace_service" {
			definition = datadogV1.ServiceSummaryWidgetDefinitionAsWidgetDefinition(buildDatadogTraceServiceDefinition(terraformWidget.Definition))
		} else if terraformWidget.Definition["type"] == "image" {
			definition = datadogV1.ImageWidgetDefinitionAsWidgetDefinition(buildDatadogImageDefinition(terraformWidget.Definition))
		} else if terraformWidget.Definition["type"] == "change" {
			definition = datadogV1.ChangeWidgetDefinitionAsWidgetDefinition(buildDatadogChangeDefinition(terraformWidget.Definition))
		} else if terraformWidget.Definition["type"] == "distribution" {
			definition = datadogV1.DistributionWidgetDefinitionAsWidgetDefinition(buildDatadogDistributionDefinition(terraformWidget.Definition))
		} else if terraformWidget.Definition["type"] == "heatmap" {
			definition = datadogV1.HeatMapWidgetDefinitionAsWidgetDefinition(buildDatadogHeatmapDefinition(terraformWidget.Definition))
		} else if terraformWidget.Definition["type"] == "scatterplot" {
			definition = datadogV1.ScatterPlotWidgetDefinitionAsWidgetDefinition(buildDatadogScatterplotDefinition(terraformWidget.Definition))
		} else if terraformWidget.Definition["type"] == "toplist" {
			definition = datadogV1.ToplistWidgetDefinitionAsWidgetDefinition(buildDatadogToplistDefinition(terraformWidget.Definition))
		} else if terraformWidget.Definition["type"] == "query_table" {
			definition = datadogV1.TableWidgetDefinitionAsWidgetDefinition(buildDatadogQueryTableDefinition(terraformWidget.Definition))
		} else if terraformWidget.Definition["type"] == "query_value" {
			definition = datadogV1.QueryValueWidgetDefinitionAsWidgetDefinition(buildDatadogQueryValueDefinition(terraformWidget.Definition))
		} else if terraformWidget.Definition["type"] == "hostmap" {
			definition = datadogV1.HostMapWidgetDefinitionAsWidgetDefinition(buildDatadogHostmapDefinition(terraformWidget.Definition))
		} else if terraformWidget.Definition["type"] == "topology_map" {
			definition = datadogV1.TopologyMapWidgetDefinitionAsWidgetDefinition(buildDatadogTopologyMapDefinition(terraformWidget.Definition))
		} else if terraformWidget.Definition["type"] == "service_level_objective" {
			definition = datadogV1.SLOWidgetDefinitionAsWidgetDefinition(buildDatadogServiceLevelObjectiveDefinition(terraformWidget.Definition))
		} else {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("this is bad: %s", terraformWidget.Definition),
			})
			continue
		}
		datadogWidget := datadogV1.NewWidget(definition)

		datadogWidgets = append(datadogWidgets, *datadogWidget)
	}
	return &datadogWidgets, diags
}

func updatePowerpackState(d *schema.ResourceData, powerpack *datadogV2.PowerpackResponse) diag.Diagnostics {
	if powerpack.Data == nil {
		return diag.Errorf("Powerpack data is empty, powerpack is: %s", powerpack)
	}
	// Set description
	if err := d.Set("description", powerpack.Data.Attributes.GetDescription()); err != nil {
		return diag.FromErr(err)
	}

	// Set widgets
	dashWidgets, diags := ppkWidgetsToDashboardWidgets(powerpack.Data.Attributes.GetGroupWidget().Definition.Widgets)
	if diags != nil {
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
