package fwprovider

import (
	"context"
	"fmt"
	"io"
	"net/http"

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
	Api  *datadogV2.ActionConnectionApi
	Auth context.Context
}

type connectionResourceModel struct {
	ID   types.String         `tfsdk:"id"`
	Name types.String         `tfsdk:"name"`
	AWS  *awsConnectionModel  `tfsdk:"aws"`
	HTTP *httpConnectionModel `tfsdk:"http"`
}

type awsConnectionModel struct {
	AssumeRole *awsAssumeRoleConnectionModel `tfsdk:"assume_role"`
}

type awsAssumeRoleConnectionModel struct {
	AccountID   types.String `tfsdk:"account_id"`
	Role        types.String `tfsdk:"role"`
	ExternalID  types.String `tfsdk:"external_id"`
	PrincipalID types.String `tfsdk:"principal_id"`
}

type httpConnectionModel struct {
	BaseURL   types.String                  `tfsdk:"base_url"`
	TokenAuth *httpTokenAuthConnectionModel `tfsdk:"token_auth"`
}

type httpTokenAuthConnectionModel struct {
	Tokens        []*httpConnectionTokenModel        `tfsdk:"token"`
	Headers       []*httpConnectionHeaderModel       `tfsdk:"header"`
	URLParameters []*httpConnectionUrlParameterModel `tfsdk:"url_parameter"`
	Body          *httpConnectionBodyModel           `tfsdk:"body"`
}

type httpConnectionTokenModel struct {
	Type  types.String `tfsdk:"type"`
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type httpConnectionHeaderModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type httpConnectionUrlParameterModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type httpConnectionBodyModel struct {
	ContentType types.String `tfsdk:"content_type"`
	Content     types.String `tfsdk:"content"`
}

func NewConnectionResource() resource.Resource {
	return &connectionResource{}
}

func (r *connectionResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetActionConnectionApiV2()
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
											Description: "Type of the token. Currently only SECRET is allowed.",
											Optional:    true,
										},
										"name": schema.StringAttribute{
											Description: "Token name",
											Optional:    true,
										},
										"value": schema.StringAttribute{
											Description: "Token value",
											Optional:    true,
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
	var plan connectionResourceModel
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	createRequest, err := connectionModelToCreateApiRequest(plan)
	if err != nil {
		response.Diagnostics.AddError("Could not build create connection request", err.Error())
		return
	}

	conn, httpResponse, err := r.Api.CreateActionConnection(r.Auth, *createRequest)
	if err != nil {
		if httpResponse != nil {
			// error body may have useful info for the user
			body, err := io.ReadAll(httpResponse.Body)
			if err != nil {
				response.Diagnostics.AddError("Could not read error response", err.Error())
				return
			}
			response.Diagnostics.AddError("Could not create connection", string(body))
		} else {
			response.Diagnostics.AddError("Could not create connection", err.Error())
		}
		return
	}

	// set computed values
	plan.ID = types.StringPointerValue(conn.Data.Id)
	if plan.AWS != nil {
		plan.AWS.AssumeRole.ExternalID = types.StringPointerValue(conn.Data.Attributes.Integration.AWSIntegration.Credentials.AWSAssumeRole.ExternalId)
		plan.AWS.AssumeRole.PrincipalID = types.StringPointerValue(conn.Data.Attributes.Integration.AWSIntegration.Credentials.AWSAssumeRole.PrincipalId)
	}

	diags = response.State.Set(ctx, &plan)
	response.Diagnostics.Append(diags...)
}

func (r *connectionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state connectionResourceModel
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	conn, httpResponse, err := r.Api.GetActionConnection(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResponse != nil {
			body, err := io.ReadAll(httpResponse.Body)
			if err != nil {
				response.Diagnostics.AddError("Could not read API error response", "")
				return
			}
			response.Diagnostics.AddError("Could not get connection", string(body))
		} else {
			response.Diagnostics.AddError("Could not get connection", err.Error())
		}
		return
	}

	connModel, err := apiResponseToConnectionModel(conn)
	if err != nil {
		response.Diagnostics.AddError(err.Error(), "")
		return
	}

	diags = response.State.Set(ctx, connModel)
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

	res, err := r.Api.DeleteActionConnection(r.Auth, state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Delete connection failed", err.Error())
		return
	}

	if res.StatusCode != http.StatusNoContent {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			response.Diagnostics.AddError("Delete connection failed", "Failed to read error")
		} else {
			response.Diagnostics.AddError("Delete connection failed", string(body))
		}
	}
}

