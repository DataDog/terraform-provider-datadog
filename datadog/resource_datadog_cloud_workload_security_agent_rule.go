package datadog

import (
	"context"

	_ "gopkg.in/warnings.v0"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const agentRuleType = "agent_rule"

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
			Required:    true,
			Description: "The description of the Agent rule.",
		},

		"enabled": {
			Type:        schema.TypeBool,
			Required:    true,
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

	// update the resource
	updateResourceDataAgentRuleFromResponse(d, response)

	return nil
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

	updateResourceDataAgentRuleFromResponse(d, agentRuleResponse)

	return nil
}

func resourceDatadogCloudWorkloadSecurityAgentRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	agentRuleId := d.Id()

	agentRuleUpdate := buildCwsAgentRuleUpdatePayload(d)

	if _, httpResponse, err := datadogClientV2.CloudWorkloadSecurityApi.UpdateCloudWorkloadSecurityAgentRule(authV2, agentRuleId, *agentRuleUpdate); err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating cloud workload security agent rule")
	}

	return nil
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

func updateResourceDataAgentRuleFromResponse(d *schema.ResourceData, agentRuleResponse datadogV2.CloudWorkloadSecurityAgentRuleResponse) {
	data := agentRuleResponse.GetData()
	d.SetId(data.GetId())

	attributes := data.GetAttributes()

	d.Set("description", attributes.GetName())
	d.Set("enabled", attributes.GetEnabled())
	d.Set("expression", attributes.GetExpression())
	d.Set("name", attributes.GetName())
}

func buildCwsAgentRuleUpdatePayload(d *schema.ResourceData) *datadogV2.CloudWorkloadSecurityAgentRuleUpdateRequest {
	payload := datadogV2.CloudWorkloadSecurityAgentRuleUpdateRequest{}
	payload.Data.Type = agentRuleType

	payload.Data.Attributes.SetDescription(d.Get("description").(string))
	payload.Data.Attributes.SetEnabled(d.Get("enabled").(bool))
	payload.Data.Attributes.SetExpression(d.Get("expression").(string))

	return &payload
}

func buildCwsAgentRuleCreatePayload(d *schema.ResourceData) *datadogV2.CloudWorkloadSecurityAgentRuleCreateRequest {
	payload := datadogV2.CloudWorkloadSecurityAgentRuleCreateRequest{}
	payload.Data.Type = agentRuleType

	payload.Data.Attributes.SetDescription(d.Get("description").(string))
	payload.Data.Attributes.SetEnabled(d.Get("enabled").(bool))
	payload.Data.Attributes.SetExpression(d.Get("expression").(string))
	payload.Data.Attributes.SetName(d.Get("name").(string))

	return &payload
}
