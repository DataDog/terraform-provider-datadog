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
	_ datasource.DataSource = &datadogDeploymentGateDataSource{}
)

type datadogDeploymentGateDataSource struct {
	Api  *datadogV2.DeploymentGatesApi
	Auth context.Context
}

type datadogDeploymentGateDataSourceModel struct {
	// Datasource ID
	ID types.String `tfsdk:"id"`
	// Computed values
	CreatedAt  types.String                  `tfsdk:"created_at"`
	DryRun     types.Bool                    `tfsdk:"dry_run"`
	Env        types.String                  `tfsdk:"env"`
	Identifier types.String                  `tfsdk:"identifier"`
	Service    types.String                  `tfsdk:"service"`
	UpdatedAt  types.String                  `tfsdk:"updated_at"`
	CreatedBy  *createdByDeploymentGateModel `tfsdk:"created_by"`
	UpdatedBy  *updatedByDeploymentGateModel `tfsdk:"updated_by"`
}

type createdByDeploymentGateModel struct {
	Handle types.String `tfsdk:"handle"`
	Id     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
}

type updatedByDeploymentGateModel struct {
	Handle types.String `tfsdk:"handle"`
	Id     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
}

func NewDatadogDeploymentGateDataSource() datasource.DataSource {
	return &datadogDeploymentGateDataSource{}
}

func (d *datadogDeploymentGateDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetDeploymentGatesApiV2()
	d.Auth = providerData.Auth
}

func (d *datadogDeploymentGateDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "deployment_gate"
}

func (d *datadogDeploymentGateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing Datadog deployment_gate.",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			// Computed values
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The `attributes` `created_at`.",
			},
			"dry_run": schema.BoolAttribute{
				Computed:    true,
				Description: "The `attributes` `dry_run`.",
			},
			"env": schema.StringAttribute{
				Computed:    true,
				Description: "The `attributes` `env`.",
			},
			"identifier": schema.StringAttribute{
				Computed:    true,
				Description: "The `attributes` `identifier`.",
			},
			"service": schema.StringAttribute{
				Computed:    true,
				Description: "The `attributes` `service`.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The `attributes` `updated_at`.",
			},
		},
		Blocks: map[string]schema.Block{
			// Computed values
			"created_by": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"handle": schema.StringAttribute{
						Computed:    true,
						Description: "The `created_by` `handle`.",
					},
					"id": schema.StringAttribute{
						Computed:    true,
						Description: "The `created_by` `id`.",
					},
					"name": schema.StringAttribute{
						Computed:    true,
						Description: "The `created_by` `name`.",
					},
				},
			},
			"updated_by": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"handle": schema.StringAttribute{
						Computed:    true,
						Description: "The `updated_by` `handle`.",
					},
					"id": schema.StringAttribute{
						Computed:    true,
						Description: "The `updated_by` `id`.",
					},
					"name": schema.StringAttribute{
						Computed:    true,
						Description: "The `updated_by` `name`.",
					},
				},
			},
		},
	}
}

func (d *datadogDeploymentGateDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state datadogDeploymentGateDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !state.ID.IsNull() {
		deploymentGateId := state.ID.ValueString()
		ddResp, _, err := d.Api.GetDeploymentGate(d.Auth, deploymentGateId)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting datadog deploymentGate"))
			return
		}

		d.updateState(ctx, &state, &ddResp)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (d *datadogDeploymentGateDataSource) updateState(ctx context.Context, state *datadogDeploymentGateDataSourceModel, deploymentGateData *datadogV2.DeploymentGateResponse) {
	state.ID = types.StringValue(*deploymentGateData.Data.Id)

	attributes := deploymentGateData.Data.GetAttributes()

	state.CreatedAt = types.StringValue(attributes.GetCreatedAt().String())
	state.DryRun = types.BoolValue(attributes.GetDryRun())
	state.Env = types.StringValue(attributes.GetEnv())
	state.Identifier = types.StringValue(attributes.GetIdentifier())
	state.Service = types.StringValue(attributes.GetService())
	state.UpdatedAt = types.StringValue(attributes.GetUpdatedAt().String())
}
