package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &syntheticsPrivateLocationResource{}
	_ resource.ResourceWithImportState = &syntheticsPrivateLocationResource{}
)

type syntheticsPrivateLocationResource struct {
	Api  *datadogV1.SyntheticsApi
	Auth context.Context
}

type syntheticsPrivateLocationModel struct {
	Id          types.String    `tfsdk:"id"`
	Config      types.String    `tfsdk:"config"`
	Description types.String    `tfsdk:"description"`
	Metadata    []metadataModel `tfsdk:"metadata"`
	Name        types.String    `tfsdk:"name"`
	Tags        types.List      `tfsdk:"tags"`
}

type metadataModel struct {
	RestrictedRoles types.Set `tfsdk:"restricted_roles"`
}

func NewSyntheticsPrivateLocationResource() resource.Resource {
	return &syntheticsPrivateLocationResource{}
}

func (r *syntheticsPrivateLocationResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetSyntheticsApiV1()
	r.Auth = providerData.Auth
}

func (r *syntheticsPrivateLocationResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "synthetics_private_location"
}

func (r *syntheticsPrivateLocationResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog SyntheticsPrivateLocation resource. This can be used to create and manage Datadog synthetics_private_location.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Synthetics private location name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Description of the private location.",
				Default:     stringdefault.StaticString(""),
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				Description: "A list of tags to associate with your synthetics private location.",
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"config": schema.StringAttribute{
				Description: "Configuration skeleton for the private location. See installation instructions of the private location on how to use this configuration.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"metadata": schema.ListNestedBlock{
				Description: "The private location metadata",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"restricted_roles": schema.SetAttribute{
							Description:        "A set of role identifiers pulled from the Roles API to restrict read and write access.",
							DeprecationMessage: "This field is no longer supported by the Datadog API. Please use `datadog_restriction_policy` instead.",
							Optional:           true,
							ElementType:        types.StringType,
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
		},
	}
}

func (r *syntheticsPrivateLocationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *syntheticsPrivateLocationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state syntheticsPrivateLocationModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	resp, httpResp, err := r.Api.GetPrivateLocation(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			// Delete the resource from the local state since it doesn't exist anymore in the actual state
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving SyntheticsPrivateLocation"))
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

func (r *syntheticsPrivateLocationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state syntheticsPrivateLocationModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildSyntheticsPrivateLocationRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreatePrivateLocation(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving SyntheticsPrivateLocation"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, resp.PrivateLocation)

	// set the config that is only returned when creating the private location
	conf, err := json.Marshal(resp.GetConfig())
	if err != nil {
		response.Diagnostics.AddError(
			"Error marshaling config to JSON",
			fmt.Sprintf("Could not marshal config: %s", err.Error()),
		)
	}

	state.Config = types.StringValue(string(conf))

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *syntheticsPrivateLocationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state syntheticsPrivateLocationModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()

	body, diags := r.buildSyntheticsPrivateLocationRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdatePrivateLocation(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving SyntheticsPrivateLocation"))
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

func (r *syntheticsPrivateLocationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state syntheticsPrivateLocationModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()

	httpResp, err := r.Api.DeletePrivateLocation(r.Auth, id)
	if err != nil {
		if httpResp == nil || httpResp.StatusCode != 404 {
			// The resource is assumed to still exist, and all prior state is preserved.
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting synthetics_private_location"))
			return
		}
	}

	// The resource is assumed to be destroyed, and all state is removed.
	response.State.RemoveResource(ctx)
}

func (r *syntheticsPrivateLocationResource) updateState(ctx context.Context, state *syntheticsPrivateLocationModel, resp *datadogV1.SyntheticsPrivateLocation) {
	state.Id = types.StringValue(resp.GetId())

	state.Description = types.StringValue(resp.GetDescription())
	state.Name = types.StringValue(resp.GetName())
	state.Tags, _ = types.ListValueFrom(ctx, types.StringType, resp.Tags)

	if metadata, ok := resp.GetMetadataOk(); ok {
		if restrictedRoles, ok := metadata.GetRestrictedRolesOk(); ok {
			localMetadata := metadataModel{}
			localMetadata.RestrictedRoles, _ = types.SetValueFrom(ctx, types.StringType, *restrictedRoles)
			state.Metadata = []metadataModel{localMetadata}
		}
	}
}

func (r *syntheticsPrivateLocationResource) buildSyntheticsPrivateLocationRequestBody(ctx context.Context, state *syntheticsPrivateLocationModel) (*datadogV1.SyntheticsPrivateLocation, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	syntheticsPrivateLocation := datadogV1.NewSyntheticsPrivateLocationWithDefaults()

	syntheticsPrivateLocation.SetName(state.Name.ValueString())

	if !state.Description.IsNull() {
		syntheticsPrivateLocation.SetDescription(state.Description.ValueString())
	}

	metadata := datadogV1.SyntheticsPrivateLocationMetadata{}
	if state.Metadata != nil {
		if len(state.Metadata) == 1 && !state.Metadata[0].RestrictedRoles.IsNull() {
			var restrictedRoles []string
			diags.Append(state.Metadata[0].RestrictedRoles.ElementsAs(ctx, &restrictedRoles, false)...)
			metadata.SetRestrictedRoles(restrictedRoles)
		}
	}
	syntheticsPrivateLocation.SetMetadata(metadata)

	tags := make([]string, 0)
	if !state.Tags.IsNull() {
		diags.Append(state.Tags.ElementsAs(ctx, &tags, false)...)
	}
	syntheticsPrivateLocation.SetTags(tags)

	return syntheticsPrivateLocation, diags
}
