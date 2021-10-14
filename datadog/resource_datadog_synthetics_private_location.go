package datadog

import (
	"context"
	"encoding/json"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogSyntheticsPrivateLocation() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog synthetics private location resource. This can be used to create and manage Datadog synthetics private locations.",
		CreateContext: resourceDatadogSyntheticsPrivateLocationCreate,
		ReadContext:   resourceDatadogSyntheticsPrivateLocationRead,
		UpdateContext: resourceDatadogSyntheticsPrivateLocationUpdate,
		DeleteContext: resourceDatadogSyntheticsPrivateLocationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Synthetics private location name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Description of the private location.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"tags": {
				Description: "A list of tags to associate with your synthetics private location.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"config": {
				Description: "Configuration skeleton for the private location. See installation instructions of the private location on how to use this configuration.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func resourceDatadogSyntheticsPrivateLocationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	syntheticsPrivateLocation := buildSyntheticsPrivateLocationStruct(d)
	createdSyntheticsPrivateLocationResponse, httpResponse, err := datadogClientV1.SyntheticsApi.CreatePrivateLocation(authV1, *syntheticsPrivateLocation)
	if err != nil {
		// Note that Id won't be set, so no state will be saved.
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating synthetics private location")
	}
	if err := utils.CheckForUnparsed(createdSyntheticsPrivateLocationResponse); err != nil {
		return diag.FromErr(err)
	}

	createdSyntheticsPrivateLocation := createdSyntheticsPrivateLocationResponse.GetPrivateLocation()
	// If the Create callback returns with or without an error without an ID set using SetId,
	// the resource is assumed to not be created, and no state is saved.
	d.SetId(createdSyntheticsPrivateLocation.GetId())

	// set the config that is only returned when creating the private location
	conf, _ := json.Marshal(createdSyntheticsPrivateLocationResponse.GetConfig())
	d.Set("config", string(conf))

	// Return the read function to ensure the state is reflected in the terraform.state file
	return resourceDatadogSyntheticsPrivateLocationRead(ctx, d, meta)
}

func resourceDatadogSyntheticsPrivateLocationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	syntheticsPrivateLocation, httpresp, err := datadogClientV1.SyntheticsApi.GetPrivateLocation(authV1, d.Id())

	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			// Delete the resource from the local state since it doesn't exist anymore in the actual state
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting synthetics private location")
	}
	if err := utils.CheckForUnparsed(syntheticsPrivateLocation); err != nil {
		return diag.FromErr(err)
	}

	return updateSyntheticsPrivateLocationLocalState(d, &syntheticsPrivateLocation)
}

func resourceDatadogSyntheticsPrivateLocationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	syntheticsPrivateLocation := buildSyntheticsPrivateLocationStruct(d)
	if _, httpResponse, err := datadogClientV1.SyntheticsApi.UpdatePrivateLocation(authV1, d.Id(), *syntheticsPrivateLocation); err != nil {
		// If the Update callback returns with or without an error, the full state is saved.
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating synthetics private location")
	}

	// Return the read function to ensure the state is reflected in the terraform.state file
	return resourceDatadogSyntheticsPrivateLocationRead(ctx, d, meta)
}

func resourceDatadogSyntheticsPrivateLocationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	if httpResponse, err := datadogClientV1.SyntheticsApi.DeletePrivateLocation(authV1, d.Id()); err != nil {
		// The resource is assumed to still exist, and all prior state is preserved.
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting synthetics private location")
	}

	// The resource is assumed to be destroyed, and all state is removed.
	return nil
}

func buildSyntheticsPrivateLocationStruct(d *schema.ResourceData) *datadogV1.SyntheticsPrivateLocation {
	syntheticsPrivateLocation := datadogV1.NewSyntheticsPrivateLocationWithDefaults()

	syntheticsPrivateLocation.SetName(d.Get("name").(string))

	if description, ok := d.GetOk("description"); ok {
		syntheticsPrivateLocation.SetDescription(description.(string))
	}

	tags := make([]string, 0)
	if attr, ok := d.GetOk("tags"); ok {
		for _, s := range attr.([]interface{}) {
			tags = append(tags, s.(string))
		}
	}
	syntheticsPrivateLocation.SetTags(tags)

	return syntheticsPrivateLocation
}

func updateSyntheticsPrivateLocationLocalState(d *schema.ResourceData, syntheticsPrivateLocation *datadogV1.SyntheticsPrivateLocation) diag.Diagnostics {
	d.Set("name", syntheticsPrivateLocation.GetName())
	d.Set("description", syntheticsPrivateLocation.GetDescription())
	d.Set("tags", syntheticsPrivateLocation.Tags)

	return nil
}
