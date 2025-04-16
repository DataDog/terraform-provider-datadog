package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ resource.Resource = &customFrameworkResource{}

type customFrameworkResource struct {
	Api  *datadogV2.SecurityMonitoringApi
	Auth context.Context
}

type customFrameworkModel struct {
	ID           types.String       `tfsdk:"id"`
	Description  types.String       `tfsdk:"description"`
	Version      types.String       `tfsdk:"version"`
	Handle       types.String       `tfsdk:"handle"`
	Name         types.String       `tfsdk:"name"`
	IconURL      types.String       `tfsdk:"icon_url"`
	Requirements []requirementModel `tfsdk:"requirements"`
}

type requirementModel struct {
	Name     types.String   `tfsdk:"name"`
	Controls []controlModel `tfsdk:"controls"`
}

type controlModel struct {
	Name    types.String `tfsdk:"name"`
	RulesID types.List   `tfsdk:"rules_id"`
}

func NewCustomFrameworkResource() resource.Resource {
	return &customFrameworkResource{}
}

func (r *customFrameworkResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "custom_framework"
}

func (r *customFrameworkResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Manages custom framework rules in Datadog.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the custom framework resource.",
				Computed:    true,
			},
			"version": schema.StringAttribute{
				Description: "The framework version.",
				Required:    true,
			},
			"handle": schema.StringAttribute{
				Description: "The framework handle.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The framework name.",
				Required:    true,
			},
			"icon_url": schema.StringAttribute{
				Description: "The URL of the icon representing the framework.",
				Optional:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the framework.",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"requirements": schema.ListNestedBlock{
				Description: "The requirements of the framework.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The name of the requirement.",
							Required:    true,
						},
					},
					Blocks: map[string]schema.Block{
						"controls": schema.ListNestedBlock{
							Description: "The controls of the requirement.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "The name of the control.",
										Required:    true,
									},
									"rules_id": schema.ListAttribute{
										Description: "The list of rules IDs for the control.",
										ElementType: types.StringType,
										Required:    true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// func (r *customFrameworkResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
// 	response.Schema = schema.Schema{
// 		Description: "Manages custom framework rules in Datadog.",
// 		Attributes: map[string]schema.Attribute{
// 			"id": schema.StringAttribute{
// 				Description: "The ID of the custom framework resource.",
// 				Computed:    true,
// 			},
// 			"version": schema.StringAttribute{
// 				Description: "The framework version.",
// 				Required:    true,
// 			},
// 			"handle": schema.StringAttribute{
// 				Description: "The framework handle.",
// 				Required:    true,
// 			},
// 			"name": schema.StringAttribute{
// 				Description: "The framework name.",
// 				Required:    true,
// 			},
// 			"icon_url": schema.StringAttribute{
// 				Description: "The URL of the icon representing the framework.",
// 				Optional:    true,
// 			},
// 			"description": schema.StringAttribute{
// 				Description: "The description of the framework.",
// 				Optional:    true,
// 			},
// 			"requirements": schema.ListNestedAttribute{
// 				Description: "The requirements of the framework.",
// 				NestedObject: schema.NestedAttributeObject{
// 					Attributes: map[string]schema.Attribute{
// 						"name": schema.StringAttribute{
// 							Description: "The name of the requirement.",
// 							Required:    true,
// 						},
// 						"controls": schema.ListNestedAttribute{
// 							Description: "The controls of the requirement.",
// 							NestedObject: schema.NestedAttributeObject{
// 								Attributes: map[string]schema.Attribute{
// 									"name": schema.StringAttribute{
// 										Description: "The name of the control.",
// 										Required:    true,
// 									},
// 									"rules_id": schema.ListAttribute{
// 										Description: "The list of rules IDs for the control.",
// 										ElementType: types.StringType,
// 										Required:    true,
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}
// }

func (r *customFrameworkResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.Auth = providerData.Auth
}

func (r *customFrameworkResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state customFrameworkModel
	diags := request.Config.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	_, _, err := r.Api.CreateCustomFramework(ctx, *buildCreateFrameworkRequest(ctx, state))
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating custom framework"))
		return
	}

	state.ID = types.StringValue(state.Handle.ValueString() + string('-') + state.Version.ValueString())
	diags = response.State.Set(ctx, &state)
	response.Diagnostics.Append(diags...)
}

func (r *customFrameworkResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state customFrameworkModel
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	_, err := r.Api.DeleteCustomFramework(ctx, state.Handle.ValueString(), state.Version.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting framework"))
		return
	}
}

