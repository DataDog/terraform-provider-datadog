package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDatadogOrganizationSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about an existing organization.",
		ReadContext: dataSourceDatadogOrganizationRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name for Organization.",
				Type:        schema.TypeString,
				Computed:    true,
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
		},
	}
}

func dataSourceDatadogOrganizationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	return updateOrganizationState(d, &org)
}
