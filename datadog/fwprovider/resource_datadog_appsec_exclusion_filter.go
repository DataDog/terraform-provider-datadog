package fwprovider

import (
	"context"
	"net/http"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

const (
	retryOnConflictTimeout = 1 * time.Minute
)

var (
	_ resource.ResourceWithConfigure   = &appsecExclusionFilterResource{}
	_ resource.ResourceWithImportState = &appsecExclusionFilterResource{}
)

type appsecExclusionFilterResource struct {
	Api  *datadogV2.ApplicationSecurityApi
	Auth context.Context
}

type appsecExclusionFilterModel struct {
	ID          types.String        `tfsdk:"id"`
	Description types.String        `tfsdk:"description"`
	Enabled     types.Bool          `tfsdk:"enabled"`
	EventQuery  types.String        `tfsdk:"event_query"`
	OnMatch     types.String        `tfsdk:"on_match"`
	PathGlob    types.String        `tfsdk:"path_glob"`
	IpList      types.List          `tfsdk:"ip_list"`
	Parameters  types.List          `tfsdk:"parameters"`
	RulesTarget []*rulesTargetModel `tfsdk:"rules_target"`
	Scope       []*scopeModel       `tfsdk:"scope"`
}

type rulesTargetModel struct {
	RuleId types.String `tfsdk:"rule_id"`
	Tags   *tagsModel   `tfsdk:"tags"`
}
type tagsModel struct {
	Category types.String `tfsdk:"category"`
	Type     types.String `tfsdk:"type"`
}

type scopeModel struct {
	Env     types.String `tfsdk:"env"`
	Service types.String `tfsdk:"service"`
}

func NewAppsecExclusionFilterResource() resource.Resource {
	return &appsecExclusionFilterResource{}
}

func (r *appsecExclusionFilterResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetApplicationSecurityApiV2()
	r.Auth = providerData.Auth
}

func (r *appsecExclusionFilterResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "appsec_exclusion_filter"
}

func (r *appsecExclusionFilterResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Application Security exclusion filter resource. This can be used to create and manage Application Security exclusion filters. Exclusion filters prevent the creation of security traces and therefore do not block originating requests.",
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				Required:    true,
				Description: "A description for the exclusion filter.",
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Indicates whether the exclusion filter is enabled.",
			},
			"event_query": schema.StringAttribute{
				Optional:    true,
				Description: "The event query matched by the legacy exclusion filter. Cannot be created nor updated.",
			},
			"on_match": schema.StringAttribute{
				Optional:    true,
				Description: "The action taken when the exclusion filter matches. When set to `monitor`, security traces are emitted but the requests are not blocked. By default, security traces are not emitted and the requests are not blocked.",
			},
			"path_glob": schema.StringAttribute{
				Optional:    true,
				Description: "The HTTP path glob expression matched by the exclusion filter.",
			},
			"ip_list": schema.ListAttribute{
				Optional:    true,
				Description: "The client IP addresses matched by the exclusion filter (CIDR notation is supported).",
				ElementType: types.StringType,
			},
			"parameters": schema.ListAttribute{
				Optional:    true,
				Description: "A list of parameters matched by the exclusion filter in the HTTP query string and HTTP request body. Nested parameters can be matched by joining fields with a dot character.",
				ElementType: types.StringType,
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"rules_target": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"rule_id": schema.StringAttribute{
							Optional:    true,
							Description: "Target a single WAF rule based on its identifier.",
						},
					},
					Blocks: map[string]schema.Block{
						"tags": schema.SingleNestedBlock{
							Attributes: map[string]schema.Attribute{
								"category": schema.StringAttribute{
									Optional:    true,
									Description: "The category of the targeted WAF rules.",
								},
								"type": schema.StringAttribute{
									Optional:    true,
									Description: "The type of the targeted WAF rules.",
								},
							},
						},
					},
				},
			},
			"scope": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"env": schema.StringAttribute{
							Optional:    true,
							Description: "Deploy on this environment.",
						},
						"service": schema.StringAttribute{
							Optional:    true,
							Description: "Deploy on this service.",
						},
					},
				},
			},
		},
	}
}

