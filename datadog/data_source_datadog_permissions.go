package datadog

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceDatadogPermissions() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDatadogPermissionsRead,

		Schema: map[string]*schema.Schema{
			// Computed values
			"permissions": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceDatadogPermissionsRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	res, _, err := datadogClientV2.RolesApi.ListPermissions(authV2).Execute()
	if err != nil {
		return translateClientError(err, "error listing permissions")
	}
	perms := res.GetData()
	permsNameToID := make(map[string]string, len(perms))
	for _, perm := range perms {
		// Don't list restricted permissions, as they cannot be granted/revoked to/from a role
		if !perm.Attributes.GetRestricted() {
			permsNameToID[perm.Attributes.GetName()] = perm.GetId()
		}
	}
	d.SetId("datadog-permissions")
	d.Set("permissions", permsNameToID)

	return nil
}
