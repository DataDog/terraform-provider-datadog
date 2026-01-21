package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &organizationDataSource{}
)

// NewOrganizationSettingsDataSource creates a new organization settings data source
func NewOrganizationSettingsDataSource() datasource.DataSource {
	return &organizationDataSource{}
}

// organizationDataSourceModel represents the Terraform state for the organization data source
type organizationDataSourceModel struct {
	ID          types.String                `tfsdk:"id"`
	Name        types.String                `tfsdk:"name"`
	PublicId    types.String                `tfsdk:"public_id"`
	Description types.String                `tfsdk:"description"`
	Settings    []organizationSettingsModel `tfsdk:"settings"`
}

// organizationSettingsModel represents the organization settings
type organizationSettingsModel struct {
	PrivateWidgetShare         types.Bool                        `tfsdk:"private_widget_share"`
	Saml                       []samlModel                       `tfsdk:"saml"`
	SamlAutocreateAccessRole   types.String                      `tfsdk:"saml_autocreate_access_role"`
	SamlAutocreateUsersDomains []samlAutocreateUsersDomainsModel `tfsdk:"saml_autocreate_users_domains"`
	SamlCanBeEnabled           types.Bool                        `tfsdk:"saml_can_be_enabled"`
	SamlIdpEndpoint            types.String                      `tfsdk:"saml_idp_endpoint"`
	SamlIdpInitiatedLogin      []samlIdpInitiatedLoginModel      `tfsdk:"saml_idp_initiated_login"`
	SamlIdpMetadataUploaded    types.Bool                        `tfsdk:"saml_idp_metadata_uploaded"`
	SamlLoginUrl               types.String                      `tfsdk:"saml_login_url"`
	SamlStrictMode             []samlStrictModeModel             `tfsdk:"saml_strict_mode"`
}

// samlModel represents SAML configuration
type samlModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

// samlAutocreateUsersDomainsModel represents SAML auto-create users domains configuration
type samlAutocreateUsersDomainsModel struct {
	Domains []types.String `tfsdk:"domains"`
	Enabled types.Bool     `tfsdk:"enabled"`
}

// samlIdpInitiatedLoginModel represents SAML IdP initiated login configuration
type samlIdpInitiatedLoginModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

// samlStrictModeModel represents SAML strict mode configuration
type samlStrictModeModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

// organizationDataSource is the data source implementation
type organizationDataSource struct {
	Api  *datadogV1.OrganizationsApi
	Auth context.Context
}

// Configure sets up the data source with provider data
func (d *organizationDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetOrganizationsApiV1()
	d.Auth = providerData.Auth
}

// Metadata returns the data source type name
func (d *organizationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "organization_settings"
}

