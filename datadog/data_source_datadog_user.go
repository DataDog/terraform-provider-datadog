package datadog

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func dataSourceDatadogUser() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about an existing user to use it in an other resources.",
		ReadContext: dataSourceDatadogUserRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"filter": {
					Description: "Filter all users by the given string.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"exact_match": {
					Description: "When true, `filter` string is exact matched against the user's `email`, followed by `name` attribute.",
					Type:        schema.TypeBool,
					Default:     false,
					Optional:    true,
				},
				// Computed values
				"email": {
					Description: "Email of the user.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"name": {
					Description: "Name of the user.",
					Type:        schema.TypeString,
					Computed:    true,
				},
			}
		},
	}
}

func dataSourceDatadogUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	filter := d.Get("filter").(string) // string | Filter all users by the given string. Defaults to no filtering. (optional) // string | Filter on status attribute. Comma separated list, with possible values `Active`, `Pending`, and `Disabled`. Defaults to no filtering. (optional)
	exactMatch := d.Get("exact_match").(bool)
	optionalParams := datadogV2.ListUsersOptionalParameters{
		Filter: &filter,
	}

	res, httpresp, err := apiInstances.GetUsersApiV2().ListUsers(auth, optionalParams)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error querying user")
	}

	users := res.GetData()
	if len(users) > 1 && !exactMatch {
		return diag.Errorf("your query returned more than one result for filter \"%s\", please try a more specific search criteria",
			filter,
		)
	} else if len(users) == 0 {
		return diag.Errorf("didn't find any user matching filter string  \"%s\"", filter)
	}

	matchedUser := users[0]
	if exactMatch {
		matchCount := 0
		for _, user := range users {
			if user.Attributes.GetEmail() == filter {
				matchedUser = user
				matchCount++
				continue
			}
			if user.Attributes.GetName() == filter {
				matchedUser = user
				matchCount++
				continue
			}
		}
		if matchCount > 1 {
			return diag.Errorf("your query returned more than one result for filter with exact match \"%s\", please try a more specific search criteria",
				filter,
			)
		}
		if matchCount == 0 {
			return diag.Errorf("didn't find any user matching filter string with exact match \"%s\"", filter)
		}
	}

	if err := utils.CheckForUnparsed(matchedUser); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(matchedUser.GetId())
	if err := d.Set("name", matchedUser.Attributes.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("email", matchedUser.Attributes.GetEmail()); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
