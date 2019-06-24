package datadog

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/zorkian/go-datadog-api"
)

const maskedSecret = "*****"

func resourceDatadogIntegrationPagerdutySO() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogIntegrationPagerdutySOCreate,
		Read:   resourceDatadogIntegrationPagerdutySORead,
		Update: resourceDatadogIntegrationPagerdutySOUpdate,
		Delete: resourceDatadogIntegrationPagerdutySODelete,
		// since the API never returns service_key, it's impossible to meaningfully import resources
		Importer: nil,

		Schema: map[string]*schema.Schema{
			"service_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"service_key": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

func buildIntegrationPagerdutySO(d *schema.ResourceData) *datadog.ServicePDRequest {
	so := &datadog.ServicePDRequest{}
	if v, ok := d.GetOk("service_name"); ok {
		so.SetServiceName(v.(string))
	}
	if v, ok := d.GetOk("service_key"); ok {
		so.SetServiceKey(v.(string))
	}

	return so
}

func resourceDatadogIntegrationPagerdutySOCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)
	integrationPdMutex.Lock()
	defer integrationPdMutex.Unlock()

	so := buildIntegrationPagerdutySO(d)

	if err := client.CreateIntegrationPDService(so); err != nil {
		// TODO: warn user that PD integration must be enabled to be able to create service objects
		return fmt.Errorf("Failed to create integration pagerduty using Datadog API: %s", err.Error())
	}
	d.SetId(so.GetServiceName())

	return resourceDatadogIntegrationPagerdutySORead(d, meta)
}

func resourceDatadogIntegrationPagerdutySORead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	so, err := client.GetIntegrationPDService(d.Id())
	if err != nil {
		return err
	}

	d.Set("service_name", so.GetServiceName())
	// Only update service_key if not set on d - the API endpoints never return
	// the keys, so this is how we recognize new values.
	if _, ok := d.GetOk("service_key"); !ok {
		d.Set("service_key", maskedSecret)
	}

	return nil
}

func resourceDatadogIntegrationPagerdutySOUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)
	integrationPdMutex.Lock()
	defer integrationPdMutex.Unlock()

	so := buildIntegrationPagerdutySO(d)

	if err := client.UpdateIntegrationPDService(so); err != nil {
		return fmt.Errorf("Failed to create integration pagerduty using Datadog API: %s", err.Error())
	}
	d.SetId(so.GetServiceName())

	return resourceDatadogIntegrationPagerdutySORead(d, meta)
}

func resourceDatadogIntegrationPagerdutySODelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)
	integrationPdMutex.Lock()
	defer integrationPdMutex.Unlock()

	if err := client.DeleteIntegrationPDService(d.Id()); err != nil {
		return fmt.Errorf("Error while deleting Pagerduty integration service object: %v", err)
	}

	return nil
}
