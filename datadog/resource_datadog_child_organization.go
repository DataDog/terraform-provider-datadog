package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceDatadogChildOrganization() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog Child Organization resource. This can be used to create Datadog Child Organizations. To manage created organization use `datadog_organization_settings`.",
		CreateContext: resourceDatadogChildOrganizationCreate,
		ReadContext:   resourceDatadogChildOrganizationRead,
		DeleteContext: resourceDatadogChildOrganizationDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Description:  "Name for Child Organization after creation.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 32),
			},
			"public_id": {
				Description: "The `public_id` of the organization you are operating within.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"description": {
				Description: "Description of the organization.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"settings": {
				Description: "Organization settings",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"private_widget_share": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether or not the organization users can share widgets outside of Datadog.",
						},
						"saml": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "SAML properties",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Whether or not SAML is enabled for this organization.",
									},
								},
							},
						},
						"saml_autocreate_access_role": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The access role of the user. Options are `st` (standard user), `adm` (admin user), or `ro` (read-only user). Allowed enum values: `st`, `adm` , `ro`, `ERROR`",
						},
						"saml_autocreate_users_domains": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of domains where the SAML automated user creation is enabled.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"domains": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "List of domains where the SAML automated user creation is enabled.",
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"enabled": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Whether or not the automated user creation based on SAML domain is enabled.",
									},
								},
							},
						},
						"saml_can_be_enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether or not SAML can be enabled for this organization.",
						},
						"saml_idp_endpoint": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Identity provider endpoint for SAML authentication.",
						},
						"saml_idp_initiated_login": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Whether or not a SAML identity provider metadata file was provided to the Datadog organization.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Whether or not a SAML identity provider metadata file was provided to the Datadog organization.",
									},
								},
							},
						},
						"saml_idp_metadata_uploaded": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether or not a SAML identity provider metadata file was provided to the Datadog organization.",
						},
						"saml_login_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "URL for SAML logging.",
						},
						"saml_strict_mode": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Whether or not the SAML strict mode is enabled. If true, all users must log in with SAML.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Whether or not the SAML strict mode is enabled. If true, all users must log in with SAML.",
									},
								},
							},
						},
					},
				},
			},

			"api_key": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Datadog API key.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:        schema.TypeString,
							Computed:    true,
							Sensitive:   true,
							Description: "API key.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of your API key.",
						},
					},
				},
			},

			"application_key": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An application key with its associated metadata.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hash": {
							Type:        schema.TypeString,
							Computed:    true,
							Sensitive:   true,
							Description: "Hash of an application key.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of an application key.",
						},
						"owner": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Owner of an application key.",
						},
					},
				},
			},

			"user": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Information about a user",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"email": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The new email of the user.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the user.",
						},
						"access_role": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The access role of the user. Options are `st` (standard user), `adm` (admin user), or `ro` (read-only user). Allowed enum values: `st`, `adm`, `ro`, `ERROR`",
						},
					},
				},
			},
		},
	}
}

func buildDatadogOrganizationCreateV1Struct(d *schema.ResourceData) *datadogV1.OrganizationCreateBody {
	body := datadogV1.NewOrganizationCreateBody(d.Get("name").(string))

	return body
}

