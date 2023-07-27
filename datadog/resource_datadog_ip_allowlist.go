package datadog

import (
	"context"
	"fmt"
	"hash/crc32"
	"net"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceDatadogIPAllowlist() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides the Datadog IP allowlist resource. This can be used to manage the Datadog IP allowlist",
		CreateContext: resourceDatadogIPAllowlistCreate,
		ReadContext:   resourceDatadogIPAllowlistRead,
		UpdateContext: resourceDatadogIPAllowlistUpdate,
		DeleteContext: resourceDatadogIPAllowlistDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"enabled": {
					Type:        schema.TypeBool,
					Required:    true,
					Description: "Whether the IP Allowlist is enabled.",
				},
				"entry": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "Set of objects containing an IP address or range of IP addresses in the allowlist and an accompanying note.",
					Elem:        GetIPAllowlistEntrySchema(),
					Set:         hashCIDR,
				},
			}
		},
	}
}

func GetIPAllowlistEntrySchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"cidr_block": {
				Type:                  schema.TypeString,
				Required:              true,
				Description:           "IP address or range of addresses.",
				ValidateDiagFunc:      cidrValidateFunc,
				DiffSuppressFunc:      diffSuppress,
				DiffSuppressOnRefresh: true,
			},
			"note": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Note accompanying IP address.",
			},
		},
	}
}

func diffSuppress(k, old, new string, d *schema.ResourceData) bool {
	return normalizeIPAddress(old) == normalizeIPAddress(new)
}

func cidrValidateFunc(cidrBlock interface{}, path cty.Path) diag.Diagnostics {
	_, errors := validation.IsCIDR(cidrBlock, cidrBlock.(string))
	if len(errors) == 0 {
		return nil
	}
	_, errors = validation.IsIPAddress(cidrBlock, cidrBlock.(string))
	if len(errors) == 0 {
		return nil
	}
	return diag.Errorf("expected %v to be a valid IP address or CIDR block", cidrBlock)
}

func normalizeIPAddress(ipAddress string) string {
	_, ipNet, err := net.ParseCIDR(ipAddress)
	if err != nil {
		ip := net.ParseIP(ipAddress)
		if ip == nil {
			return ""
		}
		// ipAddress is a single IP address
		// if it is ipv4, the prefix is 32. if ipv6, it is 128
		prefix := "32"
		if ip.DefaultMask() == nil {
			prefix = "128"
		}
		return fmt.Sprintf("%v/%v", ip, prefix)
	}
	return ipNet.String()
}

// copy of the deprecated hashcode.String function
func hashcode(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	// v == MinInt
	return 0
}

func hashCIDR(entry interface{}) int {
	ip := entry.(map[string]interface{})["cidr_block"].(string)
	note := entry.(map[string]interface{})["note"].(string)
	return hashcode(fmt.Sprintf("%s %s", normalizeIPAddress(ip), note))
}

func updateIPAllowlistState(ctx context.Context, d *schema.ResourceData, ipAllowlistAttrs *datadogV2.IPAllowlistAttributes, apiInstances *utils.ApiInstances) diag.Diagnostics {
	if ipAllowlistAttrs != nil {
		if err := d.Set("enabled", ipAllowlistAttrs.GetEnabled()); err != nil {
			return diag.FromErr(err)
		}
		entries, _ := ipAllowlistAttrs.GetEntriesOk()
		return updateIPAllowlistEntriesState(ctx, d, entries, apiInstances)
	}
	return nil
}

func updateIPAllowlistEntriesState(ctx context.Context, d *schema.ResourceData, ipAllowlistEntries *[]datadogV2.IPAllowlistEntry, apiInstances *utils.ApiInstances) diag.Diagnostics {
	var entries []map[string]string
	for _, ipAllowlistEntry := range *ipAllowlistEntries {
		ipAllowlistEntryData := ipAllowlistEntry.GetData()
		ipAllowlistEntryAttributes := ipAllowlistEntryData.GetAttributes()
		cidrBlock, okCidr := ipAllowlistEntryAttributes.GetCidrBlockOk()
		note, okNote := ipAllowlistEntryAttributes.GetNoteOk()
		if okCidr && okNote {
			entry := map[string]string{
				"cidr_block": *cidrBlock,
				"note":       *note,
			}
			entries = append(entries, entry)
		}
	}

	if err := d.Set("entry", entries); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDatadogIPAllowlistRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiInstances := meta.(*ProviderConfiguration).DatadogApiInstances
	auth := meta.(*ProviderConfiguration).Auth

	// Get the IP Allowlist
	resp, httpResp, err := apiInstances.GetIPAllowlistApiV2().GetIPAllowlist(auth)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResp, "error getting IP allowlist")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}
	ipAllowlistData := resp.GetData()
	return updateIPAllowlistState(ctx, d, ipAllowlistData.Attributes, apiInstances)
}

func resourceDatadogIPAllowlistCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiInstances := meta.(*ProviderConfiguration).DatadogApiInstances
	auth := meta.(*ProviderConfiguration).Auth

	ipAllowlistReq, _ := buildIPAllowlistUpdateRequest(d)
	resp, httpResp, err := apiInstances.GetIPAllowlistApiV2().UpdateIPAllowlist(auth, *ipAllowlistReq)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResp, "error updating IP allowlist")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}
	ipAllowlistData := resp.GetData()
	d.SetId(ipAllowlistData.GetId())
	return updateIPAllowlistState(ctx, d, ipAllowlistData.Attributes, apiInstances)
}

func resourceDatadogIPAllowlistDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiInstances := meta.(*ProviderConfiguration).DatadogApiInstances
	auth := meta.(*ProviderConfiguration).Auth

	ipAllowlistUpdateReq := datadogV2.NewIPAllowlistUpdateRequestWithDefaults()
	ipAllowlistData := datadogV2.NewIPAllowlistDataWithDefaults()

	ipAllowlistAttributes := datadogV2.NewIPAllowlistAttributesWithDefaults()
	ipAllowlistAttributes.SetEnabled(false)
	ipAllowlistAttributes.SetEntries([]datadogV2.IPAllowlistEntry{})

	ipAllowlistData.SetAttributes(*ipAllowlistAttributes)
	ipAllowlistUpdateReq.SetData(*ipAllowlistData)

	resp, httpResp, err := apiInstances.GetIPAllowlistApiV2().UpdateIPAllowlist(auth, *ipAllowlistUpdateReq)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResp, "error disabling and removing entries from IP allowlist")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDatadogIPAllowlistUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiInstances := meta.(*ProviderConfiguration).DatadogApiInstances
	auth := meta.(*ProviderConfiguration).Auth

	if d.HasChange("enabled") || d.HasChange("entry") {
		ipAllowlistReq, err := buildIPAllowlistUpdateRequest(d)
		if err != nil {
			return diag.FromErr(err)
		}
		resp, httpResp, err := apiInstances.GetIPAllowlistApiV2().UpdateIPAllowlist(auth, *ipAllowlistReq)
		if err != nil {
			return utils.TranslateClientErrorDiag(err, httpResp, "error updating IP allowlist")
		}
		if err := utils.CheckForUnparsed(resp); err != nil {
			return diag.FromErr(err)
		}
		ipAllowlistData := resp.GetData()
		if err := updateIPAllowlistState(ctx, d, ipAllowlistData.Attributes, apiInstances); err != nil {
			return err
		}
	}

	return nil
}

func buildIPAllowlistUpdateRequest(d *schema.ResourceData) (*datadogV2.IPAllowlistUpdateRequest, error) {
	ipAllowlistUpdateRequest := datadogV2.NewIPAllowlistUpdateRequestWithDefaults()
	ipAllowlistData := datadogV2.NewIPAllowlistDataWithDefaults()
	ipAllowlistAttributes := datadogV2.NewIPAllowlistAttributesWithDefaults()

	enabled := d.Get("enabled")
	ipAllowlistAttributes.SetEnabled(enabled.(bool))

	if entriesI, ok := d.GetOk("entry"); ok {
		entries := entriesI.(*schema.Set).List()
		ipAllowlistEntries := make([]datadogV2.IPAllowlistEntry, len(entries))
		for i, entryI := range entries {
			entry := entryI.(map[string]interface{})
			ipAllowlistEntry := datadogV2.NewIPAllowlistEntryWithDefaults()
			ipAllowlistEntryData := datadogV2.NewIPAllowlistEntryDataWithDefaults()
			ipAllowlistEntryAttributes := datadogV2.NewIPAllowlistEntryAttributesWithDefaults()
			ipAllowlistEntryAttributes.SetCidrBlock(entry["cidr_block"].(string))
			ipAllowlistEntryAttributes.SetNote(entry["note"].(string))
			ipAllowlistEntryData.SetAttributes(*ipAllowlistEntryAttributes)
			ipAllowlistEntry.SetData(*ipAllowlistEntryData)
			ipAllowlistEntries[i] = *ipAllowlistEntry
		}
		ipAllowlistAttributes.SetEntries(ipAllowlistEntries)
	} else {
		ipAllowlistAttributes.SetEntries([]datadogV2.IPAllowlistEntry{})
	}

	ipAllowlistData.SetAttributes(*ipAllowlistAttributes)
	ipAllowlistUpdateRequest.SetData(*ipAllowlistData)
	return ipAllowlistUpdateRequest, nil
}
