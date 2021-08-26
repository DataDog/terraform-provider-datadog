package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDatadogPermissions() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve the list of Datadog permissions by name and their corresponding ID, for use in the role resource.",
		ReadContext: dataSourceDatadogPermissionsRead,

		Schema: map[string]*schema.Schema{
			// Computed values
			"permissions": {
				Description: "Map of permissions names to their corresponding ID.",
				Type:        schema.TypeMap,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceDatadogPermissionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	res, resp, err := datadogClientV2.RolesApi.ListPermissions(authV2)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, resp, "error listing permissions")
	}
	if err := utils.CheckForUnparsed(res); err != nil {
		return diag.FromErr(err)
	}
	perms := res.GetData()
	permsNameToID := make(map[string]string, len(perms))
	for _, perm := range perms {
		// Don't list restricted permissions, as they cannot be granted/revoked to/from a role
		if perm.Attributes.GetRestricted() {
			continue
		}
		permsNameToID[perm.Attributes.GetName()] = perm.GetId()
	}
	if err := d.Set("permissions", permsNameToID); err != nil {
		return diag.FromErr(err)
	}
	d.SetId("datadog-permissions")

	return nil
}
