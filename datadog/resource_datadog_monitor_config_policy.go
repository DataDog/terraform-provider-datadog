package datadog

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

func resourceDatadogMonitorConfigPolicy() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog monitor config policy resource. This can be used to create and manage Datadog monitor config policies.",
		CreateContext: resourceDatadogMonitorConfigPolicyCreate,
		ReadContext:   resourceDatadogMonitorConfigPolicyRead,
		UpdateContext: resourceDatadogMonitorConfigPolicyUpdate,
		DeleteContext: resourceDatadogMonitorConfigPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"policy_type": {
					Description:      "The monitor config policy type",
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewMonitorConfigPolicyTypeFromValue),
				},
				"tag_policy": {
					Description:   "Config for a tag policy. Only set if `policy_type` is `tag`.",
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{"downtime_policy"},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"tag_key": {
								Type:        schema.TypeString,
								Description: "The key of the tag",
								Required:    true,
							},
							"tag_key_required": {
								Type:        schema.TypeBool,
								Description: "If a tag key is required for monitor creation",
								Required:    true,
							},
							"valid_tag_values": {
								Type:        schema.TypeList,
								Description: "Valid values for the tag",
								Required:    true,
								Elem:        &schema.Schema{Type: schema.TypeString},
							},
						},
					},
				},
				"downtime_policy": {
					Description:   "Config for a downtime duration policy. Only set if `policy_type` is `downtime`.",
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{"tag_policy"},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"max_duration_ms": {
								Type:        schema.TypeInt,
								Description: "The maximum allowed downtime duration, in milliseconds",
								Required:    true,
							},
						},
					},
				},
			}
		},
	}
}

func buildMonitorConfigPolicyCreateV2Struct(d *schema.ResourceData) *datadogV2.MonitorConfigPolicyCreateRequest {
	policyType, _ := datadogV2.NewMonitorConfigPolicyTypeFromValue(d.Get("policy_type").(string))

	return datadogV2.NewMonitorConfigPolicyCreateRequest(
		datadogV2.MonitorConfigPolicyCreateData{
			Attributes: datadogV2.MonitorConfigPolicyAttributeCreateRequest{
				PolicyType: *policyType,
				Policy:     *buildCreateRequestPolicy(d, *policyType),
			},
			Type: datadogV2.MONITORCONFIGPOLICYRESOURCETYPE_MONITOR_CONFIG_POLICY,
		})
}

func buildCreateRequestPolicy(d *schema.ResourceData, policyType datadogV2.MonitorConfigPolicyType) *datadogV2.MonitorConfigPolicyPolicyCreateRequest {
	if policyType == datadogV2.MONITORCONFIGPOLICYTYPE_TAG {
		tagKey := d.Get("tag_policy.0.tag_key").(string)
		tagKeyRequired := d.Get("tag_policy.0.tag_key_required").(bool)
		var validTagValues []string
		for _, s := range d.Get("tag_policy.0.valid_tag_values").([]interface{}) {
			validTagValues = append(validTagValues, s.(string))
		}
		return &datadogV2.MonitorConfigPolicyPolicyCreateRequest{
			MonitorConfigPolicyTagPolicyCreateRequest: &datadogV2.MonitorConfigPolicyTagPolicyCreateRequest{
				TagKey:         tagKey,
				TagKeyRequired: tagKeyRequired,
				ValidTagValues: validTagValues,
			}}
	}
	if policyType == datadogV2.MONITORCONFIGPOLICYTYPE_DOWNTIME {
		return &datadogV2.MonitorConfigPolicyPolicyCreateRequest{
			MonitorConfigPolicyDowntimePolicyCreateRequest: &datadogV2.MonitorConfigPolicyDowntimePolicyCreateRequest{
				MaxDurationMs: int64(d.Get("downtime_policy.0.max_duration_ms").(int)),
			}}
	}
	return nil
}

func buildMonitorConfigPolicyUpdateV2Struct(d *schema.ResourceData) *datadogV2.MonitorConfigPolicyEditRequest {
	id := d.Id()
	policyType, _ := datadogV2.NewMonitorConfigPolicyTypeFromValue(d.Get("policy_type").(string))
	return datadogV2.NewMonitorConfigPolicyEditRequest(
		datadogV2.MonitorConfigPolicyEditData{
			Attributes: datadogV2.MonitorConfigPolicyAttributeEditRequest{
				Policy:     *buildUpdateRequestPolicy(d, *policyType),
				PolicyType: *policyType,
			},
			Id:   id,
			Type: datadogV2.MONITORCONFIGPOLICYRESOURCETYPE_MONITOR_CONFIG_POLICY,
		},
	)
}

