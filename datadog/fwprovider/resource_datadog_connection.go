package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure        = &connectionResource{}
	_ resource.ResourceWithImportState      = &connectionResource{}
	_ resource.ResourceWithConfigValidators = &connectionResource{}
	_ resource.ResourceWithValidateConfig   = &connectionResource{}
)

type connectionResource struct {
	Api  *datadogV2.TeamsApi
	Auth context.Context
}

type connectionResourceModel struct {
	ID   types.String         `tfsdk:"id"`
	Name types.String         `tfsdk:"name"`
	AWS  *awsConnectionModel  `tfsdk:"aws"`
	HTTP *httpConnectionModel `tfsdk:"http"`
}

type awsConnectionModel struct {
	AssumeRole *awsAssumeRoleModel `tfsdk:"assume_role"`
}

type awsAssumeRoleModel struct {
	AccountID   types.String `tfsdk:"account_id"`
	Role        types.String `tfsdk:"role"`
	ExternalID  types.String `tfsdk:"external_id"`
	PrincipalID types.String `tfsdk:"principal_id"`
}

type httpConnectionModel struct {
	BaseURL   types.String        `tfsdk:"base_url"`
	TokenAuth *httpTokenAuthModel `tfsdk:"token_auth"`
}

type httpTokenAuthModel struct {
	Tokens        []*tokenModel        `tfsdk:"token"`
	Headers       []*headerModel       `tfsdk:"header"`
	URLParameters []*urlParameterModel `tfsdk:"url_parameter"`
	Body          *bodyModel           `tfsdk:"body"`
}

type tokenModel struct {
	Type  types.String `tfsdk:"type"`
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type headerModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type urlParameterModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type bodyModel struct {
	ContentType types.String `tfsdk:"content_type"`
	Content     types.String `tfsdk:"content"`
}

func NewConnectionResource() resource.Resource {
	return &connectionResource{}
}

func (r *connectionResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetTeamsApiV2()
	r.Auth = providerData.Auth
}

// contains simple validations that can be done by the framework
func (r *connectionResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.Conflicting(
			path.MatchRoot("aws"),
			path.MatchRoot("http"),
		),
	}
}