func apiResponseToConnectionModel(connection datadogV2.GetActionConnectionResponse) (*connectionResourceModel, error) {
	connModel := &connectionResourceModel{
		ID: types.StringPointerValue(connection.Data.Id),
	}

	attributes := connection.Data.Attributes
	connModel.Name = types.StringValue(attributes.Name)

	if attributes.Integration.AWSIntegration != nil {
		awsAttr := attributes.Integration.AWSIntegration
		if awsAttr.Credentials.AWSAssumeRole == nil {
			err := fmt.Errorf("this provider only supports AWS connections of the assume role type")
			return nil, err
		}

		connModel.AWS = &awsConnectionModel{
			AssumeRole: &awsAssumeRoleConnectionModel{
				AccountID:   types.StringValue(awsAttr.Credentials.AWSAssumeRole.AccountId),
				Role:        types.StringValue(awsAttr.Credentials.AWSAssumeRole.Role),
				ExternalID:  types.StringPointerValue(awsAttr.Credentials.AWSAssumeRole.ExternalId),
				PrincipalID: types.StringPointerValue(awsAttr.Credentials.AWSAssumeRole.PrincipalId),
			},
		}
	}

	if attributes.Integration.HTTPIntegration != nil {
		httpAttr := attributes.Integration.HTTPIntegration

		tokenAuth := &httpTokenAuthConnectionModel{}
		tokens := []*httpConnectionTokenModel{}
		for _, token := range httpAttr.Credentials.HTTPTokenAuth.Tokens {
			tokens = append(tokens, &httpConnectionTokenModel{
				Type: types.StringValue(string(token.Type)),
				Name: types.StringValue(token.Name),
			})
		}
		if len(tokens) > 0 {
			tokenAuth.Tokens = tokens
		}

		headers := []*httpConnectionHeaderModel{}
		for _, header := range httpAttr.Credentials.HTTPTokenAuth.Headers {
			headers = append(headers, &httpConnectionHeaderModel{
				Name:  types.StringValue(header.Name),
				Value: types.StringValue(header.Value),
			})
		}
		if len(headers) > 0 {
			tokenAuth.Headers = headers
		}

		urlParams := []*httpConnectionUrlParameterModel{}
		for _, urlParam := range httpAttr.Credentials.HTTPTokenAuth.UrlParameters {
			urlParams = append(urlParams, &httpConnectionUrlParameterModel{
				Name:  types.StringValue(urlParam.Name),
				Value: types.StringValue(urlParam.Value),
			})
		}
		if len(urlParams) > 0 {
			tokenAuth.URLParameters = urlParams
		}

		body := httpAttr.Credentials.HTTPTokenAuth.Body
		if body != nil {
			tokenAuth.Body = &httpConnectionBodyModel{}
			if body.Content != nil {
				tokenAuth.Body.Content = types.StringPointerValue(body.Content)
			}
			if body.ContentType != nil {
				tokenAuth.Body.ContentType = types.StringPointerValue(body.ContentType)
			}
		}

		connModel.HTTP = &httpConnectionModel{
			BaseURL:   types.StringValue(httpAttr.BaseUrl),
			TokenAuth: tokenAuth,
		}
	}

	return connModel, nil
}

func connectionModelToCreateApiRequest(connectionModel connectionResourceModel) (*datadogV2.CreateActionConnectionRequest, error) {
	attributes := datadogV2.NewActionConnectionAttributesWithDefaults()
	attributes.SetName(connectionModel.Name.ValueString())

	if connectionModel.AWS != nil {
		assumeRoleParams := datadogV2.NewAWSAssumeRole(
			connectionModel.AWS.AssumeRole.AccountID.ValueString(),
			connectionModel.AWS.AssumeRole.Role.ValueString(),
			datadogV2.AWSASSUMEROLETYPE_AWSASSUMEROLE,
		)

		awsIntegration := datadogV2.NewAWSIntegration(
			datadogV2.AWSAssumeRoleAsAWSCredentials(assumeRoleParams),
			datadogV2.AWSINTEGRATIONTYPE_AWS,
		)
		integration := datadogV2.AWSIntegrationAsActionConnectionIntegration(awsIntegration)
		attributes.SetIntegration(integration)
	}

	if connectionModel.HTTP != nil {
		httpTokenAuth := datadogV2.NewHTTPTokenAuth(datadogV2.HTTPTOKENAUTHTYPE_HTTPTOKENAUTH)

		tokens := connectionModel.HTTP.TokenAuth.Tokens
		for _, token := range tokens {
			tokenType, err := datadogV2.NewTokenTypeFromValue(token.Type.ValueString())
			if err != nil {
				return nil, err
			}

			tokenModel := datadogV2.NewHTTPToken(token.Name.ValueString(), *tokenType)
			tokenModel.SetValue(token.Value.ValueString())
			httpTokenAuth.Tokens = append(httpTokenAuth.Tokens, *tokenModel)
		}

		headers := connectionModel.HTTP.TokenAuth.Headers
		for _, header := range headers {
			httpTokenAuth.Headers = append(httpTokenAuth.Headers, *datadogV2.NewHTTPHeader(
				header.Name.ValueString(),
				header.Value.ValueString(),
			))
		}

		urlParams := connectionModel.HTTP.TokenAuth.URLParameters
		for _, urlParam := range urlParams {
			httpTokenAuth.UrlParameters = append(httpTokenAuth.UrlParameters, *datadogV2.NewUrlParam(
				urlParam.Name.ValueString(),
				urlParam.Value.ValueString(),
			))
		}

		if connectionModel.HTTP.TokenAuth.Body != nil {
			httpTokenAuth.Body = datadogV2.NewHTTPBody()
			if !connectionModel.HTTP.TokenAuth.Body.ContentType.IsNull() {
				httpTokenAuth.Body.SetContentType(connectionModel.HTTP.TokenAuth.Body.ContentType.ValueString())
			}
			if !connectionModel.HTTP.TokenAuth.Body.Content.IsNull() {
				httpTokenAuth.Body.SetContent(connectionModel.HTTP.TokenAuth.Body.Content.ValueString())
			}
		}

		httpCredentials := datadogV2.HTTPTokenAuthAsHTTPCredentials(httpTokenAuth)
		httpIntegration := datadogV2.NewHTTPIntegration(
			connectionModel.HTTP.BaseURL.ValueString(),
			httpCredentials,
			datadogV2.HTTPINTEGRATIONTYPE_HTTP,
		)
		integration := datadogV2.HTTPIntegrationAsActionConnectionIntegration(httpIntegration)
		attributes.SetIntegration(integration)
	}

	data := datadogV2.NewActionConnectionData(*attributes, datadogV2.ACTIONCONNECTIONDATATYPE_ACTION_CONNECTION)
	req := datadogV2.NewCreateActionConnectionRequest(*data)

	return req, nil
}
