package datadog

import (
	"context"

	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func dataSourceDatadogUser() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about an existing user to use it in an other resources.",
		ReadContext: dataSourceDatadogUserRead,

		Schema: map[string]*schema.Schema{
			"filter": {
				Description: "Filter all users by the given string.",
				Type:        schema.TypeString,
				Required:    true,
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
		},
	}
}

func dataSourceDatadogUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2
	filter := d.Get("filter").(string) // string | Filter all users by the given string. Defaults to no filtering. (optional) // string | Filter on status attribute. Comma separated list, with possible values `Active`, `Pending`, and `Disabled`. Defaults to no filtering. (optional)
	optionalParams := datadogV2.ListUsersOptionalParameters{
		Filter: &filter,
	}

	res, httpresp, err := datadogClientV2.UsersApi.ListUsers(authV2, optionalParams)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error querying user")
	}
	if err := utils.CheckForUnparsed(res); err != nil {
		return diag.FromErr(err)
	}
	users := res.GetData()
	if len(users) > 1 {
		return diag.Errorf("your query returned more than one result for filter \"%s\", please try a more specific search criteria",
			filter,
		)
	} else if len(users) == 0 {
		return diag.Errorf("didn't find any user matching filter string  \"%s\"", filter)
	}
	matchedUser := users[0]
	d.SetId(matchedUser.GetId())
	if err := d.Set("name", matchedUser.Attributes.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("email", matchedUser.Attributes.GetEmail()); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
