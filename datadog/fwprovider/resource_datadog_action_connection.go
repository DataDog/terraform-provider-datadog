package fwprovider

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure        = &actionConnectionResource{}
	_ resource.ResourceWithImportState      = &actionConnectionResource{}
	_ resource.ResourceWithConfigValidators = &actionConnectionResource{}
	_ resource.ResourceWithValidateConfig   = &actionConnectionResource{}
)

type actionConnectionResource struct {
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

func NewActionConnectionResource() resource.Resource {
	return &actionConnectionResource{}
}

func (r *actionConnectionResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetActionConnectionApiV2()
	r.Auth = providerData.Auth
}

// contains simple validations that can be done by the framework
func (r *actionConnectionResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("aws"),
			path.MatchRoot("http"),
		),
	}
}

// contains more complex validations that we need because the Schema definition isn't expressive enough for us
func (r *actionConnectionResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var conn connectionResourceModel
	diags := request.Config.Get(ctx, &conn)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if conn.AWS != nil && conn.AWS.AssumeRole == nil {
		response.Diagnostics.AddAttributeError(
			path.Root("aws"),
			"AWS credential type required",
			"You must specify a credential type block.",
		)
		return
	}

	if conn.HTTP != nil && conn.HTTP.TokenAuth == nil {
		response.Diagnostics.AddAttributeError(
			path.Root("http"),
			"HTTP credential type required",
			"You must specify a credential type block.",
		)
		return
	}
}

func (r *actionConnectionResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "action_connection"
}

