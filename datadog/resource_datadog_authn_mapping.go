package datadog

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func resourceDatadogAuthnMapping() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog AuthN Mappings resource. This feature lets you automatically assign roles to users based on their SAML attributes.",
		CreateContext: resourceDatadogAuthnMappingCreate,
		ReadContext:   resourceDatadogAuthnMappingRead,
		UpdateContext: resourceDatadogAuthnMappingUpdate,
		DeleteContext: resourceDatadogAuthnMappingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"key": {
				Description: "Identity provider key.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"value": {
				Description: "Identity provider value.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"role": {
				Description: "The ID of a role to attach to all users with the corresponding key and value.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func resourceDatadogAuthnMappingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ProviderConfiguration).DatadogClientV2
	auth := meta.(*ProviderConfiguration).AuthV2
	authNMapReq := buildAuthNMappingCreateRequest(d)
	createResp, httpResponse, err := client.AuthNMappingsApi.CreateAuthNMapping(auth, authNMapReq)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating authn mapping")
	}
	if err := utils.CheckForUnparsed(createResp); err != nil {
		return diag.FromErr(err)
	}

	var getAuthNMappingResponse datadog.AuthNMappingResponse
	var httpResponseGet *http.Response

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		getAuthNMappingResponse, httpResponseGet, err = client.AuthNMappingsApi.GetAuthNMapping(auth, createResp.Data.GetId())
		if err != nil {
			if httpResponseGet != nil && httpResponseGet.StatusCode == 404 {
				return resource.RetryableError(fmt.Errorf("SAML role mapping not created yet"))
			}

			return resource.NonRetryableError(err)
		}
		if err := utils.CheckForUnparsed(getAuthNMappingResponse); err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	authNMappingData := getAuthNMappingResponse.GetData()
	d.SetId(authNMappingData.GetId())
	return updateAuthNMappingState(d, &authNMappingData)
}

func resourceDatadogAuthnMappingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ProviderConfiguration).DatadogClientV2
	auth := meta.(*ProviderConfiguration).AuthV2

	resp, httpResponse, err := client.AuthNMappingsApi.GetAuthNMapping(auth, d.Id())
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			d.SetId((""))
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpResponse, "error getting authn mapping")
	}
	authNMappingData := resp.GetData()
	return updateAuthNMappingState(d, &authNMappingData)
}

func resourceDatadogAuthnMappingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ProviderConfiguration).DatadogClientV2
	auth := meta.(*ProviderConfiguration).AuthV2

	req := buildAuthNMappingUpdateRequest(d)
	resp, httpResponse, err := client.AuthNMappingsApi.UpdateAuthNMapping(auth, d.Id(), req)

	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating role mapping")
	}

	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}

	authNMappingData := resp.GetData()
	return updateAuthNMappingState(d, &authNMappingData)
}

func resourceDatadogAuthnMappingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ProviderConfiguration).DatadogClientV2
	auth := meta.(*ProviderConfiguration).AuthV2

	httpResponse, err := client.AuthNMappingsApi.DeleteAuthNMapping(auth, d.Id())
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting authn mapping")
	}

	return nil
}

func updateAuthNMappingState(d *schema.ResourceData, authNMapping *datadog.AuthNMapping) diag.Diagnostics {
	authNMappingAttributes := authNMapping.GetAttributes()
	authNMappingRelations := authNMapping.GetRelationships()
	authNMappingRoleRelation := authNMappingRelations.GetRole()
	authNRole := authNMappingRoleRelation.GetData()

	if err := d.Set("key", authNMappingAttributes.GetAttributeKey()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("value", authNMappingAttributes.GetAttributeValue()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("role", authNRole.GetId()); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func buildAuthNMappingCreateRequest(d *schema.ResourceData) datadog.AuthNMappingCreateRequest {
	authNMappingCreateRequest := datadog.NewAuthNMappingCreateRequestWithDefaults()
	authNMappingCreateData := datadog.NewAuthNMappingCreateDataWithDefaults()
	authNMappingCreateAttrs := datadog.NewAuthNMappingCreateAttributesWithDefaults()
	authNMappingRelations := datadog.NewAuthNMappingCreateRelationshipsWithDefaults()

	// Set AuthN mapping Attributes
	authNMappingCreateAttrs.SetAttributeKey(d.Get("key").(string))
	authNMappingCreateAttrs.SetAttributeValue(d.Get("value").(string))

	// Set AuthN mapping Relationships
	roleRelations := buildRoleRelations(d)
	authNMappingRelations.SetRole(*roleRelations)

	// Set AuthN mapping create data
	authNMappingCreateData.SetAttributes(*authNMappingCreateAttrs)
	authNMappingCreateData.SetRelationships(*authNMappingRelations)

	// Set AuthN mapping create request
	authNMappingCreateRequest.SetData(*authNMappingCreateData)
	return *authNMappingCreateRequest
}

func buildAuthNMappingUpdateRequest(d *schema.ResourceData) datadog.AuthNMappingUpdateRequest {
	authNMappingUpdateRequest := datadog.NewAuthNMappingUpdateRequestWithDefaults()
	authNMappingUpdateData := datadog.NewAuthNMappingUpdateDataWithDefaults()
	authNMappingUpdateAttrs := datadog.NewAuthNMappingUpdateAttributesWithDefaults()
	authNMappingRelations := datadog.NewAuthNMappingUpdateRelationshipsWithDefaults()

	// Set AuthN mapping Attributes
	authNMappingUpdateAttrs.SetAttributeKey(d.Get("key").(string))
	authNMappingUpdateAttrs.SetAttributeValue(d.Get("value").(string))

	// Set AuthN mapping Relationships
	roleRelations := buildRoleRelations(d)
	authNMappingRelations.SetRole(*roleRelations)

	// Set AuthN mapping update data
	authNMappingUpdateData.SetAttributes(*authNMappingUpdateAttrs)
	authNMappingUpdateData.SetRelationships(*authNMappingRelations)
	authNMappingUpdateData.SetId(d.Id())

	// Set AuthN mapping update request
	authNMappingUpdateRequest.SetData(*authNMappingUpdateData)
	return *authNMappingUpdateRequest
}

func buildRoleRelations(d *schema.ResourceData) *datadog.RelationshipToRole {
	roleRelations := datadog.NewRelationshipToRoleWithDefaults()
	roleRelationsData := datadog.NewRelationshipToRoleDataWithDefaults()
	roleRelationsData.SetId(d.Get("role").(string))
	roleRelations.SetData(*roleRelationsData)
	return roleRelations
}
