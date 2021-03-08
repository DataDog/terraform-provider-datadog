package datadog

import (
	"context"
	"log"
	"strings"
	"sync"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zorkian/go-datadog-api"
)

// creating/modifying/deleting PD integration and its service objects in parallel on one account
// is unsupported by the API right now; therefore we use the mutex to only operate on one at a time
var integrationPdMutex = sync.Mutex{}

func resourceDatadogIntegrationPagerduty() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog - PagerDuty resource. This can be used to create and manage Datadog - PagerDuty integration. See also [PagerDuty Integration Guide](https://www.pagerduty.com/docs/guides/datadog-integration-guide/).",
		CreateContext: resourceDatadogIntegrationPagerdutyCreate,
		ReadContext:   resourceDatadogIntegrationPagerdutyRead,
		UpdateContext: resourceDatadogIntegrationPagerdutyUpdate,
		DeleteContext: resourceDatadogIntegrationPagerdutyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"individual_services": {
				Description: "Boolean to specify whether or not individual service objects specified by [datadog_integration_pagerduty_service_object](https://registry.terraform.io/providers/DataDog/datadog/latest/docs/resources/integration_pagerduty_service_object) resource are to be used. Mutually exclusive with `services` key.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"services": {
				ConflictsWith: []string{"individual_services"},
				Deprecated:    "set \"individual_services\" to true and use datadog_pagerduty_integration_service_object",
				Type:          schema.TypeList,
				Optional:      true,
				Description:   "A list of service names and service keys.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"service_name": {
							Description: "Your Service name in PagerDuty.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"service_key": {
							Description: "Your Service name associated service key in Pagerduty.",
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
						},
					},
				},
			},
			"subdomain": {
				Description: "Your PagerDuty account’s personalized subdomain name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"schedules": {
				Description: "Array of your schedule URLs.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"api_token": {
				Description: "Your PagerDuty API token.",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func buildIntegrationPagerduty(d *schema.ResourceData) (*datadog.IntegrationPDRequest, error) {
	pd := &datadog.IntegrationPDRequest{}
	pd.SetSubdomain(d.Get("subdomain").(string))
	pd.SetAPIToken(d.Get("api_token").(string))

	var schedules []string
	for _, s := range d.Get("schedules").([]interface{}) {
		schedules = append(schedules, s.(string))
	}
	pd.Schedules = schedules

	var services []datadog.ServicePDRequest
	if value, ok := d.GetOk("individual_services"); ok && value.(bool) {
		services = nil
	} else {
		configServices, ok := d.GetOk("services")
		if ok {
			for _, sInterface := range configServices.([]interface{}) {
				s := sInterface.(map[string]interface{})

				service := datadog.ServicePDRequest{}
				service.SetServiceName(s["service_name"].(string))
				service.SetServiceKey(s["service_key"].(string))

				services = append(services, service)
			}
		}
	}
	pd.Services = services

	return pd, nil
}

func resourceDatadogIntegrationPagerdutyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	integrationPdMutex.Lock()
	defer integrationPdMutex.Unlock()

	pd, err := buildIntegrationPagerduty(d)
	if err != nil {
		return diag.Errorf("failed to parse resource configuration: %s", err.Error())
	}

	if err := client.CreateIntegrationPD(pd); err != nil {
		return utils.TranslateClientErrorDiag(err, "error creating PagerDuty integration")
	}

	pdIntegration, err := client.GetIntegrationPD()
	if err != nil {
		return utils.TranslateClientErrorDiag(err, "error getting PagerDuty integration")
	}

	d.SetId(pdIntegration.GetSubdomain())

	return resourceDatadogIntegrationPagerdutyRead(ctx, d, meta)
}

func resourceDatadogIntegrationPagerdutyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	pd, err := client.GetIntegrationPD()
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, "error getting PagerDuty integration")
	}

	var services []map[string]string
	if value, ok := d.GetOk("individual_services"); ok && value.(bool) {
		services = nil
	} else {
		for _, service := range pd.Services {
			services = append(services, map[string]string{
				"service_name": service.GetServiceName(),
				"service_key":  service.GetServiceKey(),
			})
		}
	}

	d.Set("services", services)
	d.Set("subdomain", pd.GetSubdomain())
	d.Set("schedules", pd.Schedules)

	return nil
}

func resourceDatadogIntegrationPagerdutyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	integrationPdMutex.Lock()
	defer integrationPdMutex.Unlock()

	pd, err := buildIntegrationPagerduty(d)
	if err != nil {
		return diag.Errorf("failed to parse resource configuration: %s", err.Error())
	}

	if err := client.UpdateIntegrationPD(pd); err != nil {
		return utils.TranslateClientErrorDiag(err, "error updating PagerDuty integration")
	}

	// if there are none currently configured services, we actually
	// have to remove them explicitly, otherwise the underlying API client
	// would not send the "services" key at all and they wouldn't get deleted
	if value, ok := d.GetOk("individual_services"); !ok || !value.(bool) {
		currentServices := d.Get("services").([]interface{})
		if len(currentServices) == 0 {
			pd, err := client.GetIntegrationPD()
			if err != nil {
				return utils.TranslateClientErrorDiag(err, "error getting PagerDuty integration")
			}
			for _, service := range pd.Services {
				if err := client.DeleteIntegrationPDService(*service.ServiceName); err != nil {
					return utils.TranslateClientErrorDiag(err, "error deleting PagerDuty integration service")
				}
			}
		}
	}
	return resourceDatadogIntegrationPagerdutyRead(ctx, d, meta)
}

func resourceDatadogIntegrationPagerdutyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Deleting the pagerduty integration isn't safe, please don't use this resource anymore")

	return nil
}
