package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &connectionDatasource{}

type connectionDatasource struct {
	Api  *datadogV2.ActionConnectionApi
	Auth context.Context
}

func NewDatadogConnectionDataSource() datasource.DataSource {
	return &connectionDatasource{}
}

func (d *connectionDatasource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetActionConnectionApiV2()
	d.Auth = providerData.Auth
}

func (d *connectionDatasource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "connection"
}

func (d *connectionDatasource) Schema(_ context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "A connection that can be used in Actions, including in the Workflow Automation and App Builder products.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID for Connection.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the connection",
			},
		},
		Blocks: map[string]schema.Block{
			"aws": schema.SingleNestedBlock{
				Description: "Configuration for an AWS connection",
				Blocks: map[string]schema.Block{
					"assume_role": schema.SingleNestedBlock{
						Description: "Configuration for an assume role AWS connection",
						Attributes: map[string]schema.Attribute{
							"external_id": schema.StringAttribute{
								Description: "External ID used to scope which connection can be used to assume the role",
								Computed:    true,
							},
							"principal_id": schema.StringAttribute{
								Description: "AWS account that will assume the role",
								Computed:    true,
							},
							"account_id": schema.StringAttribute{
								Description: "AWS account the connection is created for",
								Computed:    true,
							},
							"role": schema.StringAttribute{
								Description: "Role to assume",
								Computed:    true,
							},
						},
					},
				},
			},
			"http": schema.SingleNestedBlock{
				Description: "Configuration for an HTTP connection",
				Attributes: map[string]schema.Attribute{
					"base_url": schema.StringAttribute{
						Description: "Base HTTP url for the integration",
						Computed:    true,
					},
				},
				Blocks: map[string]schema.Block{
					"token_auth": schema.SingleNestedBlock{
						Description: "Configuration for an HTTP connection using token auth",
						Blocks: map[string]schema.Block{
							"token": schema.ListNestedBlock{
								Description: "Token for HTTP authentication",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"type": schema.StringAttribute{
											Description: "Type of the token. Currently only STRING is allowed.",
											Computed:    true,
										},
										"name": schema.StringAttribute{
											Description: "Token name",
											Computed:    true,
										},
										"value": schema.StringAttribute{
											Description: "Token value",
											Computed:    true,
											Sensitive:   true,
										},
									},
								},
							},
							"header": schema.ListNestedBlock{
								Description: "Header for HTTP authentication",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "Header name",
											Computed:    true,
										},
										"value": schema.StringAttribute{
											Description: "",
											Computed:    true,
										},
									},
								},
							},
							"url_parameter": schema.ListNestedBlock{
								Description: "URL parameter for HTTP authentication",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "URL parameter name",
											Computed:    true,
										},
										"value": schema.StringAttribute{
											Description: "URL parameter value",
											Computed:    true,
										},
									},
								},
							},
							"body": schema.SingleNestedBlock{
								Description: "Body for HTTP authentication",
								Attributes: map[string]schema.Attribute{
									"content_type": schema.StringAttribute{
										Description: "Content type of the body",
										Computed:    true,
									},
									"content": schema.StringAttribute{
										Description: "Serialized body content",
										Computed:    true,
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

func (d *connectionDatasource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state connectionResourceModel
	diags := request.Config.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// manually set values but we need to read this from API IRL
	state.Name = types.StringValue("a name")
	state.AWS = &awsConnectionModel{
		AssumeRole: &awsAssumeRoleConnectionModel{
			AccountID:   types.StringValue("accid"),
			Role:        types.StringValue("role"),
			ExternalID:  types.StringValue("extid"),
			PrincipalID: types.StringValue("principalid"),
		},
	}

	diags = response.State.Set(ctx, &state)
	response.Diagnostics.Append(diags...)
}
