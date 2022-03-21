package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDatadogSyntheticsTest() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve a Datadog Synthetic Test.",
		ReadContext: dataSourceDatadogSyntheticsTestRead,

		Schema: map[string]*schema.Schema{
			"test_id": {
				Description: "The synthetic test id to search for",
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
		},
	}
}

func dataSourceDatadogSyntheticsTestRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	tests, httpresp, err := datadogClientV1.SyntheticsApi.ListTests(authV1)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting synthetic tests")
	}
	if err := utils.CheckForUnparsed(tests); err != nil {
		return diag.FromErr(err)
	}

	searchedId := d.Get("test_id").(string)

	for _, test := range *tests.Tests {
		if test.GetPublicId() == searchedId {
			d.SetId(test.GetPublicId())
			d.Set("name", test.GetName())
			d.Set("tags", test.GetTags())
			d.Set("url", test.GetConfig().Request.GetUrl())

			return nil
		}
	}

	return diag.Errorf("Couldn't find synthetic test with id %s", searchedId)
}
