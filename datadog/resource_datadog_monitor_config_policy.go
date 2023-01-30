package datadog

import (
	"context"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func resourceDatadogMonitorConfigPolicy() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog monitor config policy resource. This can be used to create and manage Datadog monitor config policies.",
		CreateContext: resourceDatadogMonitorConfigPolicyCreate,
		ReadContext:   resourceDatadogMonitorConfigPolicyRead,
		UpdateContext: resourceDatadogMonitorConfigPolicyUpdate,
		DeleteContext: resourceDatadogMonitorConfigPolicyDelete,
		//CustomizeDiff: resourceDatadogMonitorCustomizeDiff,
		//Importer: &schema.ResourceImporter{
		//	StateContext: schema.ImportStatePassthroughContext,
		//},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of Datadog monitor.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func buildMonitorConfigPolicyCreateV2Struct(d utils.Resource) *datadogV2.MonitorConfigPolicyCreateRequest {
	tagKey := d.Get("tag_key").(string)
	tagKeyRequired := d.Get("tag_key_required").(bool)
	var validTagValues []string
	for _, s := range d.Get("valid_tag_values").([]interface{}) {
		validTagValues = append(validTagValues, s.(string))
	}

	return datadogV2.NewMonitorConfigPolicyCreateRequest(
		datadogV2.MonitorConfigPolicyCreateData{
			Attributes: datadogV2.MonitorConfigPolicyAttributeCreateRequest{
				PolicyType: datadogV2.MONITORCONFIGPOLICYTYPE_TAG,
				Policy: datadogV2.MonitorConfigPolicyPolicyCreateRequest{
					MonitorConfigPolicyTagPolicyCreateRequest: &datadogV2.MonitorConfigPolicyTagPolicyCreateRequest{
						TagKey:         tagKey,
						TagKeyRequired: tagKeyRequired,
						ValidTagValues: validTagValues,
					}},
			},
			Type: datadogV2.MONITORCONFIGPOLICYRESOURCETYPE_MONITOR_CONFIG_POLICY,
		})
}

func buildMonitorConfigPolicyUpdateV2Struct(d utils.Resource) *datadogV2.MonitorConfigPolicyEditRequest {
	id := d.Get("id").(string)
	// TODO optional vars
	tagKey := d.Get("tag_key").(string)
	tagKeyRequired := d.Get("tag_key_required").(bool)
	var validTagValues []string
	if attr, ok := d.GetOk("valid_tag_values"); ok {
		for _, s := range attr.([]interface{}) {
			validTagValues = append(validTagValues, s.(string))
		}
	}

	return datadogV2.NewMonitorConfigPolicyEditRequest(
		datadogV2.MonitorConfigPolicyEditData{
			Attributes: datadogV2.MonitorConfigPolicyAttributeEditRequest{
				Policy: datadogV2.MonitorConfigPolicyPolicy{
					MonitorConfigPolicyTagPolicy: &datadogV2.MonitorConfigPolicyTagPolicy{
						TagKey:         &tagKey,
						TagKeyRequired: &tagKeyRequired,
						ValidTagValues: validTagValues,
					}},
				PolicyType: datadogV2.MONITORCONFIGPOLICYTYPE_TAG,
			},
			Id:   id,
			Type: datadogV2.MONITORCONFIGPOLICYRESOURCETYPE_MONITOR_CONFIG_POLICY,
		},
	)
}

func resourceDatadogMonitorConfigPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	m := buildMonitorConfigPolicyCreateV2Struct(d)
	mCreated, httpResponse, err := apiInstances.GetMonitorsApiV2().CreateMonitorConfigPolicy(auth, *m)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating monitor config policy")
	}
	if err := utils.CheckForUnparsed(m); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(*mCreated.Data.Id)

	return updateMonitorConfigPolicyState(d, mCreated.Data)
}

func resourceDatadogMonitorConfigPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	var monitorConfigPolicyResponse datadogV2.MonitorConfigPolicyResponse
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		monitorConfigPolicyResponse, httpresp, err := apiInstances.GetMonitorsApiV2().GetMonitorConfigPolicy(auth, d.Id())
		if err != nil {
			if httpresp != nil {
				if httpresp.StatusCode == 404 {
					d.SetId("")
					return nil
				} else if httpresp.StatusCode == 502 {
					return resource.RetryableError(utils.TranslateClientError(err, httpresp, "error getting monitor config policy, retrying"))
				}
			}
			return resource.NonRetryableError(utils.TranslateClientError(err, httpresp, "error getting monitor config policy"))
		}
		if err := utils.CheckForUnparsed(monitorConfigPolicyResponse); err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	if d.Id() == "" {
		return nil
	}

	return updateMonitorConfigPolicyState(d, monitorConfigPolicyResponse.Data)
}

func resourceDatadogMonitorConfigPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	monitorConfigPolicy := buildMonitorConfigPolicyUpdateV2Struct(d)

	monitorConfigPolicyResp, httpresp, err := apiInstances.GetMonitorsApiV2().UpdateMonitorConfigPolicy(
		auth, monitorConfigPolicy.Data.Id, *monitorConfigPolicy,
	)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error updating monitor config policy")
	}
	if err := utils.CheckForUnparsed(monitorConfigPolicyResp); err != nil {
		return diag.FromErr(err)
	}

	return updateMonitorConfigPolicyState(d, monitorConfigPolicyResp.Data)
}

func resourceDatadogMonitorConfigPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	httpresp, err := apiInstances.GetMonitorsApiV2().DeleteMonitorConfigPolicy(auth, d.Id())
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error deleting monitor config policy")
	}

	return nil
}

func updateMonitorConfigPolicyState(d *schema.ResourceData, m *datadogV2.MonitorConfigPolicyResponseData) diag.Diagnostics {
	return nil
}
