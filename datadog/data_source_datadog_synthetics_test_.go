package datadog

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"regexp"
)

func dataSourceDatadogSyntheticsTest() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve a Datadog Synthetic Test.",
		ReadContext: dataSourceDatadogSyntheticsTestRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"test_id": {
					Description: "The synthetic test id or URL to search for",
					Type:        schema.TypeString,
					Required:    true,
				},
				"name": {
					Description: "The name of the synthetic test.",
					Type:        schema.TypeString,
					Elem:        &schema.Schema{Type: schema.TypeString},
					Computed:    true,
				},
				"tags": {
					Description: "A list of tags assigned to the synthetic test.",
					Type:        schema.TypeList,
					Elem:        &schema.Schema{Type: schema.TypeString},
					Computed:    true,
				},
				"url": {
					Description: "The start URL of the synthetic test.",
					Type:        schema.TypeString,
					Elem:        &schema.Schema{Type: schema.TypeString},
					Computed:    true,
				},
			}
		},
	}
}

func dataSourceDatadogSyntheticsTestRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	urlRegex := regexp.MustCompile(`https:\/\/(.*)\.(datadoghq|ddog-gov)\.(com|eu)\/synthetics\/details\/`)
	searchedId := urlRegex.ReplaceAllString(d.Get("test_id").(string), "")

	if test, _, err := apiInstances.GetSyntheticsApiV1().GetAPITest(auth, searchedId); err == nil {
		d.SetId(test.GetPublicId())
		d.Set("name", test.GetName())
		d.Set("tags", test.GetTags())
		d.Set("url", test.Config.Request.GetUrl())
	} else if test, _, err := apiInstances.GetSyntheticsApiV1().GetBrowserTest(auth, searchedId); err == nil {
		d.SetId(test.GetPublicId())
		d.Set("name", test.GetName())
		d.Set("tags", test.GetTags())
		d.Set("url", test.Config.Request.GetUrl())
	} else {
		return diag.Errorf("Couldn't retrieve synthetic test with id %s", searchedId)
	}

	return nil
}
