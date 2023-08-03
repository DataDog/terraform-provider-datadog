package datadog

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func dataSourceDatadogMonitorConfigPolicies() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to list existing monitor config policies for use in other resources.",
		ReadContext: dataSourceDatadogMonitorConfigPoliciesRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				// Computed values
				"monitor_config_policies": {
					Description: "List of monitor config policies",
					Type:        schema.TypeList,
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Description: "ID of the monitor config policy",
								Type:        schema.TypeString,
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
								MaxItems:    1,
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
											Type:        schema.TypeList,
											Description: "Valid values for the tag",
											Computed:    true,
											Elem:        &schema.Schema{Type: schema.TypeString},
										},
									},
								},
							},
						},
					},
				},
			}
		},
	}
}

func dataSourceDatadogMonitorConfigPoliciesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	monitorConfigPolicies, httpresp, err := apiInstances.GetMonitorsApiV2().ListMonitorConfigPolicies(auth)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error querying monitor config policies")
	}

	diags := diag.Diagnostics{}
	tfMonitorConfigPolicies := make([]map[string]interface{}, len(monitorConfigPolicies.Data))
	for i, mcp := range monitorConfigPolicies.Data {
		if err := utils.CheckForUnparsed(mcp); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("skipping monitor config policy with id: %s", mcp.GetId()),
				Detail:   fmt.Sprintf("aws logs service contains unparsed object: %v", err),
			})
			continue
		}

		attributes := mcp.GetAttributes()
		tfMonitorConfigPolicies[i] = map[string]interface{}{
			"id":          mcp.GetId(),
			"policy_type": attributes.GetPolicyType(),
		}

		policy := attributes.GetPolicy()
		if policy.MonitorConfigPolicyTagPolicy != nil {
			tfMonitorConfigPolicies[i]["tag_policy"] = []interface{}{map[string]interface{}{
				"tag_key":          policy.MonitorConfigPolicyTagPolicy.GetTagKey(),
				"tag_key_required": policy.MonitorConfigPolicyTagPolicy.GetTagKeyRequired(),
				"valid_tag_values": policy.MonitorConfigPolicyTagPolicy.GetValidTagValues(),
			}}
		}
	}

	d.SetId("monitor-config-policies")
	d.Set("monitor_config_policies", tfMonitorConfigPolicies)

	return diags
}