func updateOrganizationState(d *schema.ResourceData, org *datadogV1.Organization) diag.Diagnostics {
	d.Set("name", org.GetName())
	d.Set("public_id", org.GetPublicId())
	d.Set("description", org.GetDescription())

	settings := org.GetSettings()
	settingsMap := make(map[string]interface{})

	// `private_widget_share`
	settingsMap["private_widget_share"] = settings.GetPrivateWidgetShare()

	// `saml`
	settingsSaml := settings.GetSaml()
	settingsSamlMap := make(map[string]interface{})
	settingsSamlMap["enabled"] = settingsSaml.GetEnabled()
	settingsMap["saml"] = []map[string]interface{}{settingsSamlMap}

	// `saml_autocreate_access_role`
	settingsMap["saml_autocreate_access_role"] = settings.GetSamlAutocreateAccessRole()

	// `saml_autocreate_users_domains`
	settingsSamlAutocreateUsersDomains := settings.GetSamlAutocreateUsersDomains()
	settingsSamlAutocreateUsersDomainsMap := make(map[string]interface{})
	settingsSamlAutocreateUsersDomainsMap["domains"] = settingsSamlAutocreateUsersDomains.GetDomains()
	settingsSamlAutocreateUsersDomainsMap["enabled"] = settingsSamlAutocreateUsersDomains.GetEnabled()
	settingsMap["saml_autocreate_users_domains"] = []map[string]interface{}{settingsSamlAutocreateUsersDomainsMap}

	// `saml_can_be_enabled` & `saml_idp_endpoint`
	settingsMap["saml_can_be_enabled"] = settings.GetSamlCanBeEnabled()
	settingsMap["saml_idp_endpoint"] = settings.GetSamlIdpEndpoint()

	// `saml_idp_initiated_login`
	settingsSamlIdpInitiatedLogin := settings.GetSamlIdpInitiatedLogin()
	settingsSamlIdpInitiatedLoginMap := make(map[string]interface{})
	settingsSamlIdpInitiatedLoginMap["enabled"] = settingsSamlIdpInitiatedLogin.GetEnabled()
	settingsMap["saml_idp_initiated_login"] = []map[string]interface{}{settingsSamlIdpInitiatedLoginMap}

	// `saml_idp_metadata_uploaded` & `saml_login_url`
	settingsMap["saml_idp_metadata_uploaded"] = settings.GetSamlIdpMetadataUploaded()
	settingsMap["saml_login_url"] = settings.GetSamlLoginUrl()

	// `saml_strict_mode`
	settingsSamlStrictMode := settings.GetSamlStrictMode()
	settingsSamlStrictModeMap := make(map[string]interface{})
	settingsSamlStrictModeMap["enabled"] = settingsSamlStrictMode.GetEnabled()
	settingsMap["saml_strict_mode"] = []map[string]interface{}{settingsSamlStrictModeMap}

	if err := d.Set("settings", []interface{}{settingsMap}); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func updateOrganizationApiKeyState(d *schema.ResourceData, apiKey *datadogV1.ApiKey) diag.Diagnostics {
	apiKeyMap := make(map[string]interface{})

	apiKeyMap["key"] = apiKey.GetKey()
	apiKeyMap["name"] = apiKey.GetName()

	if err := d.Set("api_key", []interface{}{apiKeyMap}); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func updateOrganizationApplicationKeyState(d *schema.ResourceData, apiKey *datadogV1.ApplicationKey) diag.Diagnostics {
	applicationKeyMap := make(map[string]interface{})

	applicationKeyMap["hash"] = apiKey.GetHash()
	applicationKeyMap["name"] = apiKey.GetName()
	applicationKeyMap["owner"] = apiKey.GetOwner()

	if err := d.Set("application_key", []interface{}{applicationKeyMap}); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func updateOrganizationUserState(d *schema.ResourceData, apiKey *datadogV1.User) diag.Diagnostics {
	userMap := make(map[string]interface{})

	userMap["email"] = apiKey.GetEmail()
	userMap["name"] = apiKey.GetName()
	userMap["access_role"] = apiKey.GetAccessRole()

	if err := d.Set("user", []interface{}{userMap}); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDatadogChildOrganizationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	resp, httpResponse, err := datadogClientV1.OrganizationsApi.CreateChildOrg(authV1, *buildDatadogOrganizationCreateV1Struct(d))
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating organization")
	}

	org := resp.GetOrg()
	apiKey := resp.GetApiKey()
	applicationKey := resp.GetApplicationKey()
	user := resp.GetUser()

	publicId := org.GetPublicId()
	d.SetId(publicId)

	updateOrganizationApiKeyState(d, &apiKey)
	updateOrganizationApplicationKeyState(d, &applicationKey)
	updateOrganizationUserState(d, &user)

	return updateOrganizationState(d, &org)
}

func resourceDatadogChildOrganizationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Once the organization is created there is no way to get an information for it.
	return nil
}

func resourceDatadogChildOrganizationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return append(diags, diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  "Cannot delete organization.",
		Detail:   "Remove organization by contacting support (https://docs.datadoghq.com/help/).",
	})
}
