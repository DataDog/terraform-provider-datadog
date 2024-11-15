package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

type csmThreatsPolicyModel struct {
	Id          types.String `tfsdk:"id"`
	Tags        types.Set    `tfsdk:"tags"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
}

type csmThreatsPolicyResource struct {
	api  *datadogV2.CSMThreatsApi
	auth context.Context
}

func NewCSMThreatsPolicyResource() resource.Resource {
	return &csmThreatsPolicyResource{}
}

func (r *csmThreatsPolicyResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "csm_threats_policy"
}

func (r *csmThreatsPolicyResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.api = providerData.DatadogApiInstances.GetCSMThreatsApiV2()
	r.auth = providerData.Auth
}

func (r *csmThreatsPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog CSM Threats policy API resource.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the policy.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "A description for the policy.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Indicates whether the policy is enabled.",
				Computed:    true,
			},
			"tags": schema.SetAttribute{
				Optional:    true,
				Description: "Host tags that define where the policy is deployed.",
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

func (r *csmThreatsPolicyResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *csmThreatsPolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state csmThreatsPolicyModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	csmThreatsMutex.Lock()
	defer csmThreatsMutex.Unlock()

	policyPayload, err := r.buildCreateCSMThreatsPolicyPayload(&state)
	if err != nil {
		response.Diagnostics.AddError("error while parsing resource", err.Error())
	}

	res, _, err := r.api.CreateCSMThreatsAgentPolicy(r.auth, *policyPayload)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating policy"))
		return
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "response contains unparsed object"))
		return
	}

	r.updateStateFromResponse(ctx, &state, &res)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *csmThreatsPolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state csmThreatsPolicyModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	policyId := state.Id.ValueString()
	res, httpResponse, err := r.api.GetCSMThreatsAgentPolicy(r.auth, policyId)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error fetching agent policy"))
		return
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "response contains unparsed object"))
		return
	}

	r.updateStateFromResponse(ctx, &state, &res)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *csmThreatsPolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state csmThreatsPolicyModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	csmThreatsMutex.Lock()
	defer csmThreatsMutex.Unlock()

	policyPayload, err := r.buildUpdateCSMThreatsPolicyPayload(&state)
	if err != nil {
		response.Diagnostics.AddError("error while parsing resource", err.Error())
	}

	res, _, err := r.api.UpdateCSMThreatsAgentPolicy(r.auth, state.Id.ValueString(), *policyPayload)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating agent rule"))
		return
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "response contains unparsed object"))
		return
	}

	r.updateStateFromResponse(ctx, &state, &res)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *csmThreatsPolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state csmThreatsPolicyModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	csmThreatsMutex.Lock()
	defer csmThreatsMutex.Unlock()

	id := state.Id.ValueString()

	httpResp, err := r.api.DeleteCSMThreatsAgentPolicy(r.auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting agent rule"))
		return
	}
}

func (r *csmThreatsPolicyResource) buildCreateCSMThreatsPolicyPayload(state *csmThreatsPolicyModel) (*datadogV2.CloudWorkloadSecurityAgentPolicyCreateRequest, error) {
	_, name, description, enabled, tags, err := r.extractPolicyAttributesFromResource(state)
	if err != nil {
		return nil, err
	}

	attributes := datadogV2.CloudWorkloadSecurityAgentPolicyCreateAttributes{}
	attributes.Name = name
	attributes.Description = description
	attributes.Enabled = enabled
	attributes.HostTags = tags

	data := datadogV2.NewCloudWorkloadSecurityAgentPolicyCreateData(attributes, datadogV2.CLOUDWORKLOADSECURITYAGENTPOLICYTYPE_POLICY)
	return datadogV2.NewCloudWorkloadSecurityAgentPolicyCreateRequest(*data), nil
}

func (r *csmThreatsPolicyResource) buildUpdateCSMThreatsPolicyPayload(state *csmThreatsPolicyModel) (*datadogV2.CloudWorkloadSecurityAgentPolicyUpdateRequest, error) {
	policyId, name, description, enabled, tags, err := r.extractPolicyAttributesFromResource(state)
	if err != nil {
		return nil, err
	}
	attributes := datadogV2.CloudWorkloadSecurityAgentPolicyUpdateAttributes{}
	attributes.Name = &name
	attributes.Description = description
	attributes.Enabled = enabled
	attributes.HostTags = tags

	data := datadogV2.NewCloudWorkloadSecurityAgentPolicyUpdateData(attributes, datadogV2.CLOUDWORKLOADSECURITYAGENTPOLICYTYPE_POLICY)
	data.Id = &policyId
	return datadogV2.NewCloudWorkloadSecurityAgentPolicyUpdateRequest(*data), nil
}

func (r *csmThreatsPolicyResource) extractPolicyAttributesFromResource(state *csmThreatsPolicyModel) (string, string, *string, *bool, []string, error) {
	// Mandatory fields
	id := state.Id.ValueString()
	name := state.Name.ValueString()
	enabled := state.Enabled.ValueBoolPointer()
	description := state.Description.ValueStringPointer()
	var tags []string
	if !state.Tags.IsNull() && !state.Tags.IsUnknown() {
		for _, tag := range state.Tags.Elements() {
			tagStr, ok := tag.(types.String)
			if !ok {
				return "", "", nil, nil, nil, fmt.Errorf("expected item to be of type types.String, got %T", tag)
			}
			tags = append(tags, tagStr.ValueString())
		}
	}

	return id, name, description, enabled, tags, nil
}

func (r *csmThreatsPolicyResource) updateStateFromResponse(ctx context.Context, state *csmThreatsPolicyModel, res *datadogV2.CloudWorkloadSecurityAgentPolicyResponse) {
	state.Id = types.StringValue(res.Data.GetId())

	attributes := res.Data.Attributes

	state.Name = types.StringValue(attributes.GetName())
	state.Description = types.StringValue(attributes.GetDescription())
	state.Enabled = types.BoolValue(attributes.GetEnabled())
	state.Tags, _ = types.SetValueFrom(ctx, types.StringType, attributes.GetHostTags())
}
