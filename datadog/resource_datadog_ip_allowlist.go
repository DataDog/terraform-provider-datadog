package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
		Schema: map[string]*schema.Schema{
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
			},
		},
	}
}

func GetIPAllowlistEntrySchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"cidr_block": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "IP address or range of addresses.",
			},
			"note": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Note accompanying IP address.",
			},
		},
	}
}

func updateIPAllowlistState(ctx context.Context, d *schema.ResourceData, ipAllowlistAttrs *datadogV2.IPAllowlistAttributes, apiInstances *utils.ApiInstances) diag.Diagnostics {
	if ipAllowlistAttrs != nil {
		if err := d.Set("enabled", ipAllowlistAttrs.GetEnabled()); err != nil {
			return diag.FromErr(err)
		}
		return updateIPAllowlistEntriesState(ctx, d, ipAllowlistAttrs.GetEntries(), apiInstances)
	}
	return nil
}

func updateIPAllowlistEntriesState(ctx context.Context, d *schema.ResourceData, ipAllowlistEntries []datadogV2.IPAllowlistEntry, apiInstances *utils.ApiInstances) diag.Diagnostics {
	var entries []map[string]string
	for _, ipAllowlistEntry := range ipAllowlistEntries {
		ipAllowlistEntryData := ipAllowlistEntry.GetData()
		ipAllowlistEntryAttributes := ipAllowlistEntryData.GetAttributes()
		cidr_block, ok_cidr := ipAllowlistEntryAttributes.GetCidrBlockOk()
		note, ok_note := ipAllowlistEntryAttributes.GetNoteOk()
		if ok_cidr && ok_note {
			entry := map[string]string{
				"cidr_block": *cidr_block,
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
	resp, httpResp, err := apiInstances.GetIPAllowlistApiV2().UpdateIPAllowlist(auth, ipAllowlistReq)
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

	if d.HasChange("enabled") || d.HasChange("entries") {
		ipAllowlistReq, err := buildIPAllowlistUpdateRequest(d)
		if err != nil {
			return diag.FromErr(err)
		}
		resp, httpResp, err := apiInstances.GetIPAllowlistApiV2().UpdateIPAllowlist(auth, ipAllowlistReq)
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

func buildIPAllowlistUpdateRequest(d *schema.ResourceData) (datadogV2.IPAllowlistUpdateRequest, error) {
	ipAllowlistUpdateRequest := datadogV2.NewIPAllowlistUpdateRequestWithDefaults()
	ipAllowlistData := datadogV2.NewIPAllowlistDataWithDefaults()
	ipAllowlistAttributes := datadogV2.NewIPAllowlistAttributesWithDefaults()

	enabled := d.Get("enabled")
	ipAllowlistAttributes.SetEnabled(enabled.(bool))

	if entriesI, ok := d.GetOk("entries"); ok {
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
	return *ipAllowlistUpdateRequest, nil
}
