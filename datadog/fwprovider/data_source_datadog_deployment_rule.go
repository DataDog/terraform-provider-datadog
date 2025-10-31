package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datadogDeploymentRuleDataSource{}
)

type datadogDeploymentRuleDataSource struct {
	Api  *datadogV2.DeploymentGatesApi
	Auth context.Context
}

type datadogDeploymentRuleDataSourceModel struct {
	ID types.String `tfsdk:"id"`

	// Computed values
	GateId    types.String `tfsdk:"gate_id"`
	CreatedAt types.String `tfsdk:"created_at"`
	DryRun    types.Bool   `tfsdk:"dry_run"`
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"`
	UpdatedAt types.String `tfsdk:"updated_at"`

	Options *deploymentRuleOptionsModel `tfsdk:"options"`
}

type createdByModel struct {
	Handle types.String `tfsdk:"handle"`
	Id     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
}

type deploymentRuleOptionsModel struct {
	ExcludedResources types.List   `tfsdk:"excluded_resources"`
	Duration          types.Int64  `tfsdk:"duration"`
	Query             types.String `tfsdk:"query"`
}

type updatedByModel struct {
	Handle types.String `tfsdk:"handle"`
	Id     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
}

func NewDatadogDeploymentRuleDataSource() datasource.DataSource {
	return &datadogDeploymentRuleDataSource{}
}

func (d *datadogDeploymentRuleDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetDeploymentGatesApiV2()
	d.Auth = providerData.Auth
}

func (d *datadogDeploymentRuleDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "deployment_rule"
}

func (d *datadogDeploymentRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing Datadog deployment_rule.",
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
			"gate_id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the deployment gate.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The `attributes` `name`.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The `attributes` `type`.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The `attributes` `updated_at`.",
			},
		},
		Blocks: map[string]schema.Block{
			"options": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{},
				Blocks: map[string]schema.Block{
					"deployment_rule_options_faulty_deployment_detection": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"duration": schema.Int64Attribute{
								Computed:    true,
								Description: "The wait time for faulty deployment detection.",
							},
							"excluded_resources": schema.ListAttribute{
								Computed:    true,
								Description: "Resources to exclude from faulty deployment detection.",
								ElementType: types.StringType,
							},
						},
					},
					"deployment_rule_options_monitor": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"duration": schema.Int64Attribute{
								Computed:    true,
								Description: "The duration for the monitor.",
							},
							"query": schema.StringAttribute{
								Computed:    true,
								Description: "The query for the monitor.",
							},
						},
					},
				},
			},
		},
	}
}

func (d *datadogDeploymentRuleDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state datadogDeploymentRuleDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !state.ID.IsNull() {
		deploymentRuleId := state.ID.ValueString()
		ddResp, _, err := d.Api.GetDeploymentRule(d.Auth, state.GateId.String(), deploymentRuleId)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting datadog deploymentRule"))
			return
		}

		d.updateState(ctx, &state, &ddResp)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (d *datadogDeploymentRuleDataSource) updateState(ctx context.Context, state *datadogDeploymentRuleDataSourceModel, deploymentRuleData *datadogV2.DeploymentRuleResponse) {
	data := deploymentRuleData.GetData()
	attributes := data.GetAttributes()

	state.ID = types.StringValue(data.GetId())

	state.CreatedAt = types.StringValue(attributes.GetCreatedAt().String())
	state.DryRun = types.BoolValue(attributes.GetDryRun())
	state.GateId = types.StringValue(attributes.GetGateId())
	state.Name = types.StringValue(attributes.GetName())
	state.Type = types.StringValue(string(attributes.GetType()))
	state.UpdatedAt = types.StringValue(attributes.GetUpdatedAt().String())
	if attributes.GetType() == "faulty_deployment_detection" {
		options := attributes.GetOptions().DeploymentRuleOptionsFaultyDeploymentDetection
		state.Options = &deploymentRuleOptionsModel{
			ExcludedResources: stringSliceToTerraformList(options.ExcludedResources),
			Duration:          types.Int64Value(options.GetDuration()),
		}
	} else {
		options := attributes.GetOptions().DeploymentRuleOptionsMonitor
		state.Options = &deploymentRuleOptionsModel{
			Query:    types.StringValue(options.GetQuery()),
			Duration: types.Int64Value(options.GetDuration()),
		}
	}
}

func stringSliceToTerraformList(input []string) types.List {
	elems := make([]attr.Value, len(input))
	for i, v := range input {
		elems[i] = types.StringValue(v)
	}

	listValue, diags := types.ListValue(types.StringType, elems)
	if diags.HasError() {
		return types.ListNull(types.StringType)
	}

	return listValue
}
