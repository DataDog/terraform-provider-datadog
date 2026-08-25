package fwprovider

import (
	"context"
	"fmt"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.ResourceWithConfigure   = &tagRuleResource{}
	_ resource.ResourceWithImportState = &tagRuleResource{}
)

type tagRuleResource struct {
	Api  *datadogV2.TagRulesApi
	Auth context.Context
}

type tagRuleModel struct {
	ID                    types.String   `tfsdk:"id"`
	Name                  types.String   `tfsdk:"name"`
	Source                types.String   `tfsdk:"source"`
	Scope                 types.String   `tfsdk:"scope"`
	TagKey                types.String   `tfsdk:"tag_key"`
	TagValuePatterns      []types.String `tfsdk:"tag_value_patterns"`
	RuleType              types.String   `tfsdk:"rule_type"`
	Enabled               types.Bool     `tfsdk:"enabled"`
	Negated               types.Bool     `tfsdk:"negated"`
	Required              types.Bool     `tfsdk:"required"`
	ForceBlockingOnCreate types.Bool     `tfsdk:"force_blocking_on_create"`
	HardDelete            types.Bool     `tfsdk:"hard_delete"`
	Version               types.Int64    `tfsdk:"version"`
	CreatedAt             types.String   `tfsdk:"created_at"`
	CreatedBy             types.String   `tfsdk:"created_by"`
	ModifiedAt            types.String   `tfsdk:"modified_at"`
	ModifiedBy            types.String   `tfsdk:"modified_by"`
}

func NewTagRuleResource() resource.Resource {
	return &tagRuleResource{}
}

func (r *tagRuleResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	providerData, ok := request.ProviderData.(*FrameworkProvider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *FrameworkProvider, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)
		return
	}

	r.Api = providerData.DatadogApiInstances.GetTagRulesApiV2()
	r.Auth = providerData.Auth
}

func (r *tagRuleResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "tag_rule"
}

func (r *tagRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog tag rule resource. This can be used to create and manage Datadog governance tag rules.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the tag rule.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Human-readable name for the tag rule.",
				Required:    true,
			},
			"source": schema.StringAttribute{
				Description: "The telemetry source that the tag rule applies to. This field cannot be updated after creation; changing it forces a new resource.",
				Required:    true,
				Validators:  []validator.String{validators.NewEnumValidator[validator.String](datadogV2.NewTagRuleSourceFromValue)},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"scope": schema.StringAttribute{
				Description: "The scope the rule applies within. Typically an environment, team, or organization-level identifier used to limit where the rule is enforced.",
				Required:    true,
			},
			"tag_key": schema.StringAttribute{
				Description: "The tag key that the rule governs (for example, `service`).",
				Required:    true,
			},
			"tag_value_patterns": schema.ListAttribute{
				Description: "One or more patterns that valid values for the tag key must match. At least one pattern is required. These are not regular expressions: the API restricts pattern characters to `A-Za-z0-9_:-.,/*`, with `*` acting as a wildcard.",
				Required:    true,
				ElementType: types.StringType,
				Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
			},
			"rule_type": schema.StringAttribute{
				Description: "How the rule is enforced. `surfacing` only highlights non-compliant telemetry, while `blocking` rejects telemetry that violates the rule. The API only accepts `surfacing` at creation time, so creating a rule directly as `blocking` requires `force_blocking_on_create` to be set to `true`. Using `blocking` at all requires blocking tag rules to be enabled for your organization; otherwise the API returns `403 permission denied`.",
				Required:    true,
				Validators:  []validator.String{validators.NewEnumValidator[validator.String](datadogV2.NewTagRuleTypeFromValue)},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the rule is currently enforced. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
			},
			"negated": schema.BoolAttribute{
				Description: "When `true`, the rule matches tag values that do NOT match any of the supplied patterns. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
			},
			"required": schema.BoolAttribute{
				Description: "When `true`, telemetry without this tag is treated as a violation. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
			},
			"force_blocking_on_create": schema.BoolAttribute{
				Description: "Set to `true` to allow creating a rule with `rule_type` set to `blocking`. The Datadog API only accepts `surfacing` at creation time, so the provider creates the rule as `surfacing` and then immediately updates it to `blocking`, which makes the create non-atomic: if the update fails, a `surfacing` rule is left behind and the resource is marked tainted. This field is only read during creation and is not sent to the API; changing it afterwards produces a diff that has no effect.",
				Optional:    true,
			},
			"hard_delete": schema.BoolAttribute{
				Description: "Whether destroying this resource permanently deletes the tag rule. When set to `false` the rule is soft-deleted instead, which keeps it recoverable and preserves its historical compliance score data. This field is only read during deletion and is not sent to the API on create or update.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"version": schema.Int64Attribute{
				Description: "A monotonically increasing version counter that is incremented on each update.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The RFC 3339 timestamp at which the rule was created.",
				Computed:    true,
			},
			"created_by": schema.StringAttribute{
				Description: "The identifier of the user who created the rule.",
				Computed:    true,
			},
			"modified_at": schema.StringAttribute{
				Description: "The RFC 3339 timestamp at which the rule was last modified.",
				Computed:    true,
			},
			"modified_by": schema.StringAttribute{
				Description: "The identifier of the user who last modified the rule.",
				Computed:    true,
			},
		},
	}
}

