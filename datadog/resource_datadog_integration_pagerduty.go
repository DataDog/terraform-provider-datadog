package datadog

import (
	"context"
	"log"
	"strings"
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zorkian/go-datadog-api"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
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
		return diag.Errorf("error creating PagerDuty integration: %s", err.Error())
	}

	pdIntegration, err := client.GetIntegrationPD()
	if err != nil {
		return diag.Errorf("error getting PagerDuty integration: %s", err.Error())
	}
	if err := utils.CheckForUnparsed(pdIntegration); err != nil {
		return diag.FromErr(err)
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
		return diag.Errorf("error getting PagerDuty integration: %s", err.Error())
	}
	if err := utils.CheckForUnparsed(pd); err != nil {
		return diag.FromErr(err)
	}

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
		return diag.Errorf("error updating PagerDuty integration: %s", err.Error())
	}

	return resourceDatadogIntegrationPagerdutyRead(ctx, d, meta)
}

func resourceDatadogIntegrationPagerdutyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Deleting the pagerduty integration isn't safe, please don't use this resource anymore")

	return nil
}
