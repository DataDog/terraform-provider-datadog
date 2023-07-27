package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogRUMApplication() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog RUM application resource. This can be used to create and manage Datadog RUM applications.",
		CreateContext: resourceDatadogRUMApplicationCreate,
		ReadContext:   resourceDatadogRUMApplicationRead,
		UpdateContext: resourceDatadogRUMApplicationUpdate,
		DeleteContext: resourceDatadogRUMApplicationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The name of the RUM application",
				},
				"type": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "browser",
					Description: "The RUM application type. Supported values are `browser`, `ios`, `android`, `react-native`, `flutter`",
				},
				"client_token": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The client token",
				},
			}
		},
	}
}

func resourceDatadogRUMApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	resp, httpResponse, err := apiInstances.GetRumApiV2().GetRUMApplication(auth, d.Id())
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			d.SetId("")
			return nil
		}

		return utils.TranslateClientErrorDiag(err, httpResponse, "error getting RUM application")
	}

	return updateRUMApplicationState(d, resp.Data)
}

func resourceDatadogRUMApplicationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	body := datadogV2.RUMApplicationCreateRequest{
		Data: datadogV2.RUMApplicationCreate{
			Attributes: datadogV2.RUMApplicationCreateAttributes{
				Name: d.Get("name").(string),
				Type: datadog.PtrString(d.Get("type").(string)),
			},
			Type: datadogV2.RUMAPPLICATIONCREATETYPE_RUM_APPLICATION_CREATE,
		},
	}

	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	resp, httpResponse, err := apiInstances.GetRumApiV2().CreateRUMApplication(auth, body)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating RUM application")
	}

	d.SetId(resp.Data.Id)

	return updateRUMApplicationState(d, resp.Data)
}

func resourceDatadogRUMApplicationUpdate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	body := datadogV2.RUMApplicationUpdateRequest{
		Data: datadogV2.RUMApplicationUpdate{
			Attributes: &datadogV2.RUMApplicationUpdateAttributes{
				Name: datadog.PtrString(d.Get("name").(string)),
				Type: datadog.PtrString(d.Get("type").(string)),
			},
			Id:   d.Id(),
			Type: datadogV2.RUMAPPLICATIONUPDATETYPE_RUM_APPLICATION_UPDATE,
		},
	}

	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	resp, httpResponse, err := apiInstances.GetRumApiV2().UpdateRUMApplication(auth, d.Id(), body)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating RUM application")
	}

	return updateRUMApplicationState(d, resp.Data)
}

func resourceDatadogRUMApplicationDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	httpResponse, err := apiInstances.GetRumApiV2().DeleteRUMApplication(auth, d.Id())
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting RUM application")
	}

	return nil
}

func updateRUMApplicationState(d *schema.ResourceData, application *datadogV2.RUMApplication) diag.Diagnostics {
	if err := d.Set("name", application.Attributes.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", application.Attributes.Type); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("client_token", application.Attributes.ClientToken); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
