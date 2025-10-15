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
	_ resource.ResourceWithConfigure   = &deploymentGateResource{}
	_ resource.ResourceWithImportState = &deploymentGateResource{}
)

type deploymentGateResource struct {
	Api  *datadogV2.DeploymentGatesApi
	Auth context.Context
}

type deploymentGateModel struct {
	ID         types.String    `tfsdk:"id"`
	DryRun     types.Bool      `tfsdk:"dry_run"`
	Env        types.String    `tfsdk:"env"`
	Identifier types.String    `tfsdk:"identifier"`
	Service    types.String    `tfsdk:"service"`
	CreatedAt  types.String    `tfsdk:"created_at"`
	UpdatedAt  types.String    `tfsdk:"updated_at"`
	CreatedBy  *createdByModel `tfsdk:"created_by"`
	UpdatedBy  *updatedByModel `tfsdk:"updated_by"`
}

func NewDeploymentGateResource() resource.Resource {
	return &deploymentGateResource{}
}

func (r *deploymentGateResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetDeploymentGatesApiV2()
	r.Auth = providerData.Auth
}

func (r *deploymentGateResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "deployment_gate"
}

func (r *deploymentGateResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog DeploymentGate resource. This can be used to create and manage Datadog deployment_gate.",
		Attributes: map[string]schema.Attribute{
			"dry_run": schema.BoolAttribute{
				Optional:    true,
				Description: "The `attributes` `dry_run`.",
			},
			"env": schema.StringAttribute{
				Required:    true,
				Description: "The `attributes` `env`.",
			},
			"identifier": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The `attributes` `identifier`.",
			},
			"service": schema.StringAttribute{
				Required:    true,
				Description: "The `attributes` `service`.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp of the deployment gate.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Last update timestamp of the deployment gate.",
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"created_by": schema.SingleNestedBlock{
				Description: "User who created the deployment gate.",
				Attributes: map[string]schema.Attribute{
					"handle": schema.StringAttribute{
						Computed:    true,
						Description: "The user handle.",
					},
					"id": schema.StringAttribute{
						Computed:    true,
						Description: "The user ID.",
					},
					"name": schema.StringAttribute{
						Computed:    true,
						Description: "The user name.",
					},
				},
			},
			"updated_by": schema.SingleNestedBlock{
				Description: "User who last updated the deployment gate.",
				Attributes: map[string]schema.Attribute{
					"handle": schema.StringAttribute{
						Computed:    true,
						Description: "The user handle.",
					},
					"id": schema.StringAttribute{
						Computed:    true,
						Description: "The user ID.",
					},
					"name": schema.StringAttribute{
						Computed:    true,
						Description: "The user name.",
					},
				},
			},
		},
	}
}

func (r *deploymentGateResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *deploymentGateResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state deploymentGateModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()

	resp, httpResp, err := r.Api.GetDeploymentGate(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving DeploymentGate"))
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

func (r *deploymentGateResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state deploymentGateModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildDeploymentGateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateDeploymentGate(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving DeploymentGate"))
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

func (r *deploymentGateResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state deploymentGateModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body, diags := r.buildDeploymentGateUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateDeploymentGate(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving DeploymentGate"))
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

func (r *deploymentGateResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state deploymentGateModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeleteDeploymentGate(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting deployment_gate"))
		return
	}
}

func (r *deploymentGateResource) updateState(ctx context.Context, state *deploymentGateModel, resp *datadogV2.DeploymentGateResponse) {
	data := resp.GetData()
	state.ID = types.StringValue(data.GetId())

	attributes := data.GetAttributes()
	if createdAt, ok := attributes.GetCreatedAtOk(); ok {
		state.CreatedAt = types.StringValue(createdAt.String())
	}

	if dryRun, ok := attributes.GetDryRunOk(); ok {
		state.DryRun = types.BoolValue(*dryRun)
	}

	if env, ok := attributes.GetEnvOk(); ok {
		state.Env = types.StringValue(*env)
	}

	if identifier, ok := attributes.GetIdentifierOk(); ok {
		state.Identifier = types.StringValue(*identifier)
	}

	if service, ok := attributes.GetServiceOk(); ok {
		state.Service = types.StringValue(*service)
	}

	if updatedAt, ok := attributes.GetUpdatedAtOk(); ok {
		state.UpdatedAt = types.StringValue(updatedAt.String())
	}

	if createdBy, ok := attributes.GetCreatedByOk(); ok {
		createdByTf := createdByModel{}
		if handle, ok := createdBy.GetHandleOk(); ok {
			createdByTf.Handle = types.StringValue(*handle)
		} else {
			createdByTf.Handle = types.StringNull()
		}
		if id, ok := createdBy.GetIdOk(); ok {
			createdByTf.Id = types.StringValue(*id)
		} else {
			createdByTf.Id = types.StringNull()
		}
		if name, ok := createdBy.GetNameOk(); ok {
			createdByTf.Name = types.StringValue(*name)
		} else {
			createdByTf.Name = types.StringNull()
		}
		state.CreatedBy = &createdByTf
	}

	if updatedBy, ok := attributes.GetUpdatedByOk(); ok {
		updatedByTf := updatedByModel{}
		if handle, ok := updatedBy.GetHandleOk(); ok {
			updatedByTf.Handle = types.StringValue(*handle)
		} else {
			updatedByTf.Handle = types.StringNull()
		}
		if id, ok := updatedBy.GetIdOk(); ok {
			updatedByTf.Id = types.StringValue(*id)
		} else {
			updatedByTf.Id = types.StringNull()
		}
		if name, ok := updatedBy.GetNameOk(); ok {
			updatedByTf.Name = types.StringValue(*name)
		} else {
			updatedByTf.Name = types.StringNull()
		}
		state.UpdatedBy = &updatedByTf
	}
}

func (r *deploymentGateResource) buildDeploymentGateRequestBody(ctx context.Context, state *deploymentGateModel) (*datadogV2.CreateDeploymentGateParams, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewCreateDeploymentGateParamsDataAttributesWithDefaults()

	if !state.DryRun.IsNull() {
		attributes.SetDryRun(state.DryRun.ValueBool())
	}
	if !state.Env.IsNull() {
		attributes.SetEnv(state.Env.ValueString())
	}
	if !state.Identifier.IsNull() {
		attributes.SetIdentifier(state.Identifier.ValueString())
	}
	if !state.Service.IsNull() {
		attributes.SetService(state.Service.ValueString())
	}

	req := datadogV2.NewCreateDeploymentGateParamsWithDefaults()
	req.Data = datadogV2.NewCreateDeploymentGateParamsDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *deploymentGateResource) buildDeploymentGateUpdateRequestBody(ctx context.Context, state *deploymentGateModel) (*datadogV2.UpdateDeploymentGateParams, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewUpdateDeploymentGateParamsDataAttributesWithDefaults()

	if !state.DryRun.IsNull() {
		attributes.SetDryRun(state.DryRun.ValueBool())
	}

	req := datadogV2.NewUpdateDeploymentGateParamsWithDefaults()
	req.Data = datadogV2.NewUpdateDeploymentGateParamsDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
