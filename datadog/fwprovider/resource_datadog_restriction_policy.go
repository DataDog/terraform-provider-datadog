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
	_ resource.ResourceWithConfigure   = &RestrictionPolicyResource{}
	_ resource.ResourceWithImportState = &RestrictionPolicyResource{}
)

type RestrictionPolicyResource struct {
	Api  *datadogV2.RestrictionPoliciesApi
	Auth context.Context
}

type RestrictionPolicyModel struct {
	ID         types.String     `tfsdk:"id"`
	ResourceId types.String     `tfsdk:"resource_id"`
	Bindings   []*BindingsModel `tfsdk:"bindings"`
}

type BindingsModel struct {
	Relation   types.String `tfsdk:"relation"`
	Principals types.List   `tfsdk:"principals"`
}

func NewRestrictionPolicyResource() resource.Resource {
	return &RestrictionPolicyResource{}
}

func (r *RestrictionPolicyResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	providerData, ok := request.ProviderData.(*FrameworkProvider)
	if !ok {
		response.Diagnostics.AddError("Unexpected Resource Configure Type", "")
		return
	}

	r.Api = providerData.DatadogApiInstances.GetRestrictionPoliciesApiV2()
	r.Auth = providerData.Auth
}

func (r *RestrictionPolicyResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "restriction_policy"
}

func (r *RestrictionPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog RestrictionPolicy resource. This can be used to create and manage Datadog restriction_policy.",
		Attributes: map[string]schema.Attribute{
			"resource_id": schema.StringAttribute{
				Optional:    true,
				Description: "UPDATE ME",
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"bindings": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"relation": schema.StringAttribute{
							Optional:    true,
							Description: "The role/level of access.",
						},
						"principals": schema.ListAttribute{
							Optional:    true,
							Description: "An array of principals. A principal is a subject or group of subjects. Each principal is formatted as `type:id`. Supported types: `role` and `org`. The org ID can be obtained through the api/v2/current_user API.",
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (r *RestrictionPolicyResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *RestrictionPolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state RestrictionPolicyModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	resp, httpResp, err := r.Api.GetRestrictionPolicy(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving API Key"))
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

func (r *RestrictionPolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state RestrictionPolicyModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resourceId := state.ResourceId.ValueString()

	body, diags := r.buildRestrictionPolicyRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateRestrictionPolicy(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving RestrictionPolicy"))
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

func (r *RestrictionPolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state RestrictionPolicyModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeleteRestrictionPolicy(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting restriction_policy"))
		return
	}
}

func (r *RestrictionPolicyResource) updateState(ctx context.Context, state *RestrictionPolicyModel, resp *datadogV2.RestrictionPolicyResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	if bindings, ok := attributes.GetBindingsOk(); ok && len(*bindings) > 0 {
		state.Bindings = []*BindingsModel{}
		for _, bindingsDd := range *bindings {
			bindingsTfItem := BindingsModel{}
			if principals, ok := bindingsDd.GetPrincipalsOk(); ok && len(*principals) > 0 {
				bindingsTfItem.Principals, _ = types.ListValueFrom(ctx, types.StringType, *principals)
			}
			if relation, ok := bindingsDd.GetRelationOk(); ok {
				bindingsTfItem.Relation = types.StringValue(*relation)
			}

			state.Bindings = append(state.Bindings, &bindingsTfItem)
		}
	}
}

func (r *RestrictionPolicyResource) buildRestrictionPolicyRequestBody(ctx context.Context, state *RestrictionPolicyModel) (*datadogV2.RestrictionPolicyUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewRestrictionPolicyAttributesWithDefaults()

	if state.Bindings != nil {
		var bindings []datadogV2.RestrictionPolicyBinding
		for _, bindingsTFItem := range state.Bindings {
			bindingsDDItem := datadogV2.NewRestrictionPolicyBinding()

			bindingsDDItem.SetRelation(bindingsTFItem.Relation.ValueString())

			var principals []string
			diags.Append(bindingsTFItem.Principals.ElementsAs(ctx, &principals, false)...)
			bindingsDDItem.SetPrincipals(principals)
		}
		attributes.SetBindings(bindings)
	}

	req := datadogV2.NewRestrictionPolicyUpdateRequestWithDefaults()
	req.Data = *datadogV2.NewRestrictionPolicyWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
