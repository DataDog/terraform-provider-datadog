package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &deploymentRuleResource{}
	_ resource.ResourceWithImportState = &deploymentRuleResource{}
)

type deploymentRuleResource struct {
	Api  *datadogV2.DeploymentGatesApi
	Auth context.Context
}

type deploymentRuleModel struct {
	ID      types.String                `tfsdk:"id"`
	GateID  types.String                `tfsdk:"gate_id"`
	DryRun  types.Bool                  `tfsdk:"dry_run"`
	Name    types.String                `tfsdk:"name"`
	Type    types.String                `tfsdk:"type"`
	Options *deploymentRuleOptionsModel `tfsdk:"options"`
}

func NewDeploymentRuleResource() resource.Resource {
	return &deploymentRuleResource{}
}

func (r *deploymentRuleResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetDeploymentGatesApiV2()
	r.Auth = providerData.Auth
}

func (r *deploymentRuleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "deployment_rule"
}

func (r *deploymentRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog DeploymentRule resource. This can be used to create and manage Datadog deployment_rule.",
		Attributes: map[string]schema.Attribute{
			"gate_id": schema.StringAttribute{
				Required:    true,
				Description: "The deployment gate ID that this rule belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"dry_run": schema.BoolAttribute{
				Optional:    true,
				Description: "The `attributes` `dry_run`.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The `attributes` `name`.",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The `attributes` `type`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"options": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"duration": schema.Int64Attribute{
						Optional:    true,
						Description: "The duration for the rule.",
					},
					"query": schema.StringAttribute{
						Optional:    true,
						Description: "The query for monitor rules.",
					},
					"excluded_resources": schema.ListAttribute{
						Optional:    true,
						Description: "Resources to exclude from faulty deployment detection.",
						ElementType: types.StringType,
					},
				},
			},
		},
	}
}