// Schema defines the data source schema
func (d *organizationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about your Datadog organization.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Description: "Name of the organization.",
				Computed:    true,
			},
			"public_id": schema.StringAttribute{
				Description: "The public_id of the organization.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the organization.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"settings": schema.ListNestedBlock{
				Description: "Organization settings.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"private_widget_share": schema.BoolAttribute{
							Description: "Whether or not the organization users can share widgets outside of Datadog.",
							Computed:    true,
						},
						"saml_autocreate_access_role": schema.StringAttribute{
							Description: "The access role of the user. Options are `st` (standard user), `adm` (admin user), or `ro` (read-only user). Allowed enum values: `st`, `adm`, `ro`, `ERROR`.",
							Computed:    true,
						},
						"saml_can_be_enabled": schema.BoolAttribute{
							Description: "Whether or not SAML can be enabled for this organization.",
							Computed:    true,
						},
						"saml_idp_endpoint": schema.StringAttribute{
							Description: "Identity provider endpoint for SAML authentication.",
							Computed:    true,
						},
						"saml_idp_metadata_uploaded": schema.BoolAttribute{
							Description: "Whether or not a SAML identity provider metadata file was provided to the Datadog organization.",
							Computed:    true,
						},
						"saml_login_url": schema.StringAttribute{
							Description: "URL for SAML logging.",
							Computed:    true,
						},
					},
					Blocks: map[string]schema.Block{
						"saml": schema.ListNestedBlock{
							Description: "SAML properties.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"enabled": schema.BoolAttribute{
										Description: "Whether or not SAML is enabled for this organization.",
										Computed:    true,
									},
								},
							},
						},
						"saml_autocreate_users_domains": schema.ListNestedBlock{
							Description: "List of domains where the SAML automated user creation is enabled.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"domains": schema.ListAttribute{
										Description: "List of domains where the SAML automated user creation is enabled.",
										Computed:    true,
										ElementType: types.StringType,
									},
									"enabled": schema.BoolAttribute{
										Description: "Whether or not the automated user creation based on SAML domain is enabled.",
										Computed:    true,
									},
								},
							},
						},
						"saml_idp_initiated_login": schema.ListNestedBlock{
							Description: "Whether or not a SAML identity provider metadata file was provided to the Datadog organization.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"enabled": schema.BoolAttribute{
										Description: "Whether or not a SAML identity provider metadata file was provided to the Datadog organization.",
										Computed:    true,
									},
								},
							},
						},
						"saml_strict_mode": schema.ListNestedBlock{
							Description: "Whether or not the SAML strict mode is enabled. If true, all users must log in with SAML.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"enabled": schema.BoolAttribute{
										Description: "Whether or not the SAML strict mode is enabled. If true, all users must log in with SAML.",
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

// Read retrieves organization information from the Datadog API
func (d *organizationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state organizationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// List organizations accessible with current credentials
	orgsResp, httpResp, err := d.Api.ListOrgs(d.Auth)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error getting organization"), ""))
		return
	}
	if err := utils.CheckForUnparsed(orgsResp); err != nil {
		resp.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	// Get organizations list
	orgs := orgsResp.GetOrgs()
	if len(orgs) == 0 {
		resp.Diagnostics.AddError(
			"No organization found",
			"No organization was found for the current API credentials.",
		)
		return
	}

	// Use the first organization (current organization for the API credentials)
	org := orgs[0]

	// Set basic attributes
	state.ID = types.StringValue(org.GetPublicId())
	state.Name = types.StringValue(org.GetName())
	state.PublicId = types.StringValue(org.GetPublicId())
	state.Description = types.StringValue(org.GetDescription())

	// Map settings
	settings := org.GetSettings()
	settingsModel := organizationSettingsModel{
		PrivateWidgetShare:       types.BoolValue(settings.GetPrivateWidgetShare()),
		SamlAutocreateAccessRole: types.StringValue(string(settings.GetSamlAutocreateAccessRole())),
		SamlCanBeEnabled:         types.BoolValue(settings.GetSamlCanBeEnabled()),
		SamlIdpEndpoint:          types.StringValue(settings.GetSamlIdpEndpoint()),
		SamlIdpMetadataUploaded:  types.BoolValue(settings.GetSamlIdpMetadataUploaded()),
		SamlLoginUrl:             types.StringValue(settings.GetSamlLoginUrl()),
	}

	// Map SAML settings
	samlSettings := settings.GetSaml()
	settingsModel.Saml = []samlModel{
		{Enabled: types.BoolValue(samlSettings.GetEnabled())},
	}

	// Map SAML autocreate users domains
	samlAutocreateUsersDomains := settings.GetSamlAutocreateUsersDomains()
	domains := samlAutocreateUsersDomains.GetDomains()
	domainsList := make([]types.String, len(domains))
	for i, domain := range domains {
		domainsList[i] = types.StringValue(domain)
	}
	settingsModel.SamlAutocreateUsersDomains = []samlAutocreateUsersDomainsModel{
		{
			Domains: domainsList,
			Enabled: types.BoolValue(samlAutocreateUsersDomains.GetEnabled()),
		},
	}

	// Map SAML IdP initiated login
	samlIdpInitiatedLogin := settings.GetSamlIdpInitiatedLogin()
	settingsModel.SamlIdpInitiatedLogin = []samlIdpInitiatedLoginModel{
		{Enabled: types.BoolValue(samlIdpInitiatedLogin.GetEnabled())},
	}

	// Map SAML strict mode
	samlStrictMode := settings.GetSamlStrictMode()
	settingsModel.SamlStrictMode = []samlStrictModeModel{
		{Enabled: types.BoolValue(samlStrictMode.GetEnabled())},
	}

	state.Settings = []organizationSettingsModel{settingsModel}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
