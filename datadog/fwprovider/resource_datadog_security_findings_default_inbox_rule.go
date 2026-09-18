package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &securityFindingsDefaultInboxRuleResource{}
	_ resource.ResourceWithImportState = &securityFindingsDefaultInboxRuleResource{}
)

// securityFindingsDefaultInboxRuleResource manages the per-organization enabled flag of a
// Datadog-managed default inbox rule. Default rules always exist server-side: they can only be
// imported (never created), only `enabled` can be changed (via the enable/disable endpoints,
// since there is no update endpoint), and destroying the resource merely stops managing the rule.
type securityFindingsDefaultInboxRuleResource struct {
	Api  *datadogV2.SecurityMonitoringApi
	Auth context.Context
}

type securityFindingsDefaultInboxRuleModel struct {
	ID      types.String              `tfsdk:"id"`
	Enabled types.Bool                `tfsdk:"enabled"`
	Name    types.String              `tfsdk:"name"`
	Rule    *automationRuleScopeModel `tfsdk:"rule"`
	Action  *inboxRuleActionModel     `tfsdk:"action"`
}

func NewSecurityFindingsDefaultInboxRuleResource() resource.Resource {
	return &securityFindingsDefaultInboxRuleResource{}
}

func (r *securityFindingsDefaultInboxRuleResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.Auth = providerData.Auth
}

func (r *securityFindingsDefaultInboxRuleResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "security_findings_default_inbox_rule"
}

func (r *securityFindingsDefaultInboxRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog security findings default inbox rule resource. Default inbox rules are managed by Datadog and always exist: they can only be imported, not created. Only `enabled` can be set; all other attributes are read from Datadog. Default inbox rules are not part of the evaluation order managed by the `datadog_security_findings_inbox_rules_order` resource. Removing the resource from the configuration does not delete the rule, it only stops managing it.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"enabled": schema.BoolAttribute{
				Description: "Whether the default inbox rule is enabled for this organization. When not set, the rule's current server-side value is adopted; declare it to manage the value. Default inbox rules are enabled unless disabled.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the default inbox rule, defined by Datadog.",
				Computed:    true,
			},
			"rule": schema.SingleNestedAttribute{
				Description: "The scope of findings the default inbox rule applies to, defined by Datadog.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"finding_types": schema.ListAttribute{
						Description: "The list of security finding types that the default inbox rule applies to.",
						ElementType: types.StringType,
						Computed:    true,
					},
					"query": schema.StringAttribute{
						Description: "The search query that further filters the findings matched by the default inbox rule.",
						Computed:    true,
					},
				},
			},
			"action": schema.SingleNestedAttribute{
				Description: "The action taken by the default inbox rule: matching findings are pushed into the Security Inbox triage view.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"description": schema.StringAttribute{
						Description: "The description of the default inbox rule, defined by Datadog.",
						Computed:    true,
					},
				},
			},
		},
	}
}

// Create is not supported: default inbox rules always exist server-side and are defined by
// Datadog. The resource must be imported with its fixed rule ID before it can be managed.
func (r *securityFindingsDefaultInboxRuleResource) Create(_ context.Context, _ resource.CreateRequest, response *resource.CreateResponse) {
	response.Diagnostics.AddError(
		"Default inbox rule cannot be created",
		"default inbox rules are managed by Datadog and cannot be created; import the rule with its fixed ID first, e.g. `terraform import datadog_security_findings_default_inbox_rule.<label> secret_default_rule`",
	)
}

func (r *securityFindingsDefaultInboxRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	// The import ID is the fixed default rule ID (e.g. "secret_default_rule"), a plain string
	// rather than a UUID. The subsequent Read populates every other attribute from the API.
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *securityFindingsDefaultInboxRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state securityFindingsDefaultInboxRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResp, err := r.Api.GetSecurityFindingsAutomationDefaultInboxRule(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving default inbox rule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	response.Diagnostics.Append(r.updateState(ctx, &state, &resp)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *securityFindingsDefaultInboxRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan securityFindingsDefaultInboxRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	var priorState securityFindingsDefaultInboxRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &priorState)...)
	if response.Diagnostics.HasError() {
		return
	}

	// There is no update endpoint: the only mutable part is the enabled flag, set through the
	// dedicated enable/disable endpoints. Refresh ran before the plan, so priorState equals the
	// current server-side value.
	var resp datadogV2.DefaultInboxRuleResponse
	var err error
	toggled := false
	if plan.Enabled.ValueBool() != priorState.Enabled.ValueBool() {
		toggled = true
		if plan.Enabled.ValueBool() {
			resp, _, err = r.Api.EnableSecurityFindingsAutomationDefaultInboxRule(r.Auth, plan.ID.ValueString())
		} else {
			resp, _, err = r.Api.DisableSecurityFindingsAutomationDefaultInboxRule(r.Auth, plan.ID.ValueString())
		}
	}
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error toggling default inbox rule"))
		return
	}

	if toggled {
		// The toggle endpoint returned the updated rule; refresh the full state from it.
		if err := utils.CheckForUnparsed(resp); err != nil {
			response.Diagnostics.AddError("response contains unparsedObject", err.Error())
			return
		}
		response.Diagnostics.Append(r.updateState(ctx, &plan, &resp)...)
	}

	// When no toggle was needed, the plan already matches the refreshed state.
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

// Delete is a no-op: default inbox rules always exist server-side. Removing this resource from
// configuration only stops Terraform from managing the rule.
func (r *securityFindingsDefaultInboxRuleResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *securityFindingsDefaultInboxRuleResource) updateState(ctx context.Context, state *securityFindingsDefaultInboxRuleModel, resp *datadogV2.DefaultInboxRuleResponse) diag.Diagnostics {
	var diags diag.Diagnostics

	data := resp.GetData()
	attributes := data.GetAttributes()

	state.ID = types.StringValue(data.GetId())
	state.Enabled = types.BoolValue(attributes.GetEnabled())
	state.Name = types.StringValue(attributes.GetName())

	scope, d := flattenAutomationRuleScope(ctx, attributes.GetRule())
	diags.Append(d...)
	state.Rule = scope

	state.Action = flattenInboxRuleAction(attributes.GetAction())

	return diags
}
