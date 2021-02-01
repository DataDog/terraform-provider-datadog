package datadog

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDatadogIpRanges() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about Datadog's IP addresses.",
		Read:        dataSourceDatadogIPRangesRead,

		// IP ranges are divided between ipv4 and ipv6
		Schema: map[string]*schema.Schema{
			// v4
			"agents_ipv4": {
				Description: "An Array of IPv4 addresses in CIDR format specifying the A records for the Agent endpoint.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"api_ipv4": {
				Description: "An Array of IPv4 addresses in CIDR format specifying the A records for the API endpoint.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"apm_ipv4": {
				Description: "An Array of IPv4 addresses in CIDR format specifying the A records for the APM endpoint.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"logs_ipv4": {
				Description: "An Array of IPv4 addresses in CIDR format specifying the A records for the Logs endpoint.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"process_ipv4": {
				Description: "An Array of IPv4 addresses in CIDR format specifying the A records for the Process endpoint.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"synthetics_ipv4": {
				Description: "An Array of IPv4 addresses in CIDR format specifying the A records for the Synthetics endpoint.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"webhooks_ipv4": {
				Description: "An Array of IPv4 addresses in CIDR format specifying the A records for the Webhooks endpoint.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			// v6
			"agents_ipv6": {
				Description: "An Array of IPv6 addresses in CIDR format specifying the A records for the Agent endpoint.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"api_ipv6": {
				Description: "An Array of IPv6 addresses in CIDR format specifying the A records for the API endpoint.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"apm_ipv6": {
				Description: "An Array of IPv6 addresses in CIDR format specifying the A records for the APM endpoint.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"logs_ipv6": {
				Description: "An Array of IPv6 addresses in CIDR format specifying the A records for the Logs endpoint.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"process_ipv6": {
				Description: "An Array of IPv6 addresses in CIDR format specifying the A records for the Process endpoint.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"synthetics_ipv6": {
				Description: "An Array of IPv6 addresses in CIDR format specifying the A records for the Synthetics endpoint.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"webhooks_ipv6": {
				Description: "An Array of IPv6 addresses in CIDR format specifying the A records for the Webhooks endpoint.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceDatadogIPRangesRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	ipAddresses, _, err := datadogClientV1.IPRangesApi.GetIPRanges(authV1).Execute()
	if err != nil {
		return err
	}

	// v4 and v6
	ipAddressesPtr := &ipAddresses
	agents := ipAddressesPtr.GetAgents()
	api := ipAddressesPtr.GetApi()
	apm := ipAddressesPtr.GetApm()
	logs := ipAddressesPtr.GetLogs()
	process := ipAddressesPtr.GetProcess()
	synthetics := ipAddressesPtr.GetSynthetics()
	webhook := ipAddressesPtr.GetWebhooks()

	if len(agents.GetPrefixesIpv4())+len(api.GetPrefixesIpv4())+
		len(apm.GetPrefixesIpv4())+len(logs.GetPrefixesIpv4())+
		len(process.GetPrefixesIpv4())+len(synthetics.GetPrefixesIpv4())+
		len(webhook.GetPrefixesIpv4())+len(agents.GetPrefixesIpv6())+
		len(api.GetPrefixesIpv6())+len(apm.GetPrefixesIpv6())+
		len(logs.GetPrefixesIpv6())+len(process.GetPrefixesIpv6())+
		len(synthetics.GetPrefixesIpv6())+len(webhook.GetPrefixesIpv6()) > 0 {
		d.SetId("datadog-ip-ranges")
	}

	// Set ranges when the list is not empty
	// v4
	d.Set("agents_ipv4", agents.GetPrefixesIpv4())
	d.Set("api_ipv4", api.GetPrefixesIpv4())
	d.Set("apm_ipv4", apm.GetPrefixesIpv4())
	d.Set("logs_ipv4", logs.GetPrefixesIpv4())
	d.Set("process_ipv4", process.GetPrefixesIpv4())
	d.Set("synthetics_ipv4", synthetics.GetPrefixesIpv4())
	d.Set("webhooks_ipv4", webhook.GetPrefixesIpv4())

	// v6
	d.Set("agents_ipv6", agents.GetPrefixesIpv6())
	d.Set("api_ipv6", api.GetPrefixesIpv6())
	d.Set("apm_ipv6", apm.GetPrefixesIpv6())
	d.Set("logs_ipv6", logs.GetPrefixesIpv6())
	d.Set("process_ipv6", process.GetPrefixesIpv6())
	d.Set("synthetics_ipv6", synthetics.GetPrefixesIpv6())
	d.Set("webhooks_ipv6", webhook.GetPrefixesIpv6())

	return nil
}
