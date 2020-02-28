package datadog

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceDatadogIPRanges() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDatadogIPRangesRead,

		// IP ranges are divided between ipv4 and ipv6
		Schema: map[string]*schema.Schema{
			// v4
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
			// v6
			"agents_ipv6": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"api_ipv6": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"apm_ipv6": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"logs_ipv6": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"process_ipv6": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"synthetics_ipv6": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"webhooks_ipv6": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceDatadogIPRangesRead(d *schema.ResourceData, meta interface{}) error {

	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	ipAddresses, err := client.GetIPRanges()

	if err != nil {
		return err
	}

	// v4 and v6
	if len(ipAddresses.Agents["prefixes_ipv4"])+len(ipAddresses.API["prefixes_ipv4"])+
		len(ipAddresses.Apm["prefixes_ipv4"])+len(ipAddresses.Logs["prefixes_ipv4"])+
		len(ipAddresses.Process["prefixes_ipv4"])+len(ipAddresses.Synthetics["prefixes_ipv4"])+
		len(ipAddresses.Webhooks["prefixes_ipv4"])+len(ipAddresses.Agents["prefixes_ipv6"])+
		len(ipAddresses.API["prefixes_ipv6"])+len(ipAddresses.Apm["prefixes_ipv6"])+
		len(ipAddresses.Logs["prefixes_ipv6"])+len(ipAddresses.Process["prefixes_ipv6"])+
		len(ipAddresses.Synthetics["prefixes_ipv6"])+len(ipAddresses.Webhooks["prefixes_ipv6"]) > 0 {
		d.SetId("datadog-ip-ranges")
	}

	// Set ranges when the list is not empty
	// v4
	if len(ipAddresses.Agents["prefixes_ipv4"]) > 0 {
		d.Set("agents_ipv4", ipAddresses.Agents["prefixes_ipv4"])
	} else {
		d.Set("agents_ipv4", []string{})
	}

	if len(ipAddresses.API["prefixes_ipv4"]) > 0 {
		d.Set("api_ipv4", ipAddresses.API["prefixes_ipv4"])
	} else {
		d.Set("api_ipv4", []string{})
	}

	if len(ipAddresses.Apm["prefixes_ipv4"]) > 0 {
		d.Set("apm_ipv4", ipAddresses.Apm["prefixes_ipv4"])
	} else {
		d.Set("apm_ipv4", []string{})
	}

	if len(ipAddresses.Logs["prefixes_ipv4"]) > 0 {
		d.Set("logs_ipv4", ipAddresses.Logs["prefixes_ipv4"])
	} else {
		d.Set("logs_ipv4", []string{})
	}

	if len(ipAddresses.Process["prefixes_ipv4"]) > 0 {
		d.Set("process_ipv4", ipAddresses.Process["prefixes_ipv4"])
	} else {
		d.Set("process_ipv4", []string{})
	}

	if len(ipAddresses.Synthetics["prefixes_ipv4"]) > 0 {
		d.Set("synthetics_ipv4", ipAddresses.Synthetics["prefixes_ipv4"])
	} else {
		d.Set("synthetics_ipv4", []string{})
	}

	if len(ipAddresses.Webhooks["prefixes_ipv4"]) > 0 {
		d.Set("webhooks_ipv4", ipAddresses.Webhooks["prefixes_ipv4"])
	} else {
		d.Set("webhooks_ipv4", []string{})
	}

	// v6
	if len(ipAddresses.Agents["prefixes_ipv6"]) > 0 {
		d.Set("agents_ipv6", ipAddresses.Agents["prefixes_ipv6"])
	} else {
		d.Set("agents_ipv6", []string{})
	}

	if len(ipAddresses.API["prefixes_ipv6"]) > 0 {
		d.Set("api_ipv6", ipAddresses.API["prefixes_ipv6"])
	} else {
		d.Set("api_ipv6", []string{})
	}

	if len(ipAddresses.Apm["prefixes_ipv6"]) > 0 {
		d.Set("apm_ipv6", ipAddresses.Apm["prefixes_ipv6"])
	} else {
		d.Set("apm_ipv6", []string{})
	}

	if len(ipAddresses.Logs["prefixes_ipv6"]) > 0 {
		d.Set("logs_ipv6", ipAddresses.Logs["prefixes_ipv6"])
	} else {
		d.Set("logs_ipv6", []string{})
	}

	if len(ipAddresses.Process["prefixes_ipv6"]) > 0 {
		d.Set("process_ipv6", ipAddresses.Process["prefixes_ipv6"])
	} else {
		d.Set("process_ipv6", []string{})
	}

	if len(ipAddresses.Synthetics["prefixes_ipv6"]) > 0 {
		d.Set("synthetics_ipv6", ipAddresses.Synthetics["prefixes_ipv6"])
	} else {
		d.Set("synthetics_ipv6", []string{})
	}

	if len(ipAddresses.Webhooks["prefixes_ipv6"]) > 0 {
		d.Set("webhooks_ipv6", ipAddresses.Webhooks["prefixes_ipv6"])
	} else {
		d.Set("webhooks_ipv6", []string{})
	}

	return nil
}
