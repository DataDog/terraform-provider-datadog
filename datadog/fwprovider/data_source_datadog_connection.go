package fwprovider

import (
	"context"
	"fmt"
	"io"

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

	conn, httpResponse, err := d.Api.GetActionConnection(d.Auth, state.ID.ValueString())
	if err != nil {
		body, _ := io.ReadAll(httpResponse.Body)
		response.Diagnostics.AddError("Could not get connection", string(body))
		return
	}

	attributes := conn.Data.Attributes
	state.Name = types.StringValue(attributes.Name)

	if attributes.Integration.AWSIntegration != nil {
		awsAttr := attributes.Integration.AWSIntegration
		if awsAttr.GetCredentials().AWSAssumeRole == nil {
			response.Diagnostics.AddError("Unsupported connection type", "This provider only supports AWS connections of the assume role type.")
			return
		}

		state.AWS = &awsConnectionModel{
			AssumeRole: &awsAssumeRoleConnectionModel{
				AccountID:   types.StringValue(awsAttr.Credentials.AWSAssumeRole.GetAccountId()),
				Role:        types.StringValue(awsAttr.Credentials.AWSAssumeRole.GetRole()),
				ExternalID:  types.StringValue(awsAttr.Credentials.AWSAssumeRole.GetExternalId()),
				PrincipalID: types.StringValue(awsAttr.Credentials.AWSAssumeRole.GetPrincipalId()),
			},
		}
	}

	if attributes.Integration.HTTPIntegration != nil {
		httpAttr := attributes.Integration.HTTPIntegration
		// if httpAttr.GetCredentials().HTTPTokenAuth == nil {
		// 	response.Diagnostics.AddError("Unsupported connection type", "This provider only supports HTTP connections of the token auth type.")
		// 	return
		// }

		response.Diagnostics.AddWarning(fmt.Sprintf("%#v", httpAttr.GetCredentials()), "")

		tokenAuth := &httpTokenAuthConnectionModel{}
		tokens := []*httpConnectionTokenModel{}
		for _, token := range httpAttr.Credentials.HTTPTokenAuth.GetTokens() {
			tokens = append(tokens, &httpConnectionTokenModel{
				Type: types.StringValue(string(token.GetType())),
			})
		}
		if len(tokens) > 0 {
			tokenAuth.Tokens = tokens
		}

		headers := []*httpConnectionHeaderModel{}
		for _, header := range httpAttr.Credentials.HTTPTokenAuth.GetHeaders() {
			headers = append(headers, &httpConnectionHeaderModel{
				Name:  types.StringValue(header.Name),
				Value: types.StringValue(header.Value),
			})
		}
		if len(headers) > 0 {
			tokenAuth.Headers = headers
		}

		urlParams := []*httpConnectionUrlParameterModel{}
		for _, urlParam := range httpAttr.Credentials.HTTPTokenAuth.GetUrlParameters() {
			urlParams = append(urlParams, &httpConnectionUrlParameterModel{
				Name:  types.StringValue(urlParam.Name),
				Value: types.StringValue(urlParam.Value),
			})
		}
		if len(urlParams) > 0 {
			tokenAuth.URLParameters = urlParams
		}

		body := httpAttr.Credentials.HTTPTokenAuth.GetBody()
		tokenAuth.Body = &httpConnectionBodyModel{}
		if body.Content != nil {
			tokenAuth.Body.Content = types.StringValue(*body.Content)
		}
		if body.ContentType != nil {
			tokenAuth.Body.ContentType = types.StringValue(*body.ContentType)
		}

		state.HTTP = &httpConnectionModel{
			BaseURL:   types.StringValue(httpAttr.BaseUrl),
			TokenAuth: tokenAuth,
		}
	}

	diags = response.State.Set(ctx, &state)
	response.Diagnostics.Append(diags...)
}