func (r *deploymentRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *deploymentRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state deploymentRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	gateId := state.GateID.ValueString()
	id := state.ID.ValueString()

	resp, httpResp, err := r.Api.GetDeploymentRule(r.Auth, gateId, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving DeploymentRule"))
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

func (r *deploymentRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state deploymentRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Validate that rule type matches options
	response.Diagnostics.Append(r.validateTypeOptionsMatch(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	gateId := state.GateID.ValueString()

	body, diags := r.buildDeploymentRuleRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateDeploymentRule(r.Auth, gateId, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving DeploymentRule"))
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

func (r *deploymentRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state deploymentRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Validate that rule type matches options
	response.Diagnostics.Append(r.validateTypeOptionsMatch(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	gateId := state.GateID.ValueString()

	id := state.ID.ValueString()

	body, diags := r.buildDeploymentRuleUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateDeploymentRule(r.Auth, gateId, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving DeploymentRule"))
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

func (r *deploymentRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state deploymentRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	gateId := state.GateID.ValueString()

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeleteDeploymentRule(r.Auth, gateId, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting deployment_rule"))
		return
	}
}

func (r *deploymentRuleResource) updateState(ctx context.Context, state *deploymentRuleModel, resp *datadogV2.DeploymentRuleResponse) {
	data := resp.GetData()
	state.ID = types.StringValue(data.GetId())

	attributes := data.GetAttributes()

	if dryRun, ok := attributes.GetDryRunOk(); ok {
		state.DryRun = types.BoolValue(*dryRun)
	}

	if gateId, ok := attributes.GetGateIdOk(); ok {
		state.GateID = types.StringValue(*gateId)
	}

	if name, ok := attributes.GetNameOk(); ok {
		state.Name = types.StringValue(*name)
	}

	if typeVar, ok := attributes.GetTypeOk(); ok {
		state.Type = types.StringValue(string(*typeVar))
	}

	// Handle options from the response
	if options, ok := attributes.GetOptionsOk(); ok {
		if state.Options == nil {
			state.Options = &deploymentRuleOptionsModel{
				ExcludedResources: types.ListNull(types.StringType),
				Duration:          types.Int64Null(),
				Query:             types.StringNull(),
			}
		}

		// Handle options based on rule type like the data source does
		if state.Type.ValueString() == "faulty_deployment_detection" {
			if fddOptions := options.DeploymentRuleOptionsFaultyDeploymentDetection; fddOptions != nil {
				if duration, ok := fddOptions.GetDurationOk(); ok {
					state.Options.Duration = types.Int64PointerValue(duration)
				}
				if excludedResources, ok := fddOptions.GetExcludedResourcesOk(); ok {
					if excludedResources != nil && len(*excludedResources) > 0 {
						elements := make([]attr.Value, len(*excludedResources))
						for i, resource := range *excludedResources {
							elements[i] = types.StringValue(resource)
						}
						state.Options.ExcludedResources, _ = types.ListValue(types.StringType, elements)
					}
				}
			}
		} else if state.Type.ValueString() == "monitor" {
			if monitorOptions := options.DeploymentRuleOptionsMonitor; monitorOptions != nil {
				if duration, ok := monitorOptions.GetDurationOk(); ok {
					state.Options.Duration = types.Int64PointerValue(duration)
				}
				if query, ok := monitorOptions.GetQueryOk(); ok {
					state.Options.Query = types.StringValue(*query)
				}
			}
		}
	}

}

func (r *deploymentRuleResource) buildDeploymentRuleRequestBody(ctx context.Context, state *deploymentRuleModel) (*datadogV2.CreateDeploymentRuleParams, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewCreateDeploymentRuleParamsDataAttributesWithDefaults()

	if !state.DryRun.IsNull() {
		attributes.SetDryRun(state.DryRun.ValueBool())
	}
	if !state.Name.IsNull() {
		attributes.SetName(state.Name.ValueString())
	}
	if !state.Type.IsNull() {
		attributes.SetType(state.Type.ValueString())
	}

	// Handle options based on rule type
	if state.Options != nil {
		options := datadogV2.DeploymentRulesOptions{}

		if state.Type.ValueString() == "faulty_deployment_detection" {
			fddOptions := state.Options
			options.DeploymentRuleOptionsFaultyDeploymentDetection = &datadogV2.DeploymentRuleOptionsFaultyDeploymentDetection{}

			if !fddOptions.Duration.IsNull() {
				options.DeploymentRuleOptionsFaultyDeploymentDetection.Duration = fddOptions.Duration.ValueInt64Pointer()
			}
			if !fddOptions.ExcludedResources.IsNull() {
				var excludedResources []string
				diags.Append(fddOptions.ExcludedResources.ElementsAs(ctx, &excludedResources, false)...)
				if !diags.HasError() {
					options.DeploymentRuleOptionsFaultyDeploymentDetection.ExcludedResources = excludedResources
				}
			}
		} else if state.Type.ValueString() == "monitor" {
			monitorOptions := state.Options
			options.DeploymentRuleOptionsMonitor = &datadogV2.DeploymentRuleOptionsMonitor{}

			if !monitorOptions.Duration.IsNull() {
				options.DeploymentRuleOptionsMonitor.Duration = monitorOptions.Duration.ValueInt64Pointer()
			}
			if !monitorOptions.Query.IsNull() {
				options.DeploymentRuleOptionsMonitor.Query = monitorOptions.Query.ValueString()
			}
		}

		attributes.SetOptions(options)
	}

	req := datadogV2.NewCreateDeploymentRuleParamsWithDefaults()
	req.Data = datadogV2.NewCreateDeploymentRuleParamsDataWithDefaults()
	req.Data.Type = "deployment_rule"
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *deploymentRuleResource) buildDeploymentRuleUpdateRequestBody(ctx context.Context, state *deploymentRuleModel) (*datadogV2.UpdateDeploymentRuleParams, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewUpdateDeploymentRuleParamsDataAttributesWithDefaults()

	if !state.DryRun.IsNull() {
		attributes.SetDryRun(state.DryRun.ValueBool())
	}
	if !state.Name.IsNull() {
		attributes.SetName(state.Name.ValueString())
	}

	// Handle options based on rule type
	if state.Options != nil {
		options := datadogV2.DeploymentRulesOptions{}

		if state.Type.ValueString() == "faulty_deployment_detection" {
			fddOptions := state.Options
			options.DeploymentRuleOptionsFaultyDeploymentDetection = &datadogV2.DeploymentRuleOptionsFaultyDeploymentDetection{}

			if !fddOptions.Duration.IsNull() {
				options.DeploymentRuleOptionsFaultyDeploymentDetection.Duration = fddOptions.Duration.ValueInt64Pointer()
			}
			if !fddOptions.ExcludedResources.IsNull() {
				var excludedResources []string
				diags.Append(fddOptions.ExcludedResources.ElementsAs(ctx, &excludedResources, false)...)
				if !diags.HasError() {
					options.DeploymentRuleOptionsFaultyDeploymentDetection.ExcludedResources = excludedResources
				}
			}
		} else if state.Type.ValueString() == "monitor" {
			monitorOptions := state.Options
			options.DeploymentRuleOptionsMonitor = &datadogV2.DeploymentRuleOptionsMonitor{}

			if !monitorOptions.Duration.IsNull() {
				options.DeploymentRuleOptionsMonitor.Duration = monitorOptions.Duration.ValueInt64Pointer()
			}
			if !monitorOptions.Query.IsNull() {
				options.DeploymentRuleOptionsMonitor.Query = monitorOptions.Query.ValueString()
			}
		}

		attributes.SetOptions(options)
	}

	req := datadogV2.NewUpdateDeploymentRuleParamsWithDefaults()
	req.Data = *datadogV2.NewUpdateDeploymentRuleParamsDataWithDefaults()
	req.Data.Type = "deployment_rule"
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *deploymentRuleResource) validateTypeOptionsMatch(ctx context.Context, state *deploymentRuleModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if state.Options == nil {
		return diags
	}

	ruleType := state.Type.ValueString()

	// Check for faulty_deployment_detection specific options
	hasFddOptions :=
		!state.Options.ExcludedResources.IsNull()

	// Check for monitor specific options
	hasMonitorOptions := !state.Options.Query.IsNull()

	if ruleType == "faulty_deployment_detection" && hasMonitorOptions {
		diags.AddError(
			"Invalid options for deployment rule type",
			"Rule type 'faulty_deployment_detection' cannot use monitor options (query). "+
				"Use faulty deployment detection options instead: duration, excluded_resources.",
		)
	} else if ruleType == "monitor" && hasFddOptions {
		diags.AddError(
			"Invalid options for deployment rule type",
			"Rule type 'monitor' cannot use faulty deployment detection options (excluded_resources). "+
				"Use monitor options instead: duration, query.",
		)
	}

	return diags
}