func (r *actionConnectionResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "A connection that can be used in Actions, including in the Workflow Automation and App Builder products. This resource requires a [registered application key](https://registry.terraform.io/providers/DataDog/datadog/latest/docs/resources/app_key_registration).",
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
								Description: "External ID that specifies which connection can be used to assume the role",
								Computed:    true,
							},
							"principal_id": schema.StringAttribute{
								Description: "AWS account that will assume the role",
								Computed:    true,
							},
							"account_id": schema.StringAttribute{
								Description: "AWS account that the connection is created for",
								Optional:    true,
								Validators: []validator.String{
									stringvalidator.LengthAtLeast(1),
								},
							},
							"role": schema.StringAttribute{
								Description: "Role to assume",
								Optional:    true,
								Validators: []validator.String{
									stringvalidator.LengthAtLeast(1),
								},
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
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
					},
				},
				Blocks: map[string]schema.Block{
					"token_auth": schema.SingleNestedBlock{
						Description: "Configuration for an HTTP connection that uses token auth",
						Blocks: map[string]schema.Block{
							"token": schema.ListNestedBlock{
								Description: "Token for HTTP authentication",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"type": schema.StringAttribute{
											Description: "Token type",
											Optional:    true,
											Validators: []validator.String{
												stringvalidator.OneOf("SECRET"),
											},
										},
										"name": schema.StringAttribute{
											Description: "Token name",
											Optional:    true,
											Validators: []validator.String{
												stringvalidator.LengthAtLeast(1),
											},
										},
										"value": schema.StringAttribute{
											Description: "Token value",
											Optional:    true,
											Sensitive:   true,
											Validators: []validator.String{
												stringvalidator.LengthAtLeast(1),
											},
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
											Validators: []validator.String{
												stringvalidator.LengthAtLeast(1),
											},
										},
										"value": schema.StringAttribute{
											Description: "",
											Optional:    true,
											Validators: []validator.String{
												stringvalidator.LengthAtLeast(1),
											},
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
											Validators: []validator.String{
												stringvalidator.LengthAtLeast(1),
											},
										},
										"value": schema.StringAttribute{
											Description: "URL parameter value",
											Optional:    true,
											Validators: []validator.String{
												stringvalidator.LengthAtLeast(1),
											},
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
										Validators: []validator.String{
											stringvalidator.LengthAtLeast(1),
										},
									},
									"content": schema.StringAttribute{
										Description: "Serialized body content",
										Optional:    true,
										Validators: []validator.String{
											stringvalidator.LengthAtLeast(1),
										},
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

func (r *actionConnectionResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *actionConnectionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
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

func (r *actionConnectionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state connectionResourceModel
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	connModel, err := readConnection(r.Auth, r.Api, state.ID.ValueString(), state)
	if err != nil {
		response.Diagnostics.AddError("Could not read connection", err.Error())
		return
	}

	diags = response.State.Set(ctx, connModel)
	response.Diagnostics.Append(diags...)
}

func (r *actionConnectionResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan connectionResourceModel
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// current state is required so we can detect what's been deleted
	var oldState connectionResourceModel
	diags = request.State.Get(ctx, &oldState)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	updateRequest, err := connectionModelToUpdateApiRequest(plan, oldState)
	if err != nil {
		response.Diagnostics.AddError("Could not build update connection request", err.Error())
		return
	}

	res, httpResponse, err := r.Api.UpdateActionConnection(r.Auth, plan.ID.ValueString(), *updateRequest)
	if err != nil {
		if httpResponse != nil {
			// error body may have useful info for the user
			body, err := io.ReadAll(httpResponse.Body)
			if err != nil {
				response.Diagnostics.AddError("Could not read error response", err.Error())
				return
			}
			response.Diagnostics.AddError("Could not update connection", string(body))
		} else {
			response.Diagnostics.AddError("Could not update connection", err.Error())
		}
		return
	}

	// set computed values
	if plan.AWS != nil {
		plan.AWS.AssumeRole.ExternalID = types.StringPointerValue(res.Data.Attributes.Integration.AWSIntegration.Credentials.AWSAssumeRole.ExternalId)
		plan.AWS.AssumeRole.PrincipalID = types.StringPointerValue(res.Data.Attributes.Integration.AWSIntegration.Credentials.AWSAssumeRole.PrincipalId)
	}

	diags = response.State.Set(ctx, &plan)
	response.Diagnostics.Append(diags...)
}

func (r *actionConnectionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
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

		if httpAttr.Credentials.HTTPTokenAuth != nil {
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
				if body.Content == nil && body.ContentType == nil {
					tokenAuth.Body = nil
				} else {
					tokenAuth.Body = &httpConnectionBodyModel{}
					tokenAuth.Body.Content = types.StringPointerValue(body.Content)
					tokenAuth.Body.ContentType = types.StringPointerValue(body.ContentType)
				}
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

			tokenModel := datadogV2.NewHTTPToken(token.Name.ValueString(), *tokenType, token.Value.ValueString())
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

		httpTokenAuth.Body = datadogV2.NewHTTPBody()
		if connectionModel.HTTP.TokenAuth.Body != nil {
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

func connectionModelToUpdateApiRequest(plan, oldState connectionResourceModel) (*datadogV2.UpdateActionConnectionRequest, error) {
	attributes := datadogV2.NewActionConnectionAttributesUpdate()
	attributes.SetName(plan.Name.ValueString())

	if plan.AWS != nil {
		assumeRoleParams := datadogV2.NewAWSAssumeRoleUpdate(datadogV2.AWSASSUMEROLETYPE_AWSASSUMEROLE)
		assumeRoleParams.SetAccountId(plan.AWS.AssumeRole.AccountID.ValueString())
		assumeRoleParams.SetRole(plan.AWS.AssumeRole.Role.ValueString())

		awsIntegration := datadogV2.NewAWSIntegrationUpdate(datadogV2.AWSINTEGRATIONTYPE_AWS)
		awsIntegration.SetCredentials(datadogV2.AWSAssumeRoleUpdateAsAWSCredentialsUpdate(assumeRoleParams))
		integration := datadogV2.AWSIntegrationUpdateAsActionConnectionIntegrationUpdate(awsIntegration)
		attributes.SetIntegration(integration)
	}

	if plan.HTTP != nil {
		httpTokenAuth := datadogV2.NewHTTPTokenAuthUpdate(datadogV2.HTTPTOKENAUTHTYPE_HTTPTOKENAUTH)

		buildHttpDeletions(plan, oldState, httpTokenAuth)

		for _, token := range plan.HTTP.TokenAuth.Tokens {
			tokenType, err := datadogV2.NewTokenTypeFromValue(token.Type.ValueString())
			if err != nil {
				return nil, err
			}

			tokenModel := datadogV2.NewHTTPTokenUpdate(token.Name.ValueString(), *tokenType, token.Value.ValueString())
			httpTokenAuth.Tokens = append(httpTokenAuth.Tokens, *tokenModel)
		}

		for _, header := range plan.HTTP.TokenAuth.Headers {
			headerUpdate := datadogV2.NewHTTPHeaderUpdate(header.Name.ValueString())
			headerUpdate.SetValue(header.Value.ValueString())
			httpTokenAuth.Headers = append(httpTokenAuth.Headers, *headerUpdate)
		}

		for _, urlParam := range plan.HTTP.TokenAuth.URLParameters {
			paramUpdate := datadogV2.NewUrlParamUpdate(urlParam.Name.ValueString())
			paramUpdate.SetValue(urlParam.Value.ValueString())
			httpTokenAuth.UrlParameters = append(httpTokenAuth.UrlParameters, *paramUpdate)
		}

		httpTokenAuth.Body = datadogV2.NewHTTPBody()
		if plan.HTTP.TokenAuth.Body != nil {
			if !plan.HTTP.TokenAuth.Body.ContentType.IsNull() {
				httpTokenAuth.Body.SetContentType(plan.HTTP.TokenAuth.Body.ContentType.ValueString())
			}
			if !plan.HTTP.TokenAuth.Body.Content.IsNull() {
				httpTokenAuth.Body.SetContent(plan.HTTP.TokenAuth.Body.Content.ValueString())
			}
		}

		httpCredentials := datadogV2.HTTPTokenAuthUpdateAsHTTPCredentialsUpdate(httpTokenAuth)
		httpIntegration := datadogV2.NewHTTPIntegrationUpdate(datadogV2.HTTPINTEGRATIONTYPE_HTTP)
		httpIntegration.SetBaseUrl(plan.HTTP.BaseURL.ValueString())
		httpIntegration.SetCredentials(httpCredentials)

		integration := datadogV2.HTTPIntegrationUpdateAsActionConnectionIntegrationUpdate(httpIntegration)
		attributes.SetIntegration(integration)
	}

	data := datadogV2.NewActionConnectionDataUpdate(*attributes, datadogV2.ACTIONCONNECTIONDATATYPE_ACTION_CONNECTION)
	req := datadogV2.NewUpdateActionConnectionRequest(*data)

	return req, nil
}

// The connections API handles deletions of tokens, headers, and URL params with a "deleted" flag instead of
// just exclusion from the request body as is common in a PUT request. So some more work is required to
// build the API request to handle deletions. This function does that work, comparing the current state (oldState)
// to the proposed state (plan) and detecting what to mark for deletion.
func buildHttpDeletions(plan, oldState connectionResourceModel, updateModel *datadogV2.HTTPTokenAuthUpdate) {
	deletedTokens := []*httpConnectionTokenModel{}
	for _, token := range oldState.HTTP.TokenAuth.Tokens {
		foundToken := false
		for _, planToken := range plan.HTTP.TokenAuth.Tokens {
			if planToken.Name.Equal(token.Name) {
				foundToken = true
				break
			}
		}
		if !foundToken {
			deletedTokens = append(deletedTokens, token)
		}
	}

	for _, deletedToken := range deletedTokens {
		tokenUpdate := datadogV2.NewHTTPTokenUpdate(deletedToken.Name.ValueString(), datadogV2.TOKENTYPE_SECRET, deletedToken.Value.ValueString())
		tokenUpdate.SetValue("")
		tokenUpdate.SetDeleted(true)
		updateModel.Tokens = append(updateModel.Tokens, *tokenUpdate)
	}

	deletedHeaders := []*httpConnectionHeaderModel{}
	for _, header := range oldState.HTTP.TokenAuth.Headers {
		foundHeader := false
		for _, planHeader := range plan.HTTP.TokenAuth.Headers {
			if planHeader.Name.Equal(header.Name) {
				foundHeader = true
				break
			}
		}
		if !foundHeader {
			deletedHeaders = append(deletedHeaders, header)
		}
	}

	for _, deletedHeader := range deletedHeaders {
		headerUpdate := datadogV2.NewHTTPHeaderUpdate(deletedHeader.Name.ValueString())
		headerUpdate.SetValue("")
		headerUpdate.SetDeleted(true)
		updateModel.Headers = append(updateModel.Headers, *headerUpdate)
	}

	deletedUrlParams := []*httpConnectionUrlParameterModel{}
	for _, param := range oldState.HTTP.TokenAuth.URLParameters {
		foundParam := false
		for _, planParam := range plan.HTTP.TokenAuth.URLParameters {
			if planParam.Name.Equal(param.Name) {
				foundParam = true
				break
			}
		}
		if !foundParam {
			deletedUrlParams = append(deletedUrlParams, param)
		}
	}

	for _, deletedParam := range deletedUrlParams {
		paramUpdate := datadogV2.NewUrlParamUpdate(deletedParam.Name.ValueString())
		paramUpdate.SetValue("")
		paramUpdate.SetDeleted(true)
		updateModel.UrlParameters = append(updateModel.UrlParameters, *paramUpdate)
	}
}

// Read logic is shared between data source and resource
func readConnection(authCtx context.Context, api *datadogV2.ActionConnectionApi, id string, currentState connectionResourceModel) (*connectionResourceModel, error) {
	conn, httpResponse, err := api.GetActionConnection(authCtx, id)
	if err != nil {
		if httpResponse != nil {
			body, err := io.ReadAll(httpResponse.Body)
			if err != nil {
				return nil, fmt.Errorf("could not read error response")
			}
			return nil, fmt.Errorf("%s", body)
		}
		return nil, err
	}

	if _, ok := conn.GetDataOk(); !ok {
		return nil, fmt.Errorf("connection not found")
	}

	connModel, err := apiResponseToConnectionModel(conn)
	if err != nil {
		return nil, err
	}

	// The API response does not include the token value, so this code gets it from the state.
	// This is used to determine whether the token value changed since the last update.
	if currentState.HTTP != nil {
		for _, stateToken := range currentState.HTTP.TokenAuth.Tokens {
			for _, responseToken := range connModel.HTTP.TokenAuth.Tokens {
				if stateToken.Name.Equal(responseToken.Name) {
					responseToken.Value = stateToken.Value
				}
			}
		}
	}

	return connModel, nil
}
