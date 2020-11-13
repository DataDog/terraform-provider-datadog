package datadog

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceDatadogRole() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDatadogRoleRead,

		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeString,
				Required: true,
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
	filter := d.Get("filter").(string)
	req = req.Filter(filter)

	res, _, err := req.Execute()
	if err != nil {
		return translateClientError(err, "error querying roles")
	}
	roles := res.GetData()
	roleIndex := 0
	if len(roles) > 1 {
		exactMatchFound := false
		for i, role := range roles {
			if role.Attributes.GetName() == filter {
				exactMatchFound = true
				roleIndex = i
				log.Printf("[INFO] Got multiple matches for role '%s', picking the one with this exact name", filter)
				break
			}
		}
		if !exactMatchFound {
			return fmt.Errorf(
				"your query returned more than one result and no exact match for name '%s' were found, "+
					"please try a more specific search criteria",
				filter,
			)
		}
	} else if len(roles) == 0 {
		return fmt.Errorf("your query returned no result, please try a less specific search criteria")
	}

	r := roles[roleIndex]
	d.SetId(r.GetId())
	d.Set("name", r.Attributes.GetName())
	d.Set("user_count", r.Attributes.GetUserCount())

	return nil
}
