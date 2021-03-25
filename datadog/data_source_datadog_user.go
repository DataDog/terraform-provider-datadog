package datadog

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func dataSourceDatadogUser() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about an existing user to use it in an other resources.",
		Read:        dataSourceDatadogUserRead,

		Schema: map[string]*schema.Schema{
			"filter": {
				Description: "email",
				Type:        schema.TypeString,
				Required:    true,
			},
			// Computed values
			"datadog_id": {
				Description: "Id of the user.",
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

func dataSourceDatadogUserRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2
	pageNumber := int64(0)
	filter := d.Get("filter").(string)
	founded := false

	req := datadogClientV2.UsersApi.ListUsers(authV2).PageSize(20).PageNumber(pageNumber).FilterStatus("Active")
	res, _, err := req.Execute()
	totalPage := res.Meta.Page.GetTotalCount()
	if err != nil {
		return utils.TranslateClientError(err, "error querying user")
	}
	for pageNumber < totalPage {
		req = req.PageNumber(pageNumber)
		res, _, err := req.Execute()
		if len(*res.Data) == 0 { // This will skip the empty calls to datadog api  ( 50 page in my org)
			break
		}
		if err != nil {
			return utils.TranslateClientError(err, "error querying user")
		}
		for _, user := range *res.Data {
			if user.Attributes.GetEmail() == filter {
				founded = true
				d.SetId(user.Attributes.GetEmail())
				if err := d.Set("name", user.Attributes.GetEmail()); err != nil {
					return err
				}
				if err := d.Set("datadog_id", user.GetId()); err != nil {
					return err
				}
				break
			}
		}
		pageNumber += 1
	}
	if !founded {
		return fmt.Errorf("didn't founded any used mathing this email")
	}
	return nil
}
