package datadog

import (
	"context"

	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func dataSourceDatadogMonitorConfigPolicy() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about an existing monitor config policy for use in other resources.",
		ReadContext: dataSourceDatadogMonitorConfigPolicyRead,
		Schema: map[string]*schema.Schema{
			// Computed values
			"id": {
				Description: "ID of the monitor config policy",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"policy_type": {
				Description: "The monitor config policy type",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"tag_policy": {
				Description: "Config for a tag policy. Only set if `policy_type` is `tag`.",
				Type:        schema.TypeList,
				Computed:    true,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tag_key": {
							Type:        schema.TypeString,
							Description: "The key of the tag",
							Computed:    true,
						},
						"tag_key_required": {
							Type:        schema.TypeBool,
							Description: "If a tag key is required for monitor creation",
							Computed:    true,
						},
						"valid_tag_values": {
							Type:        schema.TypeString,
							Description: "Valid values for the tag",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceDatadogMonitorConfigPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	if policyId := d.Get("id").(string); policyId != "" {
		monitorConfigPolicy, httpResponse, err := apiInstances.GetMonitorsApiV2().GetMonitorConfigPolicy(auth, policyId)
		if err != nil {
			if httpResponse != nil && httpResponse.StatusCode == 404 {
				return diag.FromErr(errors.New("monitor config policy does not exist"))
			}
		}
		return utils.TranslateClientErrorDiag(err, httpResponse, "error querying monitor config policy")
		if err := utils.CheckForUnparsed(monitorConfigPolicy); err != nil {
			return diag.FromErr(err)
		}

		d.SetId(monitorConfigPolicy.Data.GetId())
		d.Set("policy_type", monitorConfigPolicy.Data.Attributes.GetPolicyType())

		attributes := monitorConfigPolicy.Data.GetAttributes()
		policy := attributes.GetPolicy()
		if policy.MonitorConfigPolicyTagPolicy != nil {
			d.Set("tag_policy", map[string]interface{}{
				"tag_key":          policy.MonitorConfigPolicyTagPolicy.GetTagKey(),
				"tag_key_required": policy.MonitorConfigPolicyTagPolicy.GetTagKeyRequired(),
				"valid_tag_values": policy.MonitorConfigPolicyTagPolicy.GetValidTagValues(),
			})
		}
	}

	return diag.Errorf("missing id parameter")
}