func (r *tagRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *tagRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan tagRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// The API only accepts `surfacing` at creation time (TagRuleCreateType has a single
	// allowed value), so a blocking rule has to be created as surfacing and then updated.
	// That makes the create non-atomic, which the practitioner has to opt into explicitly.
	wantsBlocking := plan.RuleType.ValueString() == string(datadogV2.TAGRULETYPE_BLOCKING)
	if wantsBlocking && !plan.ForceBlockingOnCreate.ValueBool() {
		response.Diagnostics.AddError(
			"Cannot create a blocking tag rule directly",
			"The Datadog API only accepts `rule_type = \"surfacing\"` when creating a tag rule. To create this rule as "+
				"`blocking`, set `force_blocking_on_create = true`, which makes the provider create the rule as `surfacing` "+
				"and then immediately update it to `blocking`. Alternatively, apply the rule as `surfacing` first and change "+
				"`rule_type` to `blocking` in a subsequent apply.",
		)
		return
	}

	attributes := datadogV2.TagRuleCreateAttributes{
		Name:             plan.Name.ValueString(),
		Scope:            plan.Scope.ValueString(),
		Source:           datadogV2.TagRuleSource(plan.Source.ValueString()),
		TagKey:           plan.TagKey.ValueString(),
		TagValuePatterns: expandTagValuePatterns(plan.TagValuePatterns),
		RuleType:         datadogV2.TAGRULECREATETYPE_SURFACING,
	}

	if !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() {
		enabled := plan.Enabled.ValueBool()
		attributes.Enabled = &enabled
	}
	if !plan.Negated.IsNull() && !plan.Negated.IsUnknown() {
		negated := plan.Negated.ValueBool()
		attributes.Negated = &negated
	}
	if !plan.Required.IsNull() && !plan.Required.IsUnknown() {
		required := plan.Required.ValueBool()
		attributes.Required = &required
	}

	body := datadogV2.TagRuleCreateRequest{
		Data: datadogV2.TagRuleCreateData{
			Type:       datadogV2.TAGRULERESOURCETYPE_TAG_RULE,
			Attributes: attributes,
		},
	}

	resp, httpResp, err := r.Api.CreateTagRule(r.Auth, body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error creating TagRule"), ""))
		return
	}

	if !wantsBlocking {
		r.updateState(&plan, &resp)
		response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
		return
	}

	// Second leg of the non-atomic create: promote the freshly created surfacing rule to
	// blocking. Persist state first so a failure here leaves a managed (tainted) resource
	// rather than an untracked rule in the org.
	r.updateState(&plan, &resp)

	promoteBody := datadogV2.TagRuleUpdateRequest{
		Data: datadogV2.TagRuleUpdateData{
			Id:         plan.ID.ValueString(),
			Type:       datadogV2.TAGRULERESOURCETYPE_TAG_RULE,
			Attributes: buildTagRuleUpdateAttributes(&plan, datadogV2.TAGRULETYPE_BLOCKING),
		},
	}

	promoted, httpResp, err := r.Api.UpdateTagRule(r.Auth, plan.ID.ValueString(), promoteBody)
	if err != nil {
		response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
		response.Diagnostics.AddError(
			"Error promoting TagRule to blocking",
			fmt.Sprintf("The tag rule was created as `surfacing` (ID %s) but could not be updated to `blocking`: %s. "+
				"The rule exists in Datadog and is tracked in state as `surfacing`; the resource is tainted. Run "+
				"`terraform apply` again to retry the update, or `terraform destroy` to remove the rule.",
				plan.ID.ValueString(), err.Error()),
		)
		return
	}

	r.updateState(&plan, &promoted)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *tagRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state tagRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResp, err := r.Api.GetTagRule(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error retrieving TagRule"), ""))
		return
	}

	// A soft-deleted rule may still be returned by the API. Treat it as gone so refresh
	// does not loop forever on a rule that was destroyed with `hard_delete = false`.
	if deletedAt := resp.Data.Attributes.DeletedAt; deletedAt.IsSet() && deletedAt.Get() != nil {
		response.State.RemoveResource(ctx)
		return
	}

	r.updateState(&state, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *tagRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan tagRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// `source` is absent from TagRuleUpdateAttributes and carries RequiresReplace, so it is
	// never part of an update.
	body := datadogV2.TagRuleUpdateRequest{
		Data: datadogV2.TagRuleUpdateData{
			Id:         plan.ID.ValueString(),
			Type:       datadogV2.TAGRULERESOURCETYPE_TAG_RULE,
			Attributes: buildTagRuleUpdateAttributes(&plan, datadogV2.TagRuleType(plan.RuleType.ValueString())),
		},
	}

	resp, httpResp, err := r.Api.UpdateTagRule(r.Auth, plan.ID.ValueString(), body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error updating TagRule"), ""))
		return
	}

	r.updateState(&plan, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *tagRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state tagRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	optionalParams := datadogV2.NewDeleteTagRuleOptionalParameters().
		WithHardDelete(state.HardDelete.ValueBool())

	httpResp, err := r.Api.DeleteTagRule(r.Auth, state.ID.ValueString(), *optionalParams)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error deleting TagRule"), ""))
		return
	}
}

func (r *tagRuleResource) updateState(state *tagRuleModel, resp *datadogV2.TagRuleResponse) {
	data := resp.Data
	attributes := data.Attributes

	state.ID = types.StringValue(data.Id)
	state.Name = types.StringValue(attributes.Name)
	state.Source = types.StringValue(string(attributes.Source))
	state.Scope = types.StringValue(attributes.Scope)
	state.TagKey = types.StringValue(attributes.TagKey)
	state.RuleType = types.StringValue(string(attributes.RuleType))
	state.Enabled = types.BoolValue(attributes.Enabled)
	state.Negated = types.BoolValue(attributes.Negated)
	state.Required = types.BoolValue(attributes.Required)
	state.Version = types.Int64Value(attributes.Version)
	state.CreatedAt = types.StringValue(attributes.CreatedAt.Format(time.RFC3339))
	state.CreatedBy = types.StringValue(attributes.CreatedBy)
	state.ModifiedAt = types.StringValue(attributes.ModifiedAt.Format(time.RFC3339))
	state.ModifiedBy = types.StringValue(attributes.ModifiedBy)

	patterns := make([]types.String, len(attributes.TagValuePatterns))
	for i, pattern := range attributes.TagValuePatterns {
		patterns[i] = types.StringValue(pattern)
	}
	state.TagValuePatterns = patterns
}

// buildTagRuleUpdateAttributes returns the complete mutable attribute set for a PATCH.
// The API rejects a partial payload with `400 invalid request`, so every updatable field
// is always sent, even when only one of them changed.
func buildTagRuleUpdateAttributes(plan *tagRuleModel, ruleType datadogV2.TagRuleType) *datadogV2.TagRuleUpdateAttributes {
	name := plan.Name.ValueString()
	scope := plan.Scope.ValueString()
	tagKey := plan.TagKey.ValueString()
	enabled := plan.Enabled.ValueBool()
	negated := plan.Negated.ValueBool()
	required := plan.Required.ValueBool()

	return &datadogV2.TagRuleUpdateAttributes{
		Name:             &name,
		Scope:            &scope,
		TagKey:           &tagKey,
		TagValuePatterns: expandTagValuePatterns(plan.TagValuePatterns),
		RuleType:         &ruleType,
		Enabled:          &enabled,
		Negated:          &negated,
		Required:         &required,
	}
}

func expandTagValuePatterns(patterns []types.String) []string {
	expanded := make([]string, len(patterns))
	for i, pattern := range patterns {
		expanded[i] = pattern.ValueString()
	}
	return expanded
}
