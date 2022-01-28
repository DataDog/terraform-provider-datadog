package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogCloudWorkloadSecurityAgentRule() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog Cloud Workload Security Agent Rule API resource for agent rules.",
		CreateContext: resourceDatadogCloudWorkloadSecurityAgentRuleCreate,
		ReadContext:   resourceDatadogCloudWorkloadSecurityAgentRuleRead,
		UpdateContext: resourceDatadogCloudWorkloadSecurityAgentRuleUpdate,
		DeleteContext: resourceDatadogCloudWorkloadSecurityAgentRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: cloudWorkloadSecurityAgentRuleSchema(),
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

func resourceDatadogCloudWorkloadSecurityAgentRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	agentRuleCreate := buildCwsAgentRuleCreatePayload(d)

	response, httpResponse, err := datadogClientV2.CloudWorkloadSecurityApi.CreateCloudWorkloadSecurityAgentRule(authV2, *agentRuleCreate)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating cloud workload security agent rule")
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		return diag.FromErr(err)
	}

	return updateCloudWorkloadSecurityAgentRuleState(d, response)
}

func resourceDatadogCloudWorkloadSecurityAgentRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	id := d.Id()
	agentRuleResponse, httpResponse, err := datadogClientV2.CloudWorkloadSecurityApi.GetCloudWorkloadSecurityAgentRule(authV2, id)
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

	return updateCloudWorkloadSecurityAgentRuleState(d, agentRuleResponse)
}

func resourceDatadogCloudWorkloadSecurityAgentRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	agentRuleId := d.Id()

	agentRuleUpdate := buildCwsAgentRuleUpdatePayload(d)

	agentRuleResponse, httpResponse, err := datadogClientV2.CloudWorkloadSecurityApi.UpdateCloudWorkloadSecurityAgentRule(authV2, agentRuleId, *agentRuleUpdate)

	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating cloud workload security agent rule")
	}

	return updateCloudWorkloadSecurityAgentRuleState(d, agentRuleResponse)
}

func resourceDatadogCloudWorkloadSecurityAgentRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	agentRuleId := d.Id()

	if httpResponse, err := datadogClientV2.CloudWorkloadSecurityApi.DeleteCloudWorkloadSecurityAgentRule(authV2, agentRuleId); err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting cloud workload security agent rule")
	}

	return nil
}

func updateCloudWorkloadSecurityAgentRuleState(d *schema.ResourceData, agentRuleResponse datadogV2.CloudWorkloadSecurityAgentRuleResponse) diag.Diagnostics {
	data := agentRuleResponse.GetData()
	d.SetId(data.GetId())

	attributes := data.GetAttributes()

	d.Set("description", attributes.GetDescription())
	d.Set("enabled", attributes.GetEnabled())
	d.Set("expression", attributes.GetExpression())
	d.Set("name", attributes.GetName())

	return nil
}

func buildCwsAgentRuleUpdatePayload(d *schema.ResourceData) *datadogV2.CloudWorkloadSecurityAgentRuleUpdateRequest {
	payload := datadogV2.CloudWorkloadSecurityAgentRuleUpdateRequest{}
	payload.Data.Type = datadogV2.CLOUDWORKLOADSECURITYAGENTRULETYPE_AGENT_RULE

	if attr, ok := d.GetOk("description"); ok {
		payload.Data.Attributes.SetDescription(attr.(string))
	}

	if attr, ok := d.GetOk("expression"); ok {
		payload.Data.Attributes.SetExpression(attr.(string))
	}

	if attr, ok := d.GetOk("enabled"); ok {
		payload.Data.Attributes.SetEnabled(attr.(bool))
	}

	return &payload
}

func buildCwsAgentRuleCreatePayload(d *schema.ResourceData) *datadogV2.CloudWorkloadSecurityAgentRuleCreateRequest {
	payload := datadogV2.CloudWorkloadSecurityAgentRuleCreateRequest{}
	payload.Data.Type = datadogV2.CLOUDWORKLOADSECURITYAGENTRULETYPE_AGENT_RULE

	payload.Data.Attributes.SetExpression(d.Get("expression").(string))
	payload.Data.Attributes.SetName(d.Get("name").(string))

	if attr, ok := d.GetOk("description"); ok {
		payload.Data.Attributes.SetDescription(attr.(string))
	}

	if attr, ok := d.GetOk("enabled"); ok {
		payload.Data.Attributes.SetEnabled(attr.(bool))
	}

	return &payload
}
