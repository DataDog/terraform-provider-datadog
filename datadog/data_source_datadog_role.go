package datadog

import (
	"fmt"
	"log"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceDatadogRole() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about an existing role for use in other resources.",
		Read:        dataSourceDatadogRoleRead,

		Schema: map[string]*schema.Schema{
			"filter": {
				Description: "A string on which to filter the roles.",
				Type:        schema.TypeString,
				Required:    true,
			},

			// Computed values
			"name": {
				Description: "Name of the role.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"user_count": {
				Description: "Number of users assigned to this role.",
				Type:        schema.TypeInt,
				Computed:    true,
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
		return utils.TranslateClientError(err, "error querying roles")
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
	if err := d.Set("name", r.Attributes.GetName()); err != nil {
		return err
	}
	if err := d.Set("user_count", r.Attributes.GetUserCount()); err != nil {
		return err
	}

	return nil
}
