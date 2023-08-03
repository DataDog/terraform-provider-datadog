package datadog

import (
	"context"
	"log"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func dataSourceDatadogRole() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about an existing role for use in other resources.",
		ReadContext: dataSourceDatadogRoleRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
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
			}
		},
	}
}

func dataSourceDatadogRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	optionalParams := datadogV2.NewListRolesOptionalParameters()
	filter := d.Get("filter").(string)
	optionalParams = optionalParams.WithFilter(filter)

	res, httpresp, err := apiInstances.GetRolesApiV2().ListRoles(auth, *optionalParams)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error querying roles")
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
			return diag.Errorf(
				"your query returned more than one result and no exact match for name '%s' were found, "+
					"please try a more specific search criteria",
				filter,
			)
		}
	} else if len(roles) == 0 {
		return diag.Errorf("your query returned no result, please try a less specific search criteria")
	}

	if err := utils.CheckForUnparsed(roles[roleIndex]); err != nil {
		return diag.FromErr(err)
	}

	r := roles[roleIndex]
	d.SetId(r.GetId())
	if err := d.Set("name", r.Attributes.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("user_count", r.Attributes.GetUserCount()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
