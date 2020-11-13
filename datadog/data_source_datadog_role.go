package datadog

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceDatadogRole() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDatadogRoleRead,

		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeString,
				Optional: true,
			},

			// Computed values
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_count": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceDatadogRoleRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	req := datadogClientV2.RolesApi.ListRoles(authV2)
	if v, ok := d.GetOk("filter"); ok {
		req = req.Filter(v.(string))
	}

	res, _, err := req.Execute()
	if err != nil {
		return translateClientError(err, "error querying monitors")
	}
	roles := res.GetData()
	if len(roles) > 1 {
		return fmt.Errorf("your query returned more than one result, please try a more specific search criteria")
	}
	if len(roles) == 0 {
		return fmt.Errorf("your query returned no result, please try a less specific search criteria")
	}

	r := roles[0]
	d.SetId(r.GetId())
	d.Set("name", r.Attributes.GetName())
	d.Set("user_count", r.Attributes.GetUserCount())

	return nil
}
