package datadog

import (
	"github.com/hashicorp/terraform/helper/schema"
	datadog "github.com/zorkian/go-datadog-api"
)

func dataSourceDatadogIPRanges() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDatadogIPRangesRead,

		Schema: map[string]*schema.Schema{
			"agents": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"api": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"apm": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"logs": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"process": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"synthetics": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"webhooks": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceDatadogIPRangesRead(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*datadog.Client)

	ipAddresses, err := client.GetIPRanges()
	if err != nil {
		return err
	}

	d.SetId("datadog-ip-ranges")

	d.Set("agents", ipAddresses.Agents["prefixes_ipv4"])
	d.Set("api", ipAddresses.API["prefixes_ipv4"])
	d.Set("apm", ipAddresses.Apm["prefixes_ipv4"])
	d.Set("logs", ipAddresses.Logs["prefixes_ipv4"])
	d.Set("process", ipAddresses.Process["prefixes_ipv4"])
	d.Set("synthetics", ipAddresses.Synthetics["prefixes_ipv4"])
	d.Set("webhooks", ipAddresses.Webhooks["prefixes_ipv4"])

	return nil
}
