package datadog

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDatadogRUMApplication() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve a Datadog RUM Application.",
		ReadContext: dataSourceDatadogRUMApplicationRead,

		Schema: map[string]*schema.Schema{
			"name_filter": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name used to search for a RUM application",
			},
			"type_filter": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The type used to search for a RUM application",
			},
			"id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "ID of the RUM application. If set, this takes precedence over name and type filters.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the RUM application",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The RUM application type. Supported values are `browser`, `ios`, `android`, `react-native`, `flutter`",
			},
			"client_token": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The client token",
			},
		},
	}
}

func dataSourceDatadogRUMApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	if searchID, ok := d.GetOk("id"); ok {
		resp, _, err := apiInstances.GetRumApiV2().GetRUMApplication(auth, searchID.(string))
		if err != nil {
			return diag.Errorf("Couldn't find RUM application with id %s", searchID)
		}

		return dataSourceDatadogRUMApplicationUpdate(d, resp.Data.GetAttributes())
	} else {
		resp, _, err := apiInstances.GetRumApiV2().GetRUMApplications(auth)
		if err != nil {
			return diag.Errorf("Couldn't retrieve list of RUM Applications")
		}

		searchName, searchNameOk := d.GetOk("name_filter")
		searchType, searchTypeOk := d.GetOk("type_filter")
		var foundRUMApplications []datadogV2.RUMApplicationAttributes
		for _, resp_data := range resp.Data {
			if rum_app, ok := resp_data.GetAttributesOk(); ok {
				nameSetAndMatched := searchNameOk && rum_app.GetName() == searchName
				typeSetAndMatched := searchTypeOk && rum_app.GetType() == searchType
				if nameSetAndMatched || typeSetAndMatched {
					foundRUMApplications = append(foundRUMApplications, *rum_app)
				}
			}
		}

		if len(foundRUMApplications) == 0 {
			return diag.Errorf("Couldn't find a RUM Application with name '%s' and type '%s'", searchName, searchType)
		} else if len(foundRUMApplications) > 1 {
			return diag.Errorf("Searching for name '%s' and type '%s' returned more than one RUM application.", searchName, searchType)
		}

		app_resp, _, app_err := apiInstances.GetRumApiV2().GetRUMApplication(auth, foundRUMApplications[0].GetApplicationId())
		if app_err != nil {
			return diag.Errorf("Couldn't find RUM application with id %s", foundRUMApplications[0].GetApplicationId())
		}
		return dataSourceDatadogRUMApplicationUpdate(d, app_resp.Data.GetAttributes())
	}
}

func dataSourceDatadogRUMApplicationUpdate(d *schema.ResourceData, rum_app datadogV2.RUMApplicationAttributes) diag.Diagnostics {
	d.SetId(rum_app.GetApplicationId())
	if err := d.Set("name", rum_app.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", rum_app.GetType()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("client_token", rum_app.GetHash()); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
