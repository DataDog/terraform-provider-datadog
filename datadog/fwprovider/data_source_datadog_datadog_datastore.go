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
	_ datasource.DataSource = &datadogDatadogDatastoreDataSource{}
)

type datadogDatadogDatastoreDataSource struct {
	Api  *datadogV2.ActionsDatastoresApi
	Auth context.Context
}

type datadogDatadogDatastoreDataSourceModel struct {
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

func NewDatadogDatadogDatastoreDataSource() datasource.DataSource {
	return &datadogDatadogDatastoreDataSource{}
}

func (d *datadogDatadogDatastoreDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetActionsDatastoresApiV2()
	d.Auth = providerData.Auth
}

func (d *datadogDatadogDatastoreDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "datadog_datastore"
}

func (d *datadogDatadogDatastoreDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing Datadog datadog_datastore.",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			// Query Parameters
			"datastore_id": schema.StringAttribute{
				Optional:    true,
				Description: "UPDATE ME",
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

func (d *datadogDatadogDatastoreDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state datadogDatadogDatastoreDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !state.DatadogDatastoreId.IsNull() {
		datadogDatastoreId := state.DatadogDatastoreId.ValueString()
		ddResp, _, err := d.Api.GetDatadogDatastore(d.Auth, datadogDatastoreId)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting datadog datadogDatastore"))
			return
		}

		d.updateState(ctx, &state, ddResp.Data)
	} else {

		optionalParams := datadogV2.ListDatadogDatastoresOptionalParameters{}

		ddResp, _, err := d.Api.ListDatadogDatastores(d.Auth, optionalParams)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing datadog datadogDatastore"))
			return
		}

		if len(ddResp.Data) > 1 {
			response.Diagnostics.AddError("filters returned more than one result, use more specific search criteria", "")
			return
		}
		if len(ddResp.Data) == 0 {
			response.Diagnostics.AddError("filters returned no results", "")
			return
		}

		d.updateStateFromListResponse(ctx, &state, &ddResp.Data[0])
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (d *datadogDatadogDatastoreDataSource) updateState(ctx context.Context, state *datadogDatadogDatastoreDataSourceModel, datadogDatastoreData *datadogV2.DatadogDatastore) {
	state.ID = types.StringValue(datadogDatastoreData.GetId())

	attributes := datadogDatastoreData.GetAttributes()
	state.CreatedAt = types.StringValue(attributes.GetCreatedAt().String())
	state.CreatorUserId = types.Int64Value(int64(attributes.GetCreatorUserId()))
	state.CreatorUserUuid = types.StringValue(attributes.GetCreatorUserUuid())
	state.Description = types.StringValue(attributes.GetDescription())
	state.ModifiedAt = types.StringValue(attributes.GetModifiedAt().String())
	state.Name = types.StringValue(attributes.GetName())
	state.OrgId = types.Int64Value(int64(attributes.GetOrgId()))
	state.PrimaryColumnName = types.StringValue(attributes.GetPrimaryColumnName())
	state.PrimaryKeyGenerationStrategy = types.StringValue(attributes.GetPrimaryKeyGenerationStrategy())
}

func (d *datadogDatadogDatastoreDataSource) updateStateFromListResponse(ctx context.Context, state *datadogDatadogDatastoreDataSourceModel, datadogDatastoreData *datadogV2.DatadogDatastore) {
	state.ID = types.StringValue(datadogDatastoreData.GetId())
	state.DatastoreId = types.StringValue(datadogDatastoreData.GetId())

	attributes := datadogDatastoreData.GetAttributes()
	state.CreatedAt = types.StringValue(attributes.GetCreatedAt().String())
	state.CreatorUserId = types.Int64Value(int64(attributes.GetCreatorUserId()))
	state.CreatorUserUuid = types.StringValue(attributes.GetCreatorUserUuid())
	state.Description = types.StringValue(attributes.GetDescription())
	state.ModifiedAt = types.StringValue(attributes.GetModifiedAt().String())
	state.Name = types.StringValue(attributes.GetName())
	state.OrgId = types.Int64Value(int64(attributes.GetOrgId()))
	state.PrimaryColumnName = types.StringValue(attributes.GetPrimaryColumnName())
	state.PrimaryKeyGenerationStrategy = types.StringValue(attributes.GetPrimaryKeyGenerationStrategy())
}
