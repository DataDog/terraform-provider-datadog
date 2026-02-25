package datadog

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/dashboardmapping"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogDashboardV2() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog dashboard resource (v2, FieldSpec engine). This can be used to create and manage Datadog dashboards.\n\n!> The `is_read_only` field is deprecated and non-functional. Use `restricted_roles` instead to define which roles are required to edit the dashboard.",
		CreateContext: resourceDatadogDashboardV2Create,
		UpdateContext: resourceDatadogDashboardV2Update,
		ReadContext:   resourceDatadogDashboardV2Read,
		DeleteContext: resourceDatadogDashboardV2Delete,
		CustomizeDiff: func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
			oldValue, newValue := diff.GetChange("dashboard_lists")
			if !oldValue.(*schema.Set).Equal(newValue.(*schema.Set)) {
				// Only calculate removed when the list change, to no create useless diffs
				removed := oldValue.(*schema.Set).Difference(newValue.(*schema.Set))
				if err := diff.SetNew("dashboard_lists_removed", removed); err != nil {
					return err
				}
			} else {
				if err := diff.Clear("dashboard_lists_removed"); err != nil {
					return err
				}
			}

			return nil
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: func() map[string]*schema.Schema {
			s := dashboardmapping.FieldSpecsToSchema(dashboardmapping.DashboardTopLevelFields)
			// widget is special: uses AllWidgetSchemasMap for full widget type dispatch
			s["widget"] = &schema.Schema{
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The list of widgets to display on the dashboard.",
				Elem:        &schema.Resource{Schema: getWidgetSchemaV2()},
			}
			return s
		},
	}
}

// resourceDatadogDashboardV2Create, resourceDatadogDashboardV2Read,
// resourceDatadogDashboardV2Update, and resourceDatadogDashboardV2Delete implement
// CRUD for the dashboard_v2 resource using the dashboardmapping engine.

func resourceDatadogDashboardV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	bodyStr, err := dashboardmapping.MarshalDashboardJSON(d)
	if err != nil {
		return diag.FromErr(err)
	}

	respByte, httpresp, err := utils.SendRequest(auth, apiInstances.HttpClient, "POST", dashboardmapping.DashboardAPIPath, &bodyStr)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error creating dashboard")
	}

	respMap, err := utils.ConvertResponseByteToMap(respByte)
	if err != nil {
		return diag.FromErr(err)
	}

	id, ok := respMap["id"]
	if !ok {
		return diag.FromErr(errors.New("error retrieving id from response"))
	}
	d.SetId(fmt.Sprintf("%v", id))

	layoutType, ok := respMap["layout_type"]
	if !ok {
		return diag.FromErr(errors.New("error retrieving layout_type from response"))
	}

	var httpResponse *http.Response
	retryErr := retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		_, httpResponse, err = utils.SendRequest(auth, apiInstances.HttpClient, "GET", dashboardmapping.DashboardAPIPath+"/"+d.Id(), nil)
		if err != nil {
			if httpResponse != nil && httpResponse.StatusCode == 404 {
				return retry.RetryableError(fmt.Errorf("dashboard not created yet"))
			}
			return retry.NonRetryableError(err)
		}
		// We only log the error, as failing to update the list shouldn't fail dashboard creation
		updateDashboardLists(d, providerConf, d.Id(), fmt.Sprintf("%v", layoutType))
		return nil
	})
	if retryErr != nil {
		return diag.FromErr(retryErr)
	}

	return dashboardmapping.UpdateDashboardEngineState(d, respMap)
}

func resourceDatadogDashboardV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	respByte, httpresp, err := utils.SendRequest(auth, apiInstances.HttpClient, "GET", dashboardmapping.DashboardAPIPath+"/"+d.Id(), nil)
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting dashboard")
	}

	respMap, err := utils.ConvertResponseByteToMap(respByte)
	if err != nil {
		return diag.FromErr(err)
	}

	return dashboardmapping.UpdateDashboardEngineState(d, respMap)
}

func resourceDatadogDashboardV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	bodyStr, err := dashboardmapping.MarshalDashboardJSON(d)
	if err != nil {
		return diag.FromErr(err)
	}

	respByte, httpresp, err := utils.SendRequest(auth, apiInstances.HttpClient, "PUT", dashboardmapping.DashboardAPIPath+"/"+d.Id(), &bodyStr)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error updating dashboard")
	}

	respMap, err := utils.ConvertResponseByteToMap(respByte)
	if err != nil {
		return diag.FromErr(err)
	}

	layoutType, ok := respMap["layout_type"]
	if !ok {
		return diag.FromErr(errors.New("error retrieving layout_type from response"))
	}

	updateDashboardLists(d, providerConf, d.Id(), fmt.Sprintf("%v", layoutType))

	return dashboardmapping.UpdateDashboardEngineState(d, respMap)
}

func resourceDatadogDashboardV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	_, httpresp, err := utils.SendRequest(auth, apiInstances.HttpClient, "DELETE", dashboardmapping.DashboardAPIPath+"/"+d.Id(), nil)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error deleting dashboard")
	}

	return nil
}

// updateDashboardLists and buildTerraformRestrictedRoles are defined in
// resource_datadog_dashboard.go (the restored original). The v2 resource
// calls those shared helpers directly.

//
// Widget helpers
//

// getWidgetSchemaV2 builds the widget schema using the FieldSpec engine's AllWidgetSchemasMap.
// The generic widget schema is a combination of the schema for a non-group widget
// and the schema for a Group Widget (which can contains only non-group widgets)
func getWidgetSchemaV2() map[string]*schema.Schema {
	s := dashboardmapping.AllWidgetSchemasMap(false)
	// Inject recursive group widget sub-schema
	groupSchema := s["group_definition"]
	if groupSchema != nil {
		groupSchema.Elem.(*schema.Resource).Schema["widget"] = &schema.Schema{
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The list of widgets in this group.",
			Elem: &schema.Resource{
				Schema: dashboardmapping.AllWidgetSchemasMap(false),
			},
		}
	}
	return s
}
