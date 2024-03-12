package datadog

import (
	"context"
	"fmt"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDatadogCloudWorkloadSecurityAgentRules() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about existing Cloud Workload Security Agent Rules for use in other resources.",
		ReadContext: dataSourceDatadogCloudWorkloadSecurityAgentRulesRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				// Computed
				"agent_rules": {
					Description: "List of Agent rules.",
					Type:        schema.TypeList,
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The id of the Agent rule.",
							},
							"description": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The description of the Agent rule.",
							},
							"enabled": {
								Type:        schema.TypeBool,
								Computed:    true,
								Description: "Whether the Agent rule is enabled.",
							},
							"expression": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The SECL expression of the Agent rule.",
							},
							"name": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The name of the Agent rule.",
							},
						},
					},
				},
			}
		},
	}
}

func dataSourceDatadogCloudWorkloadSecurityAgentRulesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	agentRules := make([]map[string]interface{}, 0)
	response, httpresp, err := apiInstances.GetCloudWorkloadSecurityApiV2().ListCloudWorkloadSecurityAgentRules(auth)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error listing agent rules")
	}

	diags := diag.Diagnostics{}
	for _, agentRule := range response.GetData() {
		if err := utils.CheckForUnparsed(agentRule); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("skipping agent rule with id: %s", agentRule.GetId()),
				Detail:   fmt.Sprintf("rule contains unparsed object: %v", err),
			})
			continue
		}

		// extract agent rule
		agentRuleTF := make(map[string]interface{})
		attributes := agentRule.GetAttributes()

		agentRuleTF["id"] = agentRule.GetId()
		agentRuleTF["name"] = attributes.GetName()
		agentRuleTF["description"] = attributes.GetDescription()
		agentRuleTF["expression"] = attributes.GetExpression()
		agentRuleTF["enabled"] = attributes.GetEnabled()

		agentRules = append(agentRules, agentRuleTF)
	}

	d.SetId("cloud-workload-security-agent-rules")
	d.Set("agent_rules", agentRules)

	return diags
}