func buildUpdateRequestPolicy(d *schema.ResourceData, policyType datadogV2.MonitorConfigPolicyType) *datadogV2.MonitorConfigPolicyPolicy {
	if policyType == datadogV2.MONITORCONFIGPOLICYTYPE_TAG {
		tagKey := d.Get("tag_policy.0.tag_key").(string)
		tagKeyRequired := d.Get("tag_policy.0.tag_key_required").(bool)
		var validTagValues []string
		for _, s := range d.Get("tag_policy.0.valid_tag_values").([]interface{}) {
			validTagValues = append(validTagValues, s.(string))
		}
		return &datadogV2.MonitorConfigPolicyPolicy{
			MonitorConfigPolicyTagPolicy: &datadogV2.MonitorConfigPolicyTagPolicy{
				TagKey:         &tagKey,
				TagKeyRequired: &tagKeyRequired,
				ValidTagValues: validTagValues,
			}}
	}
	if policyType == datadogV2.MONITORCONFIGPOLICYTYPE_DOWNTIME {
		return &datadogV2.MonitorConfigPolicyPolicy{
			MonitorConfigPolicyDowntimePolicy: &datadogV2.MonitorConfigPolicyDowntimePolicy{
				MaxDurationMs: int64(d.Get("downtime_policy.0.max_duration_ms").(int)),
			}}
	}
	return nil
}

func resourceDatadogMonitorConfigPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	var monitorConfigPolicyResponse datadogV2.MonitorConfigPolicyResponse
	monitorConfigPolicyResponse, httpresp, err := apiInstances.GetMonitorsApiV2().GetMonitorConfigPolicy(auth, d.Id())
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting monitor config policy")
	}

	return updateMonitorConfigPolicyState(d, monitorConfigPolicyResponse.Data)
}

func resourceDatadogMonitorConfigPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	err := checkPolicyConsistency(d)
	if err != nil {
		return diag.FromErr(err)
	}

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

func resourceDatadogMonitorConfigPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	err := checkPolicyConsistency(d)
	if err != nil {
		return diag.FromErr(err)
	}

	monitorConfigPolicy := buildMonitorConfigPolicyUpdateV2Struct(d)

	mUpdated, httpresp, err := apiInstances.GetMonitorsApiV2().UpdateMonitorConfigPolicy(
		auth, monitorConfigPolicy.Data.Id, *monitorConfigPolicy,
	)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error updating monitor config policy")
	}
	if err := utils.CheckForUnparsed(mUpdated); err != nil {
		return diag.FromErr(err)
	}

	return updateMonitorConfigPolicyState(d, mUpdated.Data)
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
	d.SetId(m.GetId())
	attributes := m.GetAttributes()
	d.Set("policy_type", attributes.GetPolicyType())
	policy := attributes.GetPolicy()
	if attributes.GetPolicyType() == datadogV2.MONITORCONFIGPOLICYTYPE_TAG {
		if policy.MonitorConfigPolicyTagPolicy != nil {
			d.Set("tag_policy", []interface{}{map[string]interface{}{
				"tag_key":          policy.MonitorConfigPolicyTagPolicy.GetTagKey(),
				"tag_key_required": policy.MonitorConfigPolicyTagPolicy.GetTagKeyRequired(),
				"valid_tag_values": policy.MonitorConfigPolicyTagPolicy.GetValidTagValues(),
			}})
		}
		d.Set("downtime_policy", nil)
	}
	if attributes.GetPolicyType() == datadogV2.MONITORCONFIGPOLICYTYPE_DOWNTIME {
		if policy.MonitorConfigPolicyDowntimePolicy != nil {
			d.Set("downtime_policy", []interface{}{map[string]interface{}{
				"max_duration_ms": policy.MonitorConfigPolicyDowntimePolicy.GetMaxDurationMs(),
			}})
		}
		d.Set("tag_policy", nil)
	}
	return nil
}

func checkPolicyConsistency(d *schema.ResourceData) error {
	if d.Get("policy_type") == string(datadogV2.MONITORCONFIGPOLICYTYPE_TAG) {
		if _, ok := d.GetOk("tag_policy"); !ok {
			return fmt.Errorf("tag_policy values must be set for tag policy_type")
		}
	}
	if d.Get("policy_type") == string(datadogV2.MONITORCONFIGPOLICYTYPE_DOWNTIME) {
		if _, ok := d.GetOk("downtime_policy"); !ok {
			return fmt.Errorf("downtime_policy values must be set for downtime policy_type")
		}
	}
	return nil
}
