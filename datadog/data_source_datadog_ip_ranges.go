package datadog

import (
	"github.com/hashicorp/terraform/helper/schema"
	datadog "github.com/zorkian/go-datadog-api"
)

func dataSourceDatadogIpRanges() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDatadogIpRangesRead,

		Schema: map[string]*schema.Schema{
			"agents_ipv4": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"api_ipv4": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"apm_ipv4": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"logs_ipv4": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"process_ipv4": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"synthetics_ipv4": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"webhooks_ipv4": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceDatadogIpRangesRead(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*datadog.Client)

	ipAddresses, err := client.GetIPRanges()

	if err != nil {
		return err
	}

	if len(ipAddresses.Agents)+len(ipAddresses.API)+len(ipAddresses.Apm)+len(ipAddresses.Logs)+len(ipAddresses.Process)+len(ipAddresses.Synthetics)+len(ipAddresses.Webhooks) > 0 {
		d.SetId("datadog-ip-ranges")
	}

	switch {
	case len(ipAddresses.Agents["prefixes_ipv4"]) > 0:
		d.Set("agents_ipv4", ipAddresses.Agents["prefixes_ipv4"])
	case len(ipAddresses.API["prefixes_ipv4"]) > 0:
		d.Set("api_ipv4", ipAddresses.API["prefixes_ipv4"])
	case len(ipAddresses.Apm["prefixes_ipv4"]) > 0:
		d.Set("apm_ipv4", ipAddresses.Apm["prefixes_ipv4"])
	case len(ipAddresses.Logs["prefixes_ipv4"]) > 0:
		d.Set("logs_ipv4", ipAddresses.Logs["prefixes_ipv4"])
	case len(ipAddresses.Process["prefixes_ipv4"]) > 0:
		d.Set("process_ipv4", ipAddresses.Process["prefixes_ipv4"])
	case len(ipAddresses.Synthetics["prefixes_ipv4"]) > 0:
		d.Set("synthetics_ipv4", ipAddresses.Synthetics["prefixes_ipv4"])
	case len(ipAddresses.Webhooks["prefixes_ipv4"]) > 0:
		d.Set("webhooks_ipv4", ipAddresses.Webhooks["prefixes_ipv4"])
	}

	return nil
}
