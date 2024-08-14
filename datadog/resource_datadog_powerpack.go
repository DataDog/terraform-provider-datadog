package datadog

import (
	"context"
	"fmt"
	"net/http"

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
						Schema: getNonGroupWidgetSchema(true),
					},
				},
				"layout": {
					Type:        schema.TypeList,
					MaxItems:    1,
					Optional:    true,
					Computed:    true,
					Description: "The layout of the powerpack on a free-form dashboard.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"x": {
								Description:  "The position of the widget on the x (horizontal) axis. Should be greater than or equal to 0.",
								Type:         schema.TypeInt,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.IntAtLeast(0),
							},
							"y": {
								Description:  "The position of the widget on the y (vertical) axis. Should be greater than or equal to 0.",
								Type:         schema.TypeInt,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.IntAtLeast(0),
							},
							"width": {
								Description:  "The width of the widget.",
								Type:         schema.TypeInt,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.IntAtLeast(1),
							},
							"height": {
								Description:  "The height of the widget.",
								Type:         schema.TypeInt,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.IntAtLeast(1),
							},
						},
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
			diags = append(diags, utils.TranslateClientErrorDiag(err, httpResponse, "error updating powerpack")...)
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

	width := int64(layout["width"].(int))
	x := int64(layout["x"].(int))
	if width+x > 12 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "powerpack layout contains an invalid value. sum of x and width is greater than the maximum of 12.",
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
	terraformWidgets := d.Get("widget").([]interface{})
	datadogWidgets, _ := buildDatadogWidgets(&terraformWidgets)

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
	}

	// Finally, build JSON Powerpack API compatible widgets
	powerpackWidgets, diags := dashboardWidgetsToPpkWidgets(datadogWidgets, columnWidth)

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

func dashboardWidgetsToPpkWidgets(terraformWidgets *[]datadogV1.Widget, columnWidth int64) ([]datadogV2.PowerpackInnerWidgets, diag.Diagnostics) {
	var diags diag.Diagnostics

	widgets := make([]datadogV2.PowerpackInnerWidgets, len(*terraformWidgets))
	for i, terraformWidget := range *terraformWidgets {
		dashJsonBytes, _ := terraformWidget.MarshalJSON()
		var newPowerpackWidget datadogV2.PowerpackInnerWidgets
		newPowerpackWidget.UnmarshalJSON(dashJsonBytes)
		// Explicitly set additionalProperties as nil so we don't send bad definitions
		newPowerpackWidget.AdditionalProperties = nil

		widgets[i] = newPowerpackWidget
	}

	return widgets, diags
}

func ppkWidgetsToTerraformWidgets(ppkWidgets []datadogV2.PowerpackInnerWidgets) (*[]map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	terraformWidgets := make([]map[string]interface{}, len(ppkWidgets))

	for i, ppkWidget := range ppkWidgets {
		serializedMap, err := ppkWidget.MarshalJSON()
		if err != nil {
			return nil, diag.FromErr(err)
		}

		var ddV1Widget datadogV1.Widget
		ddV1Widget.UnmarshalJSON(serializedMap)

		tfWidget, err := buildTerraformWidget(&ddV1Widget)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		terraformWidgets[i] = tfWidget
	}
	return &terraformWidgets, diags
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

	// Set tags
	if err := d.Set("live_span", powerpack.Data.Attributes.GroupWidget.GetLiveSpan()); err != nil {
		return diag.FromErr(err)
	}

	// Set template variables
	templateVariables := buildPowerpackTerraformTemplateVariables(powerpack.Data.Attributes.GetTemplateVariables())
	if err := d.Set("template_variables", templateVariables); err != nil {
		return diag.FromErr(err)
	}

	// Build layout
	if v, ok := powerpack.Data.Attributes.GroupWidget.GetLayoutOk(); ok {
		widgetLayout := map[string]interface{}{
			"x":      (*v).GetX(),
			"y":      (*v).GetY(),
			"height": (*v).GetHeight(),
			"width":  (*v).GetWidth(),
		}

		if err := d.Set("layout", []map[string]interface{}{widgetLayout}); err != nil {
			return diag.FromErr(err)
		}

	}

	// Set widgets
	dashWidgets, diags := ppkWidgetsToTerraformWidgets(powerpack.Data.Attributes.GetGroupWidget().Definition.Widgets)
	if diags.HasError() {
		return diags
	}

	if err := d.Set("widget", dashWidgets); err != nil {
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
