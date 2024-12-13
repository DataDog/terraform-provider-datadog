package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &connectionResource{}
	_ resource.ResourceWithImportState = &connectionResource{}
)

type connectionResource struct {
	Api  *datadogV2.TeamsApi
	Auth context.Context
}

type connectionResourceModel struct {
	ID          types.String                  `tfsdk:"id"`
	Name        types.String                  `tfsdk:"name"`
	Integration connectionResourceIntegration `tfsdk:"integration"`
}

type connectionResourceIntegration struct {
	Type types.String `tfsdk:"type"`
}

func NewConnectionResource() resource.Resource {
	return &connectionResource{}
}

func (r *connectionResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetTeamsApiV2()
	r.Auth = providerData.Auth
}

func (r *connectionResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "connection"
}

func (r *connectionResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Required:    true,
				Description: "",
			},
		},
		Blocks: map[string]schema.Block{
			"aws": schema.SingleNestedBlock{
				Description: "",
				Blocks: map[string]schema.Block{
					"assume_role": schema.SingleNestedBlock{
						Description: "",
						Attributes: map[string]schema.Attribute{
							"external_id": schema.StringAttribute{
								Description: "",
								Optional:    true,
							},
							"principal_id": schema.StringAttribute{
								Description: "",
								Optional:    true,
							},
							"account_id": schema.StringAttribute{
								Description: "",
								Optional:    true,
							},
							"role": schema.StringAttribute{
								Description: "",
								Optional:    true,
							},
						},
					},
				},
			},
			"http": schema.SingleNestedBlock{
				Description: "",
				Attributes: map[string]schema.Attribute{
					"base_url": schema.StringAttribute{
						Description: "",
						Optional:    true,
					},
				},
				Blocks: map[string]schema.Block{
					"http_token_auth": schema.SingleNestedBlock{
						Description: "",
						Blocks: map[string]schema.Block{
							"tokens": schema.ListNestedBlock{
								Description: "",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"type": schema.StringAttribute{
											Description: "",
											Optional:    true,
										},
										"name": schema.StringAttribute{
											Description: "",
											Optional:    true,
										},
										"value": schema.StringAttribute{
											Description: "",
											Optional:    true,
										},
									},
								},
							},
							"headers": schema.ListNestedBlock{
								Description: "",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "",
											Optional:    true,
										},
										"value": schema.StringAttribute{
											Description: "",
											Optional:    true,
										},
									},
								},
							},
							"url_parameters": schema.ListNestedBlock{
								Description: "",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "",
											Optional:    true,
										},
										"value": schema.StringAttribute{
											Description: "",
											Optional:    true,
										},
									},
								},
							},
							"body": schema.SingleNestedBlock{
								Description: "",
								Attributes: map[string]schema.Attribute{
									"content_type": schema.StringAttribute{
										Description: "",
										Optional:    true,
									},
									"content": schema.StringAttribute{
										Description: "",
										Optional:    true,
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

func (r *connectionResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *connectionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state connectionResourceModel
	diags := request.Plan.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue("created ID")

	diags = response.State.Set(ctx, &state)
	response.Diagnostics.Append(diags...)
}

func (r *connectionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state connectionResourceModel
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue("read ID")

	diags = response.State.Set(ctx, &state)
	response.Diagnostics.Append(diags...)
}

func (r *connectionResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state connectionResourceModel
	diags := request.Plan.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	diags = response.State.Set(ctx, &state)
	response.Diagnostics.Append(diags...)
}

func (r *connectionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state connectionResourceModel
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// noop
}
