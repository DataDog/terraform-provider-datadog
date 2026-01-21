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
	_ resource.ResourceWithConfigure   = &datastoreResource{}
	_ resource.ResourceWithImportState = &datastoreResource{}
)

type datastoreResource struct {
	Api  *datadogV2.ActionsDatastoresApi
	Auth context.Context
}

type datastoreModel struct {
	ID                           types.String `tfsdk:"id"`
	Description                  types.String `tfsdk:"description"`
	Name                         types.String `tfsdk:"name"`
	OrgAccess                    types.String `tfsdk:"org_access"`
	PrimaryColumnName            types.String `tfsdk:"primary_column_name"`
	PrimaryKeyGenerationStrategy types.String `tfsdk:"primary_key_generation_strategy"`
	// Computed fields
	CreatedAt       types.String `tfsdk:"created_at"`
	CreatorUserId   types.Int64  `tfsdk:"creator_user_id"`
	CreatorUserUuid types.String `tfsdk:"creator_user_uuid"`
	ModifiedAt      types.String `tfsdk:"modified_at"`
	OrgId           types.Int64  `tfsdk:"org_id"`
}

func NewDatastoreResource() resource.Resource {
	return &datastoreResource{}
}

func (r *datastoreResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetActionsDatastoresApiV2()
	r.Auth = providerData.Auth
}

func (r *datastoreResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "datastore"
}

func (r *datastoreResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Datastore resource. This can be used to create and manage Datadog datastore.",
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
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the datastore was created.",
			},
			"creator_user_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The numeric ID of the user who created the datastore.",
			},
			"creator_user_uuid": schema.StringAttribute{
				Computed:    true,
				Description: "The UUID of the user who created the datastore.",
			},
			"modified_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the datastore was last modified.",
			},
			"org_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the organization that owns this datastore.",
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *datastoreResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *datastoreResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state datastoreModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		response.Diagnostics.AddWarning("Failed to read datastore state", "An error occurred while reading the current state")
		return
	}
	id := state.ID.ValueString()

	resp, httpResp, err := r.Api.GetDatastore(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Datastore"))
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

func (r *datastoreResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state datastoreModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		response.Diagnostics.AddWarning("Failed to read datastore plan", "An error occurred while reading the planned state")
		return
	}

	body, diags := r.buildDatastoreRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		response.Diagnostics.AddWarning("Failed to build datastore request", "An error occurred while building the create request body")
		return
	}

	resp, _, err := r.Api.CreateDatastore(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating Datastore"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	// Set the ID from create response
	if data, ok := resp.GetDataOk(); ok && data != nil {
		if id, ok := data.GetIdOk(); ok && id != nil {
			state.ID = types.StringValue(*id)
		}
	}

	// Read back the full datastore to populate all fields
	readResp, _, err := r.Api.GetDatastore(r.Auth, state.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error reading created Datastore"))
		return
	}
	r.updateState(ctx, &state, &readResp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *datastoreResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state datastoreModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		response.Diagnostics.AddWarning("Failed to read datastore plan", "An error occurred while reading the planned state")
		return
	}

	id := state.ID.ValueString()

	body, diags := r.buildDatastoreUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		response.Diagnostics.AddWarning("Failed to build datastore update request", "An error occurred while building the update request body")
		return
	}

	resp, _, err := r.Api.UpdateDatastore(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Datastore"))
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

func (r *datastoreResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state datastoreModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		response.Diagnostics.AddWarning("Failed to read datastore state", "An error occurred while reading the current state")
		return
	}

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeleteDatastore(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting datastore"))
		return
	}
}

func (r *datastoreResource) updateState(ctx context.Context, state *datastoreModel, resp *datadogV2.Datastore) {
	// Get the data wrapper
	data := resp.GetData()

	// Set ID
	if id, ok := data.GetIdOk(); ok && id != nil {
		state.ID = types.StringValue(*id)
	}

	// Get attributes
	attributes := data.GetAttributes()

	if createdAt, ok := attributes.GetCreatedAtOk(); ok && createdAt != nil {
		state.CreatedAt = types.StringValue(createdAt.String())
	}

	if creatorUserId, ok := attributes.GetCreatorUserIdOk(); ok && creatorUserId != nil {
		state.CreatorUserId = types.Int64Value(*creatorUserId)
	}

	if creatorUserUuid, ok := attributes.GetCreatorUserUuidOk(); ok && creatorUserUuid != nil {
		state.CreatorUserUuid = types.StringValue(*creatorUserUuid)
	}

	if description, ok := attributes.GetDescriptionOk(); ok && description != nil {
		state.Description = types.StringValue(*description)
	}

	if modifiedAt, ok := attributes.GetModifiedAtOk(); ok && modifiedAt != nil {
		state.ModifiedAt = types.StringValue(modifiedAt.String())
	}

	if name, ok := attributes.GetNameOk(); ok && name != nil {
		state.Name = types.StringValue(*name)
	}

	if orgId, ok := attributes.GetOrgIdOk(); ok && orgId != nil {
		state.OrgId = types.Int64Value(*orgId)
	}

	if primaryColumnName, ok := attributes.GetPrimaryColumnNameOk(); ok && primaryColumnName != nil {
		state.PrimaryColumnName = types.StringValue(*primaryColumnName)
	}

	if primaryKeyGenerationStrategy, ok := attributes.GetPrimaryKeyGenerationStrategyOk(); ok && primaryKeyGenerationStrategy != nil {
		state.PrimaryKeyGenerationStrategy = types.StringValue(string(*primaryKeyGenerationStrategy))
	}
}

func (r *datastoreResource) buildDatastoreRequestBody(ctx context.Context, state *datastoreModel) (*datadogV2.CreateAppsDatastoreRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
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

func (r *datastoreResource) buildDatastoreUpdateRequestBody(ctx context.Context, state *datastoreModel) (*datadogV2.UpdateAppsDatastoreRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
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
