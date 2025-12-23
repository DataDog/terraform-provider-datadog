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
	_ resource.ResourceWithConfigure   = &datadogDatastoreResource{}
	_ resource.ResourceWithImportState = &datadogDatastoreResource{}
)

type datadogDatastoreResource struct {
	Api  *datadogV2.ActionsDatastoresApi
	Auth context.Context
}

type datadogDatastoreModel struct {
	ID                           types.String `tfsdk:"id"`
	Description                  types.String `tfsdk:"description"`
	Name                         types.String `tfsdk:"name"`
	OrgAccess                    types.String `tfsdk:"org_access"`
	PrimaryColumnName            types.String `tfsdk:"primary_column_name"`
	PrimaryKeyGenerationStrategy types.String `tfsdk:"primary_key_generation_strategy"`
}

func NewDatadogDatastoreResource() resource.Resource {
	return &datadogDatastoreResource{}
}

func (r *datadogDatastoreResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetActionsDatastoresApiV2()
	r.Auth = providerData.Auth
}

func (r *datadogDatastoreResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "datadog_datastore"
}

func (r *datadogDatastoreResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog DatadogDatastore resource. This can be used to create and manage Datadog datadog_datastore.",
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "A human-readable description about the datastore.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The display name for the new datastore.",
			},
			"org_access": schema.StringAttribute{
				Optional:    true,
				Description: "The organization access level for the datastore. For example, 'contributor'.",
			},
			"primary_column_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the primary key column for this datastore. Primary column names:   - Must abide by both [PostgreSQL naming conventions](https://www.postgresql.org/docs/7.0/syntax525.htm)   - Cannot exceed 63 characters",
			},
			"primary_key_generation_strategy": schema.StringAttribute{
				Optional:    true,
				Description: "Can be set to `uuid` to automatically generate primary keys when new items are added. Default value is `none`, which requires you to supply a primary key for each new item.",
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *datadogDatastoreResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *datadogDatastoreResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state datadogDatastoreModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()

	resp, httpResp, err := r.Api.GetDatastore(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving DatadogDatastore"))
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

func (r *datadogDatastoreResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state datadogDatastoreModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildDatadogDatastoreRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateDatastore(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving DatadogDatastore"))
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

func (r *datadogDatastoreResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state datadogDatastoreModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body, diags := r.buildDatadogDatastoreUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateDatastore(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving DatadogDatastore"))
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

func (r *datadogDatastoreResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state datadogDatastoreModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	_, httpResp, err := r.Api.DeleteDatastore(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting datadog_datastore"))
		return
	}
}

func (r *datadogDatastoreResource) updateState(ctx context.Context, state *datadogDatastoreModel, resp *datadogV2.Datastore) {
	state.ID = types.StringValue(resp.GetDatastoreId())

	if createdAt, ok := resp.GetCreatedAtOk(); ok {
		state.CreatedAt = types.StringValue(createdAt.String())
	}

	if creatorUserId, ok := resp.GetCreatorUserIdOk(); ok {
		state.CreatorUserId = types.Int64Value(int64(creatorUserId))
	}

	if creatorUserUuid, ok := resp.GetCreatorUserUuidOk(); ok {
		state.CreatorUserUuid = types.StringValue(*creatorUserUuid)
	}

	if description, ok := resp.GetDescriptionOk(); ok {
		state.Description = types.StringValue(*description)
	}

	if modifiedAt, ok := resp.GetModifiedAtOk(); ok {
		state.ModifiedAt = types.StringValue(modifiedAt.String())
	}

	if name, ok := resp.GetNameOk(); ok {
		state.Name = types.StringValue(*name)
	}

	if orgId, ok := resp.GetOrgIdOk(); ok {
		state.OrgId = types.Int64Value(int64(orgId))
	}

	state.PrimaryColumnName = types.StringValue(resp.GetPrimaryColumnName())

	if primaryKeyGenerationStrategy, ok := resp.GetPrimaryKeyGenerationStrategyOk(); ok {
		state.PrimaryKeyGenerationStrategy = types.StringValue(string(*primaryKeyGenerationStrategy))
	}
}

func (r *datadogDatastoreResource) buildDatadogDatastoreRequestBody(ctx context.Context, state *datadogDatastoreModel) (*datadogV2.CreateAppsDatastoreRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	req := &datadogV2.CreateAppsDatastoreRequest{}
	attributes := datadogV2.NewCreateAppsDatastoreRequestDataAttributesWithDefaults()

	if !state.Description.IsNull() {
		attributes.SetDescription(state.Description.ValueString())
	}
	if !state.Name.IsNull() {
		attributes.SetName(state.Name.ValueString())
	}
	if !state.OrgAccess.IsNull() {
		attributes.SetOrgAccess(datadogV2.CreateAppsDatastoreRequestDataAttributesOrgAccess(state.OrgAccess.ValueString()))
	}
	if !state.PrimaryColumnName.IsNull() {
		attributes.SetPrimaryColumnName(state.PrimaryColumnName.ValueString())
	}
	if !state.PrimaryKeyGenerationStrategy.IsNull() {
		attributes.SetPrimaryKeyGenerationStrategy(datadogV2.DatastorePrimaryKeyGenerationStrategy(state.PrimaryKeyGenerationStrategy.ValueString()))
	}

	req := datadogV2.NewCreateAppsDatastoreRequestWithDefaults()
	req.Data = datadogV2.NewCreateAppsDatastoreRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *datadogDatastoreResource) buildDatadogDatastoreUpdateRequestBody(ctx context.Context, state *datadogDatastoreModel) (*datadogV2.UpdateAppsDatastoreRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	req := &datadogV2.UpdateAppsDatastoreRequest{}
	attributes := datadogV2.NewUpdateAppsDatastoreRequestDataAttributesWithDefaults()

	if !state.Description.IsNull() {
		attributes.SetDescription(state.Description.ValueString())
	}
	if !state.Name.IsNull() {
		attributes.SetName(state.Name.ValueString())
	}

	req := datadogV2.NewUpdateAppsDatastoreRequestWithDefaults()
	req.Data = datadogV2.NewUpdateAppsDatastoreRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
