package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datadogDatastoreDataSource{}
)

type datadogDatastoreDataSource struct {
	Api  *datadogV2.ActionsDatastoresApi
	Auth context.Context
}

type datadogDatastoreDataSourceModel struct {
	// Datasource ID
	ID types.String `tfsdk:"id"`

	// Query Parameters
	DatastoreId types.String `tfsdk:"datastore_id"`

	// Computed values
	CreatedAt                    types.String `tfsdk:"created_at"`
	CreatorUserId                types.Int64  `tfsdk:"creator_user_id"`
	CreatorUserUuid              types.String `tfsdk:"creator_user_uuid"`
	Description                  types.String `tfsdk:"description"`
	ModifiedAt                   types.String `tfsdk:"modified_at"`
	Name                         types.String `tfsdk:"name"`
	OrgId                        types.Int64  `tfsdk:"org_id"`
	PrimaryColumnName            types.String `tfsdk:"primary_column_name"`
	PrimaryKeyGenerationStrategy types.String `tfsdk:"primary_key_generation_strategy"`
}

func NewDatadogDatastoreDataSource() datasource.DataSource {
	return &datadogDatastoreDataSource{}
}

func (d *datadogDatastoreDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetActionsDatastoresApiV2()
	d.Auth = providerData.Auth
}

func (d *datadogDatastoreDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "datastore"
}

func (d *datadogDatastoreDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing Datadog datastore.",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			// Query Parameters
			"datastore_id": schema.StringAttribute{
				Optional:    true,
				Description: "The unique identifier of the datastore to retrieve. If not specified, returns a single datastore from the list.",
			},
			// Computed values
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
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "A human-readable description about the datastore.",
			},
			"modified_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the datastore was last modified.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The display name of the datastore.",
			},
			"org_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the organization that owns this datastore.",
			},
			"primary_column_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the primary key column for this datastore. Primary column names:   - Must abide by both [PostgreSQL naming conventions](https://www.postgresql.org/docs/7.0/syntax525.htm)   - Cannot exceed 63 characters",
			},
			"primary_key_generation_strategy": schema.StringAttribute{
				Computed:    true,
				Description: "Can be set to `uuid` to automatically generate primary keys when new items are added. Default value is `none`, which requires you to supply a primary key for each new item.",
			},
		},
	}
}

func (d *datadogDatastoreDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state datadogDatastoreDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !state.DatastoreId.IsNull() {
		datastoreId := state.DatastoreId.ValueString()
		ddResp, _, err := d.Api.GetDatastore(d.Auth, datastoreId)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting datastore"))
			return
		}

		d.updateState(ctx, &state, ddResp.GetData())
	} else {
		ddResp, _, err := d.Api.ListDatastores(d.Auth)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing datastores"))
			return
		}

		data := ddResp.GetData()
		if len(data) > 1 {
			response.Diagnostics.AddError("filters returned more than one result, use more specific search criteria", "")
			return
		}
		if len(data) == 0 {
			response.Diagnostics.AddError("filters returned no results", "")
			return
		}

		d.updateStateFromListResponse(ctx, &state, &data[0])
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (d *datadogDatastoreDataSource) updateState(ctx context.Context, state *datadogDatastoreDataSourceModel, datastoreData datadogV2.DatastoreData) {
	if id, ok := datastoreData.GetIdOk(); ok && id != nil {
		state.ID = types.StringValue(*id)
	}

	attributes := datastoreData.GetAttributes()

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

func (d *datadogDatastoreDataSource) updateStateFromListResponse(ctx context.Context, state *datadogDatastoreDataSourceModel, datastoreData *datadogV2.DatastoreData) {
	if id, ok := datastoreData.GetIdOk(); ok && id != nil {
		state.ID = types.StringValue(*id)
		state.DatastoreId = types.StringValue(*id)
	}

	attributes := datastoreData.GetAttributes()

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
