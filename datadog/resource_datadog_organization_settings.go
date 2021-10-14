package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceDatadogOrganizationSettings() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog Organization resource. This can be used to manage your Datadog organization's settings.",
		CreateContext: resourceDatadogOrganizationSettingsCreate,
		ReadContext:   resourceDatadogOrganizationSettingsRead,
		UpdateContext: resourceDatadogOrganizationSettingsUpdate,
		DeleteContext: resourceDatadogOrganizationSettingsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description:  "Name for Organization.",
				Type:         schema.TypeString,
				Optional:     true,
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
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"private_widget_share": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether or not the organization users can share widgets outside of Datadog.",
						},
						"saml": {
							Type:        schema.TypeList,
							Required:    true,
							MaxItems:    1,
							Description: "SAML properties",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Whether or not SAML is enabled for this organization.",
									},
								},
							},
						},
						"saml_autocreate_access_role": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "st",
							Description:  "The access role of the user. Options are `st` (standard user), `adm` (admin user), or `ro` (read-only user). Allowed enum values: `st`, `adm` , `ro`, `ERROR`",
							ValidateFunc: validation.StringInSlice([]string{"st", "adm", "ro", "ERROR"}, false),
						},
						"saml_autocreate_users_domains": {
							Type:        schema.TypeList,
							Required:    true,
							MaxItems:    1,
							Description: "List of domains where the SAML automated user creation is enabled.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"domains": {
										Type:        schema.TypeList,
										Optional:    true,
										Description: "List of domains where the SAML automated user creation is enabled.",
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"enabled": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
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
							Required:    true,
							MaxItems:    1,
							Description: "Whether or not a SAML identity provider metadata file was provided to the Datadog organization.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
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
							Required:    true,
							MaxItems:    1,
							Description: "Whether or not the SAML strict mode is enabled. If true, all users must log in with SAML.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Whether or not the SAML strict mode is enabled. If true, all users must log in with SAML.",
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

func buildDatadogOrganizationUpdateV1Struct(d *schema.ResourceData) *datadogV1.Organization {
	org := datadogV1.NewOrganization()

	if v, ok := d.GetOk("name"); ok {
		org.SetName(v.(string))
	}

	// settings
	if v, ok := d.GetOk("settings"); ok {
		if settingsSetList := v.([]interface{}); len(settingsSetList) > 0 {
			settings := datadogV1.NewOrganizationSettings()
			settingsSet := settingsSetList[0].(map[string]interface{})

			// private_widget_share
			if v, ok := settingsSet["private_widget_share"]; ok {
				settings.SetPrivateWidgetShare(v.(bool))
			}

			// saml
			if v, ok := settingsSet["saml"]; ok {
				if samlIdpInitiatedLoginSetList := v.([]interface{}); len(samlIdpInitiatedLoginSetList) > 0 {
					saml := datadogV1.NewOrganizationSettingsSaml()
					samlSet := samlIdpInitiatedLoginSetList[0].(map[string]interface{})
					if v, ok := samlSet["enabled"]; ok {
						saml.SetEnabled(v.(bool))
					}
					settings.SetSaml(*saml)
				}
			}

			// saml_autocreate_access_role
			if v, ok := settingsSet["saml_autocreate_access_role"]; ok {
				settings.SetSamlAutocreateAccessRole(datadogV1.AccessRole(v.(string)))
			}

			// saml_autocreate_users_domains
			if v, ok := settingsSet["saml_autocreate_users_domains"]; ok {
				if samlAutocreateUsersDomainsSetList := v.([]interface{}); len(samlAutocreateUsersDomainsSetList) > 0 {
					samlAutocreateUsersDomainsSet := samlAutocreateUsersDomainsSetList[0].(map[string]interface{})
					samlAutocreateUsersDomains := datadogV1.NewOrganizationSettingsSamlAutocreateUsersDomains()

					// domains
					if v, ok := samlAutocreateUsersDomainsSet["domains"]; ok {
						tfDomains := v.([]interface{})
						domains := make([]string, len(tfDomains))
						for i, domain := range tfDomains {
							domains[i] = domain.(string)
						}
						samlAutocreateUsersDomains.SetDomains(domains)
					}

					// enabled
					if v, ok := samlAutocreateUsersDomainsSet["enabled"]; ok {
						samlAutocreateUsersDomains.SetEnabled(v.(bool))
					}

					settings.SetSamlAutocreateUsersDomains(*samlAutocreateUsersDomains)
				}
			}

			// saml_idp_initiated_login
			if v, ok := settingsSet["saml_idp_initiated_login"]; ok {
				if samlIdpInitiatedLoginSetList := v.([]interface{}); len(samlIdpInitiatedLoginSetList) > 0 {
					samlIdpInitiatedLogin := datadogV1.NewOrganizationSettingsSamlIdpInitiatedLogin()
					samlIdpInitiatedLoginSet := samlIdpInitiatedLoginSetList[0].(map[string]interface{})
					if v, ok := samlIdpInitiatedLoginSet["enabled"]; ok {
						samlIdpInitiatedLogin.SetEnabled(v.(bool))
					}
					settings.SetSamlIdpInitiatedLogin(*samlIdpInitiatedLogin)
				}
			}

			// saml_strict_mode
			if v, ok := settingsSet["saml_strict_mode"]; ok {
				if samlStrictModeSetList := v.([]interface{}); len(samlStrictModeSetList) > 0 {
					samlStrictMode := datadogV1.NewOrganizationSettingsSamlStrictMode()
					samlStrictModeSet := samlStrictModeSetList[0].(map[string]interface{})
					if v, ok := samlStrictModeSet["enabled"]; ok {
						samlStrictMode.SetEnabled(v.(bool))
					}
					settings.SetSamlStrictMode(*samlStrictMode)
				}
			}

			org.SetSettings(*settings)
		}
	}

	return org
}

func resourceDatadogOrganizationSettingsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	resp, httpResponse, err := datadogClientV1.OrganizationsApi.ListOrgs(authV1)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error getting organization")
	}

	orgs := resp.GetOrgs()
	if len(orgs) == 0 {
		return diag.Errorf("no organizations available")
	}

	org := orgs[0]
	d.SetId(org.GetPublicId())

	return resourceDatadogOrganizationSettingsUpdate(ctx, d, meta)
}

func resourceDatadogOrganizationSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	resp, httpResponse, err := datadogClientV1.OrganizationsApi.GetOrg(authV1, d.Id())
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpResponse, "error getting organization")
	}

	org := resp.GetOrg()
	d.SetId(org.GetPublicId())

	return updateOrganizationState(d, &org)
}

func resourceDatadogOrganizationSettingsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	resp, httpResponse, err := datadogClientV1.OrganizationsApi.UpdateOrg(authV1, d.Id(), *buildDatadogOrganizationUpdateV1Struct(d))
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating organization")
	}

	org := resp.GetOrg()

	publicId := org.GetPublicId()
	d.SetId(publicId)

	return updateOrganizationState(d, &org)
}

func resourceDatadogOrganizationSettingsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return append(diags, diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  "Cannot delete organization settings.",
		Detail:   "Remove organization by contacting support (https://docs.datadoghq.com/help/).",
	})
}
