package datadog

import (
	"context"
	"fmt"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDatadogPermissions() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve the list of Datadog permissions by name and their corresponding ID, for use in the role resource.",
		ReadContext: dataSourceDatadogPermissionsRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"include_restricted": {
					Description: "Whether to include restricted permissions. Restricted permissions are granted by default to all users of a Datadog org, and cannot be manually granted or revoked.",
					Type:        schema.TypeBool,
					Default:     false,
					Optional:    true,
				},
				// Computed values
				"permissions": {
					Description: "Map of permissions names to their corresponding ID.",
					Type:        schema.TypeMap,
					Computed:    true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			}
		},
	}
}

func dataSourceDatadogPermissionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	res, resp, err := apiInstances.GetRolesApiV2().ListPermissions(auth)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, resp, "error listing permissions")
	}

	diags := diag.Diagnostics{}
	perms := res.GetData()
	permsNameToID := make(map[string]string, len(perms))
	includeRestricted := d.Get("include_restricted").(bool)
	for _, perm := range perms {
		if !includeRestricted && perm.Attributes.GetRestricted() {
			continue
		}

		if err := utils.CheckForUnparsed(perm); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("skipping permission with id: %s", perm.GetId()),
				Detail:   fmt.Sprintf("permission contains unparsed object: %v", err),
			})
			continue
		}
		permsNameToID[perm.Attributes.GetName()] = perm.GetId()
	}
	if err := d.Set("permissions", permsNameToID); err != nil {
		return diag.FromErr(err)
	}
	d.SetId("datadog-permissions")

	return diags
}
