package datadog

import (
	"context"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDatadogCloudWorkloadSecurityAgentRules() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about existing cloud workload security agent rules for use in other resources.",
		ReadContext: dataSourceDatadogCloudWorkloadSecurityAgentRuleRead,

		Schema: map[string]*schema.Schema{
			// Computed
			"agent_rules_ids": {
				Description: "List of IDs of agent rules.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"agent_rules": {
				Description: "List of agent rules.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: cloudWorkloadSecurityAgentRuleSchema(),
				},
			},
		},
	}
}

func dataSourceDatadogCloudWorkloadSecurityAgentRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	agentRulesIds := make([]string, 0)
	agentRules := make([]map[string]interface{}, 0)

	response, httpresp, err := datadogClientV2.CloudWorkloadSecurityApi.ListCloudWorkloadSecurityAgentRules(authV2)

	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error listing agent rules")
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		return diag.FromErr(err)
	}

	for _, agentRule := range response.GetData() {
		// get agent rule id
		agentRulesIds = append(agentRulesIds, agentRule.GetId())

		// extract agent rule
		agentRuleTF := make(map[string]interface{})
		attributes := agentRule.GetAttributes()

		agentRuleTF["name"] = attributes.GetName()
		agentRuleTF["description"] = attributes.GetDescription()
		agentRuleTF["expression"] = attributes.GetExpression()
		agentRuleTF["enabled"] = attributes.GetEnabled()

		agentRules = append(agentRules, agentRuleTF)
	}

	d.SetId(strings.Join(agentRulesIds, "--"))
	d.Set("agent_rules", agentRules)
	d.Set("agent_rules_ids", agentRulesIds)

	return nil
}