// contains more complex validations that we need because the Schema definition isn't expressive enough for us
func (r *connectionResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var conn connectionResourceModel

	diags := request.Config.Get(ctx, &conn)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if conn.AWS == nil && conn.HTTP == nil {
		response.Diagnostics.AddAttributeError(
			path.Root(""),
			"Integration type required",
			"You must specify an AWS or HTTP block.",
		)
		return
	}

	if conn.AWS != nil {
		if conn.AWS.AssumeRole == nil {
			response.Diagnostics.AddAttributeError(
				path.Root("aws"),
				"AWS credential type required",
				"You must specify a credential type block.",
			)
			return
		}

		if isStringEmpty(conn.AWS.AssumeRole.AccountID) {
			response.Diagnostics.AddAttributeError(
				path.Root("aws").AtName("assume_role").AtName("account_id"),
				"AWS account_id required",
				"You must specify an AWS account ID.",
			)
		}

		if isStringEmpty(conn.AWS.AssumeRole.Role) {
			response.Diagnostics.AddAttributeError(
				path.Root("aws").AtName("assume_role").AtName("role"),
				"AWS role required",
				"You must specify an AWS role.",
			)
		}
	}

	if conn.HTTP != nil {
		if isStringEmpty(conn.HTTP.BaseURL) {
			response.Diagnostics.AddAttributeError(
				path.Root("http").AtName("base_url"),
				"Base URL required",
				"You must specify a base URL for this connection.",
			)
		}

		if conn.HTTP.TokenAuth == nil {
			response.Diagnostics.AddAttributeError(
				path.Root("http"),
				"HTTP credential type required",
				"You must specify a credential type block.",
			)
			return
		}

		if len(conn.HTTP.TokenAuth.Tokens) == 0 &&
			len(conn.HTTP.TokenAuth.Headers) == 0 &&
			len(conn.HTTP.TokenAuth.URLParameters) == 0 &&
			conn.HTTP.TokenAuth.Body == nil {
			response.Diagnostics.AddAttributeError(
				path.Root("http").AtName("token_auth"),
				"Credential information required",
				"You must specify at least one of: tokens, headers, URL parameters, body.",
			)
			return
		}

		for i, token := range conn.HTTP.TokenAuth.Tokens {
			if isStringEmpty(token.Type) {
				response.Diagnostics.AddAttributeError(
					path.Root("http").AtName("token_auth").AtName("token").AtListIndex(i).AtName("type"),
					"Token type required",
					"You must specify a token type",
				)
			}

			if isStringEmpty(token.Name) {
				response.Diagnostics.AddAttributeError(
					path.Root("http").AtName("token_auth").AtName("token").AtListIndex(i).AtName("name"),
					"Token name required",
					"You must specify a token name",
				)
			}

			if isStringEmpty(token.Value) {
				response.Diagnostics.AddAttributeError(
					path.Root("http").AtName("token_auth").AtName("token").AtListIndex(i).AtName("value"),
					"Token value required",
					"You must specify a token value",
				)
			}
		}

		for i, header := range conn.HTTP.TokenAuth.Headers {
			if isStringEmpty(header.Name) {
				response.Diagnostics.AddAttributeError(
					path.Root("http").AtName("token_auth").AtName("header").AtListIndex(i).AtName("name"),
					"Header name required",
					"You must specify a header name",
				)
			}

			if isStringEmpty(header.Value) {
				response.Diagnostics.AddAttributeError(
					path.Root("http").AtName("token_auth").AtName("header").AtListIndex(i).AtName("value"),
					"Header value required",
					"You must specify a header value",
				)
			}
		}

		for i, param := range conn.HTTP.TokenAuth.URLParameters {
			if isStringEmpty(param.Name) {
				response.Diagnostics.AddAttributeError(
					path.Root("http").AtName("token_auth").AtName("url_parameter").AtListIndex(i).AtName("name"),
					"URL parameter name required",
					"You must specify a URL parameter name",
				)
			}

			if isStringEmpty(param.Value) {
				response.Diagnostics.AddAttributeError(
					path.Root("http").AtName("token_auth").AtName("url_parameter").AtListIndex(i).AtName("value"),
					"URL parameter value required",
					"You must specify a URL parameter value",
				)
			}
		}

		if conn.HTTP.TokenAuth.Body != nil {
			if isStringEmpty(conn.HTTP.TokenAuth.Body.ContentType) {
				response.Diagnostics.AddAttributeError(
					path.Root("http").AtName("token_auth").AtName("body").AtName("content_type"),
					"Body content type required",
					"You must specify a body content type",
				)
			}

			if isStringEmpty(conn.HTTP.TokenAuth.Body.Content) {
				response.Diagnostics.AddAttributeError(
					path.Root("http").AtName("token_auth").AtName("body").AtName("content"),
					"Body content required",
					"You must specify body content",
				)
			}
		}
	}
}

func isStringEmpty(str types.String) bool {
	return str.IsNull() || str.ValueString() == ""
}

func (r *connectionResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "connection"
}

func (r *connectionResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "A connection that can be used in Actions, including in the Workflow Automation and App Builder products.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Required:    true,
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
								Optional:    true,
							},
							"role": schema.StringAttribute{
								Description: "Role to assume",
								Optional:    true,
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
						Optional:    true,
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
											Optional:    true,
										},
										"name": schema.StringAttribute{
											Description: "Token name",
											Optional:    true,
										},
										"value": schema.StringAttribute{
											Description: "Token value",
											Optional:    true,
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
											Optional:    true,
										},
										"value": schema.StringAttribute{
											Description: "",
											Optional:    true,
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
											Optional:    true,
										},
										"value": schema.StringAttribute{
											Description: "URL parameter value",
											Optional:    true,
										},
									},
								},
							},
							"body": schema.SingleNestedBlock{
								Description: "Body for HTTP authentication",
								Attributes: map[string]schema.Attribute{
									"content_type": schema.StringAttribute{
										Description: "Content type of the body",
										Optional:    true,
									},
									"content": schema.StringAttribute{
										Description: "Serialized body content",
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
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *connectionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state connectionResourceModel
	diags := request.Plan.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue("created ID")
	state.AWS.AssumeRole.ExternalID = types.StringValue("extid")
	state.AWS.AssumeRole.PrincipalID = types.StringValue("princid")

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
