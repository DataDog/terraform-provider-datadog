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

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
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
					Type:          schema.TypeString,
					Optional:      true,
					Computed:      true,
					ConflictsWith: []string{"name_filter", "type_filter"},
					Description:   "ID of the RUM application. Cannot be used with name and type filters.",
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
			}
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
		bothSet := searchNameOk && searchTypeOk

		var foundRUMApplicationIDs []string
		for _, resp_data := range resp.Data {
			if rum_app, ok := resp_data.GetAttributesOk(); ok {
				nameSetAndMatched := searchNameOk && rum_app.GetName() == searchName
				typeSetAndMatched := searchTypeOk && rum_app.GetType() == searchType
				if bothSet {
					if nameSetAndMatched && typeSetAndMatched {
						foundRUMApplicationIDs = append(foundRUMApplicationIDs, rum_app.GetApplicationId())
					}
				} else if nameSetAndMatched || typeSetAndMatched {
					foundRUMApplicationIDs = append(foundRUMApplicationIDs, rum_app.GetApplicationId())
				}
			}
		}

		if len(foundRUMApplicationIDs) == 0 {
			return diag.Errorf("Couldn't find a RUM Application with name '%s' and type '%s'", searchName, searchType)
		} else if len(foundRUMApplicationIDs) > 1 {
			return diag.Errorf("Searching for name '%s' and type '%s' returned more than one RUM application.", searchName, searchType)
		}

		app_resp, _, app_err := apiInstances.GetRumApiV2().GetRUMApplication(auth, foundRUMApplicationIDs[0])
		if app_err != nil {
			return diag.Errorf("Found RUM application with id %s, but couldn't retrieve details.", foundRUMApplicationIDs[0])
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
	if err := d.Set("client_token", rum_app.GetClientToken()); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