func (r *customFrameworkResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state customFrameworkModel
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	data, _, err := r.Api.RetrieveCustomFramework(ctx, state.Handle.String(), state.Version.String())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error reading custom framework"))
		return
	}
	// convert data to state and set it as the new state
	if err := utils.CheckForUnparsed(data); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	state.ID = types.StringValue(data.GetData().Attributes.Handle + string('-') + data.GetData().Attributes.Version)
	state.Description = types.StringValue(data.GetData().Attributes.Description)
	state.IconURL = types.StringValue(data.GetData().Attributes.IconUrl)
	state.Version = types.StringValue(data.GetData().Attributes.Version)
	state.Handle = types.StringValue(data.GetData().Attributes.Handle)
	state.Name = types.StringValue(data.GetData().Attributes.Name)
	requirements := make([]attr.Value, len(data.GetData().Attributes.Requirements))
	for i, req := range data.GetData().Attributes.Requirements {
		controls := make([]attr.Value, len(req.Controls))
		for j, ctrl := range req.Controls {
			rulesID := make([]attr.Value, len(ctrl.RulesId))
			for k, ruleID := range ctrl.RulesId {
				rulesID[k] = types.StringValue(ruleID)
			}
			controls[j] = types.ObjectValueMust(
				map[string]attr.Type{
					"name":     types.StringType,
					"rules_id": types.ListType{ElemType: types.StringType},
				},
				map[string]attr.Value{
					"name":     types.StringValue(ctrl.Name),
					"rules_id": types.ListValueMust(types.StringType, rulesID),
				},
			)
		}
		requirements[i] = types.ObjectValueMust(
			map[string]attr.Type{
				"name": types.StringType,
				"controls": types.ListType{ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":     types.StringType,
						"rules_id": types.ListType{ElemType: types.StringType},
					},
				}},
			},
			map[string]attr.Value{
				"name": types.StringValue(req.Name),
				"controls": types.ListValueMust(types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":     types.StringType,
						"rules_id": types.ListType{ElemType: types.StringType},
					},
				}, controls),
			},
		)
	}
	// state.Requirements = types.ListValueMust(
	// 	types.ObjectType{
	// 		AttrTypes: map[string]attr.Type{
	// 			"name": types.StringType,
	// 			"controls": types.ListType{ElemType: types.ObjectType{
	// 				AttrTypes: map[string]attr.Type{
	// 					"name":     types.StringType,
	// 					"rules_id": types.ListType{ElemType: types.StringType},
	// 				},
	// 			}},
	// 		},
	// 	},
	// 	requirements,
	// )

	diags = response.State.Set(ctx, &state)
	response.Diagnostics.Append(diags...)
}

func (r *customFrameworkResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state customFrameworkModel
	diags := request.Config.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	_, _, err := r.Api.UpdateCustomFramework(ctx, state.Handle.String(), state.Version.String(), *buildUpdateFrameworkRequest(ctx, state))
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating custom framework"))
		return
	}
	diags = response.State.Set(ctx, &state)
	response.Diagnostics.Append(diags...)
}

func buildCreateFrameworkRequest(ctx context.Context, state customFrameworkModel) *datadogV2.CreateCustomFrameworkRequest {
	createFrameworkRequest := datadogV2.NewCreateCustomFrameworkRequestWithDefaults()
	description := state.Description.ValueString()
	iconURL := state.IconURL.ValueString()
	createFrameworkRequest.SetData(datadogV2.CustomFrameworkData{
		Type: "custom_framework",
		Attributes: datadogV2.CustomFrameworkDataAttributes{
			Handle:      state.Handle.ValueString(),
			Name:        state.Name.ValueString(),
			Description: &description,
			IconUrl:     &iconURL,
			Version:     state.Version.ValueString(),
			Requirements: func() []datadogV2.CustomFrameworkRequirement {
				requirements := make([]datadogV2.CustomFrameworkRequirement, len(state.Requirements))
				for i, req := range state.Requirements {
					controls := make([]datadogV2.CustomFrameworkControl, len(req.Controls))
					for j, ctrl := range req.Controls {
						rulesID := make([]string, len(ctrl.RulesID.Elements()))
						for k, ruleID := range ctrl.RulesID.Elements() {
							rulesID[k] = ruleID.(types.String).ValueString()
						}
						controls[j] = datadogV2.CustomFrameworkControl{
							Name:    ctrl.Name.ValueString(),
							RulesId: rulesID,
						}
					}
					requirements[i] = datadogV2.CustomFrameworkRequirement{
						Name:     req.Name.ValueString(),
						Controls: controls,
					}
				}
				return requirements
			}(),
		},
	})
	return createFrameworkRequest
}

func buildUpdateFrameworkRequest(ctx context.Context, state customFrameworkModel) *datadogV2.UpdateCustomFrameworkRequest {
	updateFrameworkRequest := datadogV2.NewUpdateCustomFrameworkRequestWithDefaults()
	description := state.Description.ValueString()
	iconURL := state.IconURL.ValueString()
	updateFrameworkRequest.SetData(datadogV2.CustomFrameworkData{
		Type: "custom_framework",
		Attributes: datadogV2.CustomFrameworkDataAttributes{
			Handle:      state.Handle.ValueString(),
			Name:        state.Name.ValueString(),
			Description: &description,
			IconUrl:     &iconURL,
			Version:     state.Version.ValueString(),
			Requirements: func() []datadogV2.CustomFrameworkRequirement {
				requirements := make([]datadogV2.CustomFrameworkRequirement, len(state.Requirements))
				for i, req := range state.Requirements {
					controls := make([]datadogV2.CustomFrameworkControl, len(req.Controls))
					for j, ctrl := range req.Controls {
						rulesID := make([]string, len(ctrl.RulesID.Elements()))
						for k, ruleID := range ctrl.RulesID.Elements() {
							rulesID[k] = ruleID.(types.String).ValueString()
						}
						controls[j] = datadogV2.CustomFrameworkControl{
							Name:    ctrl.Name.ValueString(),
							RulesId: rulesID,
						}
					}
					requirements[i] = datadogV2.CustomFrameworkRequirement{
						Name:     req.Name.ValueString(),
						Controls: controls,
					}
				}
				return requirements
			}(),
		},
	})
	return updateFrameworkRequest
}
