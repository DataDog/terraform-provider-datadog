package datadog

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func dataSourceDatadogServiceAccount() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about an existing service account user to reference in other resources.",
		ReadContext: dataSourceDatadogServiceAccountRead,

		Schema: map[string]*schema.Schema{
			"filter": {
				Description: "Filter all service account users by the given string.",
				Type:        schema.TypeString,
				Required:    true,
			},

			// Computed values
			"email": {
				Description: "Email of the service account user.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "Name of the service account user.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceDatadogServiceAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	filter := d.Get("filter").(string)
	optionalParams := datadogV2.ListUsersOptionalParameters{
		Filter: &filter,
	}

	res, httpresp, err := apiInstances.GetUsersApiV2().ListUsers(auth, optionalParams)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error querying service account user")
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
		return diag.Errorf("didn't find any service account user matching filter string  \"%s\"", filter)
	}
	matchedUser := users[0]
	if !matchedUser.Attributes.GetServiceAccount() {
		return diag.Errorf("your query returned a human user and not a service account user")
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