func (r *appsecExclusionFilterResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *appsecExclusionFilterResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state appsecExclusionFilterModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	resp, httpResp, err := r.Api.GetApplicationSecurityExclusionFilter(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving AppsecExclusionFilter"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *appsecExclusionFilterResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state appsecExclusionFilterModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildAppsecExclusionFilterRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	var resp datadogV2.ApplicationSecurityExclusionFilterResponse
	var err error
	err = retry.RetryContext(ctx, retryOnConflictTimeout, func() *retry.RetryError {
		var httpResp *http.Response
		resp, httpResp, err = r.Api.CreateApplicationSecurityExclusionFilter(r.Auth, *body)
		if err != nil {
			if httpResp.StatusCode == http.StatusConflict {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving AppsecExclusionFilter"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *appsecExclusionFilterResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state appsecExclusionFilterModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body, diags := r.buildAppsecExclusionFilterRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	var resp datadogV2.ApplicationSecurityExclusionFilterResponse
	var err error
	err = retry.RetryContext(ctx, retryOnConflictTimeout, func() *retry.RetryError {
		var httpResp *http.Response
		resp, httpResp, err = r.Api.UpdateApplicationSecurityExclusionFilter(r.Auth, id, *body)
		if err != nil {
			if httpResp.StatusCode == http.StatusConflict {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving AppsecExclusionFilter"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *appsecExclusionFilterResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state appsecExclusionFilterModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	var httpResp *http.Response
	var err error
	err = retry.RetryContext(ctx, retryOnConflictTimeout, func() *retry.RetryError {
		httpResp, err = r.Api.DeleteApplicationSecurityExclusionFilter(r.Auth, id)
		if err != nil {
			if httpResp.StatusCode == http.StatusConflict {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting appsec_exclusion_filter"))
		return
	}
}

func (r *appsecExclusionFilterResource) updateState(ctx context.Context, state *appsecExclusionFilterModel, resp *datadogV2.ApplicationSecurityExclusionFilterResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	state.Description = types.StringValue(attributes.GetDescription())

	state.Enabled = types.BoolValue(attributes.GetEnabled())

	if eventQuery, ok := attributes.GetEventQueryOk(); ok {
		state.EventQuery = types.StringValue(*eventQuery)
	}

	if onMatch, ok := attributes.GetOnMatchOk(); ok {
		state.OnMatch = types.StringValue(string(*onMatch))
	}

	if pathGlob, ok := attributes.GetPathGlobOk(); ok {
		state.PathGlob = types.StringValue(*pathGlob)
	}

	if ipList, ok := attributes.GetIpListOk(); ok && len(*ipList) > 0 {
		state.IpList, _ = types.ListValueFrom(ctx, types.StringType, *ipList)
	}

	if parameters, ok := attributes.GetParametersOk(); ok && len(*parameters) > 0 {
		state.Parameters, _ = types.ListValueFrom(ctx, types.StringType, *parameters)
	}

	if rulesTarget, ok := attributes.GetRulesTargetOk(); ok && len(*rulesTarget) > 0 {
		state.RulesTarget = []*rulesTargetModel{}
		for _, rulesTargetDd := range *rulesTarget {
			rulesTargetTfItem := rulesTargetModel{}
			if ruleId, ok := rulesTargetDd.GetRuleIdOk(); ok {
				rulesTargetTfItem.RuleId = types.StringValue(*ruleId)
			}
			if tags, ok := rulesTargetDd.GetTagsOk(); ok {

				tagsTf := tagsModel{}
				if category, ok := tags.GetCategoryOk(); ok {
					tagsTf.Category = types.StringValue(*category)
				}
				if typeVar, ok := tags.GetTypeOk(); ok {
					tagsTf.Type = types.StringValue(*typeVar)
				}

				rulesTargetTfItem.Tags = &tagsTf
			}
			state.RulesTarget = append(state.RulesTarget, &rulesTargetTfItem)
		}
	}

	if scope, ok := attributes.GetScopeOk(); ok && len(*scope) > 0 {
		state.Scope = []*scopeModel{}
		for _, scopeDd := range *scope {
			scopeTfItem := scopeModel{}
			if env, ok := scopeDd.GetEnvOk(); ok {
				scopeTfItem.Env = types.StringValue(*env)
			}
			if service, ok := scopeDd.GetServiceOk(); ok {
				scopeTfItem.Service = types.StringValue(*service)
			}
			state.Scope = append(state.Scope, &scopeTfItem)
		}
	}
}

func (r *appsecExclusionFilterResource) buildAppsecExclusionFilterRequestBody(ctx context.Context, state *appsecExclusionFilterModel) (*datadogV2.ApplicationSecurityExclusionFilterRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewApplicationSecurityExclusionFilterAttributesWithDefaults()

	if !state.Description.IsNull() {
		attributes.SetDescription(state.Description.ValueString())
	}
	if !state.Enabled.IsNull() {
		attributes.SetEnabled(state.Enabled.ValueBool())
	}
	if !state.OnMatch.IsNull() {
		attributes.SetOnMatch(datadogV2.ApplicationSecurityExclusionFilterOnMatch(state.OnMatch.ValueString()))
	}
	if !state.PathGlob.IsNull() {
		attributes.SetPathGlob(state.PathGlob.ValueString())
	}

	if !state.IpList.IsNull() {
		var ipList []string
		diags.Append(state.IpList.ElementsAs(ctx, &ipList, false)...)
		attributes.SetIpList(ipList)
	}

	if !state.Parameters.IsNull() {
		var parameters []string
		diags.Append(state.Parameters.ElementsAs(ctx, &parameters, false)...)
		attributes.SetParameters(parameters)
	}

	if state.RulesTarget != nil {
		var rulesTarget []datadogV2.ApplicationSecurityExclusionFilterRulesTarget
		for _, rulesTargetTFItem := range state.RulesTarget {
			rulesTargetDDItem := datadogV2.NewApplicationSecurityExclusionFilterRulesTarget()

			if !rulesTargetTFItem.RuleId.IsNull() {
				rulesTargetDDItem.SetRuleId(rulesTargetTFItem.RuleId.ValueString())
			}

			if rulesTargetTFItem.Tags != nil {
				var tags datadogV2.ApplicationSecurityExclusionFilterRulesTargetTags

				if !rulesTargetTFItem.Tags.Category.IsNull() {
					tags.SetCategory(rulesTargetTFItem.Tags.Category.ValueString())
				}
				if !rulesTargetTFItem.Tags.Type.IsNull() {
					tags.SetType(rulesTargetTFItem.Tags.Type.ValueString())
				}
				rulesTargetDDItem.Tags = &tags
			}
		}
		attributes.SetRulesTarget(rulesTarget)
	}

	if state.Scope != nil {
		var scope []datadogV2.ApplicationSecurityExclusionFilterScope
		for _, scopeTFItem := range state.Scope {
			scopeDDItem := datadogV2.NewApplicationSecurityExclusionFilterScope()

			if !scopeTFItem.Env.IsNull() {
				scopeDDItem.SetEnv(scopeTFItem.Env.ValueString())
			}
			if !scopeTFItem.Service.IsNull() {
				scopeDDItem.SetService(scopeTFItem.Service.ValueString())
			}
		}
		attributes.SetScope(scope)
	}

	req := datadogV2.NewApplicationSecurityExclusionFilterRequestWithDefaults()
	req.Data = *datadogV2.NewApplicationSecurityExclusionFilterResourceWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
