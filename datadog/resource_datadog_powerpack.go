package datadog

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

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
					Type:        schema.TypeList,
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
						Schema: getWidgetSchema(),
					},
				},
			}
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
		tags := make([]string, len(v.([]interface{})))
		for i, tag := range v.([]interface{}) {
			tags[i] = tag.(string)
		}
		attributes.SetTags(tags)
	}

	// Set TemplateVariables
	if v, ok := d.GetOk("template_variables"); ok {
		templateVariables := *buildPowerpackTemplateVariables(v.([]interface{}))
		attributes.SetTemplateVariables(templateVariables)
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
	datadogWidgets, _ := buildDatadogWidgets(&requestWidgets)
	// Convert to TF widget type for easier parsing
	terraformWidgets, _ := buildTerraformWidgets(datadogWidgets, d)
	// Finally, build JSON Powerpack API compatible widgets
	powerpackWidgets, diags := dashboardWidgetsToPpkWidgets(terraformWidgets)

	if diags != nil {
		return nil, diags
	}

	// Set Widget
	definition.SetWidgets(powerpackWidgets)

	groupWidget.Definition = definition

	attributes.GroupWidget = groupWidget

	req := datadogV2.NewPowerpackWithDefaults()
	req.Data = datadogV2.NewPowerpackDataWithDefaults()
	// Set type to powerpack, which is the only acceptable value for a powerpack request
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
			// Each terraform definition contains an ID field which is unused,
			// and a widget definition which we need to process
			if widgetType == "id" {
				continue
			}
			widgetDef = terraformDefinition.([]map[string]interface{})[0]
			// The type in the dictionary is in the format <widget_type>_definition, where <widget_type> can contain
			// a type with multiple underscores. To parse a valid type name, we take a substring up until the last
			// underscore. Ex: free_text_definition -> free_text, hostmap_definition -> hostmap
			widgetDef["type"] = widgetType[:strings.LastIndex(widgetType, "_")]
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
		widgetDefinition := terraformWidget.Definition
		if widgetDefinition == nil {
			continue
		}
		// Add new powerpack-supported widgets here
		// We save Powerpack widgets as Dashboard widgets so we need to convert them to the appropriate widget definition object.
		widgetType := widgetDefinition["type"]
		switch widgetType {
		case "alert_graph":
			definition = datadogV1.AlertGraphWidgetDefinitionAsWidgetDefinition(buildDatadogAlertGraphDefinition(widgetDefinition))
		case "check_status":
			definition = datadogV1.CheckStatusWidgetDefinitionAsWidgetDefinition(buildDatadogCheckStatusDefinition(widgetDefinition))
		case "free_text":
			definition = datadogV1.FreeTextWidgetDefinitionAsWidgetDefinition(buildDatadogFreeTextDefinition(widgetDefinition))
		case "iframe":
			definition = datadogV1.IFrameWidgetDefinitionAsWidgetDefinition(buildDatadogIframeDefinition(widgetDefinition))
		case "image":
			definition = datadogV1.ImageWidgetDefinitionAsWidgetDefinition(buildDatadogImageDefinition(widgetDefinition))
		case "note":
			definition = datadogV1.NoteWidgetDefinitionAsWidgetDefinition(buildDatadogNoteDefinition(widgetDefinition))
		case "servicemap":
			definition = datadogV1.ServiceMapWidgetDefinitionAsWidgetDefinition(buildDatadogServiceMapDefinition(widgetDefinition))
		default:
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("unsupported widget type: %s", terraformWidget.Definition["type"]),
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
		return diag.Errorf("error updating powerpack: %s", powerpack)
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
