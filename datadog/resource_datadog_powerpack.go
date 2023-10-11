package datadog

import (
	"context"
	"fmt"
	"net/http"

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

	widgets := []datadogV2.PowerpackInnerWidgets{}
	widgetDef := make(map[string]interface{})
	widgetDef["type"] = "note"
	widgetDef["content"] = "test test test"
	widgetsDDItem := datadogV2.NewPowerpackInnerWidgets(widgetDef)
	widgets = append(widgets, *widgetsDDItem)

	definition.SetWidgets(widgets)

	groupWidget.Definition = definition

	attributes.GroupWidget = groupWidget

	req := datadogV2.NewPowerpackWithDefaults()
	req.Data = datadogV2.NewPowerpackDataWithDefaults()
	req.Data.SetType("powerpack")
	req.Data.SetAttributes(*attributes)

	return req, diags

}

func ppkWidgetsToDashboardWidgets(ppkWidgets []datadogV2.PowerpackInnerWidgets) (*[]datadogV1.Widget, diag.Diagnostics) {

	datadogWidgets := make([]datadogV1.Widget, len(ppkWidgets))
	for i, terraformWidget := range ppkWidgets {
		var definition datadogV1.WidgetDefinition
		if terraformWidget.Definition == nil {
			continue
		}
		if terraformWidget.Definition["type"] == "note" {
			definition = datadogV1.NoteWidgetDefinitionAsWidgetDefinition(buildDatadogNoteDefinition(terraformWidget.Definition))
		}
		datadogWidget := datadogV1.NewWidget(definition)

		datadogWidgets[i] = *datadogWidget
	}
	return &datadogWidgets, nil
}

func updatePowerpackState(d *schema.ResourceData, powerpack *datadogV2.PowerpackResponse) diag.Diagnostics {
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
