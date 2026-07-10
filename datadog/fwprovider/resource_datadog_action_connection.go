package fwprovider

import (
	"context"
	"encoding/json"
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
	ID           types.String                 `tfsdk:"id"`
	Name         types.String                 `tfsdk:"name"`
	AWS          *awsConnectionModel          `tfsdk:"aws"`
	Anthropic    *apiTokenConnectionModel     `tfsdk:"anthropic"`
	Asana        *asanaConnectionModel        `tfsdk:"asana"`
	Azure        *azureConnectionModel        `tfsdk:"azure"`
	CircleCI     *apiTokenConnectionModel     `tfsdk:"circle_ci"`
	Clickup      *apiTokenConnectionModel     `tfsdk:"clickup"`
	Cloudflare   *cloudflareConnectionModel   `tfsdk:"cloudflare"`
	ConfigCat    *configCatConnectionModel    `tfsdk:"config_cat"`
	Datadog      *datadogConnectionModel      `tfsdk:"datadog"`
	Fastly       *apiKeyConnectionModel       `tfsdk:"fastly"`
	Freshservice *freshserviceConnectionModel `tfsdk:"freshservice"`
	GCP          *gcpConnectionModel          `tfsdk:"gcp"`
	Gemini       *apiKeyConnectionModel       `tfsdk:"gemini"`
	Gitlab       *apiTokenConnectionModel     `tfsdk:"gitlab"`
	GreyNoise    *apiKeyConnectionModel       `tfsdk:"grey_noise"`
	HTTP         *httpConnectionModel         `tfsdk:"http"`
	LaunchDarkly *apiTokenConnectionModel     `tfsdk:"launch_darkly"`
	Notion       *apiTokenConnectionModel     `tfsdk:"notion"`
	Okta         *oktaConnectionModel         `tfsdk:"okta"`
	OpenAI       *apiTokenConnectionModel     `tfsdk:"openai"`
	ServiceNow   *serviceNowConnectionModel   `tfsdk:"service_now"`
	Split        *apiKeyConnectionModel       `tfsdk:"split"`
	Statsig      *apiKeyConnectionModel       `tfsdk:"statsig"`
	VirusTotal   *apiKeyConnectionModel       `tfsdk:"virus_total"`
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

type apiTokenConnectionModel struct {
	APIKey *apiTokenCredentialModel `tfsdk:"api_key"`
}

type apiTokenCredentialModel struct {
	APIToken types.String `tfsdk:"api_token"`
}

type apiKeyConnectionModel struct {
	APIKey *apiKeyCredentialModel `tfsdk:"api_key"`
}

type apiKeyCredentialModel struct {
	APIKey types.String `tfsdk:"api_key"`
}

type asanaConnectionModel struct {
	AccessToken *asanaAccessTokenCredentialModel `tfsdk:"access_token"`
}

type asanaAccessTokenCredentialModel struct {
	AccessToken types.String `tfsdk:"access_token"`
}

type azureConnectionModel struct {
	Tenant *azureTenantCredentialModel `tfsdk:"tenant"`
}

type azureTenantCredentialModel struct {
	AppClientID  types.String `tfsdk:"app_client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	CustomScopes types.String `tfsdk:"custom_scopes"`
	TenantID     types.String `tfsdk:"tenant_id"`
}

type cloudflareConnectionModel struct {
	APIToken       *cloudflareAPITokenCredentialModel       `tfsdk:"api_token"`
	GlobalAPIToken *cloudflareGlobalAPITokenCredentialModel `tfsdk:"global_api_token"`
}

type cloudflareAPITokenCredentialModel struct {
	APIToken types.String `tfsdk:"api_token"`
}

type cloudflareGlobalAPITokenCredentialModel struct {
	AuthEmail    types.String `tfsdk:"auth_email"`
	GlobalAPIKey types.String `tfsdk:"global_api_key"`
}

type configCatConnectionModel struct {
	SDKKey *configCatSDKKeyCredentialModel `tfsdk:"sdk_key"`
}

type configCatSDKKeyCredentialModel struct {
	APIPassword types.String `tfsdk:"api_password"`
	APIUsername types.String `tfsdk:"api_username"`
	SDKKey      types.String `tfsdk:"sdk_key"`
}

type datadogConnectionModel struct {
	APIKey *datadogAPIKeyCredentialModel `tfsdk:"api_key"`
}

type datadogAPIKeyCredentialModel struct {
	APIKey     types.String `tfsdk:"api_key"`
	AppKey     types.String `tfsdk:"app_key"`
	Datacenter types.String `tfsdk:"datacenter"`
	Subdomain  types.String `tfsdk:"subdomain"`
}

type freshserviceConnectionModel struct {
	APIKey *freshserviceAPIKeyCredentialModel `tfsdk:"api_key"`
}

type freshserviceAPIKeyCredentialModel struct {
	APIKey types.String `tfsdk:"api_key"`
	Domain types.String `tfsdk:"domain"`
}

type gcpConnectionModel struct {
	ServiceAccount *gcpServiceAccountCredentialModel `tfsdk:"service_account"`
}

type gcpServiceAccountCredentialModel struct {
	PrivateKey          types.String `tfsdk:"private_key"`
	ServiceAccountEmail types.String `tfsdk:"service_account_email"`
}

type oktaConnectionModel struct {
	APIToken *oktaAPITokenCredentialModel `tfsdk:"api_token"`
}

type oktaAPITokenCredentialModel struct {
	APIToken types.String `tfsdk:"api_token"`
	Domain   types.String `tfsdk:"domain"`
}

type serviceNowConnectionModel struct {
	BasicAuth *serviceNowBasicAuthCredentialModel `tfsdk:"basic_auth"`
}

type serviceNowBasicAuthCredentialModel struct {
	Instance types.String `tfsdk:"instance"`
	Password types.String `tfsdk:"password"`
	Username types.String `tfsdk:"username"`
}

type actionConnectionFieldSpec struct {
	Name        string
	Description string
	Sensitive   bool
}

type actionConnectionCredentialSpec struct {
	Name        string
	Description string
	Fields      []actionConnectionFieldSpec
}

type actionConnectionIntegrationSpec struct {
	Name        string
	Description string
	Credentials []actionConnectionCredentialSpec
}

var additionalActionConnectionSpecs = []actionConnectionIntegrationSpec{
	{
		Name: "anthropic", Description: "Configuration for an Anthropic connection",
		Credentials: []actionConnectionCredentialSpec{{
			Name: "api_key", Description: "Configuration for Anthropic API key authentication",
			Fields: []actionConnectionFieldSpec{{Name: "api_token", Description: "Anthropic API token", Sensitive: true}},
		}},
	},
	{
		Name: "asana", Description: "Configuration for an Asana connection",
		Credentials: []actionConnectionCredentialSpec{{
			Name: "access_token", Description: "Configuration for Asana access token authentication",
			Fields: []actionConnectionFieldSpec{{Name: "access_token", Description: "Asana access token", Sensitive: true}},
		}},
	},
	{
		Name: "azure", Description: "Configuration for an Azure connection",
		Credentials: []actionConnectionCredentialSpec{{
			Name: "tenant", Description: "Configuration for Azure tenant authentication",
			Fields: []actionConnectionFieldSpec{
				{Name: "app_client_id", Description: "Azure application client ID"},
				{Name: "client_secret", Description: "Azure application client secret", Sensitive: true},
				{Name: "custom_scopes", Description: "Custom scope requested when acquiring an OAuth 2 access token"},
				{Name: "tenant_id", Description: "Azure Active Directory tenant ID"},
			},
		}},
	},
	{
		Name: "circle_ci", Description: "Configuration for a CircleCI connection",
		Credentials: []actionConnectionCredentialSpec{{
			Name: "api_key", Description: "Configuration for CircleCI API key authentication",
			Fields: []actionConnectionFieldSpec{{Name: "api_token", Description: "CircleCI API token", Sensitive: true}},
		}},
	},
	{
		Name: "clickup", Description: "Configuration for a ClickUp connection",
		Credentials: []actionConnectionCredentialSpec{{
			Name: "api_key", Description: "Configuration for ClickUp API key authentication",
			Fields: []actionConnectionFieldSpec{{Name: "api_token", Description: "ClickUp API token", Sensitive: true}},
		}},
	},
	{
		Name: "cloudflare", Description: "Configuration for a Cloudflare connection",
		Credentials: []actionConnectionCredentialSpec{
			{
				Name: "api_token", Description: "Configuration for Cloudflare API token authentication",
				Fields: []actionConnectionFieldSpec{{Name: "api_token", Description: "Cloudflare API token", Sensitive: true}},
			},
			{
				Name: "global_api_token", Description: "Configuration for Cloudflare global API token authentication",
				Fields: []actionConnectionFieldSpec{
					{Name: "auth_email", Description: "Email address associated with the Cloudflare account"},
					{Name: "global_api_key", Description: "Cloudflare global API key", Sensitive: true},
				},
			},
		},
	},
	{
		Name: "config_cat", Description: "Configuration for a ConfigCat connection",
		Credentials: []actionConnectionCredentialSpec{{
			Name: "sdk_key", Description: "Configuration for ConfigCat SDK key authentication",
			Fields: []actionConnectionFieldSpec{
				{Name: "api_password", Description: "ConfigCat Public Management API password", Sensitive: true},
				{Name: "api_username", Description: "ConfigCat Public Management API username"},
				{Name: "sdk_key", Description: "ConfigCat SDK key", Sensitive: true},
			},
		}},
	},
	{
		Name: "datadog", Description: "Configuration for a Datadog connection",
		Credentials: []actionConnectionCredentialSpec{{
			Name: "api_key", Description: "Configuration for Datadog API and application key authentication",
			Fields: []actionConnectionFieldSpec{
				{Name: "api_key", Description: "Datadog API key", Sensitive: true},
				{Name: "app_key", Description: "Datadog application key", Sensitive: true},
				{Name: "datacenter", Description: "Datadog site datacenter"},
				{Name: "subdomain", Description: "Custom subdomain used for URLs generated with this connection"},
			},
		}},
	},
	{
		Name: "fastly", Description: "Configuration for a Fastly connection",
		Credentials: []actionConnectionCredentialSpec{{
			Name: "api_key", Description: "Configuration for Fastly API key authentication",
			Fields: []actionConnectionFieldSpec{{Name: "api_key", Description: "Fastly API key", Sensitive: true}},
		}},
	},
	{
		Name: "freshservice", Description: "Configuration for a Freshservice connection",
		Credentials: []actionConnectionCredentialSpec{{
			Name: "api_key", Description: "Configuration for Freshservice API key authentication",
			Fields: []actionConnectionFieldSpec{
				{Name: "api_key", Description: "Freshservice API key", Sensitive: true},
				{Name: "domain", Description: "Freshservice domain"},
			},
		}},
	},
	{
		Name: "gcp", Description: "Configuration for a Google Cloud connection",
		Credentials: []actionConnectionCredentialSpec{{
			Name: "service_account", Description: "Configuration for Google Cloud service account authentication",
			Fields: []actionConnectionFieldSpec{
				{Name: "private_key", Description: "Google Cloud service account private key", Sensitive: true},
				{Name: "service_account_email", Description: "Google Cloud service account email"},
			},
		}},
	},
	{
		Name: "gemini", Description: "Configuration for a Gemini connection",
		Credentials: []actionConnectionCredentialSpec{{
			Name: "api_key", Description: "Configuration for Gemini API key authentication",
			Fields: []actionConnectionFieldSpec{{Name: "api_key", Description: "Gemini API key", Sensitive: true}},
		}},
	},
	{
		Name: "gitlab", Description: "Configuration for a GitLab connection",
		Credentials: []actionConnectionCredentialSpec{{
			Name: "api_key", Description: "Configuration for GitLab API key authentication",
			Fields: []actionConnectionFieldSpec{{Name: "api_token", Description: "GitLab API token", Sensitive: true}},
		}},
	},
	{
		Name: "grey_noise", Description: "Configuration for a GreyNoise connection",
		Credentials: []actionConnectionCredentialSpec{{
			Name: "api_key", Description: "Configuration for GreyNoise API key authentication",
			Fields: []actionConnectionFieldSpec{{Name: "api_key", Description: "GreyNoise API key", Sensitive: true}},
		}},
	},
	{
		Name: "launch_darkly", Description: "Configuration for a LaunchDarkly connection",
		Credentials: []actionConnectionCredentialSpec{{
			Name: "api_key", Description: "Configuration for LaunchDarkly API key authentication",
			Fields: []actionConnectionFieldSpec{{Name: "api_token", Description: "LaunchDarkly API token", Sensitive: true}},
		}},
	},
	{
		Name: "notion", Description: "Configuration for a Notion connection",
		Credentials: []actionConnectionCredentialSpec{{
			Name: "api_key", Description: "Configuration for Notion API key authentication",
			Fields: []actionConnectionFieldSpec{{Name: "api_token", Description: "Notion API token", Sensitive: true}},
		}},
	},
	{
		Name: "okta", Description: "Configuration for an Okta connection",
		Credentials: []actionConnectionCredentialSpec{{
			Name: "api_token", Description: "Configuration for Okta API token authentication",
			Fields: []actionConnectionFieldSpec{
				{Name: "api_token", Description: "Okta API token", Sensitive: true},
				{Name: "domain", Description: "Okta domain"},
			},
		}},
	},
	{
		Name: "openai", Description: "Configuration for an OpenAI connection",
		Credentials: []actionConnectionCredentialSpec{{
			Name: "api_key", Description: "Configuration for OpenAI API key authentication",
			Fields: []actionConnectionFieldSpec{{Name: "api_token", Description: "OpenAI API token", Sensitive: true}},
		}},
	},
	{
		Name: "service_now", Description: "Configuration for a ServiceNow connection",
		Credentials: []actionConnectionCredentialSpec{{
			Name: "basic_auth", Description: "Configuration for ServiceNow basic authentication",
			Fields: []actionConnectionFieldSpec{
				{Name: "instance", Description: "ServiceNow instance"},
				{Name: "password", Description: "ServiceNow password", Sensitive: true},
				{Name: "username", Description: "ServiceNow username"},
			},
		}},
	},
	{
		Name: "split", Description: "Configuration for a Split connection",
		Credentials: []actionConnectionCredentialSpec{{
			Name: "api_key", Description: "Configuration for Split API key authentication",
			Fields: []actionConnectionFieldSpec{{Name: "api_key", Description: "Split API key", Sensitive: true}},
		}},
	},
	{
		Name: "statsig", Description: "Configuration for a Statsig connection",
		Credentials: []actionConnectionCredentialSpec{{
			Name: "api_key", Description: "Configuration for Statsig API key authentication",
			Fields: []actionConnectionFieldSpec{{Name: "api_key", Description: "Statsig API key", Sensitive: true}},
		}},
	},
	{
		Name: "virus_total", Description: "Configuration for a VirusTotal connection",
		Credentials: []actionConnectionCredentialSpec{{
			Name: "api_key", Description: "Configuration for VirusTotal API key authentication",
			Fields: []actionConnectionFieldSpec{{Name: "api_key", Description: "VirusTotal API key", Sensitive: true}},
		}},
	},
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
			path.MatchRoot("anthropic"),
			path.MatchRoot("asana"),
			path.MatchRoot("azure"),
			path.MatchRoot("circle_ci"),
			path.MatchRoot("clickup"),
			path.MatchRoot("cloudflare"),
			path.MatchRoot("config_cat"),
			path.MatchRoot("datadog"),
			path.MatchRoot("fastly"),
			path.MatchRoot("freshservice"),
			path.MatchRoot("gcp"),
			path.MatchRoot("gemini"),
			path.MatchRoot("gitlab"),
			path.MatchRoot("grey_noise"),
			path.MatchRoot("http"),
			path.MatchRoot("launch_darkly"),
			path.MatchRoot("notion"),
			path.MatchRoot("okta"),
			path.MatchRoot("openai"),
			path.MatchRoot("service_now"),
			path.MatchRoot("split"),
			path.MatchRoot("statsig"),
			path.MatchRoot("virus_total"),
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

	if conn.Cloudflare != nil {
		credentialCount := 0
		if conn.Cloudflare.APIToken != nil {
			credentialCount++
		}
		if conn.Cloudflare.GlobalAPIToken != nil {
			credentialCount++
		}
		if credentialCount != 1 {
			response.Diagnostics.AddAttributeError(
				path.Root("cloudflare"),
				"Exactly one Cloudflare credential type required",
				"You must specify exactly one of the api_token or global_api_token credential blocks.",
			)
			return
		}
	}

	if integrationName := missingAdditionalConnectionCredential(conn); integrationName != "" {
		response.Diagnostics.AddAttributeError(
			path.Root(integrationName),
			"Credential type required",
			"You must specify a credential type block.",
		)
	}
}

func missingAdditionalConnectionCredential(conn connectionResourceModel) string {
	switch {
	case conn.Anthropic != nil && conn.Anthropic.APIKey == nil:
		return "anthropic"
	case conn.Asana != nil && conn.Asana.AccessToken == nil:
		return "asana"
	case conn.Azure != nil && conn.Azure.Tenant == nil:
		return "azure"
	case conn.CircleCI != nil && conn.CircleCI.APIKey == nil:
		return "circle_ci"
	case conn.Clickup != nil && conn.Clickup.APIKey == nil:
		return "clickup"
	case conn.ConfigCat != nil && conn.ConfigCat.SDKKey == nil:
		return "config_cat"
	case conn.Datadog != nil && conn.Datadog.APIKey == nil:
		return "datadog"
	case conn.Fastly != nil && conn.Fastly.APIKey == nil:
		return "fastly"
	case conn.Freshservice != nil && conn.Freshservice.APIKey == nil:
		return "freshservice"
	case conn.GCP != nil && conn.GCP.ServiceAccount == nil:
		return "gcp"
	case conn.Gemini != nil && conn.Gemini.APIKey == nil:
		return "gemini"
	case conn.Gitlab != nil && conn.Gitlab.APIKey == nil:
		return "gitlab"
	case conn.GreyNoise != nil && conn.GreyNoise.APIKey == nil:
		return "grey_noise"
	case conn.LaunchDarkly != nil && conn.LaunchDarkly.APIKey == nil:
		return "launch_darkly"
	case conn.Notion != nil && conn.Notion.APIKey == nil:
		return "notion"
	case conn.Okta != nil && conn.Okta.APIToken == nil:
		return "okta"
	case conn.OpenAI != nil && conn.OpenAI.APIKey == nil:
		return "openai"
	case conn.ServiceNow != nil && conn.ServiceNow.BasicAuth == nil:
		return "service_now"
	case conn.Split != nil && conn.Split.APIKey == nil:
		return "split"
	case conn.Statsig != nil && conn.Statsig.APIKey == nil:
		return "statsig"
	case conn.VirusTotal != nil && conn.VirusTotal.APIKey == nil:
		return "virus_total"
	default:
		return ""
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

	for _, integrationSpec := range additionalActionConnectionSpecs {
		response.Schema.Blocks[integrationSpec.Name] = actionConnectionResourceBlock(integrationSpec)
	}
}

func actionConnectionResourceBlock(integrationSpec actionConnectionIntegrationSpec) schema.Block {
	credentialBlocks := make(map[string]schema.Block, len(integrationSpec.Credentials))
	for _, credentialSpec := range integrationSpec.Credentials {
		attributes := make(map[string]schema.Attribute, len(credentialSpec.Fields))
		for _, fieldSpec := range credentialSpec.Fields {
			attributes[fieldSpec.Name] = schema.StringAttribute{
				Description: fieldSpec.Description,
				Optional:    true,
				Sensitive:   fieldSpec.Sensitive,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			}
		}
		credentialBlocks[credentialSpec.Name] = schema.SingleNestedBlock{
			Description: credentialSpec.Description,
			Attributes:  attributes,
		}
	}

	return schema.SingleNestedBlock{
		Description: integrationSpec.Description,
		Blocks:      credentialBlocks,
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

		connModel.HTTP = &httpConnectionModel{
			BaseURL:   types.StringValue(httpAttr.BaseUrl),
			TokenAuth: tokenAuth,
		}
	}

	if err := setAdditionalConnectionModelFromAPI(connModel, attributes.Integration); err != nil {
		return nil, err
	}

	return connModel, nil
}

func setAdditionalConnectionModelFromAPI(connModel *connectionResourceModel, integration datadogV2.ActionConnectionIntegration) error {
	integrationJSON, err := json.Marshal(integration)
	if err != nil {
		return fmt.Errorf("could not serialize connection integration: %w", err)
	}

	var integrationData struct {
		Type        string                 `json:"type"`
		Credentials map[string]interface{} `json:"credentials"`
	}
	if err := json.Unmarshal(integrationJSON, &integrationData); err != nil {
		return fmt.Errorf("could not deserialize connection integration: %w", err)
	}

	credential := integrationData.Credentials
	switch integrationData.Type {
	case "AWS", "HTTP":
		return nil
	case "Anthropic":
		connModel.Anthropic = &apiTokenConnectionModel{APIKey: &apiTokenCredentialModel{
			APIToken: actionConnectionCredentialString(credential, "api_token"),
		}}
	case "Asana":
		connModel.Asana = &asanaConnectionModel{AccessToken: &asanaAccessTokenCredentialModel{
			AccessToken: actionConnectionCredentialString(credential, "access_token"),
		}}
	case "Azure":
		connModel.Azure = &azureConnectionModel{Tenant: &azureTenantCredentialModel{
			AppClientID:  actionConnectionCredentialString(credential, "app_client_id"),
			ClientSecret: actionConnectionCredentialString(credential, "client_secret"),
			CustomScopes: actionConnectionCredentialString(credential, "custom_scopes"),
			TenantID:     actionConnectionCredentialString(credential, "tenant_id"),
		}}
	case "CircleCI":
		connModel.CircleCI = &apiTokenConnectionModel{APIKey: &apiTokenCredentialModel{
			APIToken: actionConnectionCredentialString(credential, "api_token"),
		}}
	case "Clickup":
		connModel.Clickup = &apiTokenConnectionModel{APIKey: &apiTokenCredentialModel{
			APIToken: actionConnectionCredentialString(credential, "api_token"),
		}}
	case "Cloudflare":
		connModel.Cloudflare = &cloudflareConnectionModel{}
		credentialType, _ := credential["type"].(string)
		switch credentialType {
		case "CloudflareAPIToken":
			connModel.Cloudflare.APIToken = &cloudflareAPITokenCredentialModel{
				APIToken: actionConnectionCredentialString(credential, "api_token"),
			}
		case "CloudflareGlobalAPIToken":
			connModel.Cloudflare.GlobalAPIToken = &cloudflareGlobalAPITokenCredentialModel{
				AuthEmail:    actionConnectionCredentialString(credential, "auth_email"),
				GlobalAPIKey: actionConnectionCredentialString(credential, "global_api_key"),
			}
		default:
			return fmt.Errorf("unsupported Cloudflare credential type %q", credentialType)
		}
	case "ConfigCat":
		connModel.ConfigCat = &configCatConnectionModel{SDKKey: &configCatSDKKeyCredentialModel{
			APIPassword: actionConnectionCredentialString(credential, "api_password"),
			APIUsername: actionConnectionCredentialString(credential, "api_username"),
			SDKKey:      actionConnectionCredentialString(credential, "sdk_key"),
		}}
	case "Datadog":
		connModel.Datadog = &datadogConnectionModel{APIKey: &datadogAPIKeyCredentialModel{
			APIKey:     actionConnectionCredentialString(credential, "api_key"),
			AppKey:     actionConnectionCredentialString(credential, "app_key"),
			Datacenter: actionConnectionCredentialString(credential, "datacenter"),
			Subdomain:  actionConnectionCredentialString(credential, "subdomain"),
		}}
	case "Fastly":
		connModel.Fastly = &apiKeyConnectionModel{APIKey: &apiKeyCredentialModel{
			APIKey: actionConnectionCredentialString(credential, "api_key"),
		}}
	case "Freshservice":
		connModel.Freshservice = &freshserviceConnectionModel{APIKey: &freshserviceAPIKeyCredentialModel{
			APIKey: actionConnectionCredentialString(credential, "api_key"),
			Domain: actionConnectionCredentialString(credential, "domain"),
		}}
	case "GCP":
		connModel.GCP = &gcpConnectionModel{ServiceAccount: &gcpServiceAccountCredentialModel{
			PrivateKey:          actionConnectionCredentialString(credential, "private_key"),
			ServiceAccountEmail: actionConnectionCredentialString(credential, "service_account_email"),
		}}
	case "Gemini":
		connModel.Gemini = &apiKeyConnectionModel{APIKey: &apiKeyCredentialModel{
			APIKey: actionConnectionCredentialString(credential, "api_key"),
		}}
	case "Gitlab":
		connModel.Gitlab = &apiTokenConnectionModel{APIKey: &apiTokenCredentialModel{
			APIToken: actionConnectionCredentialString(credential, "api_token"),
		}}
	case "GreyNoise":
		connModel.GreyNoise = &apiKeyConnectionModel{APIKey: &apiKeyCredentialModel{
			APIKey: actionConnectionCredentialString(credential, "api_key"),
		}}
	case "LaunchDarkly":
		connModel.LaunchDarkly = &apiTokenConnectionModel{APIKey: &apiTokenCredentialModel{
			APIToken: actionConnectionCredentialString(credential, "api_token"),
		}}
	case "Notion":
		connModel.Notion = &apiTokenConnectionModel{APIKey: &apiTokenCredentialModel{
			APIToken: actionConnectionCredentialString(credential, "api_token"),
		}}
	case "Okta":
		connModel.Okta = &oktaConnectionModel{APIToken: &oktaAPITokenCredentialModel{
			APIToken: actionConnectionCredentialString(credential, "api_token"),
			Domain:   actionConnectionCredentialString(credential, "domain"),
		}}
	case "OpenAI":
		connModel.OpenAI = &apiTokenConnectionModel{APIKey: &apiTokenCredentialModel{
			APIToken: actionConnectionCredentialString(credential, "api_token"),
		}}
	case "ServiceNow":
		connModel.ServiceNow = &serviceNowConnectionModel{BasicAuth: &serviceNowBasicAuthCredentialModel{
			Instance: actionConnectionCredentialString(credential, "instance"),
			Password: actionConnectionCredentialString(credential, "password"),
			Username: actionConnectionCredentialString(credential, "username"),
		}}
	case "Split":
		connModel.Split = &apiKeyConnectionModel{APIKey: &apiKeyCredentialModel{
			APIKey: actionConnectionCredentialString(credential, "api_key"),
		}}
	case "Statsig":
		connModel.Statsig = &apiKeyConnectionModel{APIKey: &apiKeyCredentialModel{
			APIKey: actionConnectionCredentialString(credential, "api_key"),
		}}
	case "VirusTotal":
		connModel.VirusTotal = &apiKeyConnectionModel{APIKey: &apiKeyCredentialModel{
			APIKey: actionConnectionCredentialString(credential, "api_key"),
		}}
	default:
		return fmt.Errorf("unsupported connection integration type %q", integrationData.Type)
	}

	return nil
}

func actionConnectionCredentialString(credential map[string]interface{}, name string) types.String {
	value, ok := credential[name].(string)
	if !ok {
		return types.StringNull()
	}
	return types.StringValue(value)
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

	if connectionModel.AWS == nil && connectionModel.HTTP == nil {
		integration, err := additionalCreateActionConnectionIntegration(connectionModel)
		if err != nil {
			return nil, err
		}
		attributes.SetIntegration(*integration)
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

	if plan.AWS == nil && plan.HTTP == nil {
		integration, err := additionalUpdateActionConnectionIntegration(plan)
		if err != nil {
			return nil, err
		}
		attributes.SetIntegration(*integration)
	}

	data := datadogV2.NewActionConnectionDataUpdate(*attributes, datadogV2.ACTIONCONNECTIONDATATYPE_ACTION_CONNECTION)
	req := datadogV2.NewUpdateActionConnectionRequest(*data)

	return req, nil
}

func additionalCreateActionConnectionIntegration(connectionModel connectionResourceModel) (*datadogV2.ActionConnectionIntegration, error) {
	integrationData, err := additionalActionConnectionIntegrationData(connectionModel)
	if err != nil {
		return nil, err
	}
	integrationJSON, err := json.Marshal(integrationData)
	if err != nil {
		return nil, fmt.Errorf("could not serialize connection integration: %w", err)
	}

	var integration datadogV2.ActionConnectionIntegration
	if err := json.Unmarshal(integrationJSON, &integration); err != nil {
		return nil, fmt.Errorf("could not build connection integration: %w", err)
	}
	if integration.UnparsedObject != nil {
		return nil, fmt.Errorf("could not build connection integration for the configured credential type")
	}
	return &integration, nil
}

func additionalUpdateActionConnectionIntegration(connectionModel connectionResourceModel) (*datadogV2.ActionConnectionIntegrationUpdate, error) {
	integrationData, err := additionalActionConnectionIntegrationData(connectionModel)
	if err != nil {
		return nil, err
	}
	integrationJSON, err := json.Marshal(integrationData)
	if err != nil {
		return nil, fmt.Errorf("could not serialize connection integration update: %w", err)
	}

	var integration datadogV2.ActionConnectionIntegrationUpdate
	if err := json.Unmarshal(integrationJSON, &integration); err != nil {
		return nil, fmt.Errorf("could not build connection integration update: %w", err)
	}
	if integration.UnparsedObject != nil {
		return nil, fmt.Errorf("could not build connection integration update for the configured credential type")
	}
	return &integration, nil
}

func additionalActionConnectionIntegrationData(connectionModel connectionResourceModel) (map[string]interface{}, error) {
	switch {
	case connectionModel.Anthropic != nil && connectionModel.Anthropic.APIKey != nil:
		return actionConnectionIntegrationData("Anthropic", "AnthropicAPIKey", map[string]types.String{
			"api_token": connectionModel.Anthropic.APIKey.APIToken,
		}), nil
	case connectionModel.Asana != nil && connectionModel.Asana.AccessToken != nil:
		return actionConnectionIntegrationData("Asana", "AsanaAccessToken", map[string]types.String{
			"access_token": connectionModel.Asana.AccessToken.AccessToken,
		}), nil
	case connectionModel.Azure != nil && connectionModel.Azure.Tenant != nil:
		return actionConnectionIntegrationData("Azure", "AzureTenant", map[string]types.String{
			"app_client_id": connectionModel.Azure.Tenant.AppClientID,
			"client_secret": connectionModel.Azure.Tenant.ClientSecret,
			"custom_scopes": connectionModel.Azure.Tenant.CustomScopes,
			"tenant_id":     connectionModel.Azure.Tenant.TenantID,
		}), nil
	case connectionModel.CircleCI != nil && connectionModel.CircleCI.APIKey != nil:
		return actionConnectionIntegrationData("CircleCI", "CircleCIAPIKey", map[string]types.String{
			"api_token": connectionModel.CircleCI.APIKey.APIToken,
		}), nil
	case connectionModel.Clickup != nil && connectionModel.Clickup.APIKey != nil:
		return actionConnectionIntegrationData("Clickup", "ClickupAPIKey", map[string]types.String{
			"api_token": connectionModel.Clickup.APIKey.APIToken,
		}), nil
	case connectionModel.Cloudflare != nil && connectionModel.Cloudflare.APIToken != nil:
		return actionConnectionIntegrationData("Cloudflare", "CloudflareAPIToken", map[string]types.String{
			"api_token": connectionModel.Cloudflare.APIToken.APIToken,
		}), nil
	case connectionModel.Cloudflare != nil && connectionModel.Cloudflare.GlobalAPIToken != nil:
		return actionConnectionIntegrationData("Cloudflare", "CloudflareGlobalAPIToken", map[string]types.String{
			"auth_email":     connectionModel.Cloudflare.GlobalAPIToken.AuthEmail,
			"global_api_key": connectionModel.Cloudflare.GlobalAPIToken.GlobalAPIKey,
		}), nil
	case connectionModel.ConfigCat != nil && connectionModel.ConfigCat.SDKKey != nil:
		return actionConnectionIntegrationData("ConfigCat", "ConfigCatSDKKey", map[string]types.String{
			"api_password": connectionModel.ConfigCat.SDKKey.APIPassword,
			"api_username": connectionModel.ConfigCat.SDKKey.APIUsername,
			"sdk_key":      connectionModel.ConfigCat.SDKKey.SDKKey,
		}), nil
	case connectionModel.Datadog != nil && connectionModel.Datadog.APIKey != nil:
		return actionConnectionIntegrationData("Datadog", "DatadogAPIKey", map[string]types.String{
			"api_key":    connectionModel.Datadog.APIKey.APIKey,
			"app_key":    connectionModel.Datadog.APIKey.AppKey,
			"datacenter": connectionModel.Datadog.APIKey.Datacenter,
			"subdomain":  connectionModel.Datadog.APIKey.Subdomain,
		}), nil
	case connectionModel.Fastly != nil && connectionModel.Fastly.APIKey != nil:
		return actionConnectionIntegrationData("Fastly", "FastlyAPIKey", map[string]types.String{
			"api_key": connectionModel.Fastly.APIKey.APIKey,
		}), nil
	case connectionModel.Freshservice != nil && connectionModel.Freshservice.APIKey != nil:
		return actionConnectionIntegrationData("Freshservice", "FreshserviceAPIKey", map[string]types.String{
			"api_key": connectionModel.Freshservice.APIKey.APIKey,
			"domain":  connectionModel.Freshservice.APIKey.Domain,
		}), nil
	case connectionModel.GCP != nil && connectionModel.GCP.ServiceAccount != nil:
		return actionConnectionIntegrationData("GCP", "GCPServiceAccount", map[string]types.String{
			"private_key":           connectionModel.GCP.ServiceAccount.PrivateKey,
			"service_account_email": connectionModel.GCP.ServiceAccount.ServiceAccountEmail,
		}), nil
	case connectionModel.Gemini != nil && connectionModel.Gemini.APIKey != nil:
		return actionConnectionIntegrationData("Gemini", "GeminiAPIKey", map[string]types.String{
			"api_key": connectionModel.Gemini.APIKey.APIKey,
		}), nil
	case connectionModel.Gitlab != nil && connectionModel.Gitlab.APIKey != nil:
		return actionConnectionIntegrationData("Gitlab", "GitlabAPIKey", map[string]types.String{
			"api_token": connectionModel.Gitlab.APIKey.APIToken,
		}), nil
	case connectionModel.GreyNoise != nil && connectionModel.GreyNoise.APIKey != nil:
		return actionConnectionIntegrationData("GreyNoise", "GreyNoiseAPIKey", map[string]types.String{
			"api_key": connectionModel.GreyNoise.APIKey.APIKey,
		}), nil
	case connectionModel.LaunchDarkly != nil && connectionModel.LaunchDarkly.APIKey != nil:
		return actionConnectionIntegrationData("LaunchDarkly", "LaunchDarklyAPIKey", map[string]types.String{
			"api_token": connectionModel.LaunchDarkly.APIKey.APIToken,
		}), nil
	case connectionModel.Notion != nil && connectionModel.Notion.APIKey != nil:
		return actionConnectionIntegrationData("Notion", "NotionAPIKey", map[string]types.String{
			"api_token": connectionModel.Notion.APIKey.APIToken,
		}), nil
	case connectionModel.Okta != nil && connectionModel.Okta.APIToken != nil:
		return actionConnectionIntegrationData("Okta", "OktaAPIToken", map[string]types.String{
			"api_token": connectionModel.Okta.APIToken.APIToken,
			"domain":    connectionModel.Okta.APIToken.Domain,
		}), nil
	case connectionModel.OpenAI != nil && connectionModel.OpenAI.APIKey != nil:
		return actionConnectionIntegrationData("OpenAI", "OpenAIAPIKey", map[string]types.String{
			"api_token": connectionModel.OpenAI.APIKey.APIToken,
		}), nil
	case connectionModel.ServiceNow != nil && connectionModel.ServiceNow.BasicAuth != nil:
		return actionConnectionIntegrationData("ServiceNow", "ServiceNowBasicAuth", map[string]types.String{
			"instance": connectionModel.ServiceNow.BasicAuth.Instance,
			"password": connectionModel.ServiceNow.BasicAuth.Password,
			"username": connectionModel.ServiceNow.BasicAuth.Username,
		}), nil
	case connectionModel.Split != nil && connectionModel.Split.APIKey != nil:
		return actionConnectionIntegrationData("Split", "SplitAPIKey", map[string]types.String{
			"api_key": connectionModel.Split.APIKey.APIKey,
		}), nil
	case connectionModel.Statsig != nil && connectionModel.Statsig.APIKey != nil:
		return actionConnectionIntegrationData("Statsig", "StatsigAPIKey", map[string]types.String{
			"api_key": connectionModel.Statsig.APIKey.APIKey,
		}), nil
	case connectionModel.VirusTotal != nil && connectionModel.VirusTotal.APIKey != nil:
		return actionConnectionIntegrationData("VirusTotal", "VirusTotalAPIKey", map[string]types.String{
			"api_key": connectionModel.VirusTotal.APIKey.APIKey,
		}), nil
	default:
		return nil, fmt.Errorf("connection credential type is missing or unsupported")
	}
}

func actionConnectionIntegrationData(integrationType, credentialType string, fields map[string]types.String) map[string]interface{} {
	credential := map[string]interface{}{"type": credentialType}
	for name, value := range fields {
		if !value.IsNull() && !value.IsUnknown() {
			credential[name] = value.ValueString()
		}
	}
	return map[string]interface{}{
		"type":        integrationType,
		"credentials": credential,
	}
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

	// The API does not return SECRET token values, and may omit tokens entirely
	// from the GET response. Since token values are write-only (sensitive), we
	// must preserve token information from the prior state.
	if currentState.HTTP != nil && currentState.HTTP.TokenAuth != nil &&
		connModel.HTTP != nil && connModel.HTTP.TokenAuth != nil {
		if len(connModel.HTTP.TokenAuth.Tokens) > 0 {
			// API returned tokens — copy values from state for matching tokens
			for _, stateToken := range currentState.HTTP.TokenAuth.Tokens {
				for _, responseToken := range connModel.HTTP.TokenAuth.Tokens {
					if stateToken.Name.Equal(responseToken.Name) {
						responseToken.Value = stateToken.Value
					}
				}
			}
		} else {
			// API did not return tokens — preserve them entirely from state
			connModel.HTTP.TokenAuth.Tokens = currentState.HTTP.TokenAuth.Tokens
		}
	}

	preserveAdditionalConnectionSecrets(connModel, currentState)

	return connModel, nil
}

func preserveAdditionalConnectionSecrets(connModel *connectionResourceModel, currentState connectionResourceModel) {
	if currentState.Anthropic != nil && currentState.Anthropic.APIKey != nil &&
		connModel.Anthropic != nil && connModel.Anthropic.APIKey != nil {
		connModel.Anthropic.APIKey.APIToken = currentState.Anthropic.APIKey.APIToken
	}
	if currentState.Asana != nil && currentState.Asana.AccessToken != nil &&
		connModel.Asana != nil && connModel.Asana.AccessToken != nil {
		connModel.Asana.AccessToken.AccessToken = currentState.Asana.AccessToken.AccessToken
	}
	if currentState.Azure != nil && currentState.Azure.Tenant != nil &&
		connModel.Azure != nil && connModel.Azure.Tenant != nil {
		connModel.Azure.Tenant.ClientSecret = currentState.Azure.Tenant.ClientSecret
	}
	if currentState.CircleCI != nil && currentState.CircleCI.APIKey != nil &&
		connModel.CircleCI != nil && connModel.CircleCI.APIKey != nil {
		connModel.CircleCI.APIKey.APIToken = currentState.CircleCI.APIKey.APIToken
	}
	if currentState.Clickup != nil && currentState.Clickup.APIKey != nil &&
		connModel.Clickup != nil && connModel.Clickup.APIKey != nil {
		connModel.Clickup.APIKey.APIToken = currentState.Clickup.APIKey.APIToken
	}
	if currentState.Cloudflare != nil && connModel.Cloudflare != nil {
		if currentState.Cloudflare.APIToken != nil && connModel.Cloudflare.APIToken != nil {
			connModel.Cloudflare.APIToken.APIToken = currentState.Cloudflare.APIToken.APIToken
		}
		if currentState.Cloudflare.GlobalAPIToken != nil && connModel.Cloudflare.GlobalAPIToken != nil {
			connModel.Cloudflare.GlobalAPIToken.GlobalAPIKey = currentState.Cloudflare.GlobalAPIToken.GlobalAPIKey
		}
	}
	if currentState.ConfigCat != nil && currentState.ConfigCat.SDKKey != nil &&
		connModel.ConfigCat != nil && connModel.ConfigCat.SDKKey != nil {
		connModel.ConfigCat.SDKKey.APIPassword = currentState.ConfigCat.SDKKey.APIPassword
		connModel.ConfigCat.SDKKey.SDKKey = currentState.ConfigCat.SDKKey.SDKKey
	}
	if currentState.Datadog != nil && currentState.Datadog.APIKey != nil &&
		connModel.Datadog != nil && connModel.Datadog.APIKey != nil {
		connModel.Datadog.APIKey.APIKey = currentState.Datadog.APIKey.APIKey
		connModel.Datadog.APIKey.AppKey = currentState.Datadog.APIKey.AppKey
	}
	if currentState.Fastly != nil && currentState.Fastly.APIKey != nil &&
		connModel.Fastly != nil && connModel.Fastly.APIKey != nil {
		connModel.Fastly.APIKey.APIKey = currentState.Fastly.APIKey.APIKey
	}
	if currentState.Freshservice != nil && currentState.Freshservice.APIKey != nil &&
		connModel.Freshservice != nil && connModel.Freshservice.APIKey != nil {
		connModel.Freshservice.APIKey.APIKey = currentState.Freshservice.APIKey.APIKey
	}
	if currentState.GCP != nil && currentState.GCP.ServiceAccount != nil &&
		connModel.GCP != nil && connModel.GCP.ServiceAccount != nil {
		connModel.GCP.ServiceAccount.PrivateKey = currentState.GCP.ServiceAccount.PrivateKey
	}
	if currentState.Gemini != nil && currentState.Gemini.APIKey != nil &&
		connModel.Gemini != nil && connModel.Gemini.APIKey != nil {
		connModel.Gemini.APIKey.APIKey = currentState.Gemini.APIKey.APIKey
	}
	if currentState.Gitlab != nil && currentState.Gitlab.APIKey != nil &&
		connModel.Gitlab != nil && connModel.Gitlab.APIKey != nil {
		connModel.Gitlab.APIKey.APIToken = currentState.Gitlab.APIKey.APIToken
	}
	if currentState.GreyNoise != nil && currentState.GreyNoise.APIKey != nil &&
		connModel.GreyNoise != nil && connModel.GreyNoise.APIKey != nil {
		connModel.GreyNoise.APIKey.APIKey = currentState.GreyNoise.APIKey.APIKey
	}
	if currentState.LaunchDarkly != nil && currentState.LaunchDarkly.APIKey != nil &&
		connModel.LaunchDarkly != nil && connModel.LaunchDarkly.APIKey != nil {
		connModel.LaunchDarkly.APIKey.APIToken = currentState.LaunchDarkly.APIKey.APIToken
	}
	if currentState.Notion != nil && currentState.Notion.APIKey != nil &&
		connModel.Notion != nil && connModel.Notion.APIKey != nil {
		connModel.Notion.APIKey.APIToken = currentState.Notion.APIKey.APIToken
	}
	if currentState.Okta != nil && currentState.Okta.APIToken != nil &&
		connModel.Okta != nil && connModel.Okta.APIToken != nil {
		connModel.Okta.APIToken.APIToken = currentState.Okta.APIToken.APIToken
	}
	if currentState.OpenAI != nil && currentState.OpenAI.APIKey != nil &&
		connModel.OpenAI != nil && connModel.OpenAI.APIKey != nil {
		connModel.OpenAI.APIKey.APIToken = currentState.OpenAI.APIKey.APIToken
	}
	if currentState.ServiceNow != nil && currentState.ServiceNow.BasicAuth != nil &&
		connModel.ServiceNow != nil && connModel.ServiceNow.BasicAuth != nil {
		connModel.ServiceNow.BasicAuth.Password = currentState.ServiceNow.BasicAuth.Password
	}
	if currentState.Split != nil && currentState.Split.APIKey != nil &&
		connModel.Split != nil && connModel.Split.APIKey != nil {
		connModel.Split.APIKey.APIKey = currentState.Split.APIKey.APIKey
	}
	if currentState.Statsig != nil && currentState.Statsig.APIKey != nil &&
		connModel.Statsig != nil && connModel.Statsig.APIKey != nil {
		connModel.Statsig.APIKey.APIKey = currentState.Statsig.APIKey.APIKey
	}
	if currentState.VirusTotal != nil && currentState.VirusTotal.APIKey != nil &&
		connModel.VirusTotal != nil && connModel.VirusTotal.APIKey != nil {
		connModel.VirusTotal.APIKey.APIKey = currentState.VirusTotal.APIKey.APIKey
	}
}
