package fwprovider

import (
	"context"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &securityMonitoringCriticalAssetResource{}
	_ resource.ResourceWithImportState = &securityMonitoringCriticalAssetResource{}
)

type securityMonitoringCriticalAssetModel struct {
	Id        types.String `tfsdk:"id"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	Query     types.String `tfsdk:"query"`
	RuleQuery types.String `tfsdk:"rule_query"`
	Severity  types.String `tfsdk:"severity"`
	Tags      types.List   `tfsdk:"tags"`
}

type securityMonitoringCriticalAssetResource struct {
	api  *datadogV2.SecurityMonitoringApi
	auth context.Context
}

func NewSecurityMonitoringCriticalAssetResource() resource.Resource {
	return &securityMonitoringCriticalAssetResource{}
}

func (r *securityMonitoringCriticalAssetResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "security_monitoring_critical_asset"
}

func (r *securityMonitoringCriticalAssetResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.auth = providerData.Auth
}

func (r *securityMonitoringCriticalAssetResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Security Monitoring Critical Asset resource. It can be used to create and manage Datadog security monitoring critical assets to modify signal severity based on asset importance.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the critical asset is enabled.",
			},
			"query": schema.StringAttribute{
				Required:    true,
				Description: "The query used to match a critical asset and the associated signals. Uses the same syntax as the search bar in the Security Signals Explorer.",
			},
			"rule_query": schema.StringAttribute{
				Required:    true,
				Description: "The rule query to filter which detection rules this critical asset applies to. Uses the same syntax as the search bar for detection rules.",
			},
			"severity": schema.StringAttribute{
				Required:    true,
				Description: "The severity change applied to signals matching this critical asset.",
				Validators: []validator.String{
					stringvalidator.OneOf("critical", "high", "medium", "low", "info", "increase", "decrease"),
				},
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "A list of tags associated with the critical asset.",
			},
		},
	}
}

func (r *securityMonitoringCriticalAssetResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *securityMonitoringCriticalAssetResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state securityMonitoringCriticalAssetModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	criticalAssetPayload, err := r.buildCreatePayload(ctx, &state)
	if err != nil {
		response.Diagnostics.AddError("error while parsing resource", err.Error())
		return
	}

	res, _, err := r.api.CreateSecurityMonitoringCriticalAsset(r.auth, *criticalAssetPayload)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating security monitoring critical asset"))
		return
	}
	if err := utils.CheckForUnparsed(res); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "response contains unparsed object"))
		return
	}

	r.updateStateFromResponse(ctx, &state, &res)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *securityMonitoringCriticalAssetResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state securityMonitoringCriticalAssetModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	criticalAssetId := state.Id.ValueString()

	res, httpResponse, err := r.api.GetSecurityMonitoringCriticalAsset(r.auth, criticalAssetId)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error fetching security monitoring critical asset"))
		return
	}
	if err := utils.CheckForUnparsed(res); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "response contains unparsed object"))
		return
	}

	r.updateStateFromResponse(ctx, &state, &res)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *securityMonitoringCriticalAssetResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan securityMonitoringCriticalAssetModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	criticalAssetPayload, err := r.buildUpdatePayload(ctx, &plan)
	if err != nil {
		response.Diagnostics.AddError("error while parsing resource", err.Error())
		return
	}

	res, _, err := r.api.UpdateSecurityMonitoringCriticalAsset(r.auth, plan.Id.ValueString(), *criticalAssetPayload)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating security monitoring critical asset"))
		return
	}
	if err := utils.CheckForUnparsed(res); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "response contains unparsed object"))
		return
	}

	r.updateStateFromResponse(ctx, &plan, &res)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *securityMonitoringCriticalAssetResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state securityMonitoringCriticalAssetModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()

	httpResp, err := r.api.DeleteSecurityMonitoringCriticalAsset(r.auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting critical asset"))
		return
	}
}

func (r *securityMonitoringCriticalAssetResource) buildCreatePayload(ctx context.Context, state *securityMonitoringCriticalAssetModel) (*datadogV2.SecurityMonitoringCriticalAssetCreateRequest, error) {
	enabled := state.Enabled.ValueBool()
	query := state.Query.ValueString()
	ruleQuery := state.RuleQuery.ValueString()
	severity := datadogV2.SecurityMonitoringCriticalAssetSeverity(state.Severity.ValueString())

	var tags []string
	if !state.Tags.IsNull() {
		tags = make([]string, 0)
		state.Tags.ElementsAs(ctx, &tags, false)
	}

	attributes := datadogV2.NewSecurityMonitoringCriticalAssetCreateAttributes(query, ruleQuery, severity)
	attributes.SetEnabled(enabled)
	if tags != nil {
		attributes.SetTags(tags)
	}

	data := datadogV2.NewSecurityMonitoringCriticalAssetCreateData(*attributes, datadogV2.SECURITYMONITORINGCRITICALASSETTYPE_CRITICAL_ASSETS)
	return datadogV2.NewSecurityMonitoringCriticalAssetCreateRequest(*data), nil
}

func (r *securityMonitoringCriticalAssetResource) buildUpdatePayload(ctx context.Context, state *securityMonitoringCriticalAssetModel) (*datadogV2.SecurityMonitoringCriticalAssetUpdateRequest, error) {
	enabled := state.Enabled.ValueBool()
	query := state.Query.ValueString()
	ruleQuery := state.RuleQuery.ValueString()
	severity := datadogV2.SecurityMonitoringCriticalAssetSeverity(state.Severity.ValueString())

	var tags []string
	if !state.Tags.IsNull() {
		tags = make([]string, 0)
		state.Tags.ElementsAs(ctx, &tags, false)
	} else {
		tags = make([]string, 0)
	}

	attributes := datadogV2.NewSecurityMonitoringCriticalAssetUpdateAttributes()
	attributes.SetEnabled(enabled)
	attributes.SetQuery(query)
	attributes.SetRuleQuery(ruleQuery)
	attributes.SetSeverity(severity)

	if tags != nil {
		attributes.SetTags(tags)
	} else {
		attributes.SetTags(make([]string, 0))
	}

	data := datadogV2.NewSecurityMonitoringCriticalAssetUpdateData(*attributes, datadogV2.SECURITYMONITORINGCRITICALASSETTYPE_CRITICAL_ASSETS)
	return datadogV2.NewSecurityMonitoringCriticalAssetUpdateRequest(*data), nil
}

func (r *securityMonitoringCriticalAssetResource) updateStateFromResponse(ctx context.Context, state *securityMonitoringCriticalAssetModel, res *datadogV2.SecurityMonitoringCriticalAssetResponse) {
	state.Id = types.StringValue(res.Data.GetId())

	attributes := res.Data.Attributes

	state.Enabled = types.BoolValue(attributes.GetEnabled())
	state.Query = types.StringValue(attributes.GetQuery())
	state.RuleQuery = types.StringValue(attributes.GetRuleQuery())
	state.Severity = types.StringValue(string(attributes.GetSeverity()))

	if len(attributes.GetTags()) == 0 {
		state.Tags = types.ListNull(types.StringType)
	} else {
		state.Tags, _ = types.ListValueFrom(ctx, types.StringType, attributes.GetTags())
	}
}
