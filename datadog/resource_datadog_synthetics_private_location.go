package datadog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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
		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
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
				"metadata": {
					Type:        schema.TypeList,
					MaxItems:    1,
					Optional:    true,
					Description: "The private location metadata",
					Elem: &schema.Resource{
						Schema: syntheticsPrivateLocationMetadata(),
					},
				},
			}
		},
	}
}

func syntheticsPrivateLocationMetadata() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"restricted_roles": {
			Description: "A list of role identifiers pulled from the Roles API to restrict read and write access.",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
	}
}

func resourceDatadogSyntheticsPrivateLocationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	syntheticsPrivateLocation := buildSyntheticsPrivateLocationStruct(d)
	createdSyntheticsPrivateLocationResponse, httpResponse, err := apiInstances.GetSyntheticsApiV1().CreatePrivateLocation(auth, *syntheticsPrivateLocation)
	if err != nil {
		// Note that Id won't be set, so no state will be saved.
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating synthetics private location")
	}
	if err := utils.CheckForUnparsed(createdSyntheticsPrivateLocationResponse); err != nil {
		return diag.FromErr(err)
	}

	var getSyntheticsPrivateLocationRespone datadogV1.SyntheticsPrivateLocation
	var httpResponseGet *http.Response
	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		getSyntheticsPrivateLocationRespone, httpResponseGet, err = apiInstances.GetSyntheticsApiV1().GetPrivateLocation(auth, *createdSyntheticsPrivateLocationResponse.PrivateLocation.Id)
		if err != nil {
			if httpResponseGet != nil && httpResponseGet.StatusCode == 404 {
				return retry.RetryableError(fmt.Errorf("synthetics private location not created yet"))
			}

			return retry.NonRetryableError(err)
		}
		if err := utils.CheckForUnparsed(getSyntheticsPrivateLocationRespone); err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(getSyntheticsPrivateLocationRespone.GetId())

	// set the config that is only returned when creating the private location
	conf, _ := json.Marshal(createdSyntheticsPrivateLocationResponse.GetConfig())
	d.Set("config", string(conf))

	// Return the read function to ensure the state is reflected in the terraform.state file
	return resourceDatadogSyntheticsPrivateLocationRead(ctx, d, meta)
}

func resourceDatadogSyntheticsPrivateLocationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	syntheticsPrivateLocation, httpresp, err := apiInstances.GetSyntheticsApiV1().GetPrivateLocation(auth, d.Id())

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
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	syntheticsPrivateLocation := buildSyntheticsPrivateLocationStruct(d)
	if _, httpResponse, err := apiInstances.GetSyntheticsApiV1().UpdatePrivateLocation(auth, d.Id(), *syntheticsPrivateLocation); err != nil {
		// If the Update callback returns with or without an error, the full state is saved.
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating synthetics private location")
	}

	// Return the read function to ensure the state is reflected in the terraform.state file
	return resourceDatadogSyntheticsPrivateLocationRead(ctx, d, meta)
}

func resourceDatadogSyntheticsPrivateLocationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	if httpResponse, err := apiInstances.GetSyntheticsApiV1().DeletePrivateLocation(auth, d.Id()); err != nil {
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

	if metadata, ok := d.GetOk("metadata"); ok {
		if metadataMap, ok := metadata.([]interface{})[0].(map[string]interface{}); ok {
			privateLocationMetadata := datadogV1.NewSyntheticsPrivateLocationMetadataWithDefaults()
			// MaxItems is set to 1 so we are sure there is only one metadata to check
			if roles, ok := metadataMap["restricted_roles"].(*schema.Set); ok {
				restricted_roles := []string{}
				for _, role := range roles.List() {
					restricted_roles = append(restricted_roles, role.(string))
				}
				privateLocationMetadata.SetRestrictedRoles(restricted_roles)
			}
			syntheticsPrivateLocation.SetMetadata(*privateLocationMetadata)
		}
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
	localMetadata := make(map[string][]string)
	metadata := syntheticsPrivateLocation.GetMetadata()
	restrictedRoles := metadata.GetRestrictedRoles()
	if len(restrictedRoles) > 0 {
		localMetadata["restricted_roles"] = restrictedRoles
		d.Set("metadata", []map[string][]string{localMetadata})
	}

	return nil
}
