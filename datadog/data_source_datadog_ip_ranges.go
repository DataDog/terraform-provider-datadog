package datadog

import (
	"context"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func dataSourceDatadogIPRanges() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about Datadog's IP addresses.",
		ReadContext: dataSourceDatadogIPRangesRead,

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
			"synthetics_ipv4_by_location": {
				Description: "A map of IPv4 prefixes (string of concatenated IPs, delimited by ',') by location.",
				Type:        schema.TypeMap,
				Computed:    true,
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
			"synthetics_ipv6_by_location": {
				Description: "A map of IPv6 prefixes (string of concatenated IPs, delimited by ',') by location.",
				Type:        schema.TypeMap,
				Computed:    true,
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

func dataSourceDatadogIPRangesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	ipAddresses, _, err := datadogClientV1.IPRangesApi.GetIPRanges(authV1)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := utils.CheckForUnparsed(ipAddresses); err != nil {
		return diag.FromErr(err)
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
		len(synthetics.GetPrefixesIpv6())+len(webhook.GetPrefixesIpv6())+
		len(synthetics.GetPrefixesIpv4ByLocation())+len(synthetics.GetPrefixesIpv6ByLocation()) > 0 {
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

	ipv4PrefixesByLocationMap := make(map[string]string)
	ipv6PrefixesByLocationMap := make(map[string]string)

	ipv4PrefixesByLocation := synthetics.GetPrefixesIpv4ByLocation()
	ipv6PrefixesByLocation := synthetics.GetPrefixesIpv6ByLocation()

	for key, value := range ipv4PrefixesByLocation {
		ipv4PrefixesByLocationMap[key] = strings.Join(value, ",")
	}

	for key, value := range ipv6PrefixesByLocation {
		ipv6PrefixesByLocationMap[key] = strings.Join(value, ",")
	}

	err = d.Set("synthetics_ipv4_by_location", ipv4PrefixesByLocationMap)
	if err != nil {
		log.Printf("[DEBUG] Error setting IPv4 prefixes by location: %s", err)
		return diag.FromErr(err)
	}
	err = d.Set("synthetics_ipv6_by_location", ipv6PrefixesByLocationMap)
	if err != nil {
		log.Printf("[DEBUG] Error setting IPv6 prefixes by location: %s", err)
		return diag.FromErr(err)
	}

	return nil
}
