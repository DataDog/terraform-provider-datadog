package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogCloudWorkloadSecurityAgentRule() *schema.Resource {
	return &schema.Resource{
		Description:        "Provides a Datadog Cloud Workload Security Agent Rule API resource for agent rules. Deprecated, use datadog_csm_threats_agent_rule resource instead: https://registry.terraform.io/providers/DataDog/datadog/latest/docs/resources/csm_threats_agent_rule. The API no longer creates, updates or deletes these agent rules, so this resource can only read and import existing ones.",
		DeprecationMessage: "This resource is deprecated and can no longer create, update or delete agent rules. Use the datadog_csm_threats_agent_rule resource instead: https://registry.terraform.io/providers/DataDog/datadog/latest/docs/resources/csm_threats_agent_rule",
		CreateContext:      resourceDatadogCloudWorkloadSecurityAgentRuleCreate,
		ReadContext:        resourceDatadogCloudWorkloadSecurityAgentRuleRead,
		UpdateContext:      resourceDatadogCloudWorkloadSecurityAgentRuleUpdate,
		DeleteContext:      resourceDatadogCloudWorkloadSecurityAgentRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return cloudWorkloadSecurityAgentRuleSchema()
		},
	}
}

func cloudWorkloadSecurityAgentRuleSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"description": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "The description of the Agent rule.",
		},
		"enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Whether the Agent rule is enabled.",
		},
		"expression": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The SECL expression of the Agent rule.",
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The name of the Agent rule.",
		},
	}
}

// legacyAgentRuleWritesRemoved explains why the API no longer takes this
// resource's writes, and where agent rules are managed now.
const legacyAgentRuleWritesRemoved = "The legacy Workload Protection agent rule API no longer creates, updates or deletes agent rules. Manage agent rules with the datadog_csm_threats_agent_rule resource instead: https://registry.terraform.io/providers/DataDog/datadog/latest/docs/resources/csm_threats_agent_rule"

func resourceDatadogCloudWorkloadSecurityAgentRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Diagnostics{{Severity: diag.Error, Summary: "cannot create a datadog_cloud_workload_security_agent_rule", Detail: legacyAgentRuleWritesRemoved}}
}

func resourceDatadogCloudWorkloadSecurityAgentRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	id := d.Id()
	agentRuleResponse, httpResponse, err := apiInstances.GetCSMThreatsApiV2().GetCloudWorkloadSecurityAgentRule(auth, id)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpResponse, "error fetching cloud workload security agent rule")
	}
	if err := utils.CheckForUnparsed(agentRuleResponse); err != nil {
		return diag.FromErr(err)
	}

	return updateCloudWorkloadSecurityAgentRuleState(d, &agentRuleResponse)
}

func resourceDatadogCloudWorkloadSecurityAgentRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Diagnostics{{Severity: diag.Error, Summary: "cannot update a datadog_cloud_workload_security_agent_rule", Detail: legacyAgentRuleWritesRemoved}}
}

// resourceDatadogCloudWorkloadSecurityAgentRuleDelete only removes the rule from
// state, so the resource can still be dropped from a configuration.
func resourceDatadogCloudWorkloadSecurityAgentRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId("")
	return diag.Diagnostics{{
		Severity: diag.Warning,
		Summary:  "the agent rule was removed from Terraform state but not deleted in Datadog",
		Detail:   legacyAgentRuleWritesRemoved,
	}}
}

func updateCloudWorkloadSecurityAgentRuleState(d *schema.ResourceData, agentRuleResponse *datadogV2.CloudWorkloadSecurityAgentRuleResponse) diag.Diagnostics {
	data := agentRuleResponse.GetData()
	d.SetId(data.GetId())

	attributes := data.GetAttributes()

	d.Set("description", attributes.GetDescription())
	d.Set("enabled", attributes.GetEnabled())
	d.Set("expression", attributes.GetExpression())
	d.Set("name", attributes.GetName())

	return nil
}
