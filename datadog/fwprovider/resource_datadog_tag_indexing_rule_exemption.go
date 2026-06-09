package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &tagIndexingRuleExemptionResource{}
	_ resource.ResourceWithImportState = &tagIndexingRuleExemptionResource{}
)

type tagIndexingRuleExemptionResource struct {
	Api  *datadogV2.MetricsApi
	Auth context.Context
}

type tagIndexingRuleExemptionModel struct {
	ID              types.String `tfsdk:"id"`
	MetricName      types.String `tfsdk:"metric_name"`
	Reason          types.String `tfsdk:"reason"`
	Kind            types.String `tfsdk:"kind"`
	CreatedAt       types.String `tfsdk:"created_at"`
	CreatedByHandle types.String `tfsdk:"created_by_handle"`
}

func NewTagIndexingRuleExemptionResource() resource.Resource {
	return &tagIndexingRuleExemptionResource{}
}

func (r *tagIndexingRuleExemptionResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetMetricsApiV2()
	r.Auth = providerData.Auth
}

func (r *tagIndexingRuleExemptionResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "tag_indexing_rule_exemption"
}

func (r *tagIndexingRuleExemptionResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Tag Indexing Rule Exemption resource. Exempts a metric from all tag indexing rules, preserving its current tag indexing behavior regardless of which rules match it.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"metric_name": schema.StringAttribute{
				Required:    true,
				Description: "The metric name to exempt. Changing this value forces a new resource to be created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"reason": schema.StringAttribute{
				Required:    true,
				Description: "The reason the metric is exempt from tag indexing rules. Changing this value forces a new resource to be created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"kind": schema.StringAttribute{
				Computed:    true,
				Description: "Discriminates between an explicit exemption (`exemption`) and a pre-existing legacy tag configuration acting as an implicit exclusion (`legacy_tag_configuration`). A value of `legacy_tag_configuration` means this resource does not own the exemption state.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the exemption was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_by_handle": schema.StringAttribute{
				Computed:    true,
				Description: "Handle of the user who created the exemption.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *tagIndexingRuleExemptionResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *tagIndexingRuleExemptionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state tagIndexingRuleExemptionModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	metricName := state.MetricName.ValueString()
	resp, httpResp, err := r.Api.GetTagIndexingRuleExemption(r.Auth, metricName)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving tag indexing rule exemption"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	// The v2 API returns legacy_tag_configuration when the metric is excluded due to a
	// pre-existing tag configuration rather than an explicit exemption. This resource only
	// manages explicit exemptions, so treat legacy_tag_configuration the same as not-found.
	if data, ok := resp.GetDataOk(); ok && data != nil {
		if attrs, ok := data.GetAttributesOk(); ok {
			if kind, ok := attrs.GetKindOk(); ok && *kind == "legacy_tag_configuration" {
				response.State.RemoveResource(ctx)
				return
			}
		}
	}

	r.updateState(&state, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *tagIndexingRuleExemptionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state tagIndexingRuleExemptionModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	metricName := state.MetricName.ValueString()

	attrs := datadogV2.NewTagIndexingRuleExemptionCreateAttributesWithDefaults()
	attrs.SetReason(state.Reason.ValueString())

	data := datadogV2.NewTagIndexingRuleExemptionCreateDataWithDefaults()
	data.SetAttributes(*attrs)

	body := datadogV2.NewTagIndexingRuleExemptionCreateRequestWithDefaults()
	body.SetData(*data)

	resp, _, err := r.Api.CreateTagIndexingRuleExemption(r.Auth, metricName, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating tag indexing rule exemption"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(&state, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Update is not needed: both metric_name and reason are RequiresReplace, so Terraform
// will always destroy+create rather than update.
func (r *tagIndexingRuleExemptionResource) Update(_ context.Context, _ resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.AddError("unexpected update", "tag_indexing_rule_exemption does not support in-place updates")
}

func (r *tagIndexingRuleExemptionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state tagIndexingRuleExemptionModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	metricName := state.MetricName.ValueString()
	httpResp, err := r.Api.DeleteTagIndexingRuleExemption(r.Auth, metricName)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return // idempotent
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting tag indexing rule exemption"))
	}
}

func (r *tagIndexingRuleExemptionResource) updateState(state *tagIndexingRuleExemptionModel, resp *datadogV2.TagIndexingRuleExemptionResponse) {
	data := resp.GetData()
	// The resource ID is the metric name (the data.id field in the JSON:API response)
	state.ID = types.StringValue(data.GetId())
	state.MetricName = types.StringValue(data.GetId())

	attrs := data.GetAttributes()
	if kind, ok := attrs.GetKindOk(); ok && kind != nil {
		state.Kind = types.StringValue(*kind)
	}
	if reason, ok := attrs.GetReasonOk(); ok && reason != nil {
		state.Reason = types.StringValue(*reason)
	}
	if v, ok := attrs.GetCreatedAtOk(); ok && v != nil {
		state.CreatedAt = types.StringValue(v.String())
	}
	if v, ok := attrs.GetCreatedByHandleOk(); ok && v != nil {
		state.CreatedByHandle = types.StringValue(*v)
	}
}
